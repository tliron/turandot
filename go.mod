module github.com/tliron/turandot

go 1.14

replace github.com/tliron/puccini => /Depot/Projects/RedHat/puccini

// replace github.com/tliron/kubernetes-registry-spooler => /Depot/Projects/RedHat/kubernetes-registry-spooler

require (
	github.com/gofrs/flock v0.7.1
	github.com/google/go-containerregistry v0.1.1
	github.com/google/uuid v1.1.1
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/klauspost/pgzip v1.2.4
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/tebeka/atexit v0.3.0
	github.com/tliron/kubernetes-registry-spooler v1.0.6
	github.com/tliron/puccini v0.0.0-00010101000000-000000000000
	github.com/tliron/yamlkeys v1.3.3
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	k8s.io/api v0.18.4
	k8s.io/apiextensions-apiserver v0.18.4
	k8s.io/apimachinery v0.18.4
	k8s.io/apiserver v0.18.4 // indirect
	k8s.io/client-go v0.18.4
)
