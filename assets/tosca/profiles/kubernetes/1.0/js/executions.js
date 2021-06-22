
const traversal = require('tosca.lib.traversal');
const tosca = require('tosca.lib.utils');

// TODO: not here
for (let vertexId in clout.vertexes) {
	let vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	let nodeTemplate = vertex.properties;

	for (let artifactName in nodeTemplate.artifacts) {
		let artifact = nodeTemplate.artifacts[artifactName];

		if ('cloud.puccini.turandot.orchestration::Key' in artifact.types)
			artifact.$artifact = puccini.loadString(artifact.sourcePath);
	}
}

traversal.coerce();

let executions = [];

for (let vertexId in clout.vertexes) {
	let vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	let nodeTemplate = vertex.properties;

	// Sort names
	let interfaceNames = [];
	for (let interfaceName in nodeTemplate.interfaces)
		interfaceNames.push(interfaceName);
	interfaceNames.sort();

	for (let i = 0, l = interfaceNames.length; i < l; i++) {
		let interfaceName = interfaceNames[i];
		let interface_ = nodeTemplate.interfaces[interfaceName];

		if ('cloud.puccini.turandot.orchestration::Execution' in interface_.types) {
			let execution = {nodeTemplate: nodeTemplate.name};

			if (interface_.inputs.mode)
				execution.mode = interface_.inputs.mode;
			else {
				let last = interfaceName.lastIndexOf('.');
				if (last !== -1)
					execution.mode = interfaceName.substring(0, last);
				else
					execution.mode = interfaceName;
			}

			if (interface_.inputs.requirements)
				execution.requirements = interface_.inputs.requirements;

			if ('cloud.puccini.turandot.orchestration::Scriptlet' in interface_.types) {
				execution.type = 'clout';
				execution.scriptlet = interface_.inputs.scriptlet;
				execution.arguments = {
					service: puccini.arguments.service,
					nodeTemplate: nodeTemplate.name
				};

				// TODO: verify that the scriptlet exists
				if (interface_.inputs.arguments)
					for (let k in interface_.inputs.arguments)
						execution.arguments[k] = interface_.inputs.arguments[k];
			} else if ('cloud.puccini.kubernetes::Command' in interface_.types) {
				let artifacts = getArtifacts(nodeTemplate, interface_.inputs.artifacts);
				if (artifacts)
					execution.artifacts = artifacts;

				if ('cloud.puccini.kubernetes::ContainerCommand' in interface_.types) {
					execution.type = 'container';

					if (interface_.inputs.container)
						execution.container = interface_.inputs.container;

					let metadata = getKubernetesMetadata(nodeTemplate);
					if (metadata.namespace)
						execution.namespace = metadata.namespace;

					if (interface_.inputs.selector)
						execution.selector = interface_.inputs.selector;
					else {
						if (metadata)
							execution.selector = {matchLabels: metadata.labels};
							// TODO: matchExpressions
						else
							throw 'no pod selector for execution';
					}

					execution.pods = interface_.inputs.pods;
				} else if ('cloud.puccini.kubernetes::SSHCommand' in interface_.types) {
					execution.type = 'ssh';
					execution.host = interface_.inputs.host || '';
					execution.username = interface_.inputs.username || '';
					execution.key = interface_.inputs.key || '';
				}

				// Process special "$$" command arguments
				execution.command = interface_.inputs.command.slice();
				for (let ii = 1, ll = execution.command.length; ii < ll; ii++) {
					let arg = execution.command[ii];
					if (arg.substring(0, 2) === '$$') {
						arg = execution[arg.substring(2)];
						if (arg !== undefined)
							execution.command[ii] = arg;
					}
				}
			}

			executions.push(execution);
		}
	}
}

puccini.write({executions: executions});

function getKubernetesMetadata(nodeTemplate) {
	for (let capabilityName in nodeTemplate.capabilities) {
		let capability = nodeTemplate.capabilities[capabilityName];
		if ('cloud.puccini.kubernetes::Metadata' in capability.types)
			return puccini.deepCopy(capability.properties);
	}
	return null;
}

function getArtifacts(nodeTemplate, artifactNames) {
	let artifacts = [];
	if (artifactNames && nodeTemplate.artifacts)
		for (let i = 0, l = artifactNames.length; i < l; i++) {
			let artifactName = artifactNames[i];
			let artifact = nodeTemplate.artifacts[artifactName];
			if (artifact === undefined)
				throw puccini.sprintf('artifact not found: %s', artifactName);
			let info = {
				sourceUrl: artifact.sourcePath,
				targetPath: artifact.targetPath
			};
			if (('cloud.puccini.turandot.orchestration::Deployable' in artifact.types) && (artifact.properties.permissions !== undefined))
				info.permissions = artifact.properties.permissions;
			artifacts.push(info);
		}
	return artifacts;
}
