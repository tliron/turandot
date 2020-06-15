
clout.exec('tosca.lib.traversal');

var states = {};

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	var nodeTemplate = vertex.properties;

	if (vertex.metadata.turandot && vertex.metadata.turandot.states)
		for (var serviceName in vertex.metadata.turandot.states) {
			var nodeState = vertex.metadata.turandot.states[serviceName];
			var serviceStates = states[serviceName];
			if (serviceStates === undefined)
				serviceStates = states[serviceName] = {};
			serviceStates[nodeTemplate.name] = nodeState;
		}
}

puccini.write(states);
