package parser

import (
	"encoding/json"

	"github.com/tliron/puccini/ard"
	puccinicommon "github.com/tliron/puccini/common"
	"github.com/tliron/turandot/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//
// KubernetesResourceMapping
//

type KubernetesResourceMapping struct {
	Capability        string            `json:"capability"`
	APIVersion        string            `json:"apiVersion"`
	Kind              string            `json:"kind"`
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	AttributeMappings map[string]string `json:"attributes,omitempty"`
}

//
// KubernetesResourceMappingList
//

type KubernetesResourceMappingList []*KubernetesResourceMapping

func ParseKubernetesResourceMappingList(value ard.Value) (KubernetesResourceMappingList, bool) {
	if list, ok := value.(ard.List); ok {
		var mappings KubernetesResourceMappingList
		for _, element := range list {
			node := ard.NewNode(element)
			if capability, ok := node.Get("capability").String(false); ok {
				if apiVersion, ok := node.Get("apiVersion").String(false); ok {
					if kind, ok := node.Get("kind").String(false); ok {
						if name, ok := node.Get("name").String(false); ok {
							if namespace, ok := node.Get("namespace").String(false); ok {
								attributeMappings := make(map[string]string)
								if attributes, ok := node.Get("attributes").Map(false); ok {
									for key, value := range attributes {
										if key_, ok := key.(string); ok {
											if value_, ok := value.(string); ok {
												attributeMappings[key_] = value_
											} else {
												return nil, false
											}
										} else {
											return nil, false
										}
									}
								}
								if len(attributeMappings) == 0 {
									attributeMappings = nil
								}

								mapping := &KubernetesResourceMapping{
									Capability:        capability,
									APIVersion:        apiVersion,
									Kind:              kind,
									Name:              name,
									Namespace:         namespace,
									AttributeMappings: attributeMappings,
								}
								mappings = append(mappings, mapping)
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
		return mappings, true
	} else {
		return nil, false
	}
}

func (self *KubernetesResourceMapping) GVK() (schema.GroupVersionKind, error) {
	if gvk, err := common.ParseGVK(self.APIVersion, self.Kind); err == nil {
		return gvk, nil
	} else {
		return schema.GroupVersionKind{}, err
	}
}

//
// KubernetesResourceMappings
//

type KubernetesResourceMappings map[string]KubernetesResourceMappingList

func NewKubernetesResourceMappings() KubernetesResourceMappings {
	return make(KubernetesResourceMappings)
}

func ParseKubernetesResourceMappings(value ard.Value) (KubernetesResourceMappings, bool) {
	if mappings, ok := value.(ard.Map); ok {
		mappings_ := NewKubernetesResourceMappings()
		for vertexId, mappingList := range mappings {
			if vertexId_, ok := vertexId.(string); ok {
				if mappings_[vertexId_], ok = ParseKubernetesResourceMappingList(mappingList); !ok {
					return nil, false
				}
			} else {
				return nil, false
			}
		}
		return mappings_, true
	}
	return nil, false
}

func (self KubernetesResourceMappings) Add(vertexId string, capability string, apiVersion string, kind string, name string, namespace string, attributeMappings map[string]string) {
	mapping := &KubernetesResourceMapping{
		Capability:        capability,
		APIVersion:        apiVersion,
		Kind:              kind,
		Name:              name,
		Namespace:         namespace,
		AttributeMappings: attributeMappings,
	}
	self[vertexId] = append(self[vertexId], mapping)
}

func (self KubernetesResourceMappings) StringMap() map[string]string {
	map_ := make(map[string]string)
	for vertexId, list := range self {
		if bytes, err := json.Marshal(list); err == nil {
			map_[vertexId] = puccinicommon.BytesToString(bytes)
		}
	}
	return map_
}
