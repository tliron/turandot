
clout.exec('tosca.lib.traversal');

tosca.coerce();

var artifacts = [];

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	var nodeTemplate = vertex.properties;

	for (var artifactName in nodeTemplate.artifacts) {
		var artifact = nodeTemplate.artifacts[artifactName];

		if ('cloud.puccini.kubernetes::Registry' in artifact.types)
			artifacts.push({
				name: artifact.properties.name,
				registry: artifact.properties.registry,
				sourcePath: artifact.sourcePath
			});
	}
}

puccini.write({artifacts: artifacts});
