
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

	if ('cloud.puccini.mariadb::Cluster' in nodeTemplate.types)
		generateMariaDb(nodeTemplate, metadata);
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
