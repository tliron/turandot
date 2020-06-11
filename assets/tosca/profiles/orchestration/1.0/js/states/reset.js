
clout.exec('tosca.lib.utils');

var serviceName = puccini.arguments.service;
var mode = puccini.arguments.mode;

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	if (!vertex.metadata.turandot)
		continue;
	if (!vertex.metadata.turandot.states)
		continue;

	var state = vertex.metadata.turandot.states[serviceName];
	if ((state !== undefined) && (state.mode === mode))
		delete vertex.metadata.turandot.states[serviceName];
}

puccini.write(clout);
