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

func (self *Client) CreateRepository(namespace string, repositoryName string, url string, serviceName string, secretName string) (*resources.Repository, error) {
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
			URL:     url,
			Service: serviceName,
			Secret:  secretName,
		},
	}

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

func (self *Client) GetRepositoryCertificate(repository *resources.Repository) (*x509.Certificate, error) {
	return self.GetSecretCertificate(repository.Namespace, repository.Spec.Secret)
}

func (self *Client) GetRepositoryHTTPRoundTripper(repository *resources.Repository) (http.RoundTripper, error) {
	if certificate, err := self.GetRepositoryCertificate(repository); err == nil {
		certPool := x509.NewCertPool()
		certPool.AddCert(certificate)
		return util.NewForceHTTPSRoundTripper(&http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		}), nil
	} else {
		return nil, err
	}
}

// Utils

func GetRepositoryURLForCSAR(repository *resources.Repository, imageName string) string {
	return fmt.Sprintf("docker://%s/%s?format=csar", repository.Spec.URL, imageName)
}

func GetRepositoryServiceTemplateURL(repository *resources.Repository, serviceTemplateName string) string {
	return GetRepositoryURLForCSAR(repository, RepositoryImageNameForServiceTemplateName(serviceTemplateName))
}

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
