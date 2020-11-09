package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/tliron/kutil/kubernetes"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const serviceTemplateArtifactCategory = "service-templates"

func (self *Client) GetRepository(namespace string, repositoryName string) (*resources.Repository, error) {
	// Default to same namespace as operator
	if namespace == "" {
		namespace = self.Namespace
	}

	if repository, err := self.Turandot.TurandotV1alpha1().Repositories(namespace).Get(self.Context, repositoryName, meta.GetOptions{}); err == nil {
		// When retrieved from cache the GVK may be empty
		if repository.Kind == "" {
			repository = repository.DeepCopy()
			repository.APIVersion, repository.Kind = resources.RepositoryGVK.ToAPIVersionAndKind()
		}
		return repository, nil
	} else {
		return nil, err
	}
}

func (self *Client) ListRepositories() (*resources.RepositoryList, error) {
	// TODO: all repositories in cluster mode
	return self.Turandot.TurandotV1alpha1().Repositories(self.Namespace).List(self.Context, meta.ListOptions{})
}

func (self *Client) CreateRepositoryDirect(namespace string, repositoryName string, host string, tlsSecretName string, tlsSecretDataKey string, authSecretName string) (*resources.Repository, error) {
	// Default to same namespace as operator
	if namespace == "" {
		namespace = self.Namespace
	}

	repository := &resources.Repository{
		ObjectMeta: meta.ObjectMeta{
			Name:      repositoryName,
			Namespace: namespace,
		},
		Spec: resources.RepositorySpec{
			Type: resources.RepositoryTypeRegistry,
			Direct: &resources.RepositoryDirect{
				Host: host,
			},
			TLSSecret:        tlsSecretName,
			TLSSecretDataKey: tlsSecretDataKey,
			AuthSecret:       authSecretName,
		},
	}

	return self.createRepository(namespace, repositoryName, repository)
}

func (self *Client) CreateRepositoryIndirect(namespace string, repositoryName string, serviceNamespace string, serviceName string, port uint64, tlsSecretName string, tlsSecretDataKey string, authSecretName string) (*resources.Repository, error) {
	// Default to same namespace as operator
	if namespace == "" {
		namespace = self.Namespace
	}

	repository := &resources.Repository{
		ObjectMeta: meta.ObjectMeta{
			Name:      repositoryName,
			Namespace: namespace,
		},
		Spec: resources.RepositorySpec{
			Type: resources.RepositoryTypeRegistry,
			Indirect: &resources.RepositoryIndirect{
				Namespace: serviceNamespace,
				Service:   serviceName,
				Port:      port,
			},
			TLSSecret:        tlsSecretName,
			TLSSecretDataKey: tlsSecretDataKey,
			AuthSecret:       authSecretName,
		},
	}

	return self.createRepository(namespace, repositoryName, repository)
}

func (self *Client) createRepository(namespace string, repositoryName string, repository *resources.Repository) (*resources.Repository, error) {
	if repository, err := self.Turandot.TurandotV1alpha1().Repositories(namespace).Create(self.Context, repository, meta.CreateOptions{}); err == nil {
		return repository, nil
	} else if errors.IsAlreadyExists(err) {
		return self.Turandot.TurandotV1alpha1().Repositories(namespace).Get(self.Context, repositoryName, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) UpdateRepositoryStatus(repository *resources.Repository) (*resources.Repository, error) {
	if repository_, err := self.Turandot.TurandotV1alpha1().Repositories(repository.Namespace).UpdateStatus(self.Context, repository, meta.UpdateOptions{}); err == nil {
		// When retrieved from cache the GVK may be empty
		if repository_.Kind == "" {
			repository_ = repository_.DeepCopy()
			repository_.APIVersion, repository_.Kind = resources.RepositoryGVK.ToAPIVersionAndKind()
		}
		return repository_, nil
	} else {
		return repository, err
	}
}

func (self *Client) DeleteRepository(namespace string, repositoryName string) error {
	// Default to same namespace as operator
	if namespace == "" {
		namespace = self.Namespace
	}

	return self.Turandot.TurandotV1alpha1().Repositories(namespace).Delete(self.Context, repositoryName, meta.DeleteOptions{})
}

func (self *Client) GetRepositoryHost(repository *resources.Repository) (string, error) {
	if (repository.Spec.Direct != nil) && (repository.Spec.Direct.Host != "") {
		return repository.Spec.Direct.Host, nil
	} else if (repository.Spec.Indirect != nil) && (repository.Spec.Indirect.Service != "") {
		serviceNamespace := repository.Spec.Indirect.Namespace
		if serviceNamespace == "" {
			// Default to repository namespace
			serviceNamespace = repository.Namespace
		}

		if service, err := self.Kubernetes.CoreV1().Services(serviceNamespace).Get(self.Context, repository.Spec.Indirect.Service, meta.GetOptions{}); err == nil {
			return fmt.Sprintf("%s:%d", service.Spec.ClusterIP, repository.Spec.Indirect.Port), nil
		} else {
			return "", err
		}
	} else {
		return "", fmt.Errorf("malformed repository: %s", repository.Name)
	}
}

func (self *Client) GetRepositoryCertificatePath(repository *resources.Repository) string {
	if repository.Spec.TLSSecret != "" {
		secretDataKey := repository.Spec.TLSSecretDataKey
		if secretDataKey == "" {
			secretDataKey = core.TLSCertKey
		}
		return fmt.Sprintf("%s/%s", tlsMountPath, secretDataKey)
	} else {
		return ""
	}
}

func (self *Client) GetRepositoryTLSCertPool(repository *resources.Repository) (*x509.CertPool, error) {
	if repository.Spec.TLSSecret != "" {
		return self.GetSecretTLSCertPool(repository.Namespace, repository.Spec.TLSSecret, repository.Spec.TLSSecretDataKey)
	} else {
		return nil, nil
	}
}

// TODO: GetServiceHTTPRoundTripper, GetServiceAuth, GetServiceRemoteOptions

func (self *Client) GetRepositoryHTTPRoundTripper(repository *resources.Repository) (string, http.RoundTripper, error) {
	if certPool, err := self.GetRepositoryTLSCertPool(repository); err == nil {
		if certPool != nil {
			if host, err := self.GetRepositoryHost(repository); err == nil {
				roundTripper := util.NewForceHTTPSRoundTripper(&http.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs: certPool,
					},
				})
				return host, roundTripper, nil
			} else {
				return "", nil, err
			}
		} else {
			return "", nil, nil
		}
	} else {
		return "", nil, err
	}
}

func (self *Client) GetRepositoryAuth(repository *resources.Repository) (string, string, string, string, error) {
	if host, err := self.GetRepositoryHost(repository); err == nil {
		if repository.Spec.AuthSecret != "" {
			if authSecret, err := self.Kubernetes.CoreV1().Secrets(self.Namespace).Get(self.Context, repository.Spec.AuthSecret, meta.GetOptions{}); err == nil {
				switch authSecret.Type {
				case core.SecretTypeServiceAccountToken:
					if data, ok := authSecret.Data[core.ServiceAccountTokenKey]; ok {
						// OpenShift: you can also get a valid token from "oc whoami -t"
						token := util.BytesToString(data)
						return host, "", "", token, nil
					} else {
						return "", "", "", "", fmt.Errorf("malformed %q secret: %s", core.ServiceAccountTokenKey, authSecret.Data)
					}

				case core.SecretTypeDockerConfigJson, core.SecretTypeDockercfg:
					if table, err := kubernetes.NewRegistryCredentialsTableFromSecret(authSecret); err == nil {
						if credentials, ok := table[host]; ok {
							return host, credentials.Username, credentials.Password, "", nil
						}
					} else {
						return "", "", "", "", err
					}
				}
			} else {
				return "", "", "", "", err
			}
		}
	} else {
		return "", "", "", "", err
	}

	return "", "", "", "", nil
}

func (self *Client) GetRepositoryRemoteOptions(repository *resources.Repository) ([]remote.Option, error) {
	var options []remote.Option

	if _, roundTripper, err := self.GetRepositoryHTTPRoundTripper(repository); err == nil {
		if roundTripper != nil {
			options = append(options, remote.WithTransport(roundTripper))
		}
	} else {
		return nil, err
	}

	if _, username, password, token, err := self.GetRepositoryAuth(repository); err == nil {
		if (username != "") || (token != "") {
			authenticator := authn.FromConfig(authn.AuthConfig{
				Username:      username,
				Password:      password,
				RegistryToken: token,
			})
			options = append(options, remote.WithAuth(authenticator))
		}
	} else {
		return nil, err
	}

	return options, nil
}

func (self *Client) UpdateRepositoryURLContext(repository *resources.Repository, urlContext *urlpkg.Context) error {
	if host, roundTripper, err := self.GetRepositoryHTTPRoundTripper(repository); err == nil {
		if roundTripper != nil {
			urlContext.SetHTTPRoundTripper(host, roundTripper)
		}
	}

	if host, username, password, token, err := self.GetRepositoryAuth(repository); err == nil {
		if (username != "") || (token != "") {
			urlContext.SetCredentials(host, username, password, token)
		}
	} else {
		return err
	}

	return nil
}

func (self *Client) GetRepositoryURLForCSAR(repository *resources.Repository, artifactName string) (string, error) {
	if address, err := self.GetRepositoryHost(repository); err == nil {
		return fmt.Sprintf("docker://%s/%s?format=csar", address, artifactName), nil
	} else {
		return "", err
	}
}

func (self *Client) GetRepositoryServiceTemplateURL(repository *resources.Repository, serviceTemplateName string) (string, error) {
	return self.GetRepositoryURLForCSAR(repository, self.RepositoryArtifactNameForServiceTemplateName(serviceTemplateName))
}

// Utils

func (self *Client) RepositoryArtifactNameForServiceTemplateName(serviceTemplateName string) string {
	// Note: OpenShift registry permissions require the namespace as the tag category
	return fmt.Sprintf("%s/%s-%s", self.Namespace, serviceTemplateArtifactCategory, serviceTemplateName)
}

func (self *Client) ServiceTemplateNameForRepositoryArtifactName(artifactName string) (string, bool) {
	prefix := fmt.Sprintf("%s/%s-", self.Namespace, serviceTemplateArtifactCategory)
	if strings.HasPrefix(artifactName, prefix) {
		return artifactName[len(prefix):], true
	} else {
		return "", false
	}
}
