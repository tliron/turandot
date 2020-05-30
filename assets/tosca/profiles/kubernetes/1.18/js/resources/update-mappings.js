
clout.exec('tosca.lib.utils');

for (var vertexId in puccini.arguments) {
	var mappings = puccini.arguments[vertexId];
	var vertex = clout.vertexes[vertexId];
	if (vertex === undefined)
		continue;

	if (!vertex.metadata.turandot)
		vertex.metadata.turandot = {};
	vertex.metadata.turandot.version = '1.0';
	vertex.metadata.turandot.resources = JSON.parse(mappings);
}

tosca.addHistory('kubernetes.resources.update-mappings');
puccini.write(clout);
