
for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	var nodeTemplate = vertex.properties;

	for (var artifactName in nodeTemplate.artifacts) {
		var artifact = nodeTemplate.artifacts[artifactName];

		if ('cloud.puccini.kubevirt::CloudConfig' in artifact.types) {
			var cloudConfig = puccini.loadString(artifact.sourcePath);

			var variables = artifact.properties.variables;
			if (variables !== undefined) {
				variables = clout.newCoercible(variables, vertex);
				variables = variables.coerce();
				if (variables)
					for (var name in variables) {
						var r = new RegExp(escapeRegExp(name), 'g');
						cloudConfig = cloudConfig.replace(r, variables[name]);
					}
			}

			if (artifact.properties.base64)
				cloudConfig = puccini.btoa(cloudConfig);

			artifact.$artifact = cloudConfig;
		}
	}
}

function escapeRegExp(s) {
	// See: https://stackoverflow.com/a/3561711
	return s.replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&');
}
