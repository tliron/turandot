package controller

import (
	"github.com/tliron/puccini/ard"
	cloutpkg "github.com/tliron/puccini/clout"
	puccinicommon "github.com/tliron/puccini/common"
	"github.com/tliron/turandot/controller/parser"
	errorspkg "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (self *Controller) UpdateCloutResources(clout *cloutpkg.Clout, resources parser.KubernetesResources) {
	history := ard.StringMap{
		"description": "kubernetes-resources",
		"timestamp":   puccinicommon.Timestamp(false),
	}
	ard.NewNode(clout.Metadata).Get("puccini-tosca").Get("history").Append(history)

	for vertexId, _ := range resources {
		if vertex, ok := clout.Vertexes[vertexId]; ok {
			vertex.Metadata["turandot-kubernetes"] = resources.ARD(vertexId)
		}
	}
}

func (self *Controller) processResources(objects []ard.StringMap, owner meta.Object) (parser.KubernetesResources, error) {
	resources := parser.NewKubernetesResources()

	for _, object := range objects {
		object_ := &unstructured.Unstructured{Object: object}
		self.Log.Infof("creating resource %s/%s %s", object_.GetAPIVersion(), object_.GetKind(), object_.GetName())
		var err error
		if object_, err = self.Dynamic.CreateControlledResource(object_, owner, self.Processors, self.StopChannel); err == nil {
			if vertexId, ok := object_.GetAnnotations()["puccini.cloud/vertex"]; ok {
				if capability, ok := object_.GetAnnotations()["puccini.cloud/capability"]; ok {
					self.Log.Infof("adding resource %s %s %s to %s/capability", object_.GetAPIVersion(), object_.GetKind(), object_.GetName(), vertexId, capability)
					resources.Add(vertexId, object_.GetAPIVersion(), object_.GetKind(), object_.GetName(), capability)
				}
			}
			//self.Log.Errorf("!!!!!!!!!!!!!!!!! %s", unstructured.GetUID())
		} else if errorspkg.IsAlreadyExists(err) {
			self.Log.Infof("%s", err.Error())
		} else {
			return nil, err
		}
	}

	// TODO: delete other resources owned by us

	return resources, nil
}

func (self *Controller) updateAttributes(clout *cloutpkg.Clout) {
	for vertexId, vertex := range clout.Vertexes {
		if resources, ok := vertex.Metadata["turandot-kubernetes"]; ok {
			if resources_, ok := parser.ParseKubernetesResources(resources); ok {
				for _, resource := range resources_ {
					self.Log.Infof("updating attributes for %s/%s from resource %s %s %s", vertexId, resource.Capability, resource.APIVersion, resource.Kind, resource.Name)
				}
			} else {
				self.Log.Errorf("could not parse Kubernetes resources for: %s", vertexId)
			}
		}
	}
}
