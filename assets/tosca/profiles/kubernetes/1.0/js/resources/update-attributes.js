
const tosca = require('tosca.lib.utils');

let changed = false;

for (let vertexId in puccini.arguments) {
	let vertex = clout.vertexes[vertexId];
	if (vertex === undefined)
		continue;
	if (!tosca.isNodeTemplate(vertex))
		continue;
	let nodeTemplate = vertex.properties;

	let values = JSON.parse(puccini.arguments[vertexId]);
	for (let i = 0, l = values.length; i < l; i++) {
		let value = values[i];
		let capability = nodeTemplate.capabilities[value.capability];
		if (capability === undefined)
			continue;
		let attribute = getParameter(capability.attributes, value.attribute)
		if (setValue(attribute, value.value)) {
			puccini.log.infof('set capability "%s" attribute "%s" to %s', value.capability, value.attribute, JSON.stringify(value.value));
			changed = true;
		}
	}
}

if (changed) {
	tosca.addHistory('kubernetes.resources.update-attributes');
	puccini.write(clout);
}

function getParameter(o, name) {
	let path = name.split('.');
	for (let i = 0, l = path.length; i < l; i++) {
		o = o[path[i]];
		if (o === undefined)
			throw 'could not find parameter ' + name;
	}
	return o;
}

function setValue(parameter, value) {
	let coercible = toCoercible(value);
	if (coercible.$list !== undefined) {
		if (!puccini.deepEquals(parameter.$list, coercible.$list)) {
			parameter.$list = coercible.$list;
			delete parameter.$value;
			delete parameter.$map;
			delete parameter.$functionCall;
			return true;
		}
	} else if (coercible.$map !== undefined) {
		if (!puccini.deepEquals(parameter.$map, coercible.$map)) {
			parameter.$map = coercible.$map;
			delete parameter.$value;
			delete parameter.$list;
			delete parameter.$functionCall;
			return true;
		}
	} else {
		if (!puccini.deepEquals(parameter.$value, coercible.$value)) {
			parameter.$value = coercible.$value;
			delete parameter.$list;
			delete parameter.$map;
			delete parameter.$functionCall;
			return true;
		}
	}
	return false;
}

function toCoercible(value) {
	if (Array.isArray(value)) {
		let list = [];
		for (let i = 0, l = value.length; i < l; i++)
			list.push(toCoercible(value[i]));
		return {$list: list};
	} else if ((value !== null) && (typeof value === 'object')) {
		let map = [];
		for (let k in value) {
			let entry = toCoercible(value[k]);
			entry.$key = toCoercible(k);
			map.push(entry);
		}
		return {$map: map};
	} else
		return {$value: value};
}
