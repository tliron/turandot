package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/gofrs/flock"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (self *Controller) UpdateServiceInstantiationState(service *resources.Service, state resources.ServiceInstantiationState) (*resources.Service, error) {
	self.Log.Infof("updating instantiation status to %q for service: %s/%s", state, service.Namespace, service.Name)

	for {
		service = service.DeepCopy()
		service.Status.InstantiationState = state

		service_, err, retry := self.updateServiceStatus(service)
		if retry {
			service = service_
		} else {
			return service_, err
		}
	}
}

func (self *Controller) UpdateServiceStatusClout(service *resources.Service, cloutPath string, cloutHash string) (*resources.Service, error) {
	self.Log.Infof("updating Clout status for service: %s/%s", service.Namespace, service.Name)

	for {
		service = service.DeepCopy()
		service.Status.ServiceTemplateURL = service.Spec.ServiceTemplateURL
		if service.Spec.Inputs != nil {
			service.Status.Inputs = make(map[string]string)
			for key, input := range service.Spec.Inputs {
				service.Status.Inputs[key] = input
			}
		}
		service.Status.CloutPath = cloutPath
		service.Status.CloutHash = cloutHash

		service_, err, retry := self.updateServiceStatus(service)
		if retry {
			service = service_
		} else {
			return service_, err
		}
	}
}

func (self *Controller) UpdateServiceStatusMode(service *resources.Service) (*resources.Service, error) {
	self.Log.Infof("updating mode to %q for service: %s/%s", service.Spec.Mode, service.Namespace, service.Name)

	for {
		service = service.DeepCopy()
		service.Status.Mode = service.Spec.Mode
		if service.Status.NodeStates != nil {
			for nodeTemplateName, nodeState := range service.Status.NodeStates {
				if nodeState.Mode == service.Status.Mode {
					delete(service.Status.NodeStates, nodeTemplateName)
				}
			}
		}

		service_, err, retry := self.updateServiceStatus(service)
		if retry {
			service = service_
		} else {
			return service_, err
		}
	}
}

func (self *Controller) UpdateServiceStatusOutputs(service *resources.Service, outputs map[string]string) (*resources.Service, error) {
	self.Log.Infof("updating outputs for service: %s/%s", service.Namespace, service.Name)

	for {
		service = service.DeepCopy()
		service.Status.Outputs = outputs

		service_, err, retry := self.updateServiceStatus(service)
		if retry {
			service = service_
		} else {
			return service_, err
		}
	}
}

func (self *Controller) UpdateServiceStatusNodeStates(service *resources.Service, states parser.OrchestrationNodeStates) (*resources.Service, error) {
	self.Log.Infof("updating node states for service: %s/%s", service.Namespace, service.Name)

	for {
		service = service.DeepCopy()
		if service.Status.NodeStates == nil {
			service.Status.NodeStates = make(map[string]resources.ServiceNodeModeState)
		}
		for nodeTemplateName, nodeState := range states {
			service.Status.NodeStates[nodeTemplateName] = resources.ServiceNodeModeState{
				Mode:    nodeState.Mode,
				State:   resources.ModeState(nodeState.State),
				Message: nodeState.Message,
			}
		}

		service_, err, retry := self.updateServiceStatus(service)
		if retry {
			service = service_
		} else {
			return service_, err
		}
	}
}

func (self *Controller) updateServiceStatus(service *resources.Service) (*resources.Service, error, bool) {
	if service_, err := self.Client.UpdateServiceStatus(service); err == nil {
		return service_, nil, false
	} else if errors.IsConflict(err) {
		self.Log.Warningf("retrying status update for service: %s/%s", service.Namespace, service.Name)
		if service_, err := self.Client.GetService(service.Namespace, service.Name); err == nil {
			return service_, nil, true
		} else {
			return service, err, false
		}
	} else {
		return service, err, false
	}
}

func (self *Controller) processService(service *resources.Service) (bool, error) {
	// The "Instantiating" status is used as a processing lock
	if service.Status.InstantiationState == resources.ServiceInstantiating {
		return true, nil
	}

	if instantiated, err := self.isServiceInstanceOfCurrentClout(service); err == nil {
		if instantiated {
			if err := self.updateService(service); err == nil {
				return true, nil
			} else {
				return false, err
			}
		} else {
			return self.instantiateService(service)
		}
	} else {
		return false, err
	}
}

func (self *Controller) instantiateService(service *resources.Service) (bool, error) {
	self.Log.Infof("instantiating service: %s/%s", service.Namespace, service.Name)

	// The "Instantiating" status is used as a processing lock
	var err error
	if service, err = self.UpdateServiceInstantiationState(service, resources.ServiceInstantiating); err != nil {
		return false, err
	}

	cloutPath := service.Status.CloutPath
	if cloutPath == "" {
		cloutPath = fmt.Sprintf("%s-%s-%s.yaml", service.Namespace, service.Name, service.UID)
		cloutPath = util.SanitizeFilename(cloutPath)
		cloutPath = filepath.Join(self.CachePath, "clout", cloutPath)
	}

	urlContext := urlpkg.NewContext()
	defer urlContext.Release()

	// Compile
	var cloutHash string
	if cloutHash, err = self.CompileServiceTemplate(service.Spec.ServiceTemplateURL, service.Spec.Inputs, cloutPath, urlContext); err == nil {
		lock := flock.New(cloutPath)
		if err := lock.Lock(); err == nil {
			defer lock.Unlock()
		} else {
			return false, err
		}

		self.EventCompiled(service)
		if service, err = self.UpdateServiceStatusClout(service, cloutPath, cloutHash); err != nil {
			_, err := self.UpdateServiceInstantiationState(service, resources.ServiceNotInstantiated)
			return true, err
		}
	} else {
		self.EventCompilationError(service, err)
		_, err := self.UpdateServiceInstantiationState(service, resources.ServiceNotInstantiated)
		// Note that we return true to avoid unnecessary recompilation of the same service template
		return true, err
	}

	// Instantiate
	if service, err = self.instantiateClout(service, urlContext); err == nil {
		self.EventInstantiated(service)
		_, err := self.UpdateServiceInstantiationState(service, resources.ServiceInstantiated)
		return true, err
	} else {
		self.EventInstantiationError(service, err)
		_, err := self.UpdateServiceInstantiationState(service, resources.ServiceNotInstantiated)
		return false, err
	}
}

func (self *Controller) updateService(service *resources.Service) error {
	lock := flock.New(service.Status.CloutPath)
	if err := lock.Lock(); err == nil {
		defer lock.Unlock()
	} else {
		return err
	}

	self.Log.Infof("updating service: %s/%s", service.Namespace, service.Name)

	urlContext := urlpkg.NewContext()
	defer urlContext.Release()

	if _, err := self.updateClout(service, urlContext); err == nil {
		return nil
	} else {
		self.EventCloutUpdateError(service, err)
		return err
	}
}

func (self *Controller) isServiceInstanceOfCurrentClout(service *resources.Service) (bool, error) {
	if service.Status.InstantiationState != resources.ServiceInstantiated {
		return false, nil
	} else if service.Status.CloutPath == "" {
		self.Log.Infof("no Clout for service %s/%s", service.Namespace, service.Name)
		return false, nil
	} else if service.Spec.ServiceTemplateURL != service.Status.ServiceTemplateURL {
		self.Log.Infof("service template URL has changed for service %s/%s: from %q to %q", service.Namespace, service.Name, service.Status.ServiceTemplateURL, service.Spec.ServiceTemplateURL)
		return false, nil
	} else if !reflect.DeepEqual(service.Spec.Inputs, service.Status.Inputs) {
		self.Log.Infof("inputs have changed for service %s/%s", service.Namespace, service.Name)
		return false, nil
	} else if service.Status.CloutHash == "" {
		self.Log.Infof("no Clout hash for service %s/%s", service.Namespace, service.Name)
		return false, nil
	} else {
		// Get Clout hash
		if _, err := os.Stat(service.Status.CloutPath); os.IsNotExist(err) {
			self.Log.Infof("Clout disappeared for service %s/%s: %s", service.Namespace, service.Name, service.Status.CloutPath)
			return false, nil
		} else {
			if cloutHash, err := util.GetFileHash(service.Status.CloutPath); err == nil {
				if cloutHash == service.Status.CloutHash {
					self.Log.Infof("Clout has not changed for service %s/%s: %s", service.Namespace, service.Name, service.Status.CloutPath)
					return true, nil
				} else {
					self.Log.Infof("Clout has changed for service %s/%s: %s", service.Namespace, service.Name, service.Status.CloutPath)
					return false, nil
				}
			} else {
				return true, err
			}
		}
	}
}
