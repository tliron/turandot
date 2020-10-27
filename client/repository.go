package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"

	"github.com/tliron/kutil/util"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const serviceTemplateCategory = "service-templates"

var serviceTemplateImageNamePrefix = fmt.Sprintf("%s/", serviceTemplateCategory)

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

func (self *Client) CreateRepositoryDirect(namespace string, repositoryName string, url string, secretName string) (*resources.Repository, error) {
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
			Direct: resources.RepositoryDirect{
				URL: url,
			},
			Secret: secretName,
		},
	}

	return self.createRepository(namespace, repositoryName, repository)
}

func (self *Client) CreateRepositoryIndirect(namespace string, repositoryName string, serviceNamespace string, serviceName string, port uint64, secretName string) (*resources.Repository, error) {
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
			Indirect: resources.RepositoryIndirect{
				Namespace: serviceNamespace,
				Service:   serviceName,
				Port:      port,
			},
			Secret: secretName,
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

func (self *Client) GetRepositoryURL(repository *resources.Repository) (string, error) {
	if repository.Spec.Direct.URL != "" {
		return repository.Spec.Direct.URL, nil
	} else if repository.Spec.Indirect.Service != "" {
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

func (self *Client) GetRepositoryCertificate(repository *resources.Repository) (*x509.Certificate, error) {
	if repository.Spec.Secret != "" {
		return self.GetSecretCertificate(repository.Namespace, repository.Spec.Secret)
	} else {
		return nil, nil
	}
}

func (self *Client) GetRepositoryHTTPRoundTripper(repository *resources.Repository) (http.RoundTripper, error) {
	if certificate, err := self.GetRepositoryCertificate(repository); err == nil {
		if certificate != nil {
			certPool := x509.NewCertPool()
			certPool.AddCert(certificate)
			return util.NewForceHTTPSRoundTripper(&http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: certPool,
				},
			}), nil
		} else {
			return nil, nil
		}
	} else {
		return nil, err
	}
}

func (self *Client) GetRepositoryURLForCSAR(repository *resources.Repository, imageName string) (string, error) {
	if url, err := self.GetRepositoryURL(repository); err == nil {
		return fmt.Sprintf("docker://%s/%s?format=csar", url, imageName), nil
	} else {
		return "", err
	}
}

func (self *Client) GetRepositoryServiceTemplateURL(repository *resources.Repository, serviceTemplateName string) (string, error) {
	return self.GetRepositoryURLForCSAR(repository, RepositoryImageNameForServiceTemplateName(serviceTemplateName))
}

// Utils

func RepositoryImageNameForServiceTemplateName(serviceTemplateName string) string {
	return fmt.Sprintf("%s%s", serviceTemplateImageNamePrefix, serviceTemplateName)
}

func ServiceTemplateNameForRepositoryImageName(imageName string) (string, bool) {
	if strings.HasPrefix(imageName, serviceTemplateImageNamePrefix) {
		return imageName[len(serviceTemplateImageNamePrefix):], true
	} else {
		return "", false
	}
}
