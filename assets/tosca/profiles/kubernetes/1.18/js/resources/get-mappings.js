
clout.exec('tosca.lib.utils');

var mappings = {};

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!vertex.metadata.turandot ||
		(vertex.metadata.turandot.version !== '1.0'))
		continue;

	if (vertex.metadata.turandot.resources)
		mappings[vertexId] = vertex.metadata.turandot.resources;
}

puccini.write(mappings);
