
clout.exec('tosca.lib.traversal');

// Run plugins
clout.execAll('kubernetes.plugins.resources.pre-get');

var valueInformation = tosca.getValueInformation();

tosca.coerce();

var resources = [];

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	var nodeTemplate = vertex.properties;

	var kubernetesMetadata = {
		name: nodeTemplate.name,
		annotations: {'puccini.cloud/vertex': vertexId}
	};

	// Find shared metadata
	var hasMetadataCapability = false;
	for (var capabilityName in nodeTemplate.capabilities) {
		var capability = nodeTemplate.capabilities[capabilityName];
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
	for (var capabilityName in nodeTemplate.capabilities) {
		var capability = nodeTemplate.capabilities[capabilityName];
		for (var typeName in capability.types) {
			var type = capability.types[typeName];
			var metadata = type.metadata;
			if (metadata && metadata['turandot.apiVersion']) {
				var kind = metadata['turandot.kind'];
				if (!kind)
					kind = typeName.split('::')[1];
				var pathPrefix = puccini.sprintf('nodeTemplates.%s.capabilities.%s', nodeTemplate.name, capabilityName);
				var capabilityKubernetesMetadata = puccini.deepCopy(kubernetesMetadata);
				var metadataNamePostfix = capability.properties['metadataNamePostfix'];
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
clout.execAll('kubernetes.plugins.resources.get');

puccini.write(resources);

function generateResource(capability, apiVersion, kind, metadata, pathPrefix, kubernetesMetadata) {
	// Convert attribute mappings to annotation
	var attributeMappings = {};
	for (var attributeName in capability.attributes) {
		var information = valueInformation[puccini.sprintf('%s.attributes.%s', pathPrefix, attributeName)];
		if (information && information.definition && information.definition.metadata) {
			var mapping = information.definition.metadata['turandot.mapping'];
			if (mapping) {
				puccini.log.infof('mapping: %s -> %s', mapping, attributeName);
				attributeMappings[mapping] = attributeName;
			}
		}
	}
	if (Object.keys(attributeMappings).length > 0)
		kubernetesMetadata.annotations['clout.puccini.cloud/attributeMappings'] = JSON.stringify(attributeMappings);

	var resource = {
		apiVersion: apiVersion,
		kind: kind,
		metadata: kubernetesMetadata,
		spec: {}
	};

	// Copy properties into spec
	for (var propertyName in capability.properties) {
		var information = valueInformation[puccini.sprintf('%s.properties.%s', pathPrefix, propertyName)];
		if (information && information.definition && information.definition.metadata && (information.definition.metadata['turandot.ignore'] === 'true'))
			continue;
		var value = capability.properties[propertyName];
		resource.spec[propertyName] = processValue(value, information);
	}

	var keys = Object.keys(metadata).sort();

	// Process moves
	for (var k = 0, l = keys.length; k < l; k++)
		if (keys[k].substring(0, 13) === 'turandot.move') {
			var value = metadata[keys[k]].split('->');
			var from = value[0];
			var to = value[1];
			puccini.log.infof('move: %s -> %s', from, to);
			var from_ = from.split('.');
			value = getValue(resource, from_);
			deleteValue(resource, from_);
			setValue(resource, to.split('.'), value);
		}

	// Process copies
	for (var k = 0, l = keys.length; k < l; k++)
		if (keys[k].substring(0, 13) === 'turandot.copy') {
			var value = metadata[keys[k]].split('->');
			var from = value[0];
			var to = value[1];
			puccini.log.infof('copy: %s -> %s', from, to);
			var to_ = to.split('.');
			value = getValue(resource, to_);
			puccini.log.infof('current: %s = %v', to, value);
			if (!value) {
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
	var first = path[0];
	var property = object[first];
	if ((property === null) || (property === undefined))
		property = object[first] = {};
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

function processValue(value, information) {
	if (value === null)
		return null;

	if (information) {
		if (information.type && (typeof information.type === 'object')) {
			switch (information.type.name) {
			case 'scalar-unit.size':
			case 'scalar-unit.time':
			case 'scalar-unit.frequency':
			case 'scalar-unit.bitrate':
				puccini.log.infof('%s: %s -> %v', information.type.name, value.$originalString, value.$number);
				value = value.$number;
				break;
			}

			if (information.type.metadata) {
				var format = information.type.metadata['turandot.format'];
				switch (format) {
				case 'percentage':
					var percentage = (value * 100) + '%';
					puccini.log.infof('format percentage: %s -> %s', String(value), percentage);
					value = percentage;
					break;

				case 'json':
					value = JSON.stringify(value, null, ' ');
					puccini.log.infof('format JSON: %s', JSON.stringify(value));
					break;
				}
			}
		}

		if (information.properties && (typeof value === 'object')) {
			for (var key in value) {
				var propertyInformation = information.properties[key];
				if (propertyInformation)
					value[key] = processValue(value[key], propertyInformation);
			}
		}
	}

	return value;
}
