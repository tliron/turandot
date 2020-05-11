
/*
for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	var nodeTemplate = vertex.properties;

	// Find metadata
	var metadata = {};
	for (var capabilityName in nodeTemplate.capabilities) {
		var capability = nodeTemplate.capabilities[capabilityName];
		if ('cloud.puccini.kubernetes::Metadata' in capability.types) {
			metadata = capability.properties;
			break;
		}
	}

	// Generate specs
	for (var capabilityName in nodeTemplate.capabilities) {
		var capability = nodeTemplate.capabilities[capabilityName];
		if ('cloud.puccini.kubevirt::VirtualMachine' in capability.types)
			generateVirtualMachine(capability, metadata);
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
*/
