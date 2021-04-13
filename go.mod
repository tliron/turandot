module github.com/tliron/turandot

go 1.16

// replace github.com/tliron/kutil => /Depot/Projects/RedHat/kutil

// replace github.com/tliron/reposure => /Depot/Projects/RedHat/reposure

// replace github.com/tliron/puccini => /Depot/Projects/RedHat/puccini

require (
	github.com/gofrs/flock v0.8.0
	github.com/google/uuid v1.2.0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jetstack/cert-manager v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/tliron/kutil v0.1.25
	github.com/tliron/puccini v0.18.0
	github.com/tliron/reposure v0.1.3
	github.com/tliron/yamlkeys v1.3.5
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.21.0
	k8s.io/apiextensions-apiserver v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/klog/v2 v2.8.0
)
