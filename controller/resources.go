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
		self.log.Infof("creating resource \"%s/%s %s\"", object_.GetAPIVersion(), object_.GetKind(), object_.GetName())
		if _, err := self.dynamic.CreateControlledResource(object_, owner, self.processors, self.stopChannel); err != nil {
			if errorspkg.IsAlreadyExists(err) {
				self.log.Infof("%s", err.Error())
			} else {
				return err
			}
		}
	}

	// TODO: delete other resources owned by us

	return nil
}
