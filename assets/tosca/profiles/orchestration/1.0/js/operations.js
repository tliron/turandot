
clout.exec('tosca.lib.traversal');

tosca.coerce();

var operations = [];

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex))
		continue;
	var nodeTemplate = vertex.properties;

	for (var interfaceName in nodeTemplate.interfaces) {
		var interface_ = nodeTemplate.interfaces[interfaceName];
		if ('cloud.puccini.turandot.orchestration::ScriptletOrchestration' in interface_.types) {
			for (var operationName in interface_.operations) {
				var operation = interface_.operations[operationName];
				generateScriptletOperation(vertexId, interface_, operation);
			}
		}
	}
}

puccini.write({operations: operations});

function generateScriptletOperation(vertexId, interface_, operation) {
	if (operation.implementation) {
		operations.push({
			vertexId: vertexId,
			type: 'scriptlet',
			scriptletName: operation.implementation
		});
	}
}
