package client

import (
	"fmt"
	"strings"

	reposure "github.com/tliron/reposure/resources/reposure.puccini.cloud/v1alpha1"
)

const serviceTemplateArtifactCategory = "service-templates"

func (self *Client) GetRegistryURLForCSAR(registry *reposure.Registry, artifactName string) (string, error) {
	if address, err := self.Reposure.RegistryClient().GetHost(registry); err == nil {
		return fmt.Sprintf("docker://%s/%s?format=csar", address, artifactName), nil
	} else {
		return "", err
	}
}

func (self *Client) GetRegistryServiceTemplateURL(registry *reposure.Registry, serviceTemplateName string) (string, error) {
	return self.GetRegistryURLForCSAR(registry, self.RegistryImageNameForServiceTemplateName(serviceTemplateName))
}

// Utils

func (self *Client) RegistryImageNameForServiceTemplateName(serviceTemplateName string) string {
	// Note: OpenShift registry permissions require the namespace as the tag category
	return fmt.Sprintf("%s/%s-%s", self.Namespace, serviceTemplateArtifactCategory, serviceTemplateName)
}

func (self *Client) ServiceTemplateNameForRegistryImageName(artifactName string) (string, bool) {
	prefix := fmt.Sprintf("%s/%s-", self.Namespace, serviceTemplateArtifactCategory)
	if strings.HasPrefix(artifactName, prefix) {
		return artifactName[len(prefix):], true
	} else {
		return "", false
	}
}
