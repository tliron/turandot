
clout.exec('tosca.lib.traversal');

tosca.coerce();

var policies = {};

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isTosca(vertex, 'Policy'))
		continue;
	var policy = vertex.properties;

	if ('cloud.puccini.turandot.orchestration::Provisioning' in policy.types)
		generatePolicy(policy, tosca.getPolicyTargets(vertex), 'provisioning');
}

puccini.write(policies);

function generatePolicy(policy, targets, type) {
	for (var t = 0, l = targets.length; t < l; t++) {
		var target = targets[t];
 		var targetPolicies = policies[target.name];
 		if (targetPolicies === undefined)
 			targetPolicies = policies[target.name] = [];
 		targetPolicies.push({
 			type: type,
 			properties: policy.properties,
 		});
	}
}