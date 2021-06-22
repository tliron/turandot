
const tosca = require('tosca.lib.utils');

exports.plugin = function(resources) {
	for (let vertexId in clout.vertexes) {
		let vertex = clout.vertexes[vertexId];
		if (!tosca.isNodeTemplate(vertex, 'cloud.puccini.helm::Release'))
			continue;
		let nodeTemplate = vertex.properties;

		let tmp = puccini.temporaryFile('helm-chart-*.tar.gz');

		try {
			puccini.download(nodeTemplate.properties.chart, tmp);

			let name = nodeTemplate.properties.name || nodeTemplate.name;
			let namespace = nodeTemplate.properties.namespace;

			let args = [
				'install',
				'--dry-run',
				'--output', 'json'
			];

			if (nodeTemplate.properties.version)
				args.push('--version', nodeTemplate.properties.version);

			if (namespace)
				args.push('--namespace', namespace);

			if (nodeTemplate.properties.valuesUrl)
				args.push('--values', nodeTemplate.properties.valuesUrl);

			let values = nodeTemplate.properties.values;
			if (values)
				for (let key in values)
					args.push('--set-string', puccini.sprintf('%s=%s', key, values[key]));

			args.push(name)
			args.push(tmp);

			puccini.log.infof('helm %s', args.join(' '));
			let output = puccini.exec('helm', args);
			output = JSON.parse(output);
			//puccini.log.infof('%s', JSON.stringify(output, null, ' '));
			if (output.manifest) {
				let manifests = puccini.decode(output.manifest, 'yaml', true);
				for (let m = 0, l = manifests.length; m < l; m++) {
					let manifest = manifests[m];
					processManifest(manifest, name, namespace);
					resources.push(manifest);
				}
			}
			if (nodeTemplate.properties.hooks && output.hooks) {
				for (let h = 0, l = output.hooks.length; h < l; h++) {
					let manifest = puccini.decode(output.hooks[h].manifest, 'yaml');
					processManifest(manifest, name, namespace);
					resources.push(manifest);
				}
			}
		} finally {
			puccini.exec('rm', ['--force', tmp]);
		}
	}
};

function processManifest(manifest, name, namespace) {
	if (!manifest.metadata)
		manifest.metadata = {};
	if (!manifest.metadata.annotations)
		manifest.metadata.annotations = {};
	if (!manifest.metadata.annotations['meta.helm.sh/release-name'])
		manifest.metadata.annotations['meta.helm.sh/release-name'] = name;
	if (!manifest.metadata.annotations['meta.helm.sh/release-namespace'] && namespace)
		manifest.metadata.annotations['meta.helm.sh/release-namespace'] = namespace;
}
