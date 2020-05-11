package common

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/op/go-logging"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

//
// Processor
//

type GetControllerObjectFunc = func(name string, namespace string) (interface{}, error)

type ProcessFunc = func(object interface{}) (bool, error)

type Processor struct {
	Name                string
	GVK                 schema.GroupVersionKind
	Informer            cache.SharedIndexInformer
	Workqueue           workqueue.RateLimitingInterface
	Period              time.Duration
	GetControllerObject GetControllerObjectFunc
	Process             ProcessFunc
	Log                 *logging.Logger
}

func NewProcessor(name string, informer cache.SharedIndexInformer, period time.Duration, get GetControllerObjectFunc, process ProcessFunc) *Processor {
	self := Processor{
		Name:                name,
		Informer:            informer,
		Workqueue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), name),
		Period:              period,
		GetControllerObject: get,
		Process:             process,
		Log:                 logging.MustGetLogger(fmt.Sprintf("turandot.processor.%s", name)),
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: self.EnqueueFor,
		UpdateFunc: func(old interface{}, new interface{}) {
			// TODO: the informer's periodic resync will send "update" events on all resources
			// even if they didn't change
			self.EnqueueFor(new)
		},
	})

	return &self
}

// cache.InformerSynced signature
func (self *Processor) HasSynced() bool {
	return self.Informer.HasSynced()
}

func (self *Processor) Start(concurrency uint, stopChannel <-chan struct{}) {
	var i uint
	for i = 0; i < concurrency; i++ {
		go wait.Until(self.worker, self.Period, stopChannel)
	}
}

func (self *Processor) EnqueueFor(object interface{}) {
	if key, err := cache.MetaNamespaceKeyFunc(object); err == nil {
		self.Workqueue.Add(key)
	} else {
		utilruntime.HandleError(err)
	}
}

func (self *Processor) worker() {
	for self.nextWorkItem() {
	}
}

func (self *Processor) nextWorkItem() bool {
	if item, shutdown := self.Workqueue.Get(); !shutdown {
		defer self.Workqueue.Done(item)
		self.processWorkItem(item)
		return true
	} else {
		return false
	}
}

func (self *Processor) processWorkItem(item interface{}) {
	if key, ok := item.(string); ok {
		if namespace, name, err := cache.SplitMetaNamespaceKey(key); err == nil {
			if object, err := self.GetControllerObject(name, namespace); err == nil {
				if finished, err := self.Process(object); finished {
					utilruntime.HandleError(err)
					self.Log.Infof("finished work item: \"%s %s\"", namespace, name)
					self.Workqueue.Forget(item)
				} else {
					utilruntime.HandleError(err)
					self.Log.Infof("requeuing unfinished work item: \"%s %s\"", namespace, name)
					self.Workqueue.AddRateLimited(key)
				}
			} else if kuberneteserrors.IsNotFound(err) {
				self.Log.Infof("swallowing stale work item: \"%s %s\"", namespace, name)
				self.Workqueue.Forget(item)
			} else {
				utilruntime.HandleError(err)
				self.Log.Infof("requeuing failed work item: \"%s %s\"", namespace, name)
				self.Workqueue.AddRateLimited(key)
			}
		} else {
			utilruntime.HandleError(fmt.Errorf("work item in wrong format: %v", key))
			self.Workqueue.Forget(item)
		}
	} else {
		utilruntime.HandleError(fmt.Errorf("work item not a string: %v", item))
		self.Workqueue.Forget(item)
	}
}

//
// Processors
//

type Processors struct {
	processors         map[schema.GroupVersionKind]*Processor
	controlledGvks     map[schema.GroupVersionKind]bool
	controlledGvksLock sync.Mutex
	log                *logging.Logger
}

func NewProcessors() *Processors {
	return &Processors{
		processors:     make(map[schema.GroupVersionKind]*Processor),
		controlledGvks: make(map[schema.GroupVersionKind]bool),
		log:            logging.MustGetLogger("turandot.processors"),
	}
}

func (self *Processors) Add(gvk schema.GroupVersionKind, processor *Processor) {
	self.processors[gvk] = processor
}

func (self *Processors) Get(name string) (*Processor, bool) {
	for _, processor := range self.processors {
		if processor.Name == name {
			return processor, true
		}
	}
	return nil, false
}

func (self *Processors) Start(concurrency uint, stopChannel <-chan struct{}) {
	for _, processor := range self.processors {
		processor.Start(concurrency, stopChannel)
	}
}

func (self *Processors) ShutDown() {
	for _, processor := range self.processors {
		processor.Workqueue.ShutDown()
	}
}

func (self *Processors) HasSynced() []cache.InformerSynced {
	var hasSynced []cache.InformerSynced
	for _, processor := range self.processors {
		hasSynced = append(hasSynced, processor.HasSynced)
	}
	return hasSynced
}

func (self *Processors) WaitForCacheSync(stopChannel <-chan struct{}) error {
	// This should be called *before* Start()!
	if ok := cache.WaitForCacheSync(stopChannel, self.HasSynced()...); ok {
		return nil
	} else {
		return errors.New("interrupted by shutdown while waiting for informer caches to sync")
	}
}

func (self *Processors) Control(dynamic *Dynamic, controlledGvk schema.GroupVersionKind, stopChannel <-chan struct{}) error {
	self.controlledGvksLock.Lock()
	defer self.controlledGvksLock.Unlock()

	if _, ok := self.controlledGvks[controlledGvk]; !ok {
		// We'll add a change handler once and only once per controlled GVK
		self.controlledGvks[controlledGvk] = true
		return dynamic.AddUnstructuredResourceChangeHandler(controlledGvk, stopChannel, self.onObjectChanged)
	} else {
		return nil
	}
}

// OnChangedFunc signature
func (self *Processors) onObjectChanged(object *unstructured.Unstructured) error {
	if metaObject, err := GetMetaObject(object); err == nil {
		if gvk, name, err := GetControllerOf(metaObject); err == nil {
			if name != "" {
				if processor, ok := self.processors[gvk]; ok {
					if controllerObject, err := processor.GetControllerObject(name, metaObject.GetNamespace()); err == nil {
						processor.EnqueueFor(controllerObject)
					} else {
						// Could happen if controller object was deleted but controlled object was not yet garbage collected
						self.log.Infof("\"%s\" %s controller does not exist for object: %s", name, gvk.Kind, metaObject.GetSelfLink())
					}
				}
			}
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}
