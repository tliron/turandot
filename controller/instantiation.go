package controller

import (
	"fmt"

	core "k8s.io/api/core/v1"
)

type Instantiation struct {
	cloutPath   string
	serviceName string
	namespace   string
}

func (self *Controller) EnqueueInstantiation(cloutPath string, serviceName string, namespace string) {
	self.Log.Infof("enqueuing instantiation for: %s", cloutPath)
	self.InstantiationWork <- Instantiation{
		cloutPath,
		serviceName,
		namespace,
	}
}

func (self *Controller) StopInstantiator() {
	close(self.InstantiationWork)
}

func (self *Controller) RunInstantiator() {
	for {
		if instantiation, ok := <-self.InstantiationWork; ok {
			self.Log.Infof("processing instantiation for: %s", instantiation.cloutPath)
			if err := self.processInstantiation(instantiation.cloutPath, instantiation.serviceName, instantiation.namespace); err != nil {
				self.Log.Errorf("%s", err.Error())
			}
		} else {
			self.Log.Warning("no more instantiations")
			break
		}
	}
}

func (self *Controller) processInstantiation(cloutPath string, serviceName string, namespace string) error {
	if service, err := self.GetService(serviceName, namespace); err == nil {
		if err := self.processClout(cloutPath, service); err == nil {
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
