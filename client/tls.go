package client

import (
	"crypto/x509"
	"encoding/pem"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (self *Client) GetSecret(namespace string, secretName string) (*core.Secret, error) {
	return self.Kubernetes.CoreV1().Secrets(namespace).Get(self.Context, secretName, meta.GetOptions{})
}

func (self *Client) GetSecretCertificate(namespace string, secretName string) (*x509.Certificate, error) {
	if secret, err := self.GetSecret(namespace, secretName); err == nil {
		bytes := secret.Data[core.TLSCertKey]
		block, _ := pem.Decode(bytes)
		return x509.ParseCertificate(block.Bytes)
	} else {
		return nil, err
	}
}
