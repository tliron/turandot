package controller

import (
	"fmt"
	"io"

	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/common/format"
	problemspkg "github.com/tliron/puccini/common/problems"
	"github.com/tliron/puccini/tosca/compiler"
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/common"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func (self *Controller) ReadClout(cloutPath string, resolve bool, coerce bool, urlContext *urlpkg.Context) (*cloutpkg.Clout, error) {
	if url, err := urlpkg.NewURL(cloutPath, urlContext); err == nil {
		if reader, err := url.Open(); err == nil {
			defer reader.Close()
			if clout, err := common.ReadClout(reader, urlContext); err == nil {
				problems := &problemspkg.Problems{}

				if resolve {
					if compiler.Resolve(clout, problems, urlContext, "yaml", false, true, true); !problems.Empty() {
						return nil, fmt.Errorf("could not resolve Clout\n%s", problems.ToString(true))
					}
				}

				if coerce {
					if compiler.Coerce(clout, problems, urlContext, "yaml", false, true, true); !problems.Empty() {
						return nil, fmt.Errorf("could not coerce Clout\n%s", problems.ToString(true))
					}
				}

				return clout, nil
			} else {
				return nil, err
			}
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
		return self.UpdateServiceClout(service, service.Status.CloutPath, cloutHash)
	} else {
		return nil, err
	}
}

func (self *Controller) instantiateClout(service *resources.Service, urlContext *urlpkg.Context) (*resources.Service, error) {
	// TODO: update attributes in clout

	// Artifacts
	artifactMappings := make(map[string]string)
	self.Log.Infof("processing artifacts for Clout: %s", service.Status.CloutPath)
	if clout, err := self.ReadClout(service.Status.CloutPath, false, false, urlContext); err == nil {
		if yaml, err := common.ExecScriptlet(clout, "kubernetes.artifacts", urlContext); err == nil {
			if artifacts, err := format.DecodeYAML(yaml); err == nil {
				if artifactMappings, err = self.pushArtifactsToInventory(artifacts, service, urlContext); err != nil {
					return nil, err
				}
			} else if err != io.EOF {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	if len(artifactMappings) > 0 {
		self.Log.Infof("updating artifacts for Clout: %s", service.Status.CloutPath)
		if clout, err := self.ReadClout(service.Status.CloutPath, false, false, urlContext); err == nil {
			self.UpdateCloutArtifacts(clout, artifactMappings)
			if service, err = self.UpdateClout(clout, service); err != nil {
				return nil, err
			}
		}
	}

	// Policies
	orchestrationPolicies := make(parser.OrchestrationPolicies)
	self.Log.Infof("processing policies for Clout: %s", service.Status.CloutPath)
	if clout, err := self.ReadClout(service.Status.CloutPath, false, false, urlContext); err == nil {
		if yaml, err := common.ExecScriptlet(clout, "orchestration.policies", urlContext); err == nil {
			if policies, err := format.DecodeYAML(yaml); err == nil {
				if orchestrationPolicies, err = self.processPolicies(policies); err != nil {
					return nil, err
				}
			} else if err != io.EOF {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	// Substitution
	for nodeTemplateName, policies := range orchestrationPolicies {
		for _, policy := range policies {
			if policy.Substitutable {
				for _, site := range policy.Sites {
					if err := self.Substitute(service.Namespace, nodeTemplateName, policy.SubstitutionInputs, site, urlContext); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	// Kubernetes resources
	// TODO: need to filter only non-substituted and instantiable node templates
	self.Log.Infof("creating Kubernetes resources for Clout: %s", service.Status.CloutPath)
	if clout, err := self.ReadClout(service.Status.CloutPath, false, false, urlContext); err == nil {
		if yaml, err := common.ExecScriptlet(clout, "kubernetes.resources", urlContext); err == nil {
			if objects, err := common.DecodeAllYAML(yaml); err == nil {
				if resources, err := self.createResources(objects, service); err == nil {
					if clout, err := self.ReadClout(service.Status.CloutPath, false, false, urlContext); err == nil {
						self.UpdateCloutResourcesMetadata(clout, resources)
						if service, err = self.UpdateClout(clout, service); err != nil {
							return nil, err
						}
					} else {
						return nil, err
					}
				} else {
					return nil, err
				}
			} else if err != io.EOF {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	return self.updateCloutOutputs(service, urlContext)
}

func (self *Controller) updateCloutOutputs(service *resources.Service, urlContext *urlpkg.Context) (*resources.Service, error) {
	self.Log.Infof("processing outputs for Clout: %s", service.Status.CloutPath)
	if clout, err := self.ReadClout(service.Status.CloutPath, true, true, urlContext); err == nil {
		if outputs, ok := parser.GetOutputs(clout); ok {
			if service, err = self.UpdateServiceOutputs(service, outputs); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("could not parse outputs for Clout: %s", service.Status.CloutPath)
		}
	} else {
		return nil, err
	}

	return service, nil
}
