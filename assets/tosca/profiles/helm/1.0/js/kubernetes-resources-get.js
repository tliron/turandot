
for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex, 'cloud.puccini.helm::Release'))
		continue;
	var nodeTemplate = vertex.properties;

	var tmp = '/tmp/chart.tar.gz';

	try {
		puccini.download(nodeTemplate.properties.chart, tmp);

		var name = nodeTemplate.properties.name || nodeTemplate.name;
		var namespace = nodeTemplate.properties.namespace;

		var args = [
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

		var values = nodeTemplate.properties.values;
		if (values)
			for (var key in values)
				args.push('--set-string', puccini.sprintf('%s=%s', key, values[key]));

		args.push(name)
		args.push(tmp);

		puccini.log.infof('helm %s', args.join(' '));
		var output = puccini.exec('helm', args);
		output = JSON.parse(output);
		//puccini.log.infof('%s', JSON.stringify(output, null, ' '));
		if (output.manifest) {
			var manifests = puccini.decode(output.manifest, 'yaml', true);
			for (var m = 0, l = manifests.length; m < l; m++) {
				var manifest = manifests[m];
				processManifest(manifest, name, namespace);
				resources.push(manifest);
			}
		}
		if (nodeTemplate.properties.hooks && output.hooks) {
			for (var h = 0, l = output.hooks.length; h < l; h++) {
				var manifest = puccini.decode(output.hooks[h].manifest, 'yaml');
				processManifest(manifest, name, namespace);
				resources.push(manifest);
			}
		}
	} finally {
		puccini.exec('rm', ['--force', tmp]);
	}
}

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
