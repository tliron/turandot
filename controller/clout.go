package controller

import (
	"fmt"
	"io"
	"reflect"

	"github.com/tliron/puccini/ard"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
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
					if compiler.Resolve(clout, problems, urlContext, false, "yaml", false, true, true); !problems.Empty() {
						return nil, fmt.Errorf("could not resolve Clout\n%s", problems.ToString(true))
					}
				}

				if coerce {
					if compiler.Coerce(clout, problems, urlContext, false, "yaml", false, true, true); !problems.Empty() {
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

func (self *Controller) WriteClout(yaml string, cloutPath string) (string, error) {
	if file, err := format.OpenFileForWrite(cloutPath); err == nil {
		defer file.Close()
		if _, err := file.WriteString(yaml); err == nil {
			return common.GetFileHash(cloutPath)
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}

func (self *Controller) UpdateClout(yaml string, service *resources.Service) (*resources.Service, error) {
	if cloutHash, err := self.WriteClout(yaml, service.Status.CloutPath); err == nil {
		return self.UpdateServiceStatusClout(service, service.Status.CloutPath, cloutHash)
	} else {
		return nil, err
	}
}

func (self *Controller) executeCloutGet(service *resources.Service, urlContext *urlpkg.Context, scriptletName string, arguments map[string]string) (ard.Value, error) {
	if clout, err := self.ReadClout(service.Status.CloutPath, false, false, urlContext); err == nil {
		if yaml, err := common.ExecScriptlet(clout, scriptletName, arguments, urlContext); err == nil {
			if value, err := format.DecodeYAML(yaml); err == nil {
				return value, nil
			} else if err != io.EOF {
				return nil, err
			} else {
				return nil, nil
			}
		} else if js.IsScriptletNotFoundError(err) {
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Controller) executeCloutGetAll(service *resources.Service, urlContext *urlpkg.Context, scriptletName string, arguments map[string]string) ([]ard.StringMap, error) {
	if clout, err := self.ReadClout(service.Status.CloutPath, false, false, urlContext); err == nil {
		if yaml, err := common.ExecScriptlet(clout, scriptletName, arguments, urlContext); err == nil {
			if value, err := common.DecodeAllYAML(yaml); err == nil {
				return value, nil
			} else if err != io.EOF {
				return nil, err
			} else {
				return nil, nil
			}
		} else if js.IsScriptletNotFoundError(err) {
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Controller) executeCloutUpdate(service *resources.Service, urlContext *urlpkg.Context, scriptletName string, arguments map[string]string) (*resources.Service, error) {
	if clout, err := self.ReadClout(service.Status.CloutPath, false, false, urlContext); err == nil {
		if yaml, err := common.ExecScriptlet(clout, scriptletName, arguments, urlContext); err == nil {
			if yaml != "" {
				return self.UpdateClout(yaml, service)
			} else {
				return service, nil
			}
		} else if js.IsScriptletNotFoundError(err) {
			return service, nil
		} else {
			return service, err
		}
	} else {
		return service, err
	}
}

func (self *Controller) instantiateClout(service *resources.Service, urlContext *urlpkg.Context) (*resources.Service, error) {
	var err error

	// Get artifacts
	self.Log.Infof("getting artifacts from Clout: %s", service.Status.CloutPath)
	var artifacts ard.Value
	if artifacts, err = self.executeCloutGet(service, urlContext, "kubernetes.artifacts.get", nil); err != nil {
		return service, err
	}

	// Push artifacts
	var artifactMappings map[string]string
	if artifacts != nil {
		if artifacts_, ok := parser.ParseKubernetesArtifacts(artifacts); ok {
			if artifactMappings, err = self.pushArtifactsToInventory(artifacts_, service, urlContext); err != nil {
				return service, err
			}
		} else {
			return service, fmt.Errorf("could not parse artifacts: %s", artifacts)
		}
	}

	// Update artifacts
	if artifactMappings != nil {
		self.Log.Infof("updating artifacts in Clout: %s", service.Status.CloutPath)
		if service, err = self.executeCloutUpdate(service, urlContext, "kubernetes.artifacts.update", artifactMappings); err != nil {
			return service, err
		}
	}

	// Get policies
	var policies ard.Value
	self.Log.Infof("getting policies from Clout: %s", service.Status.CloutPath)
	if policies, err = self.executeCloutGet(service, urlContext, "orchestration.policies", nil); err != nil {
		return service, err
	}

	// Process policies
	if policies != nil {
		if orchestrationPolicies, ok := parser.ParseOrchestrationPolicies(policies); ok {
			if err := self.processPolicies(orchestrationPolicies, service, urlContext); err != nil {
				return service, err
			}
		} else {
			return service, fmt.Errorf("could not parse policies: %v", policies)
		}
	}

	// Get resources
	// TODO: need to filter only non-substituted and instantiable node templates
	self.Log.Infof("getting Kubernetes resources from Clout: %s", service.Status.CloutPath)
	var objects []ard.StringMap
	if objects, err = self.executeCloutGetAll(service, urlContext, "kubernetes.resources.get", nil); err != nil {
		return service, err
	}

	// Create resources
	var resourceMappings parser.KubernetesResourceMappings
	if objects != nil {
		if resourceMappings, err = self.createResources(objects, service); err != nil {
			return service, err
		}
	}

	// Update resource mappings
	if resourceMappings != nil {
		self.Log.Infof("updating resource mappings in Clout: %s", service.Status.CloutPath)
		if service, err = self.executeCloutUpdate(service, urlContext, "kubernetes.resources.update-mappings", resourceMappings.StringMap()); err != nil {
			return service, err
		}
	}

	// Get executions
	// TODO: recompilation error if this fails???
	var executions ard.Value
	self.Log.Infof("getting executions for Clout: %s", service.Status.CloutPath)
	if executions, err = self.executeCloutGet(service, urlContext, "orchestration.executions", nil); err != nil {
		return service, err
	}

	// Process executions
	if executions != nil {
		if orchestrationExecutions, ok := parser.ParseOrchestrationExecutions(executions); ok {
			if service, err = self.processExecutions(orchestrationExecutions, service, urlContext); err != nil {
				return service, err
			}
		} else {
			return service, fmt.Errorf("could not parse executions: %v", executions)
		}
	}

	return self.updateCloutOutputs(service, urlContext)
}

func (self *Controller) updateCloutFromResources(service *resources.Service, urlContext *urlpkg.Context) (*resources.Service, error) {
	var err error

	self.Log.Infof("get resource mappings from Clout: %s", service.Status.CloutPath)

	var mappings ard.Value
	if mappings, err = self.executeCloutGet(service, urlContext, "kubernetes.resources.get-mappings", nil); err != nil {
		return service, err
	}

	var attributeValues parser.CloutAttributeValues
	if mappings != nil {
		if resourceMappings, ok := parser.ParseKubernetesResourceMappings(mappings); ok {
			if attributeValues, err = self.GetAttributesFromResources(resourceMappings); err != nil {
				return service, err
			}
		} else {
			return service, fmt.Errorf("could not parse resource mappings: %v", mappings)
		}
	}

	if attributeValues != nil {
		self.Log.Infof("updating attributes in Clout: %s", service.Status.CloutPath)
		if service, err = self.executeCloutUpdate(service, urlContext, "kubernetes.resources.update-attributes", attributeValues.StringMap()); err != nil {
			return service, err
		}
	}

	return self.updateCloutOutputs(service, urlContext)
}

func (self *Controller) updateCloutOutputs(service *resources.Service, urlContext *urlpkg.Context) (*resources.Service, error) {
	self.Log.Infof("processing outputs for Clout: %s", service.Status.CloutPath)
	if clout, err := self.ReadClout(service.Status.CloutPath, true, true, urlContext); err == nil {
		if outputs, ok := parser.GetOutputs(clout); ok {
			if !reflect.DeepEqual(outputs, service.Status.Outputs) {
				return self.UpdateServiceStatusOutputs(service, outputs)
			} else {
				return service, nil
			}
		} else {
			return service, fmt.Errorf("could not parse outputs for Clout: %s", service.Status.CloutPath)
		}
	} else {
		return service, err
	}
}
