package common

import (
	"fmt"

	"github.com/op/go-logging"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

func GetMetaObject(object interface{}, log *logging.Logger) (meta.Object, error) {
	switch o := object.(type) {
	case meta.Object:
		return o, nil

	case meta.ObjectMetaAccessor:
		return o.GetObjectMeta(), nil

	case cache.DeletedFinalStateUnknown:
		switch oo := o.Obj.(type) {
		case meta.Object:
			log.Infof("recovered deleted object '%s' from tombstone", oo.GetName())
			return oo, nil

		default:
			return nil, fmt.Errorf("error decoding object tombstone, invalid type: %T", o)
		}

	default:
		return nil, fmt.Errorf("error decoding object, invalid type: %T", object)
	}
}

func GetControllerOf(metaObject meta.Object) (schema.GroupVersionKind, string, error) {
	if ownerReference := meta.GetControllerOf(metaObject); ownerReference != nil {
		if gvk, err := ToGVK(ownerReference.APIVersion, ownerReference.Kind); err == nil {
			return gvk, ownerReference.Name, nil
		} else {
			return schema.GroupVersionKind{}, "", err
		}
	} else {
		return schema.GroupVersionKind{}, "", nil
	}
}

func ToGVK(apiVersion string, kind string) (schema.GroupVersionKind, error) {
	// Improvement of schema.FromAPIVersionAndKind
	if gv, err := schema.ParseGroupVersion(apiVersion); err == nil {
		gvk := gv.WithKind(kind)
		return gvk, nil
	} else {
		return schema.GroupVersionKind{}, err
	}
}
