
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
		if ('cloud.puccini.kubevirt::VirtualMachine' in capability.types)
			generateVirtualMachine(capability, capabilityMetadata);
	}
}

function generateVirtualMachine(capability, metadata) {
	var spec = {
		apiVersion: 'kubevirt.io/v1alpha3',
		kind: 'VirtualMachine',
		metadata: metadata,
		spec: {
			running: true,
			template: {
				metadata: metadata,
				spec: {}
			}
		}
	};

	for (var propertyName in capability.properties.template) {
		var v = capability.properties.template[propertyName];
		spec.spec.template.spec[propertyName] = v;
	}

	specs.push(spec);
}
