package controller

import (
	"fmt"
	"path/filepath"

	urlpkg "github.com/tliron/puccini/url"
	core "k8s.io/api/core/v1"
)

type Instantiation struct {
	serviceName string
	namespace   string
}

func (self *Controller) EnqueueInstantiation(serviceName string, namespace string) {
	self.Log.Infof("enqueuing instantiation for: %s/%s", namespace, serviceName)
	self.InstantiationWork <- Instantiation{
		serviceName,
		namespace,
	}
}

func (self *Controller) StartInstantiator(concurrency uint, stopChannel <-chan struct{}) {
	var i uint
	for i = 0; i < concurrency; i++ {
		go self.runInstantiator(stopChannel)
	}
}

func (self *Controller) StopInstantiator() {
	close(self.InstantiationWork)
}

func (self *Controller) runInstantiator(stopChannel <-chan struct{}) {
	for {
		select {
		case <-stopChannel:
			self.Log.Warning("no more instantiations")
			return

		case instantiation := <-self.InstantiationWork:
			self.Log.Infof("processing instantiation for: %s/%s", instantiation.namespace, instantiation.serviceName)
			if err := self.processInstantiation(instantiation.serviceName, instantiation.namespace); err != nil {
				self.Log.Errorf("%s", err.Error())
			}
		}
	}
}

func (self *Controller) processInstantiation(serviceName string, namespace string) error {
	if service, err := self.GetService(serviceName, namespace); err == nil {
		if dirty, err := self.serviceDirty(service); err == nil {
			if !dirty {
				return nil
			}
		} else {
			return err
		}

		cloutPath := service.Status.CloutPath
		if cloutPath == "" {
			cloutPath = filepath.Join(self.CachePath, "clout", fmt.Sprintf("clout-%s.yaml", service.UID))
		}

		urlContext := urlpkg.NewContext()
		defer urlContext.Release()

		var cloutHash string
		if cloutHash, err = self.CompileServiceTemplate(service.Spec.ServiceTemplateURL, service.Spec.Inputs, cloutPath, urlContext); err == nil {
			self.Events.Event(service, core.EventTypeNormal, "Compiled", "Service template compiled successfully")
		} else {
			self.Events.Event(service, core.EventTypeWarning, "CompilationError", fmt.Sprintf("Service template compilation error: %s", err.Error()))
			return err
		}

		if service, err = self.UpdateServiceStatus(service, cloutPath, cloutHash); err != nil {
			return err
		}

		if err := self.processClout(service, urlContext); err == nil {
			self.Events.Event(service, core.EventTypeNormal, "Instantiated", "Service instantiated successfully")
			return nil
		} else {
			self.Events.Event(service, core.EventTypeWarning, "InstantiationError", fmt.Sprintf("Service instantiation error: %s", err.Error()))
			return err
		}
	} else {
		return err
	}
}
