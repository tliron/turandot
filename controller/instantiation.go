package controller

import (
	"fmt"
	"io"

	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/common/format"
	"github.com/tliron/turandot/common"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	core "k8s.io/api/core/v1"
)

type Instantiation struct {
	cloutPath   string
	serviceName string
	namespace   string
}

func (self *Controller) enqueueInstantiation(cloutPath string, serviceName string, namespace string) {
	self.log.Infof("enqueuing instantiation for: %s", cloutPath)
	self.instantiationWork <- Instantiation{
		cloutPath,
		serviceName,
		namespace,
	}
}

func (self *Controller) stopInstantiator() {
	close(self.instantiationWork)
}

func (self *Controller) runInstantiator() {
	for {
		if instantiation, ok := <-self.instantiationWork; ok {
			self.log.Infof("processing instantiation for: %s", instantiation.cloutPath)
			if err := self.processInstantiation(instantiation.cloutPath, instantiation.serviceName, instantiation.namespace); err != nil {
				self.log.Errorf("%s", err.Error())
			}
		} else {
			self.log.Warning("no more instantiations")
			break
		}
	}
}

func (self *Controller) processInstantiation(cloutPath string, serviceName string, namespace string) error {
	if service, err := self.getService(serviceName, namespace); err == nil {
		if err := self.processClout(cloutPath, service); err == nil {
			self.events.Event(service, core.EventTypeNormal, "Instantiated", "Service instantiated successfully")
			return nil
		} else {
			self.events.Event(service, core.EventTypeWarning, "InstantiationError", fmt.Sprintf("Service instantiation error: %s", err.Error()))
			return err
		}
	} else {
		return err
	}
}

func (self *Controller) processClout(cloutPath string, service *resources.Service) error {
	artifactMappings := make(map[string]string)
	orchestrationPolicies := make(OrchestrationPolicies)

	// Artifacts
	self.log.Infof("processing artifacts for: %s", cloutPath)
	if clout, err := common.ReadClout(cloutPath); err == nil {
		if yaml, err := common.ExecScriptlet(clout, "kubernetes.artifacts"); err == nil {
			if artifacts, err := format.DecodeYAML(yaml); err == nil {
				if artifactMappings, err = self.processArtifacts(artifacts, service); err != nil {
					return err
				}
			} else if err != io.EOF {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}

	if len(artifactMappings) > 0 {
		self.log.Infof("updating artifacts for: %s", cloutPath)
		if clout, err := common.ReadClout(cloutPath); err == nil {
			self.updateCloutArtifacts(clout, artifactMappings)
			if _, err = self.updateClout(clout, cloutPath, service); err != nil {
				return err
			}
		}
	}

	// Policies
	self.log.Infof("processing policies for: %s", cloutPath)
	if clout, err := common.ReadClout(cloutPath); err == nil {
		if yaml, err := common.ExecScriptlet(clout, "orchestration.policies"); err == nil {
			if policies, err := format.DecodeYAML(yaml); err == nil {
				if orchestrationPolicies, err = self.processPolicies(policies); err != nil {
					return err
				}
			} else if err != io.EOF {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}

	// Substitution
	for nodeTemplateName, policies := range orchestrationPolicies {
		for _, policy := range policies {
			if policy.Substitutable {
				for _, site := range policy.Sites {
					if err := self.substitute(service.Namespace, nodeTemplateName, policy.SubstitutionInputs, site); err != nil {
						return err
					}
				}
			}
		}
	}

	// Kubernetes resources
	// TODO: need to filter only non-substituted and instantiable node templates
	self.log.Infof("processing Kubernetes resources for: %s", cloutPath)
	if clout, err := common.ReadClout(cloutPath); err == nil {
		if yaml, err := common.ExecScriptlet(clout, "kubernetes.generate"); err == nil {
			if objects, err := common.DecodeAllYAML(yaml); err == nil {
				if err := self.processResources(objects, service); err != nil {
					return err
				}
			} else if err != io.EOF {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}

	return nil
}

func (self *Controller) updateClout(clout *cloutpkg.Clout, cloutPath string, service *resources.Service) (*resources.Service, error) {
	if err := common.UpdateClout(clout, cloutPath); err == nil {
		if cloutHash, err := common.GetFileHash(cloutPath); err == nil {
			return self.updateServiceStatus(service, cloutPath, cloutHash)
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
