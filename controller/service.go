package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gofrs/flock"
	puccinicommon "github.com/tliron/puccini/common"
	"github.com/tliron/puccini/common/format"
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/common"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	errorspkg "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (self *Controller) GetService(name string, namespace string) (*resources.Service, error) {
	if service, err := self.Services.Services(namespace).Get(name); err == nil {
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

func (self *Controller) CreateService(namespace string, name string, url urlpkg.URL, inputs map[string]interface{}) (*resources.Service, error) {
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
			Name:      name,
			Namespace: namespace,
		},
		Spec: resources.ServiceSpec{
			ServiceTemplateURL: url.String(),
			Inputs:             inputs_,
		},
		Status: resources.ServiceStatus{
			InstantiationState: resources.ServiceNotInstantiated,
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

func (self *Controller) UpdateServiceInstantiationState(service *resources.Service, state resources.ServiceInstantiationState) (*resources.Service, error) {
	self.Log.Infof("updating instantiation status to %q for service: %s/%s", state, service.Namespace, service.Name)

	service.Status.InstantiationState = state

	return self.updateServiceStatus(service)
}

func (self *Controller) UpdateServiceStatusClout(service *resources.Service, cloutPath string, cloutHash string) (*resources.Service, error) {
	self.Log.Infof("updating Clout status for service: %s/%s", service.Namespace, service.Name)

	service.Status.ServiceTemplateURL = service.Spec.ServiceTemplateURL
	if service.Spec.Inputs != nil {
		service.Status.Inputs = make(map[string]string)
		for key, input := range service.Spec.Inputs {
			service.Status.Inputs[key] = input
		}
	}
	service.Status.CloutPath = cloutPath
	service.Status.CloutHash = cloutHash

	return self.updateServiceStatus(service)
}

func (self *Controller) UpdateServiceStatusOutputs(service *resources.Service, outputs map[string]string) (*resources.Service, error) {
	self.Log.Infof("updating outputs for service: %s/%s", service.Namespace, service.Name)

	service.Status.Outputs = outputs

	return self.updateServiceStatus(service)
}

func (self *Controller) UpdateServiceStatusNodeStates(service *resources.Service, states parser.OrchestrationNodeStates) (*resources.Service, error) {
	self.Log.Infof("updating node states for service: %s/%s", service.Namespace, service.Name)

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

	return self.updateServiceStatus(service)
}

func (self *Controller) updateServiceStatus(service *resources.Service) (*resources.Service, error) {
	service = service.DeepCopy()
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
		cloutPath = puccinicommon.SanitizeFilename(cloutPath)
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

	if _, err := self.composeClout(service, urlContext); err == nil {
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
			if cloutHash, err := common.GetFileHash(service.Status.CloutPath); err == nil {
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
