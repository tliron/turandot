
clout.exec('tosca.lib.traversal');

tosca.coerce();

var executions = [];

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	var nodeTemplate = vertex.properties;

	for (var interfaceName in nodeTemplate.interfaces) {
		var interface_ = nodeTemplate.interfaces[interfaceName];
		if ('cloud.puccini.turandot.orchestration::Execution' in interface_.types) {
			var operation = interface_.operations.execute;
			if (operation && operation.implementation) {
				var execution = {
					nodeTemplate: nodeTemplate.name,
					phase: operation.inputs.phase || interfaceName,
					pods: operation.inputs.pods
				};

				if (operation.inputs.requirements)
					execution.requirements = operation.inputs.requirements;

				if ('cloud.puccini.turandot.orchestration::CloutExecution' in interface_.types) {
					execution.type = 'clout';
					execution.scriptlet = operation.implementation;
					// TODO: verify that the scriptlet exists
					if (operation.inputs.arguments)
						execution.arguments = operation.inputs.arguments;
				} else if ('cloud.puccini.turandot.orchestration::CommandExecution' in interface_.types) {
					execution.command = [operation.implementation];
					if (operation.inputs.arguments)
						execution.command = execution.command.concat(operation.inputs.arguments);
					var artifacts = getArtifacts(nodeTemplate, operation.inputs.artifacts);
					if (artifacts)
						execution.artifacts = artifacts;
					if ('cloud.puccini.turandot.orchestration::ContainerExecution' in interface_.types) {
						execution.type = 'container';
						execution.container = interface_.inputs.container;
					} else if ('cloud.puccini.turandot.orchestration::SSHExecution' in interface_.types) {
						execution.type = 'ssh';
						execution.host = operation.inputs.host;
						execution.username = operation.inputs.username;
						execution.key = operation.inputs.key;
					}
				}

				executions.push(execution);
			}
		}
	}
}

puccini.write({executions: executions});

function getArtifacts(nodeTemplate, artifactNames) {
	var artifacts = [];
	if (artifactNames && nodeTemplate.artifacts)
		for (var i = 0, l = artifactNames.length; i < l; i++) {
			var artifactName = artifactNames[i];
			var artifact = nodeTemplate.artifacts[artifactName];
			if (artifact !== undefined)
				var info = {
					sourcePath: artifact.sourcePath,
					targetPath: artifact.targetPath,
				};
				if (('cloud.puccini.turandot.orchestration::Deployable' in artifact.types) && (artifact.properties.permissions !== undefined))
					info.permissions = artifact.properties.permissions;
				artifacts.push(info);
		}
	return artifacts;
}
