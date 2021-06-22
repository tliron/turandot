
let mappings = {};

for (let vertexId in clout.vertexes) {
	let vertex = clout.vertexes[vertexId];
	if (!vertex.metadata.turandot ||
		(vertex.metadata.turandot.version !== '1.0'))
		continue;

	if (vertex.metadata.turandot.resources)
		mappings[vertexId] = vertex.metadata.turandot.resources;
}

puccini.write(mappings);
