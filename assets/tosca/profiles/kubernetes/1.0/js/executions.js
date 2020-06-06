
clout.exec('tosca.lib.traversal');

tosca.coerce();

var executions = [];

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	var nodeTemplate = vertex.properties;

	// Sort names
	var interfaceNames = [];
	for (var interfaceName in nodeTemplate.interfaces)
		interfaceNames.push(interfaceName);
	interfaceNames.sort();

	for (var i = 0, l = interfaceNames.length; i < l; i++) {
		var interfaceName = interfaceNames[i]; 
		var interface_ = nodeTemplate.interfaces[interfaceName];

		if ('cloud.puccini.turandot.orchestration::Execution' in interface_.types) {
			var operation = interface_.operations.execute;
			if (operation && operation.implementation) {
				var execution = {nodeTemplate: nodeTemplate.name};

				if (operation.inputs.mode)
					execution.mode = operation.inputs.mode;
				else {
					var last = interfaceName.lastIndexOf('.');
					if (last !== -1)
						execution.mode = interfaceName.substring(0, last);
					else
						execution.mode = interfaceName;
				}

				if (operation.inputs.requirements)
					execution.requirements = operation.inputs.requirements;

				if ('cloud.puccini.turandot.orchestration::CloutExecution' in interface_.types) {
					execution.type = 'clout';
					execution.scriptlet = operation.implementation;
					execution.arguments = {nodeTemplate: nodeTemplate.name};
					// TODO: verify that the scriptlet exists
					if (operation.inputs.arguments)
						for (var k in operation.inputs.arguments)
							execution.arguments[k] = operation.inputs.arguments[k];
				} else if ('cloud.puccini.kubernetes::Execution' in interface_.types) {
					var metadata = getKubernetesMetadata(nodeTemplate);
					if (metadata.namespace)
						execution.namespace = metadata.namespace;
					if (operation.inputs.selector)
						execution.selector = operation.inputs.selector;
					else {
						if (metadata)
							execution.selector = {matchLabels: metadata.labels};
							// TODO: matchExpressions
						else
							throw 'no pod selector for execution';
					}

					execution.pods = operation.inputs.pods;
					execution.command = [operation.implementation];
					if (operation.inputs.arguments)
						execution.command = execution.command.concat(operation.inputs.arguments);
					var artifacts = getArtifacts(nodeTemplate, operation.inputs.artifacts);
					if (artifacts)
						execution.artifacts = artifacts;

					if ('cloud.puccini.kubernetes::ContainerExecution' in interface_.types) {
						execution.type = 'container';
						if (operation.inputs.container)
							execution.container = operation.inputs.container;
					} else if ('cloud.puccini.kubernetes::SSHExecution' in interface_.types) {
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

function getKubernetesMetadata(nodeTemplate) {
	for (var capabilityName in nodeTemplate.capabilities) {
		var capability = nodeTemplate.capabilities[capabilityName];
		if ('cloud.puccini.kubernetes::Metadata' in capability.types)
			return puccini.deepCopy(capability.properties);
	}
	return null;
}

function getArtifacts(nodeTemplate, artifactNames) {
	var artifacts = [];
	if (artifactNames && nodeTemplate.artifacts)
		for (var i = 0, l = artifactNames.length; i < l; i++) {
			var artifactName = artifactNames[i];
			var artifact = nodeTemplate.artifacts[artifactName];
			if (artifact === undefined)
				throw puccini.sprintf('artifact not found: %s', artifactName);
			var info = {
				sourcePath: artifact.sourcePath,
				targetPath: artifact.targetPath
			};
			if (('cloud.puccini.turandot.orchestration::Deployable' in artifact.types) && (artifact.properties.permissions !== undefined))
				info.permissions = artifact.properties.permissions;
			artifacts.push(info);
		}
	return artifacts;
}
