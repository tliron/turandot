
const tosca = require('tosca.lib.utils');

exports.plugin = function() {
	for (let vertexId in clout.vertexes) {
		let vertex = clout.vertexes[vertexId];
		if (!tosca.isNodeTemplate(vertex))
			continue;
		let nodeTemplate = vertex.properties;

		for (let artifactName in nodeTemplate.artifacts) {
			let artifact = nodeTemplate.artifacts[artifactName];

			if ('cloud.puccini.kubevirt::CloudConfig' in artifact.types) {
				let cloudConfig = puccini.loadString(artifact.sourcePath);

				let variables = artifact.properties.variables;
				if (variables !== undefined) {
					variables = clout.newCoercible(variables, vertex);
					variables = variables.coerce();
					if (variables)
						for (let name in variables) {
							let r = new RegExp(escapeRegExp(name), 'g');
							cloudConfig = cloudConfig.replace(r, variables[name]);
						}
				}

				if (artifact.properties.base64)
					cloudConfig = puccini.btoa(cloudConfig);

				artifact.$artifact = cloudConfig;
			}
		}
	}
}

function escapeRegExp(s) {
	// See: https://stackoverflow.com/a/3561711
	return s.replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&');
}
