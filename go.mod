module github.com/tliron/turandot

go 1.16

// replace github.com/tliron/kutil => /Depot/Projects/RedHat/kutil

// replace github.com/tliron/reposure => /Depot/Projects/RedHat/reposure

// replace github.com/tliron/puccini => /Depot/Projects/RedHat/puccini

require (
	github.com/gofrs/flock v0.8.1
	github.com/google/uuid v1.3.0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jetstack/cert-manager v1.4.0
	github.com/kubernetes-sigs/reference-docs/gen-apidocs v0.0.0-20210707015243-9142bb8fe078
	github.com/mitchellh/go-wordwrap v1.0.1
	github.com/spf13/cobra v1.2.1
	github.com/tliron/kutil v0.1.47
	github.com/tliron/puccini v0.19.0
	github.com/tliron/reposure v0.1.5
	github.com/tliron/yamlkeys v1.3.5
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	k8s.io/klog/v2 v2.10.0
)
