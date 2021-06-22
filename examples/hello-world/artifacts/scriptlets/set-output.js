
const tosca = require('tosca.lib.utils');

let nodeTemplateName = puccini.arguments.nodeTemplate;
let name = puccini.arguments.name;
let value = puccini.arguments.value;

puccini.log.infof('execution for node template %q, setting output: %s -> %s', nodeTemplateName, name, value);

if (tosca.setOutputValue(name, value))
	puccini.write(clout);
