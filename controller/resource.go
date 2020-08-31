package controller

import (
	"encoding/json"
	"strings"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/util"
	"github.com/tliron/turandot/controller/parser"
	errorspkg "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (self *Controller) GetResource(resourceMapping *parser.KubernetesResourceMapping) (*unstructured.Unstructured, error) {
	if gvk, err := resourceMapping.GVK(); err == nil {
		return self.Dynamic.GetResource(gvk, resourceMapping.Name, resourceMapping.Namespace)
	} else {
		return nil, err
	}
}

func (self *Controller) GetAttributesFromResources(resourceMappings parser.KubernetesResourceMappings) (parser.CloutAttributeValues, error) {
	attributeValues := parser.NewCloutAttributeValues()

	for vertexId, resourceMappingList := range resourceMappings {
		for _, resourceMapping := range resourceMappingList {
			if resourceMapping.AttributeMappings != nil {
				if unstructured, err := self.GetResource(resourceMapping); err == nil {
					for from, attributeName := range resourceMapping.AttributeMappings {
						fromNode := ard.NewNode(unstructured.Object)
						for _, element := range strings.Split(from, ".") {
							fromNode = fromNode.Get(element)
						}

						attributeValues.Set(vertexId, resourceMapping.Capability, attributeName, fromNode.Data)
					}
				} else {
					return nil, err
				}
			}
		}
	}

	if len(attributeValues) == 0 {
		return nil, nil
	} else {
		return attributeValues, nil
	}
}

func (self *Controller) createResources(objects []ard.StringMap, owner meta.Object) (parser.KubernetesResourceMappings, error) {
	resourceMappings := parser.NewKubernetesResourceMappings()

	for _, object := range objects {
		object_ := &unstructured.Unstructured{Object: object}
		self.Log.Infof("creating resource %s/%s %s/%s", object_.GetAPIVersion(), object_.GetKind(), object_.GetNamespace(), object_.GetName())
		var err error
		if object_, err = self.Dynamic.CreateControlledResource(object_, owner, self.Processors, self.StopChannel); err == nil {
			if vertexId, ok := object_.GetAnnotations()["clout.puccini.cloud/vertex"]; ok {
				if capability, ok := object_.GetAnnotations()["clout.puccini.cloud/capability"]; ok {
					var attributesMappings map[string]string
					if attributeMappings_, ok := object_.GetAnnotations()["clout.puccini.cloud/attributeMappings"]; ok {
						if err := json.Unmarshal(util.StringToBytes(attributeMappings_), &attributesMappings); err != nil {
							return nil, err
						}
					}

					self.Log.Infof("adding resource mapping %s %s %s %s to %s/%s", object_.GetAPIVersion(), object_.GetKind(), object_.GetName(), object_.GetNamespace(), vertexId, capability)
					resourceMappings.Add(vertexId, capability, object_.GetAPIVersion(), object_.GetKind(), object_.GetName(), object_.GetNamespace(), attributesMappings)
				}
			}
		} else if errorspkg.IsAlreadyExists(err) {
			self.Log.Infof("%s", err.Error())
		} else {
			return nil, err
		}
	}

	// TODO: delete other resources owned by us

	if len(resourceMappings) == 0 {
		return nil, nil
	} else {
		return resourceMappings, nil
	}
}
