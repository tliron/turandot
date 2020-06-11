
clout.exec('tosca.lib.utils');

var serviceName = puccini.arguments.service;
var nodeTemplateName = puccini.arguments.nodeTemplate;

var mode = puccini.arguments.mode;
var state = puccini.arguments.state;
var message = puccini.arguments.message;

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	var nodeTemplate = vertex.properties;
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
