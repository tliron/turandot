
const tosca = require('tosca.lib.utils');

let serviceName = puccini.arguments.service;
let mode = puccini.arguments.mode;

for (let vertexId in clout.vertexes) {
	let vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	if (!vertex.metadata.turandot)
		continue;
	if (!vertex.metadata.turandot.states)
		continue;

	let state = vertex.metadata.turandot.states[serviceName];
	if ((state !== undefined) && (state.mode === mode))
		delete vertex.metadata.turandot.states[serviceName];
}

puccini.write(clout);
