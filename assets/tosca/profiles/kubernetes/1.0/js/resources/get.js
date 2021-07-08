
const traversal = require('tosca.lib.traversal');
const tosca = require('tosca.lib.utils');

// TODO: not here
for (let vertexId in clout.vertexes) {
	let vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	let nodeTemplate = vertex.properties;

	for (let artifactName in nodeTemplate.artifacts) {
		let artifact = nodeTemplate.artifacts[artifactName];

		if ('cloud.puccini.turandot.orchestration::Key' in artifact.types)
			artifact.$artifact = puccini.loadString(artifact.sourcePath);
	}
}

// Run plugins
clout.callAll('kubernetes.plugins.resources.pre-get', 'plugin');

let valueInformation = traversal.getValueInformation();

traversal.coerce();

let resources = [];

for (let vertexId in clout.vertexes) {
	let vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	let nodeTemplate = vertex.properties;

	let kubernetesMetadata = {
		name: nodeTemplate.name,
		annotations: {'puccini.cloud/vertex': vertexId}
	};

	// Find shared metadata
	let hasMetadataCapability = false;
	for (let capabilityName in nodeTemplate.capabilities) {
		let capability = nodeTemplate.capabilities[capabilityName];
		if ('cloud.puccini.kubernetes::Metadata' in capability.types) {
			if (hasMetadataCapability)
				throw puccini.sprintf('node template %s has more than one capability of type cloud.puccini.kubernetes::Metadata', nodeTemplate.name);
			hasMetadataCapability = true;

			kubernetesMetadata = puccini.deepCopy(capability.properties);
			if (!kubernetesMetadata.name)
				kubernetesMetadata.name = nodeTemplate.name;
			if (!kubernetesMetadata.annotations)
				kubernetesMetadata.annotations = {};
			kubernetesMetadata.annotations['clout.puccini.cloud/vertex'] = vertexId;
		}
	}

	// Generate resources
	for (let capabilityName in nodeTemplate.capabilities) {
		let capability = nodeTemplate.capabilities[capabilityName];
		for (let typeName in capability.types) {
			let type = capability.types[typeName];
			let metadata = type.metadata;
			if (metadata && metadata['turandot.apiVersion']) {
				let kind = metadata['turandot.kind'];
				if (!kind)
					kind = typeName.split('::')[1];
				let pathPrefix = puccini.sprintf('nodeTemplates.%s.capabilities.%s', nodeTemplate.name, capabilityName);
				let capabilityKubernetesMetadata = puccini.deepCopy(kubernetesMetadata);
				let metadataNamePostfix = capability.properties['metadataNamePostfix'];
				if (metadataNamePostfix === undefined)
					metadataNamePostfix = capabilityName;
				if (metadataNamePostfix !== '')
					capabilityKubernetesMetadata.name = puccini.sprintf('%s-%s', kubernetesMetadata.name, metadataNamePostfix);
				capabilityKubernetesMetadata.annotations['clout.puccini.cloud/capability'] = capabilityName;
				generateResource(capability, metadata['turandot.apiVersion'], kind, metadata, pathPrefix, capabilityKubernetesMetadata);
			}
		}
	}
}

// Run plugins
clout.callAll('kubernetes.plugins.resources.get', 'plugin', resources);

puccini.write(resources);

function generateResource(capability, apiVersion, kind, metadata, pathPrefix, kubernetesMetadata) {
	// Convert attribute mappings to annotation
	let attributeMappings = {};
	for (let attributeName in capability.attributes) {
		let mapping = getInformation(pathPrefix, 'attributes', attributeName, 'turandot.mapping');
		if (mapping) {
			puccini.log.infof('mapping: %s -> %s', mapping, attributeName);
			attributeMappings[mapping] = attributeName;
		}
	}
	if (Object.keys(attributeMappings).length > 0)
		kubernetesMetadata.annotations['clout.puccini.cloud/attributeMappings'] = JSON.stringify(attributeMappings);

	let resource = {
		apiVersion: apiVersion,
		kind: kind,
		metadata: kubernetesMetadata
	};

	// Copy properties into resource
	for (let propertyName in capability.properties) {
		if (getInformation(pathPrefix, 'properties', propertyName, 'turandot.ignore') === 'true')
			continue;
		let value = capability.properties[propertyName];
		resource[propertyName] = value;
	}

	let keys = Object.keys(metadata).sort();

	// Process moves
	for (let k = 0, l = keys.length; k < l; k++)
		if (keys[k].substring(0, 13) === 'turandot.move') {
			let value = metadata[keys[k]].split('->');
			let from = value[0];
			let to = value[1];
			puccini.log.infof('move: %s -> %s', from, to);
			let from_ = from.split('.');
			value = getValue(resource, from_);
			deleteValue(resource, from_);
			setValue(resource, to.split('.'), value);
		}

	// Process copies
	for (let k = 0, l = keys.length; k < l; k++)
		if (keys[k].substring(0, 13) === 'turandot.copy') {
			let value = metadata[keys[k]].split('->');
			let from = value[0];
			let to = value[1];
			puccini.log.infof('copy: %s -> %s', from, to);
			let to_ = to.split('.');
			value = getValue(resource, to_);
			puccini.log.infof('current: %s = %v', to, value);
			if ((value === null) || (value === undefined)) {
				value = getValue(resource, from.split('.'));
				puccini.log.infof('set: %s = %v', to, value);
				setValue(resource, to_, value);
			}
		}

	resources.push(resource);
}

function getValue(object, path) {
	if (!object) {
		return null;
	}
	switch (path.length) {
	case 0:
		return null;
	case 1:
		return object[path[0]];
	}
	return getValue(object[path[0]], path.slice(1));
}

function setValue(object, path, value) {
	if ((object === null) || (typeof object !== 'object')) {
		return;
	}
	switch (path.length) {
	case 0:
		return;
	case 1:
		object[path[0]] = value;
		return;
	}
	let first = path[0];
	let property = object[first];
	if (typeof property !== 'object') {
		object[first] = {};
		property = object[first];
	}
	setValue(property, path.slice(1), value);
}

function deleteValue(object, path) {
	if ((object === null) || (typeof object !== 'object')) {
		return;
	}
	switch (path.length) {
	case 0:
		return;
	case 1:
		delete object[path[0]];
		return;
	}
	deleteValue(object[path[0]], path.slice(1));
}

function getInformation(pathPrefix, type, fieldName, name) {
	let information = valueInformation[puccini.sprintf('%s.%s.%s', pathPrefix, type, fieldName)];
	if (information && information.definition && information.definition.metadata)
		return information.definition.metadata[name];
	return undefined;
}
