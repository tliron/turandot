
var nodeTemplateName = puccini.arguments.nodeTemplate;
var name = puccini.arguments.name;
var value = puccini.arguments.value;

puccini.log.infof('node template: %s', nodeTemplateName);
puccini.log.infof('output name: %s', name);
puccini.log.infof('output value: %s', value);

if (setOutput(name, value))
	puccini.write(clout);

function setOutput(name, value) {
	var output = clout.properties.tosca.outputs[name];
	if (output === undefined)
		return false;

	if (output.$information && output.$information.type)
		switch (output.$information.type.name) {
		case 'boolean':
			value = (value === 'true');
			break;
		case 'integer':
			value = parseInt(value);
			break;
		case 'float':
			value = parseFloat(value);
			break;
		}

	output.$value = value;
	return true;
}