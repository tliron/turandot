package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/tliron/turandot/common"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (self *Controller) getService(name string, namespace string) (*resources.Service, error) {
	if service, err := self.services.Services(namespace).Get(name); err == nil {
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

func (self *Controller) processService(service *resources.Service) (bool, error) {
	cloutPath := filepath.Join(self.cachePath, "clout", fmt.Sprintf("clout-%s.yaml", service.UID))
	cloutHash := ""

	// Get clout hash
	if _, err := os.Stat(cloutPath); os.IsNotExist(err) {
		self.log.Infof("clout does not exist: %s", cloutPath)
	} else {
		if cloutHash, err = common.GetFileHash(cloutPath); err == nil {
			if cloutHash == service.Status.CloutHash {
				self.log.Infof("clout has not changed: %s", cloutPath)
			} else {
				self.log.Infof("clout has changed: %s", cloutPath)
			}
		} else {
			return false, err
		}
	}

	// Check if we need to compile or recompile
	if (cloutHash == "") || (service.Spec.ServiceTemplateURL != service.Status.ServiceTemplateURL) || (!reflect.DeepEqual(service.Spec.Inputs, service.Status.Inputs)) {
		var err error
		if cloutHash, err = self.compileServiceTemplate(service.Spec.ServiceTemplateURL, service.Spec.Inputs, cloutPath); err == nil {
			self.events.Event(service, core.EventTypeNormal, "Compiled", "Service template compiled successfully")
		} else {
			self.events.Event(service, core.EventTypeWarning, "CompilationError", fmt.Sprintf("Service template compilation error: %s", err.Error()))
			return false, err
		}
	}

	if (service.Status.CloutPath != cloutPath) || (service.Status.CloutHash != cloutHash) {
		if _, err := self.updateServiceStatus(service, cloutPath, cloutHash); err == nil {
			self.events.Event(service, core.EventTypeNormal, "Synced", "Service synced successfully")
		} else {
			// TODO: really return true?
			return true, err
		}
		self.enqueueInstantiation(cloutPath, service.Name, service.Namespace)
		return true, nil
	}

	return true, nil
}

func (self *Controller) updateServiceStatus(service *resources.Service, cloutPath string, cloutHash string) (*resources.Service, error) {
	self.log.Infof("updating status for service: \"%s %s\"", service.Namespace, service.Name)
	service = service.DeepCopy()
	service.Status.ServiceTemplateURL = service.Spec.ServiceTemplateURL
	service.Status.Inputs = make(map[string]string)
	// TODO: outputs
	service.Status.Outputs = make(map[string]string)
	for key, input := range service.Spec.Inputs {
		service.Status.Inputs[key] = input
	}
	service.Status.CloutPath = cloutPath
	service.Status.CloutHash = cloutHash
	// TODO: check: does update return an error if there was no change?
	return self.turandot.TurandotV1alpha1().Services(service.Namespace).UpdateStatus(self.context, service, meta.UpdateOptions{})
}
