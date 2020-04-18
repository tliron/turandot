package controller

import (
	"strings"

	"github.com/tliron/puccini/common/format"
	urlpkg "github.com/tliron/puccini/url"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	errorspkg "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (self *Controller) substitute(namespace string, nodeTemplateName string, inputs map[string]interface{}, site string) error {
	// hacky ;)
	var serviceTemplateName string
	switch nodeTemplateName {
	case "central-pbx":
		serviceTemplateName = "asterisk-vnf"
	case "edge-pbx":
		serviceTemplateName = "asterisk-cnf"
	case "data-plane":
		serviceTemplateName = "simple-data-plane"
	}
	serviceName := serviceTemplateName

	if url, err := self.GetInventoryServiceTemplateURL(namespace, serviceTemplateName); err == nil {
		defer url.Release()

		if (site == "") || (site == self.site) {
			// Local
			if _, err := self.createService(namespace, serviceName, url, inputs); err != nil {
				return err
			}
		} else {
			// Delegate
			if client, spooler, err := self.NewDelegate(site, "/cache"); err == nil {
				if err := client.Install(site, "docker.io", true); err != nil {
					return err
				}

				if err := client.DeployServiceFromContent(serviceName, spooler, url, inputs); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	} else {
		return err
	}

	return nil
}

func (self *Controller) createService(namespace string, name string, url urlpkg.URL, inputs map[string]interface{}) (*resources.Service, error) {
	// Encode inputs
	inputs_ := make(map[string]string)
	for key, input := range inputs {
		var err error
		if inputs_[key], err = format.EncodeYAML(input, " ", false); err == nil {
			inputs_[key] = strings.TrimRight(inputs_[key], "\n")
		} else {
			return nil, err
		}
	}

	service := &resources.Service{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: resources.ServiceSpec{
			ServiceTemplateURL: url.String(),
			Inputs:             inputs_,
		},
	}

	if service, err := self.turandot.TurandotV1alpha1().Services(namespace).Create(self.context, service, meta.CreateOptions{}); err == nil {
		return service, nil
	} else if errorspkg.IsAlreadyExists(err) {
		return self.turandot.TurandotV1alpha1().Services(namespace).Get(self.context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}
