
clout.exec('tosca.lib.traversal');

// Run plugins
clout.execAll('kubernetes.pre-resources-plugins');

tosca.coerce();

var specs = [];

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	var nodeTemplate = vertex.properties;

	// Find metadata
	var metadata = {annotations: {}};
	for (var capabilityName in nodeTemplate.capabilities) {
		var capability = nodeTemplate.capabilities[capabilityName];
		if ('cloud.puccini.kubernetes::Metadata' in capability.types) {
			metadata = puccini.deepCopy(capability.properties);
			if (!metadata.name)
				metadata.name = nodeTemplate.name;
			if (!metadata.annotations)
				metadata.annotations = {};
			metadata.annotations['puccini.cloud/vertex'] = vertexId;
			break;
		}
	}

	// Generate specs
	for (var capabilityName in nodeTemplate.capabilities) {
		var capability = nodeTemplate.capabilities[capabilityName];
		var capabilityMetadata = puccini.deepCopy(metadata);
		capabilityMetadata.annotations['puccini.cloud/capability'] = capabilityName;
		if ('cloud.puccini.kubernetes::Service' in capability.types)
			generateService(capabilityName, capability, capabilityMetadata);
		else if ('cloud.puccini.kubernetes::Deployment' in capability.types)
			generateDeployment(capability, capabilityMetadata);
		else if ('cloud.puccini.kubernetes::NetworkAttachmentDefinition' in capability.types)
			generateNetworkAttachmentDefinition(capability, capabilityMetadata);
	}
}

// Run plugins
clout.execAll('kubernetes.resources-plugins');

puccini.write(specs);

function generateService(capabilityName, capability, metadata) {
	metadata.name = puccini.sprintf('%s-%s', metadata.name, capabilityName);

	var spec = {
		apiVersion: 'v1',
		kind: 'Service',
		metadata: metadata,
		spec: {}
	};

	for (var propertyName in capability.properties) {
		var v = capability.properties[propertyName];
		spec.spec[propertyName] = v;
	}

	// Default selector
	if (spec.spec.selector === undefined)
		spec.spec.selector = metadata.labels;

	specs.push(spec);
}

function generateDeployment(capability, metadata) {
	var spec = {
		apiVersion: 'apps/v1',
		kind: 'Deployment',
		metadata: metadata,
		spec: {}
	};

	for (var propertyName in capability.properties) {
		var v = capability.properties[propertyName];
		switch (propertyName) {
		case 'minReadySeconds':
		case 'progressDeadlineSeconds':
			v = convertScalarUnit(v);
			break;
		case 'strategy':
			var s = {
				type: v.type
			};
			if (v.type === 'RollingUpdate') {
				s.rollingUpdate = {
					maxSurge: convertAmount(v.maxSurge),
					maxUnavailable: convertAmount(v.maxUnavailable)
				};
			}
			v = s;
			break;
		case 'template':
			var s = {};
			for (var t in v) {
				var vv = v[t];
				switch (t) {
				case 'activeDeadlineSeconds':
				case 'terminationGracePeriodSeconds':
					vv = convertScalarUnit(vv);
					break;
				}
				s[t] = vv;
			}
			v = {
				metadata: metadata,
				spec: s
			};
		}
		spec.spec[propertyName] = v;
	}

	// Default selector
	if ((spec.spec.selector.matchExpressions == undefined) && (spec.spec.selector.matchLabels === undefined))
		spec.spec.selector.matchLabels = metadata.labels;

	specs.push(spec);
}

function generateNetworkAttachmentDefinition(capability, metadata) {
	var spec = {
		apiVersion: 'k8s.cni.cncf.io/v1',
		kind: 'NetworkAttachmentDefinition',
		metadata: metadata,
		spec: {
			config: JSON.stringify(capability.properties.config, null, '  ')
		}
	};
	
	specs.push(spec);
}

function convertScalarUnit(v) {
	return v.$number;
}

function convertAmount(v) {
	if (v.factor !== undefined)
		return (v.factor * 100) + '%';
	return v.count;
}
