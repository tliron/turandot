package controller

import (
	"io"

	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/common/format"
	"github.com/tliron/turandot/common"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func (self *Controller) UpdateClout(clout *cloutpkg.Clout, cloutPath string, service *resources.Service) (*resources.Service, error) {
	if err := common.UpdateClout(clout, cloutPath); err == nil {
		if cloutHash, err := common.GetFileHash(cloutPath); err == nil {
			return self.UpdateServiceStatus(service, cloutPath, cloutHash)
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Controller) processClout(cloutPath string, service *resources.Service) error {
	artifactMappings := make(map[string]string)
	orchestrationPolicies := make(parser.OrchestrationPolicies)

	// Artifacts
	self.Log.Infof("processing artifacts for: %s", cloutPath)
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
		self.Log.Infof("updating artifacts for: %s", cloutPath)
		if clout, err := common.ReadClout(cloutPath); err == nil {
			self.UpdateCloutArtifacts(clout, artifactMappings)
			if _, err = self.UpdateClout(clout, cloutPath, service); err != nil {
				return err
			}
		}
	}

	// Policies
	self.Log.Infof("processing policies for: %s", cloutPath)
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
					if err := self.Substitute(service.Namespace, nodeTemplateName, policy.SubstitutionInputs, site); err != nil {
						return err
					}
				}
			}
		}
	}

	// Kubernetes resources
	// TODO: need to filter only non-substituted and instantiable node templates
	self.Log.Infof("processing Kubernetes resources for: %s", cloutPath)
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
