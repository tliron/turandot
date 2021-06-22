package client

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tliron/kutil/format"
	"github.com/tliron/kutil/kubernetes"
	urlpkg "github.com/tliron/kutil/url"
	reposure "github.com/tliron/reposure/resources/reposure.puccini.cloud/v1alpha1"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
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

func (self *Client) ListServicesForImageName(registryName string, imageName string, urlContext *urlpkg.Context) ([]string, error) {
	if services, err := self.ListServices(); err == nil {
		var serviceNames []string
		for _, service := range services.Items {
			if (service.Spec.ServiceTemplate.Indirect != nil) && (service.Spec.ServiceTemplate.Indirect.Registry == registryName) && (service.Spec.ServiceTemplate.Indirect.Name == imageName) {
				serviceNames = append(serviceNames, service.Name)
			}
			// TODO: direct
		}
		return serviceNames, nil
	} else {
		return nil, err
	}
}

func (self *Client) CreateServiceDirect(namespace string, serviceName string, url urlpkg.URL, tlsSecretName string, authSecretName string, inputs map[string]interface{}, mode string) (*resources.Service, error) {
	// Default to same namespace as operator
	if namespace == "" {
		namespace = self.Namespace
	}

	// Encode inputs
	var inputs_ map[string]string
	var err error
	if inputs_, err = encodeServiceInputs(inputs); err != nil {
		return nil, err
	}

	service := &resources.Service{
		ObjectMeta: meta.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
		},
		Spec: resources.ServiceSpec{
			ServiceTemplate: resources.ServiceTemplate{
				Direct: &resources.ServiceTemplateDirect{
					URL:        url.String(),
					TLSSecret:  tlsSecretName,
					AuthSecret: authSecretName,
				},
			},
			Inputs: inputs_,
			Mode:   mode,
		},
	}

	return self.createService(namespace, serviceName, service)
}

func (self *Client) CreateServiceIndirect(namespace string, serviceName string, registryName string, imageName string, inputs map[string]interface{}, mode string) (*resources.Service, error) {
	// Default to same namespace as operator
	if namespace == "" {
		namespace = self.Namespace
	}

	if colon := strings.Index(imageName, ":"); colon == -1 {
		// Must have a tag
		imageName += ":latest"
	}

	// Encode inputs
	var inputs_ map[string]string
	var err error
	if inputs_, err = encodeServiceInputs(inputs); err != nil {
		return nil, err
	}

	service := &resources.Service{
		ObjectMeta: meta.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
		},
		Spec: resources.ServiceSpec{
			ServiceTemplate: resources.ServiceTemplate{
				Indirect: &resources.ServiceTemplateIndirect{
					Registry: registryName,
					Name:     imageName,
				},
			},
			Inputs: inputs_,
			Mode:   mode,
		},
	}

	return self.createService(namespace, serviceName, service)
}

func (self *Client) CreateServiceFromURL(namespace string, serviceName string, url string, inputs map[string]interface{}, mode string, urlContext *urlpkg.Context) (*resources.Service, error) {
	if url_, err := urlpkg.NewURL(url, urlContext); err == nil {
		return self.CreateServiceDirect(namespace, serviceName, url_, "", "", inputs, mode)
	} else {
		return nil, err
	}
}

func (self *Client) CreateServiceFromTemplate(namespace string, serviceName string, registry *reposure.Registry, serviceTemplateName string, inputs map[string]interface{}, mode string) (*resources.Service, error) {
	imageName := self.RegistryImageNameForServiceTemplateName(serviceTemplateName)
	return self.CreateServiceIndirect(namespace, serviceName, registry.Name, imageName, inputs, mode)
}

func (self *Client) CreateServiceFromContent(namespace string, serviceName string, registry *reposure.Registry, url urlpkg.URL, inputs map[string]interface{}, mode string) (*resources.Service, error) {
	spooler := self.Reposure.SurrogateSpoolerClient(registry)
	serviceTemplateName := fmt.Sprintf("%s-%s", serviceName, uuid.New().String())
	imageName := self.RegistryImageNameForServiceTemplateName(serviceTemplateName)
	if err := spooler.PushTarballFromURL(imageName, url); err == nil {
		return self.CreateServiceIndirect(namespace, serviceName, registry.Name, imageName, inputs, mode)
	} else {
		return nil, err
	}
}

func (self *Client) createService(namespace string, serviceName string, service *resources.Service) (*resources.Service, error) {
	if service, err := self.Turandot.TurandotV1alpha1().Services(namespace).Create(self.Context, service, meta.CreateOptions{}); err == nil {
		return service, nil
	} else if errors.IsAlreadyExists(err) {
		return self.Turandot.TurandotV1alpha1().Services(namespace).Get(self.Context, serviceName, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) GetServiceRegistry(service *resources.Service) (*reposure.Registry, error) {
	if (service.Spec.ServiceTemplate.Indirect != nil) && (service.Spec.ServiceTemplate.Indirect.Registry != "") {
		namespace := service.Spec.ServiceTemplate.Indirect.Namespace
		if namespace == "" {
			namespace = service.Namespace
		}
		return self.Reposure.RegistryClient().Get(namespace, service.Spec.ServiceTemplate.Indirect.Registry)
	} else {
		return nil, nil
	}
}

func (self *Client) GetServiceTemplateURL(service *resources.Service) (string, error) {
	if registry, err := self.GetServiceRegistry(service); err == nil {
		if registry != nil {
			return self.GetRegistryURLForCSAR(registry, service.Spec.ServiceTemplate.Indirect.Name)
		} else if (service.Spec.ServiceTemplate.Direct != nil) && (service.Spec.ServiceTemplate.Direct.URL != "") {
			return service.Spec.ServiceTemplate.Direct.URL, nil
		} else {
			return "", fmt.Errorf("malformed service: %s", service.Name)
		}
	} else {
		return "", err
	}
}

func (self *Client) UpdateServiceURLContext(service *resources.Service, urlContext *urlpkg.Context) error {
	if registry, err := self.GetServiceRegistry(service); err == nil {
		if registry != nil {
			return self.Reposure.RegistryClient().UpdateURLContext(registry, urlContext)
		} else {
			return nil
		}
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

func (self *Client) GetServiceClout(namespace string, serviceName string) (string, error) {
	if service, err := self.GetService(namespace, serviceName); err == nil {
		appName := fmt.Sprintf("%s-operator", self.NamePrefix)

		if podName, err := kubernetes.GetFirstPodName(self.Context, self.Kubernetes, self.Namespace, appName); err == nil {
			var builder strings.Builder
			if err := self.Exec(self.Namespace, podName, "operator", nil, &builder, "cat", service.Status.CloutPath); err == nil {
				return strings.TrimRight(builder.String(), "\n"), nil
			} else {
				return "", err
			}
		} else {
			return "", err
		}
	} else {
		return "", nil
	}
}

// Utils

func encodeServiceInputs(inputs map[string]interface{}) (map[string]string, error) {
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
	return inputs_, nil
}
