
const tosca = require('tosca.lib.utils');

let states = {};

for (let vertexId in clout.vertexes) {
	let vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	let nodeTemplate = vertex.properties;

	if (vertex.metadata.turandot && vertex.metadata.turandot.states)
		for (let serviceName in vertex.metadata.turandot.states) {
			let nodeState = vertex.metadata.turandot.states[serviceName];
			let serviceStates = states[serviceName];
			if (serviceStates === undefined)
				serviceStates = states[serviceName] = {};
			serviceStates[nodeTemplate.name] = nodeState;
		}
}

puccini.write(states);
