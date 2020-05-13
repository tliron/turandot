package parser

import (
	"github.com/tliron/puccini/ard"
)

//
// KubernetesResource
//

type KubernetesResource struct {
	APIVersion string
	Kind       string
	Name       string
	Capability string
}

func (self *KubernetesResource) ARD() ard.Map {
	map_ := make(ard.Map)
	map_["apiVersion"] = self.APIVersion
	map_["kind"] = self.Kind
	map_["name"] = self.Name
	map_["capability"] = self.Capability
	return map_
}

func ParseKubernetesResources(value ard.Value) ([]*KubernetesResource, bool) {
	if list, ok := value.(ard.List); ok {
		var resources []*KubernetesResource
		for _, entry := range list {
			node := ard.NewNode(entry)
			if apiVersion, ok := node.Get("apiVersion").String(false); ok {
				if kind, ok := node.Get("kind").String(false); ok {
					if name, ok := node.Get("name").String(false); ok {
						if capability, ok := node.Get("capability").String(false); ok {
							resource := &KubernetesResource{
								APIVersion: apiVersion,
								Kind:       kind,
								Name:       name,
								Capability: capability,
							}
							resources = append(resources, resource)
						}
					}
				}
			}
		}
		return resources, true
	}
	return nil, false
}

//
// KubernetesResources
//

type KubernetesResources map[string][]*KubernetesResource

func NewKubernetesResources() KubernetesResources {
	return make(KubernetesResources)
}

func (self KubernetesResources) Add(vertexId string, apiVersion string, kind string, name string, capability string) {
	resource := &KubernetesResource{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		Capability: capability,
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
