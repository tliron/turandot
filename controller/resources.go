package controller

import (
	"github.com/tliron/puccini/ard"
	errorspkg "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (self *Controller) processResources(objects []ard.StringMap, owner meta.Object) error {
	for _, object := range objects {
		object_ := &unstructured.Unstructured{Object: object}
		self.Log.Infof("creating resource %s/%s %s", object_.GetAPIVersion(), object_.GetKind(), object_.GetName())
		if unstructured, err := self.Dynamic.CreateControlledResource(object_, owner, self.Processors, self.StopChannel); err == nil {
			//self.Log.Errorf("!!!!!!!!!!!!!!!!! %s", unstructured.GetUID())
		} else if errorspkg.IsAlreadyExists(err) {
			self.Log.Infof("%s", err.Error())
		} else {
			return err
		}
	}

	// TODO: delete other resources owned by us

	return nil
}
