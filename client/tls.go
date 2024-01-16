package client

import (
	"crypto/x509"
	"fmt"

	"github.com/tliron/kutil/util"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const tlsMountPath = "/tls"

var tlsCertificatePath = fmt.Sprintf("%s/%s", tlsMountPath, core.TLSCertKey)
var tlsKeyPath = fmt.Sprintf("%s/%s", tlsMountPath, core.TLSPrivateKeyKey)

func (self *Client) GetSecret(namespace string, secretName string) (*core.Secret, error) {
	return self.Kubernetes.CoreV1().Secrets(namespace).Get(self.Context, secretName, meta.GetOptions{})
}

func (self *Client) GetSecretTLSCertPool(namespace string, secretName string, secretDataKey string) (*x509.CertPool, error) {
	if secret, err := self.GetSecret(namespace, secretName); err == nil {
		if secretDataKey == "" {
			secretDataKey = core.TLSCertKey
		}

		switch secret.Type {
		case core.SecretTypeTLS, core.SecretTypeServiceAccountToken:
			if bytes, ok := secret.Data[secretDataKey]; ok {
				return util.ParseX509CertificatePool(bytes)
			} else {
				return nil, fmt.Errorf("no data key %q in %q secret: %s", secretDataKey, secret.Type, secret.Data)
			}

		default:
			return nil, fmt.Errorf("unsupported TLS secret type: %s", secret.Type)
		}
	} else {
		return nil, err
	}
}
