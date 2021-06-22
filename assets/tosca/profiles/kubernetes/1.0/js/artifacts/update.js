
const tosca = require('tosca.lib.utils');

for (let vertexId in clout.vertexes) {
	let vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	let nodeTemplate = vertex.properties;

	for (let artifactName in nodeTemplate.artifacts) {
		let artifact = nodeTemplate.artifacts[artifactName];
		if (artifact.sourcePath) {
			let url = puccini.arguments[artifact.sourcePath];
			if (url !== undefined)
				artifact.$artifact = url;
		}
	}
}

tosca.addHistory('kubernetes.artifacts.update');
puccini.write(clout);
