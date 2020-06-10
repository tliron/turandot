
clout.exec('tosca.lib.traversal');

var states = {};

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	var nodeTemplate = vertex.properties;

	if (vertex.metadata.turandot && vertex.metadata.turandot.states)
		states[nodeTemplate.name] = vertex.metadata.turandot.states;
}

puccini.write(states);
