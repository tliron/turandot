package common

import (
	"io/ioutil"
	"strings"
)

// See: https://github.com/kubernetes/kubernetes/pull/63707
func GetServiceAccountNamespace() string {
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if namespace := strings.TrimSpace(string(data)); len(namespace) > 0 {
			return namespace
		}
	}
	return ""
}
