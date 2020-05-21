
clout.exec('tosca.lib.traversal');

tosca.coerce();

for (var vertexId in clout.vertexes) {
	var vertex = clout.vertexes[vertexId];
	if (!tosca.isNodeTemplate(vertex, 'cloud.puccini.helm::Release'))
		continue;
	var nodeTemplate = vertex.properties;

	var tmp = '/tmp/chart.tar.gz';

	try {
		puccini.download(nodeTemplate.properties.chart, tmp);

		var args = ['install'];

		var valuesUrl = nodeTemplate.properties.valuesUrl;
		if (valuesUrl) {
			args.push('--values');
			args.push(valueUrl);
		}

		var values = nodeTemplate.properties.values;
		if (values)
			for (var key in values) {
				var value = values[key];
				args.push('--set');
				args.push(puccini.sprintf('%s=%s', key, value));
			}

		args.push(nodeTemplate.properties.name || nodeTemplate.name)
		args.push(tmp);
		puccini.log.infof('helm %s', args.join(' '));

		puccini.write(puccini.exec('helm', args));
	} finally {
		puccini.exec('rm', ['--force', tmp]);
	}
}
