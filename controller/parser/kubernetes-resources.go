package parser

import (
	"github.com/tliron/puccini/ard"
	"github.com/tliron/turandot/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//
// KubernetesResource
//

type KubernetesResource struct {
	Capability        string
	APIVersion        string
	Kind              string
	Name              string
	Namespace         string
	AttributeMappings map[string]string
}

func ParseKubernetesResourceList(list ard.List) ([]*KubernetesResource, bool) {
	var resources []*KubernetesResource
	for _, entry := range list {
		node := ard.NewNode(entry)
		if capability, ok := node.Get("capability").String(false); ok {
			if apiVersion, ok := node.Get("apiVersion").String(false); ok {
				if kind, ok := node.Get("kind").String(false); ok {
					if name, ok := node.Get("name").String(false); ok {
						if namespace, ok := node.Get("namespace").String(false); ok {
							attributeMappings := make(map[string]string)
							if attributes, ok := node.Get("attributes").StringMap(false); ok {
								for key, value := range attributes {
									if value_, ok := value.(string); ok {
										attributeMappings[key] = value_
									} else {
										return nil, false
									}
								}
							}
							if len(attributeMappings) == 0 {
								attributeMappings = nil
							}

							resource := &KubernetesResource{
								Capability:        capability,
								APIVersion:        apiVersion,
								Kind:              kind,
								Name:              name,
								Namespace:         namespace,
								AttributeMappings: attributeMappings,
							}
							resources = append(resources, resource)
						} else {
							return nil, false
						}
					} else {
						return nil, false
					}
				} else {
					return nil, false
				}
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	return resources, true
}

func (self *KubernetesResource) ARD() ard.Map {
	map_ := make(ard.Map)
	map_["capability"] = self.Capability
	map_["apiVersion"] = self.APIVersion
	map_["kind"] = self.Kind
	map_["name"] = self.Name
	map_["namespace"] = self.Namespace
	if (self.AttributeMappings != nil) && (len(self.AttributeMappings) > 0) {
		map_["attributes"] = self.AttributeMappings
	}
	return map_
}

func (self *KubernetesResource) GVK() (schema.GroupVersionKind, error) {
	if gvk, err := common.ParseGVK(self.APIVersion, self.Kind); err == nil {
		return gvk, nil
	} else {
		return schema.GroupVersionKind{}, err
	}
}

//
// KubernetesResources
//

type KubernetesResources map[string][]*KubernetesResource

func NewKubernetesResources() KubernetesResources {
	return make(KubernetesResources)
}

func (self KubernetesResources) Add(vertexId string, capability string, apiVersion string, kind string, name string, namespace string, attributeMappings map[string]string) {
	resource := &KubernetesResource{
		Capability:        capability,
		APIVersion:        apiVersion,
		Kind:              kind,
		Name:              name,
		Namespace:         namespace,
		AttributeMappings: attributeMappings,
	}
	self[vertexId] = append(self[vertexId], resource)
}

func (self KubernetesResources) ARD(vertexId string) ard.List {
	var list ard.List
	if resources, ok := self[vertexId]; ok {
		list = make(ard.List, len(resources))
		for index, resource := range resources {
			list[index] = resource.ARD()
		}
	}
	return list
}
