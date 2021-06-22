
const traversal = require('tosca.lib.traversal');
const tosca = require('tosca.lib.utils');

traversal.coerce();

let policies = {};

for (let vertexId in clout.vertexes) {
	let vertex = clout.vertexes[vertexId];
	if (!tosca.isTosca(vertex, 'Policy'))
		continue;
	let policy = vertex.properties;

	if ('cloud.puccini.turandot.orchestration::Provisioning' in policy.types)
		generatePolicy(policy, tosca.getPolicyTargets(vertex), 'provisioning');
}

puccini.write(policies);

function generatePolicy(policy, targets, type) {
	for (let t = 0, l = targets.length; t < l; t++) {
		let target = targets[t];
 		let targetPolicies = policies[target.name];
 		if (targetPolicies === undefined)
 			targetPolicies = policies[target.name] = [];
 		targetPolicies.push({
 			type: type,
 			properties: policy.properties,
 		});
	}
}