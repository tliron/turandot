package controller

import (
	"encoding/json"
	"strings"

	"github.com/tliron/puccini/ard"
	cloutpkg "github.com/tliron/puccini/clout"
	puccinicommon "github.com/tliron/puccini/common"
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	errorspkg "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (self *Controller) UpdateCloutResourcesMetadata(clout *cloutpkg.Clout, resources parser.KubernetesResources) {
	history := ard.StringMap{
		"description": "kubernetes-resources-metadata",
		"timestamp":   puccinicommon.Timestamp(false),
	}
	ard.NewNode(clout.Metadata).Get("history").Append(history)

	for vertexId := range resources {
		if vertex, ok := clout.Vertexes[vertexId]; ok {
			self.Log.Infof("updating resource metadata for Clout vertex: %s", vertexId)

			var turandot ard.StringMap
			if turandot, ok = ard.NewNode(vertex.Metadata).Get("turandot").StringMap(false); !ok {
				turandot = make(ard.StringMap)
				turandot["version"] = "1.0"
				vertex.Metadata["turandot"] = turandot
			}

			turandot["resources"] = resources.ARD(vertexId)
		}
	}
}

func (self *Controller) UpdateCloutAttributesFromResources(service *resources.Service, urlContext *urlpkg.Context) (*resources.Service, error) {
	self.Log.Infof("updating resource attributes for Clout: %s", service.Status.CloutPath)
	if clout, err := self.ReadClout(service.Status.CloutPath, false, false, urlContext); err == nil {
		changed := false

		for vertexId, vertex := range clout.Vertexes {
			if resources, ok := ard.NewNode(vertex.Metadata).Get("turandot").Get("resources").List(false); ok {
				if resources_, ok := parser.ParseKubernetesResourceList(resources); ok {
					for _, resource := range resources_ {
						self.Log.Infof("updating attributes for %s/%s in Clout from resource %s/%s %s/%s", vertexId, resource.Capability, resource.APIVersion, resource.Kind, resource.Namespace, resource.Name)
						if resource.AttributeMappings != nil {
							for from, attributeName := range resource.AttributeMappings {
								self.Log.Infof("updating attribute %s from %s", attributeName, from)

								if gvk, err := resource.GVK(); err == nil {
									if unstructured, err := self.Dynamic.GetResource(gvk, resource.Name, resource.Namespace); err == nil {
										if attributes, ok := ard.NewNode(vertex.Properties).Get("capabilities").Get(resource.Capability).Get("attributes").StringMap(false); ok {
											fromNode := ard.NewNode(unstructured.Object)
											for _, element := range strings.Split(from, ".") {
												fromNode = fromNode.Get(element)
											}

											newValue := parser.ToCloutCoercible(fromNode.Data)
											if currentValue, ok := attributes[attributeName]; ok {
												if currentValue_, ok := currentValue.(ard.StringMap); ok {
													if !parser.CloutCoerciblesMerged(currentValue_, newValue) {
														self.Log.Infof("merging attribute: %v", newValue)
														parser.MergeCloutCoercibles(currentValue_, newValue)
														changed = true
													} else {
														self.Log.Infof("attribute not changed: %v", currentValue)
													}
												} else {
													self.Log.Errorf("malformed attribute coercible: %v", currentValue)
												}
											} else {
												self.Log.Infof("setting attribute: %v", newValue)
												attributes[attributeName] = newValue
												changed = true
											}
										} else {
											self.Log.Errorf("no attributes for capability: %s", resource.Capability)
										}
									} else {
										return nil, err
									}
								} else {
									return nil, err
								}
							}
						}
					}
				}
			}
		}

		if changed {
			if service, err := self.UpdateClout(clout, service); err == nil {
				return self.updateCloutOutputs(service, urlContext)
			} else {
				return nil, err
			}
		}
	} else {
		return nil, err
	}

	return service, nil
}

func (self *Controller) createResources(objects []ard.StringMap, owner meta.Object) (parser.KubernetesResources, error) {
	resources := parser.NewKubernetesResources()

	for _, object := range objects {
		object_ := &unstructured.Unstructured{Object: object}
		self.Log.Infof("creating resource %s/%s %s/%s", object_.GetAPIVersion(), object_.GetKind(), object_.GetNamespace(), object_.GetName())
		var err error
		if object_, err = self.Dynamic.CreateControlledResource(object_, owner, self.Processors, self.StopChannel); err == nil {
			if vertexId, ok := object_.GetAnnotations()["clout.puccini.cloud/vertex"]; ok {
				if capability, ok := object_.GetAnnotations()["clout.puccini.cloud/capability"]; ok {
					var attributesMappings map[string]string
					if attributeMappings_, ok := object_.GetAnnotations()["clout.puccini.cloud/attributeMappings"]; ok {
						if err := json.Unmarshal(puccinicommon.StringToBytes(attributeMappings_), &attributesMappings); err != nil {
							return nil, err
						}
					}

					self.Log.Infof("adding resource %s %s %s %s to %s/%s", object_.GetAPIVersion(), object_.GetKind(), object_.GetName(), object_.GetNamespace(), vertexId, capability)
					resources.Add(vertexId, capability, object_.GetAPIVersion(), object_.GetKind(), object_.GetName(), object_.GetNamespace(), attributesMappings)
				}
			}
		} else if errorspkg.IsAlreadyExists(err) {
			self.Log.Infof("%s", err.Error())
		} else {
			return nil, err
		}
	}

	// TODO: delete other resources owned by us

	return resources, nil
}
