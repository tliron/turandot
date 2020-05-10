package delegate

import (
	"fmt"
	neturlpkg "net/url"
	"strings"

	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/common"
)

const serviceTemplateCategory = "service-templates"

var serviceTemplateImageNamePrefix = fmt.Sprintf("%s/", serviceTemplateCategory)

func GetInventoryImageName(serviceTemplateName string) string {
	return fmt.Sprintf("%s/%s", serviceTemplateCategory, serviceTemplateName)
}

func ServiceTemplateNameFromInventoryImageName(imageName string) (string, bool) {
	if strings.HasPrefix(imageName, serviceTemplateImageNamePrefix) {
		return imageName[len(serviceTemplateImageNamePrefix):], true
	} else {
		return "", false
	}
}

func (self *Client) GetInventoryServiceTemplateURL(serviceTemplateName string, urlContext *urlpkg.Context) (*urlpkg.DockerURL, error) {
	appName := fmt.Sprintf("%s-inventory", self.NamePrefix)
	if ip, err := common.GetFirstPodIP(self.Context, self.Kubernetes, self.Namespace, appName); err == nil {
		imageName := GetInventoryImageName(serviceTemplateName)
		url := fmt.Sprintf("docker://%s:5000/%s?format=csar", ip, imageName)
		if url_, err := neturlpkg.ParseRequestURI(url); err == nil {
			return urlpkg.NewDockerURL(url_, urlContext), nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
