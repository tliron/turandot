package client

import (
	"strings"

	"github.com/google/uuid"
	spoolerpkg "github.com/tliron/kubernetes-registry-spooler/client"
	"github.com/tliron/kutil/format"
	urlpkg "github.com/tliron/kutil/url"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	"github.com/tliron/turandot/tools"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (self *Client) GetService(namespace string, serviceName string) (*resources.Service, error) {
	// Default to same namespace as operator
	if namespace == "" {
		namespace = self.Namespace
	}

	if service, err := self.Turandot.TurandotV1alpha1().Services(namespace).Get(self.Context, serviceName, meta.GetOptions{}); err == nil {
		// When retrieved from cache the GVK may be empty
		if service.Kind == "" {
			service = service.DeepCopy()
			service.APIVersion, service.Kind = resources.ServiceGVK.ToAPIVersionAndKind()
		}
		return service, nil
	} else {
		return nil, err
	}
}

func (self *Client) ListServices() (*resources.ServiceList, error) {
	// TODO: all services in cluster mode
	return self.Turandot.TurandotV1alpha1().Services(self.Namespace).List(self.Context, meta.ListOptions{})
}

func (self *Client) ListServicesForImage(imageName string, urlContext *urlpkg.Context) ([]string, error) {
	if serviceTemplateUrl, err := self.GetInventoryURLForCSAR(self.Namespace, imageName, urlContext); err == nil {
		serviceTemplateUrl_ := serviceTemplateUrl.String()
		if services, err := self.ListServices(); err == nil {
			var serviceNames []string
			for _, service := range services.Items {
				if service.Spec.ServiceTemplateURL == serviceTemplateUrl_ {
					serviceNames = append(serviceNames, service.Name)
				}
			}
			return serviceNames, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Client) CreateService(namespace string, serviceName string, url urlpkg.URL, inputs map[string]interface{}, mode string) (*resources.Service, error) {
	// Default to same namespace as operator
	if namespace == "" {
		namespace = self.Namespace
	}

	// Encode inputs
	var inputs_ map[string]string
	if (inputs != nil) && len(inputs) > 0 {
		inputs_ = make(map[string]string)
		for key, input := range inputs {
			var err error
			if inputs_[key], err = format.EncodeYAML(input, " ", false); err == nil {
				inputs_[key] = strings.TrimRight(inputs_[key], "\n")
			} else {
				return nil, err
			}
		}
	}

	service := &resources.Service{
		ObjectMeta: meta.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
		},
		Spec: resources.ServiceSpec{
			ServiceTemplateURL: url.String(),
			Inputs:             inputs_,
			Mode:               mode,
		},
	}

	if service, err := self.Turandot.TurandotV1alpha1().Services(namespace).Create(self.Context, service, meta.CreateOptions{}); err == nil {
		return service, nil
	} else if errors.IsAlreadyExists(err) {
		return self.Turandot.TurandotV1alpha1().Services(namespace).Get(self.Context, serviceName, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) CreateServiceFromTemplate(namespace string, serviceName string, serviceTemplateName string, inputs map[string]interface{}, mode string, urlContext *urlpkg.Context) error {
	if url, err := self.GetInventoryServiceTemplateURL(namespace, serviceTemplateName, urlContext); err == nil {
		_, err := self.CreateService(namespace, serviceName, url, inputs, mode)
		return err
	} else {
		return err
	}
}

func (self *Client) CreateServiceFromURL(namespace string, serviceName string, url string, inputs map[string]interface{}, mode string, urlContext *urlpkg.Context) error {
	if url_, err := urlpkg.NewURL(url, urlContext); err == nil {
		_, err = self.CreateService(namespace, serviceName, url_, inputs, mode)
		return err
	} else {
		return err
	}
}

func (self *Client) CreateServiceFromContent(namespace string, serviceName string, spooler *spoolerpkg.Client, url urlpkg.URL, inputs map[string]interface{}, mode string, urlContext *urlpkg.Context) error {
	serviceTemplateName := uuid.New().String()
	imageName := InventoryImageNameForServiceTemplateName(serviceTemplateName)
	if err := tools.PublishOnRegistry(imageName, url, spooler); err == nil {
		return self.CreateServiceFromTemplate(namespace, serviceName, serviceTemplateName, inputs, mode, urlContext)
	} else {
		return err
	}
}

func (self *Client) UpdateServiceSpec(service *resources.Service) (*resources.Service, error) {
	if service_, err := self.Turandot.TurandotV1alpha1().Services(service.Namespace).Update(self.Context, service, meta.UpdateOptions{}); err == nil {
		// When retrieved from cache the GVK may be empty
		if service_.Kind == "" {
			service_ = service_.DeepCopy()
			service_.APIVersion, service_.Kind = resources.ServiceGVK.ToAPIVersionAndKind()
		}
		return service_, nil
	} else {
		return service, err
	}
}

func (self *Client) UpdateServiceStatus(service *resources.Service) (*resources.Service, error) {
	if service_, err := self.Turandot.TurandotV1alpha1().Services(service.Namespace).UpdateStatus(self.Context, service, meta.UpdateOptions{}); err == nil {
		// When retrieved from cache the GVK may be empty
		if service_.Kind == "" {
			service_ = service_.DeepCopy()
			service_.APIVersion, service_.Kind = resources.ServiceGVK.ToAPIVersionAndKind()
		}
		return service_, nil
	} else {
		return service, err
	}
}

func (self *Client) UpdateServiceMode(service *resources.Service, mode string) (*resources.Service, error) {
	if service.Spec.Mode != mode {
		service = service.DeepCopy()
		service.Spec.Mode = mode
		return self.UpdateServiceSpec(service)
	} else {
		return service, nil
	}
}

func (self *Client) DeleteService(namespace string, serviceName string) error {
	// Default to same namespace as operator
	if namespace == "" {
		namespace = self.Namespace
	}

	return self.Turandot.TurandotV1alpha1().Services(namespace).Delete(self.Context, serviceName, meta.DeleteOptions{})
}
