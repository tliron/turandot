package controller

import (
	"os"
	"reflect"
	"strings"

	"github.com/tliron/puccini/common/format"
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/common"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	errorspkg "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (self *Controller) GetService(name string, namespace string) (*resources.Service, error) {
	if service, err := self.Services.Services(namespace).Get(name); err == nil {
		// BUG: when retrieved from cache the gvk may be empty
		if service.Kind == "" {
			service = service.DeepCopy()
			service.APIVersion, service.Kind = resources.ServiceGVK.ToAPIVersionAndKind()
		}
		return service, nil
	} else {
		return nil, err
	}
}

func (self *Controller) CreateService(namespace string, name string, url urlpkg.URL, inputs map[string]interface{}) (*resources.Service, error) {
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

	if service, err := self.Turandot.TurandotV1alpha1().Services(namespace).Create(self.Context, service, meta.CreateOptions{}); err == nil {
		return service, nil
	} else if errorspkg.IsAlreadyExists(err) {
		return self.Turandot.TurandotV1alpha1().Services(namespace).Get(self.Context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Controller) UpdateServiceStatus(service *resources.Service, cloutPath string, cloutHash string) (*resources.Service, error) {
	self.Log.Infof("updating status for service: \"%s %s\"", service.Namespace, service.Name)

	service = service.DeepCopy()
	service.Status.ServiceTemplateURL = service.Spec.ServiceTemplateURL
	service.Status.Inputs = make(map[string]string)
	for key, input := range service.Spec.Inputs {
		service.Status.Inputs[key] = input
	}
	service.Status.Outputs = make(map[string]string)
	// TODO: outputs
	service.Status.CloutPath = cloutPath
	service.Status.CloutHash = cloutHash
	// TODO: check: does update return an error if there was no change?

	if service, err := self.Turandot.TurandotV1alpha1().Services(service.Namespace).UpdateStatus(self.Context, service, meta.UpdateOptions{}); err == nil {
		service.APIVersion, service.Kind = resources.ServiceGVK.ToAPIVersionAndKind()
		return service, nil
	} else {
		return nil, err
	}
}

func (self *Controller) processService(service *resources.Service) (bool, error) {
	/*
		cloutPath := filepath.Join(self.CachePath, "clout", fmt.Sprintf("clout-%s.yaml", service.UID))
		cloutHash := ""

		// Get clout hash
		if _, err := os.Stat(cloutPath); os.IsNotExist(err) {
			self.Log.Infof("clout does not exist: %s", cloutPath)
		} else {
			if cloutHash, err = common.GetFileHash(cloutPath); err == nil {
				if cloutHash == service.Status.CloutHash {
					self.Log.Infof("clout has not changed: %s", cloutPath)
				} else {
					self.Log.Infof("clout has changed: %s", cloutPath)
				}
			} else {
				return false, err
			}
		}

		// Check if we need to compile or recompile
		if (cloutHash == "") || (service.Spec.ServiceTemplateURL != service.Status.ServiceTemplateURL) || (!reflect.DeepEqual(service.Spec.Inputs, service.Status.Inputs)) {
			var err error
			if cloutHash, err = self.CompileServiceTemplate(service.Spec.ServiceTemplateURL, service.Spec.Inputs, cloutPath); err == nil {
				self.Events.Event(service, core.EventTypeNormal, "Compiled", "Service template compiled successfully")
			} else {
				self.Events.Event(service, core.EventTypeWarning, "CompilationError", fmt.Sprintf("Service template compilation error: %s", err.Error()))
				return false, err
			}
		}

		if (service.Status.CloutPath != cloutPath) || (service.Status.CloutHash != cloutHash) {
			if _, err := self.UpdateServiceStatus(service, cloutPath, cloutHash); err == nil {
				self.Events.Event(service, core.EventTypeNormal, "Synced", "Service synced successfully")
			} else {
				// TODO: really return true?
				return true, err
			}
			self.EnqueueInstantiation(cloutPath, service.Name, service.Namespace)
			return true, nil
		}*/

	if dirty, err := self.serviceDirty(service); err == nil {
		if dirty {
			self.EnqueueInstantiation(service.Name, service.Namespace)
		}
	} else {
		return false, err
	}

	return true, nil
}

func (self *Controller) serviceDirty(service *resources.Service) (bool, error) {
	if (service.Spec.ServiceTemplateURL != service.Status.ServiceTemplateURL) || (!reflect.DeepEqual(service.Spec.Inputs, service.Status.Inputs)) {
		return true, nil
	} else if service.Status.CloutPath == "" {
		return true, nil
	} else {
		// Get clout hash
		if _, err := os.Stat(service.Status.CloutPath); os.IsNotExist(err) {
			self.Log.Infof("clout does not exist: %s", service.Status.CloutPath)
			return true, nil
		} else {
			if cloutHash, err := common.GetFileHash(service.Status.CloutPath); err == nil {
				if cloutHash == service.Status.CloutHash {
					self.Log.Infof("clout has not changed: %s", service.Status.CloutPath)
					return false, nil
				} else {
					self.Log.Infof("clout has changed: %s", service.Status.CloutPath)
					return true, nil
				}
			} else {
				return false, err
			}
		}
	}
}
