
const traversal = require('tosca.lib.traversal');
const tosca = require('tosca.lib.utils');

traversal.coerce();

let artifacts = [];

for (let vertexId in clout.vertexes) {
	let vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	let nodeTemplate = vertex.properties;

	for (let artifactName in nodeTemplate.artifacts) {
		let artifact = nodeTemplate.artifacts[artifactName];

		if ('cloud.puccini.kubernetes::Registry' in artifact.types)
			artifacts.push({
				name: artifact.properties.name,
				registry: artifact.properties.registry,
				sourcePath: artifact.sourcePath
			});
	}
}

puccini.write({artifacts: artifacts});
