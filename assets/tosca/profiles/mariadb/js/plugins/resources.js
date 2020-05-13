
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

	for (var capabilityName in nodeTemplate.capabilities) {
		var capability = nodeTemplate.capabilities[capabilityName];
		var capabilityMetadata = puccini.deepCopy(metadata);
		capabilityMetadata.annotations['puccini.cloud/capability'] = capabilityName;
		if ('cloud.puccini.kubevirt::Cluster' in capability.types)
			generateMariaDb(capability, capabilityMetadata);
	}
}

function generateMariaDb(nodeTemplate, metadata) {
	var spec = {
		apiVersion: 'mariadb.persistentsys/v1alpha1',
		kind: 'MariaDB',
		metadata: metadata,
		spec: nodeTemplate.properties,
	};

	specs.push(spec);
}
