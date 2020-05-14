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
	Capability string
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
}

func ParseKubernetesResources(value ard.Value) ([]*KubernetesResource, bool) {
	if list, ok := value.(ard.List); ok {
		var resources []*KubernetesResource
		for _, entry := range list {
			node := ard.NewNode(entry)
			if capability, ok := node.Get("capability").String(false); ok {
				if apiVersion, ok := node.Get("apiVersion").String(false); ok {
					if kind, ok := node.Get("kind").String(false); ok {
						if name, ok := node.Get("name").String(false); ok {
							if namespace, ok := node.Get("namespace").String(false); ok {
								resource := &KubernetesResource{
									Capability: capability,
									APIVersion: apiVersion,
									Kind:       kind,
									Name:       name,
									Namespace:  namespace,
								}
								resources = append(resources, resource)
							}
						}
					}
				}
			}
		}
		return resources, true
	}
	return nil, false
}

func (self *KubernetesResource) ARD() ard.Map {
	map_ := make(ard.Map)
	map_["capability"] = self.Capability
	map_["apiVersion"] = self.APIVersion
	map_["kind"] = self.Kind
	map_["name"] = self.Name
	map_["namespace"] = self.Namespace
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

func (self KubernetesResources) Add(vertexId string, capability string, apiVersion string, kind string, name string, namespace string) {
	resource := &KubernetesResource{
		Capability: capability,
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		Namespace:  namespace,
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
