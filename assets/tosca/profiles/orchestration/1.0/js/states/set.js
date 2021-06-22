
const tosca = require('tosca.lib.utils');

let serviceName = puccini.arguments.service;
let nodeTemplateName = puccini.arguments.nodeTemplate;

let mode = puccini.arguments.mode;
let state = puccini.arguments.state;
let message = puccini.arguments.message;

for (let vertexId in clout.vertexes) {
	let vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	let nodeTemplate = vertex.properties;
	if (nodeTemplate.name !== nodeTemplateName)
		continue;

	if (!vertex.metadata.turandot)
		vertex.metadata.turandot = {};
	if (!vertex.metadata.turandot.states)
		vertex.metadata.turandot.states = {}

	vertex.metadata.turandot.states[serviceName] = {
		mode: mode,
		state: state,
		message: message
	};

	puccini.write(clout);
	break;
}
