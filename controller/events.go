package controller

import (
	"fmt"

	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	core "k8s.io/api/core/v1"
)

const (
	EventCompiled           = "Compiled"
	EventCompilationError   = "CompilationError"
	EventInstantiated       = "Instantiated"
	EventInstantiationError = "InstantiationError"
	EventCloutUpdateError   = "CloutUpdateError"
)

func (self *Controller) EventCompiled(service *resources.Service) {
	self.Events.Event(service, core.EventTypeNormal, EventCompiled, "Service template compiled successfully")
}

func (self *Controller) EventCompilationError(service *resources.Service, err error) {
	self.Events.Event(service, core.EventTypeWarning, EventCompilationError, fmt.Sprintf("Service template compilation error: %s", err.Error()))
}

func (self *Controller) EventInstantiated(service *resources.Service) {
	self.Events.Event(service, core.EventTypeNormal, EventInstantiated, "Service instantiated successfully")
}

func (self *Controller) EventInstantiationError(service *resources.Service, err error) {
	self.Events.Event(service, core.EventTypeWarning, EventInstantiationError, fmt.Sprintf("Service instantiation error: %s", err.Error()))
}

func (self *Controller) EventCloutUpdateError(service *resources.Service, err error) {
	self.Events.Event(service, core.EventTypeWarning, EventCloutUpdateError, fmt.Sprintf("Service Clout update error: %s", err.Error()))
}
