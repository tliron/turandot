package controller

import (
	"fmt"
	"io"

	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/common/format"
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/common"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func (self *Controller) ReadClout(cloutPath string, urlContext *urlpkg.Context) (*cloutpkg.Clout, error) {
	if url, err := urlpkg.NewURL(cloutPath, urlContext); err == nil {
		if reader, err := url.Open(); err == nil {
			defer reader.Close()
			return common.ReadClout(reader, urlContext)
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Controller) WriteClout(clout *cloutpkg.Clout, cloutPath string) (string, error) {
	if file, err := format.OpenFileForWrite(cloutPath); err == nil {
		defer file.Close()
		if err := common.WriteClout(clout, file); err == nil {
			return common.GetFileHash(cloutPath)
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}

func (self *Controller) UpdateClout(clout *cloutpkg.Clout, service *resources.Service) (*resources.Service, error) {
	if cloutHash, err := self.WriteClout(clout, service.Status.CloutPath); err == nil {
		return self.UpdateServiceStatus(service, service.Status.CloutPath, cloutHash)
	} else {
		return nil, err
	}
}

func (self *Controller) processClout(service *resources.Service, urlContext *urlpkg.Context) error {
	// TODO: update attributes in clout

	// Artifacts
	artifactMappings := make(map[string]string)
	self.Log.Infof("processing artifacts for: %s", service.Status.CloutPath)
	if clout, err := self.ReadClout(service.Status.CloutPath, urlContext); err == nil {
		if yaml, err := common.ExecScriptlet(clout, "kubernetes.artifacts", urlContext); err == nil {
			if artifacts, err := format.DecodeYAML(yaml); err == nil {
				if artifactMappings, err = self.processArtifacts(artifacts, service, urlContext); err != nil {
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
		self.Log.Infof("updating artifacts for: %s", service.Status.CloutPath)
		if clout, err := self.ReadClout(service.Status.CloutPath, urlContext); err == nil {
			self.UpdateCloutArtifacts(clout, artifactMappings)
			if service, err = self.UpdateClout(clout, service); err != nil {
				return err
			}
		}
	}

	// Policies
	orchestrationPolicies := make(parser.OrchestrationPolicies)
	self.Log.Infof("processing policies for: %s", service.Status.CloutPath)
	if clout, err := self.ReadClout(service.Status.CloutPath, urlContext); err == nil {
		if yaml, err := common.ExecScriptlet(clout, "orchestration.policies", urlContext); err == nil {
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
					if err := self.Substitute(service.Namespace, nodeTemplateName, policy.SubstitutionInputs, site, urlContext); err != nil {
						return err
					}
				}
			}
		}
	}

	// Kubernetes resources
	// TODO: need to filter only non-substituted and instantiable node templates
	self.Log.Infof("processing Kubernetes resources for: %s", service.Status.CloutPath)
	if clout, err := self.ReadClout(service.Status.CloutPath, urlContext); err == nil {
		if yaml, err := common.ExecScriptlet(clout, "kubernetes.resources", urlContext); err == nil {
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

	// Outputs
	self.Log.Infof("processing outputs for: %s", service.Status.CloutPath)
	if clout, err := self.ReadClout(service.Status.CloutPath, urlContext); err == nil {
		if yaml, err := common.ExecScriptlet(clout, "orchestration.outputs", urlContext); err == nil {
			if outputs, err := format.DecodeYAML(yaml); err == nil {
				if outputs_, ok := parser.ParseOutputs(outputs); ok {
					if service, err = self.UpdateServiceOutputs(service, outputs_); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("could not parse outputs", outputs)
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
