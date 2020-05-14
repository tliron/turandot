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

func (self *Controller) UpdateCloutResourcesMetadata(clout *cloutpkg.Clout, resources parser.KubernetesResources) {
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

func (self *Controller) createResources(objects []ard.StringMap, owner meta.Object) (parser.KubernetesResources, error) {
	resources := parser.NewKubernetesResources()

	for _, object := range objects {
		object_ := &unstructured.Unstructured{Object: object}
		self.Log.Infof("creating resource %s/%s %s", object_.GetAPIVersion(), object_.GetKind(), object_.GetName())
		var err error
		if object_, err = self.Dynamic.CreateControlledResource(object_, owner, self.Processors, self.StopChannel); err == nil {
			if vertexId, ok := object_.GetAnnotations()["puccini.cloud/vertex"]; ok {
				if capability, ok := object_.GetAnnotations()["puccini.cloud/capability"]; ok {
					self.Log.Infof("adding resource %s %s %s %s to %s/capability", object_.GetAPIVersion(), object_.GetKind(), object_.GetName(), object_.GetNamespace(), vertexId, capability)
					resources.Add(vertexId, capability, object_.GetAPIVersion(), object_.GetKind(), object_.GetName(), object_.GetNamespace())
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
