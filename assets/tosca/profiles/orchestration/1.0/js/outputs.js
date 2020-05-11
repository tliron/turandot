
clout.exec('tosca.lib.traversal');

tosca.coerce();

var outputs = [];

if (clout.properties.tosca && clout.properties.tosca.outputs) {
	outputs = clout.properties.tosca.outputs;
}

puccini.write({outputs: outputs});
