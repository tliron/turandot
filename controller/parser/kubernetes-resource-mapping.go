package parser

import (
	"github.com/tliron/kutil/format"
	"github.com/tliron/kutil/kubernetes"
	"github.com/tliron/kutil/util"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//
// KubernetesResourceMapping
//

type KubernetesResourceMapping struct {
	Capability        string            `yaml:"capability" json:"capability"`
	APIVersion        string            `yaml:"apiVersion" json:"apiVersion"`
	Kind              string            `yaml:"kind" json:"kind"`
	Name              string            `yaml:"name" json:"name"`
	Namespace         string            `yaml:"namespace" json:"namespace"`
	AttributeMappings map[string]string `yaml:"attributes,omitempty" json:"attributes,omitempty"`
}

//
// KubernetesResourceMappingList
//

type KubernetesResourceMappingList []*KubernetesResourceMapping

func (self *KubernetesResourceMapping) GVK() (schema.GroupVersionKind, error) {
	if gvk, err := kubernetes.ParseGVK(self.APIVersion, self.Kind); err == nil {
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

func DecodeKubernetesResourceMappings(code string) (KubernetesResourceMappings, bool) {
	var self KubernetesResourceMappings
	if err := yaml.Unmarshal(util.StringToBytes(code), &self); err == nil {
		return self, true
	} else {
		return nil, false
	}
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

func (self KubernetesResourceMappings) JSON() map[string]string {
	map_ := make(map[string]string)
	for vertexId, list := range self {
		if value, err := format.EncodeJSON(list, ""); err == nil {
			map_[vertexId] = value
		}
	}
	return map_
}
