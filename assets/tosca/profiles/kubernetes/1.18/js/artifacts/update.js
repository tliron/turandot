
clout.exec('tosca.lib.utils');

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	var nodeTemplate = vertex.properties;

	for (var artifactName in nodeTemplate.artifacts) {
		var artifact = nodeTemplate.artifacts[artifactName];
		if (artifact.sourcePath) {
			var url = puccini.arguments[artifact.sourcePath];
			if (url !== undefined)
				artifact.$artifact = url;
		}
	}
}

tosca.addHistory('kubernetes.artifacts.update');
puccini.write(clout);
