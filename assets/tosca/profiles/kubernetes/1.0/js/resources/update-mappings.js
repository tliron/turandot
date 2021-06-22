
const tosca = require('tosca.lib.utils');

for (let vertexId in puccini.arguments) {
	let mappings = puccini.arguments[vertexId];
	let vertex = clout.vertexes[vertexId];
	if (vertex === undefined)
		continue;

	if (!vertex.metadata.turandot)
		vertex.metadata.turandot = {};
	vertex.metadata.turandot.version = '1.0';
	vertex.metadata.turandot.resources = JSON.parse(mappings);
}

tosca.addHistory('kubernetes.resources.update-mappings');
puccini.write(clout);
