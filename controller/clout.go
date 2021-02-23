package controller

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/format"
	problemspkg "github.com/tliron/kutil/problems"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
	"github.com/tliron/puccini/tosca/compiler"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	"github.com/tliron/yamlkeys"
)

func (self *Controller) ReadClout(cloutPath string, resolve bool, coerce bool, urlContext *urlpkg.Context) (*cloutpkg.Clout, error) {
	if url, err := urlpkg.NewURL(cloutPath, urlContext); err == nil {
		if reader, err := url.Open(); err == nil {
			defer reader.Close()
			if clout, err := ReadClout(reader, urlContext); err == nil {
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
			return util.GetFileHash(cloutPath)
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}

func (self *Controller) WriteServiceClout(yaml string, service *resources.Service) (*resources.Service, error) {
	if cloutHash, err := self.WriteClout(yaml, service.Status.CloutPath); err == nil {
		return self.UpdateServiceStatusClout(service, service.Status.CloutPath, cloutHash)
	} else {
		return service, err
	}
}

func (self *Controller) executeCloutGet(service *resources.Service, urlContext *urlpkg.Context, scriptletName string, arguments map[string]string) (ard.Value, error) {
	if clout, err := self.ReadClout(service.Status.CloutPath, false, false, urlContext); err == nil {
		if yaml, err := ExecCloutScriptlet(clout, scriptletName, arguments, urlContext); err == nil {
			if value, _, err := ard.DecodeYAML(yaml, false); err == nil {
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
		if yaml, err := ExecCloutScriptlet(clout, scriptletName, arguments, urlContext); err == nil {
			if values, err := yamlkeys.DecodeAll(strings.NewReader(yaml)); err == nil {
				list := make([]ard.StringMap, len(values))
				for index, value := range values {
					value, _ = ard.MapsToStringMaps(value)
					if value_, ok := value.(ard.StringMap); ok {
						list[index] = value_
					} else {
						return nil, fmt.Errorf("not a map[string]interface{}: %T", value)
					}
				}
				return list, nil
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
		if yaml, err := ExecCloutScriptlet(clout, scriptletName, arguments, urlContext); err == nil {
			if yaml != "" {
				return self.WriteServiceClout(yaml, service)
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

	// Publish artifacts
	var artifactMappings map[string]string
	if artifacts != nil {
		if artifacts_, ok := parser.ParseKubernetesArtifacts(artifacts); ok {
			if artifactMappings, err = self.publishArtifactsToRegistry(artifacts_, service, urlContext); err != nil {
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
			return service, fmt.Errorf("could not parse policies: %+v", policies)
		}
	}

	// Get Kubernetes resources
	// TODO: need to filter only non-substituted and instantiable node templates
	self.Log.Infof("getting Kubernetes resources from Clout: %s", service.Status.CloutPath)
	var objects []ard.StringMap
	if objects, err = self.executeCloutGetAll(service, urlContext, "kubernetes.resources.get", nil); err != nil {
		return service, err
	}

	// Create Kubernetes resources
	var resourceMappings parser.KubernetesResourceMappings
	if objects != nil {
		if resourceMappings, err = self.createResources(objects, service); err != nil {
			return service, err
		}
	}

	// Update Kubernetes resource mappings
	if resourceMappings != nil {
		self.Log.Infof("updating resource mappings in Clout: %s", service.Status.CloutPath)
		if service, err = self.executeCloutUpdate(service, urlContext, "kubernetes.resources.update-mappings", resourceMappings.StringMap()); err != nil {
			return service, err
		}
	}

	// TODO: debug weird recompilation namespace errors

	return self.updateServiceStatusFromClout(service, urlContext)
}

func (self *Controller) updateClout(service *resources.Service, urlContext *urlpkg.Context) (*resources.Service, error) {
	var err error

	// Change mode?
	if service.Status.Mode != service.Spec.Mode {
		if service, err = self.UpdateServiceStatusMode(service); err != nil {
			return service, err
		}
		self.Log.Infof("resetting node states in Clout: %s", service.Status.CloutPath)
		if service, err = self.executeCloutUpdate(service, urlContext, "kubernetes.executions.reset", map[string]string{
			"service": service.Name,
			"mode":    service.Status.Mode,
		}); err != nil {
			return service, err
		}
	}

	// Get executions
	var executions ard.Value
	self.Log.Infof("getting executions for Clout: %s", service.Status.CloutPath)
	if executions, err = self.executeCloutGet(service, urlContext, "kubernetes.executions", map[string]string{
		"service": service.Name,
	}); err != nil {
		return service, err
	}

	// Process executions
	if executions != nil {
		if orchestrationExecutions, ok := parser.ParseOrchestrationExecutions(executions); ok {
			if orchestrationExecutions != nil {
				if service, err = self.processExecutions(orchestrationExecutions, service, urlContext); err != nil {
					return service, err
				}
			}
		} else {
			return service, fmt.Errorf("could not parse executions: %+v", executions)
		}
	}

	// Get Kubernetes resource mappings
	self.Log.Infof("get resource mappings from Clout: %s", service.Status.CloutPath)
	var mappings ard.Value
	if mappings, err = self.executeCloutGet(service, urlContext, "kubernetes.resources.get-mappings", nil); err != nil {
		return service, err
	}

	// Get Clout attribute values from Kubernetes resources
	var attributeValues parser.CloutAttributeValues
	if mappings != nil {
		if resourceMappings, ok := parser.ParseKubernetesResourceMappings(mappings); ok {
			if attributeValues, err = self.GetAttributesFromResources(resourceMappings); err != nil {
				return service, err
			}
		} else {
			return service, fmt.Errorf("could not parse resource mappings: %+v", mappings)
		}
	}

	// Update attributes in Clout
	if attributeValues != nil {
		self.Log.Infof("updating attributes in Clout: %s", service.Status.CloutPath)
		if service, err = self.executeCloutUpdate(service, urlContext, "kubernetes.resources.update-attributes", attributeValues.StringMap()); err != nil {
			return service, err
		}
	}

	return self.updateServiceStatusFromClout(service, urlContext)
}

func (self *Controller) updateServiceStatusFromClout(service *resources.Service, urlContext *urlpkg.Context) (*resources.Service, error) {
	var err error

	// Get outputs
	self.Log.Infof("getting outputs from Clout: %s", service.Status.CloutPath)
	var outputs map[string]string
	if clout, err := self.ReadClout(service.Status.CloutPath, true, true, urlContext); err == nil {
		var ok bool
		if outputs, ok = parser.GetOutputs(clout); !ok {
			return service, fmt.Errorf("could not parse outputs for Clout: %s", service.Status.CloutPath)
		}
	} else {
		return service, err
	}

	// Update outputs in status
	if !reflect.DeepEqual(outputs, service.Status.Outputs) {
		if service, err = self.UpdateServiceStatusOutputs(service, outputs); err != nil {
			return service, err
		}
	}

	// Get node states
	self.Log.Infof("get node states from Clout: %s", service.Status.CloutPath)
	var states ard.Value
	if states, err = self.executeCloutGet(service, urlContext, "orchestration.states.get", nil); err != nil {
		return service, err
	}

	// Process node states
	if states != nil {
		if orchestrationStates, ok := parser.ParseOrchestrationStates(states); ok {
			if serviceStates, ok := orchestrationStates[service.Name]; ok {
				if service, err = self.UpdateServiceStatusNodeStates(service, serviceStates); err != nil {
					return service, err
				}
			}
		} else {
			return service, fmt.Errorf("could not parse node states: %+v", states)
		}
	}

	return service, nil
}
