package common

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/op/go-logging"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	dynamicpkg "k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

//
// Dynamic
//

type OnChangedFunc = func(object *unstructured.Unstructured) error

type Dynamic struct {
	Dynamic         dynamicpkg.Interface
	Discovery       discovery.DiscoveryInterface
	InformerFactory dynamicinformer.DynamicSharedInformerFactory
	Log             *logging.Logger

	informers     map[schema.GroupVersionResource]cache.SharedIndexInformer
	informersLock sync.Mutex
	context       context.Context
}

func NewDynamic(dynamic dynamicpkg.Interface, discovery discovery.DiscoveryInterface, namespace string, context context.Context) *Dynamic {
	var informerFactory dynamicinformer.DynamicSharedInformerFactory
	if namespace == "" {
		informerFactory = dynamicinformer.NewDynamicSharedInformerFactory(dynamic, time.Second)
	} else {
		informerFactory = dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamic, time.Second, namespace, nil)
	}

	return &Dynamic{
		Dynamic:         dynamic,
		Discovery:       discovery,
		InformerFactory: informerFactory,
		Log:             logging.MustGetLogger(fmt.Sprintf("dynamic.%s", namespace)),
		informers:       make(map[schema.GroupVersionResource]cache.SharedIndexInformer),
		context:         context,
	}
}

func (self *Dynamic) GetResource(gvk schema.GroupVersionKind, name string, namespace string) (*unstructured.Unstructured, error) {
	if gvr, err := FindResourceForKind(self.Discovery, gvk, "get", "create"); err == nil {
		return self.Dynamic.Resource(gvr).Namespace(namespace).Get(self.context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Dynamic) CreateResource(object *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	if gvr, err := FindResourceForUnstructured(self.Discovery, object, "create"); err == nil {
		if object, err = self.Dynamic.Resource(gvr).Namespace(object.GetNamespace()).Create(self.context, object, meta.CreateOptions{}); err == nil {
			return object, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

// controllerObject must also support schema.ObjectKind interface
func (self *Dynamic) CreateControlledResource(object *unstructured.Unstructured, controllerObject meta.Object, processors *Processors, stopChannel <-chan struct{}) (*unstructured.Unstructured, error) {
	if _, ok := controllerObject.(schema.ObjectKind); !ok {
		return nil, fmt.Errorf("controller object does not support schema.ObjectKind interface: %v", controllerObject)
	}

	if gvk, err := GetUnstructuredGVK(object); err == nil {
		if err := processors.Control(self, gvk, stopChannel); err == nil {
			if object.GetNamespace() == "" {
				// If namespace not specified then create in same namespace as controller object
				object.SetNamespace(controllerObject.GetNamespace())
			}

			if err := SetControllerOfUnstructured(object, controllerObject); err == nil {
				return self.CreateResource(object)
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Dynamic) UpdateResource(object *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	if gvr, err := FindResourceForUnstructured(self.Discovery, object, "update", "create"); err == nil {
		if object, err = self.Dynamic.Resource(gvr).Namespace(object.GetNamespace()).Update(self.context, object, meta.UpdateOptions{}); err == nil {
			return object, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Dynamic) GetInformers(gvk schema.GroupVersionKind) ([]cache.SharedInformer, []cache.SharedInformer, error) {
	// We can only get informers for resources that support the "watch" verb
	if gvrs, err := FindResourcesForKind(self.Discovery, gvk, "watch", "create"); err == nil {
		var informers []cache.SharedInformer
		var newInformers []cache.SharedInformer

		self.informersLock.Lock()
		defer self.informersLock.Unlock()

		for _, gvr := range gvrs {
			if informer, ok := self.informers[gvr]; ok {
				informers = append(informers, informer)
			} else {
				informer = self.InformerFactory.ForResource(gvr).Informer()
				self.informers[gvr] = informer
				informers = append(informers, informer)
				newInformers = append(newInformers, informer)
			}
		}

		return informers, newInformers, nil
	} else {
		return nil, nil, err
	}
}

func (self *Dynamic) AddResourceEventHandler(gvk schema.GroupVersionKind, stopChannel <-chan struct{}, handler cache.ResourceEventHandler) error {
	if informers, newInformers, err := self.GetInformers(gvk); err == nil {
		if len(informers) > 0 {
			// Will only start informers that have not yet been started
			self.InformerFactory.Start(stopChannel)

			// Event handlers should be added *before* syncing informer cache
			for _, informer := range informers {
				informer.AddEventHandler(handler)
			}

			// Informers should be synced before using them for the first time
			if len(newInformers) > 0 {
				var hasSynced []cache.InformerSynced
				for _, informer := range newInformers {
					hasSynced = append(hasSynced, informer.HasSynced)
				}

				self.Log.Infof("waiting for dynamic informer caches to sync for \"%s\"", gvk.String())
				if ok := cache.WaitForCacheSync(stopChannel, hasSynced...); !ok {
					return errors.New("interrupted by shutdown while waiting for informer caches to sync")
				}
				self.Log.Infof("dynamic informer caches synced for \"%s\"", gvk.String())
			}

			return nil
		} else {
			return fmt.Errorf("informers not found for: \"%s\"", gvk.String())
		}
	} else {
		return err
	}
}

func (self *Dynamic) AddUnstructuredResourceEventHandlerFuncs(gvk schema.GroupVersionKind, stopChannel <-chan struct{}, onAdded OnAddedFunc, onUpdated OnUpdatedFunc, onDeleted OnDeletedFunc) error {
	return self.AddResourceEventHandler(gvk, stopChannel, NewUnstructuredResourceEventHandler(onAdded, onUpdated, onDeleted))
}

func (self *Dynamic) AddUnstructuredResourceChangeHandler(gvk schema.GroupVersionKind, stopChannel <-chan struct{}, onChanged OnChangedFunc) error {
	return self.AddUnstructuredResourceEventHandlerFuncs(gvk, stopChannel,
		onChanged,
		func(old *unstructured.Unstructured, new *unstructured.Unstructured) error {
			// Note: the informer's periodic resync will send "update" events on all resources
			// So we'll process only resources that have changed
			if new.GetResourceVersion() != old.GetResourceVersion() {
				return onChanged(new)
			}
			return nil
		},
		onChanged, // TODO: really, deletion?
	)
}
