module github.com/tliron/turandot

go 1.15

replace github.com/tliron/puccini => /Depot/Projects/RedHat/puccini

// replace github.com/tliron/reposure => /Depot/Projects/RedHat/reposure

require (
	github.com/gofrs/flock v0.8.0
	github.com/google/go-containerregistry v0.2.1
	github.com/google/uuid v1.1.2
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jetstack/cert-manager v1.1.0
	github.com/klauspost/pgzip v1.2.5
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/spf13/cobra v1.1.1
	github.com/tebeka/atexit v0.3.0
	github.com/tliron/kutil v0.1.8
	github.com/tliron/puccini v0.0.0-00010101000000-000000000000
	github.com/tliron/reposure v0.1.0
	github.com/tliron/yamlkeys v1.3.5
	golang.org/x/crypto v0.0.0-20201203163018-be400aefbc4c
	k8s.io/api v0.19.4
	k8s.io/apiextensions-apiserver v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
)
