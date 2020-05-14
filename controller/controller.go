package controller

import (
	contextpkg "context"
	"time"

	"github.com/op/go-logging"
	turandotclientset "github.com/tliron/turandot/apis/clientset/versioned"
	turandotinformers "github.com/tliron/turandot/apis/informers/externalversions"
	turandotlisters "github.com/tliron/turandot/apis/listers/turandot.puccini.cloud/v1alpha1"
	"github.com/tliron/turandot/common"
	turandotresources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	dynamicpkg "k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	restpkg "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
)

//
// Controller
//

type Controller struct {
	Site string

	Dynamic     *common.Dynamic
	Kubernetes  kubernetes.Interface
	Turandot    turandotclientset.Interface
	Config      *restpkg.Config
	CachePath   string
	StopChannel <-chan struct{}

	Processors *common.Processors
	Events     record.EventRecorder

	KubernetesInformerFactory informers.SharedInformerFactory
	TurandotInformerFactory   turandotinformers.SharedInformerFactory

	Services turandotlisters.ServiceLister

	Context contextpkg.Context
	Log     *logging.Logger
}

func NewController(toolName string, site string, cluster bool, namespace string, dynamic dynamicpkg.Interface, kubernetes kubernetes.Interface, turandot turandotclientset.Interface, config *restpkg.Config, cachePath string, informerResyncPeriod time.Duration, stopChannel <-chan struct{}) *Controller {
	context := contextpkg.TODO()

	if cluster {
		namespace = ""
	}

	log := logging.MustGetLogger("turandot.controller")

	self := Controller{
		Site:       site,
		Config:     config,
		Dynamic:    common.NewDynamic(dynamic, kubernetes.Discovery(), namespace, context),
		Kubernetes: kubernetes,
		Turandot:   turandot,
		CachePath:  cachePath,
		Processors: common.NewProcessors(),
		Events:     common.CreateEventRecorder(kubernetes, "Turandot", log),
		Context:    context,
		Log:        log,
	}

	if cluster {
		self.KubernetesInformerFactory = informers.NewSharedInformerFactory(kubernetes, informerResyncPeriod)
		self.TurandotInformerFactory = turandotinformers.NewSharedInformerFactory(turandot, informerResyncPeriod)
	} else {
		self.KubernetesInformerFactory = informers.NewSharedInformerFactoryWithOptions(kubernetes, informerResyncPeriod, informers.WithNamespace(namespace))
		self.TurandotInformerFactory = turandotinformers.NewSharedInformerFactoryWithOptions(turandot, informerResyncPeriod, turandotinformers.WithNamespace(namespace))
	}

	// Informers
	serviceInformer := self.TurandotInformerFactory.Turandot().V1alpha1().Services()

	// Listers
	self.Services = serviceInformer.Lister()

	// Processors

	processorPeriod := 5 * time.Second

	self.Processors.Add(turandotresources.ServiceGVK, common.NewProcessor(
		"services",
		serviceInformer.Informer(),
		processorPeriod,
		func(name string, namespace string) (interface{}, error) {
			return self.GetService(name, namespace)
		},
		func(object interface{}) (bool, error) {
			return self.processService(object.(*turandotresources.Service))
		},
	))

	return &self
}

func (self *Controller) Run(concurrency uint) error {
	defer utilruntime.HandleCrash()

	self.Log.Info("starting informer factories")
	self.KubernetesInformerFactory.Start(self.StopChannel)
	self.TurandotInformerFactory.Start(self.StopChannel)

	self.Log.Info("waiting for processor informer caches to sync")
	utilruntime.HandleError(self.Processors.WaitForCacheSync(self.StopChannel))

	self.Log.Infof("starting processors (concurrency=%d)", concurrency)
	self.Processors.Start(concurrency, self.StopChannel)
	defer self.Processors.ShutDown()

	<-self.StopChannel

	self.Log.Info("shutting down")

	return nil
}
