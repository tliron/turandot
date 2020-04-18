package controller

import (
	contextpkg "context"
	"fmt"
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
	site string

	dynamic     *common.Dynamic
	kubernetes  kubernetes.Interface
	turandot    turandotclientset.Interface
	config      *restpkg.Config
	cachePath   string
	stopChannel <-chan struct{}

	processors        *common.Processors
	instantiationWork chan Instantiation
	events            record.EventRecorder

	kubernetesInformerFactory informers.SharedInformerFactory
	turandotInformerFactory   turandotinformers.SharedInformerFactory

	services turandotlisters.ServiceLister

	context contextpkg.Context
	log     *logging.Logger
}

func NewController(toolName string, site string, cluster bool, namespace string, dynamic dynamicpkg.Interface, kubernetes kubernetes.Interface, turandot turandotclientset.Interface, config *restpkg.Config, cachePath string, informerResyncPeriod time.Duration, stopChannel <-chan struct{}) *Controller {
	context := contextpkg.TODO()

	if cluster {
		namespace = ""
	}

	self := Controller{
		site:              site,
		config:            config,
		dynamic:           common.NewDynamic(dynamic, kubernetes.Discovery(), namespace, context),
		kubernetes:        kubernetes,
		turandot:          turandot,
		cachePath:         cachePath,
		processors:        common.NewProcessors(),
		instantiationWork: make(chan Instantiation, 10),
		events:            common.CreateEventRecorder(kubernetes, toolName),
		context:           context,
		log:               logging.MustGetLogger(fmt.Sprintf("turandot.controller.%s", toolName)),
	}

	if cluster {
		self.kubernetesInformerFactory = informers.NewSharedInformerFactory(kubernetes, informerResyncPeriod)
		self.turandotInformerFactory = turandotinformers.NewSharedInformerFactory(turandot, informerResyncPeriod)
	} else {
		self.kubernetesInformerFactory = informers.NewSharedInformerFactoryWithOptions(kubernetes, informerResyncPeriod, informers.WithNamespace(namespace))
		self.turandotInformerFactory = turandotinformers.NewSharedInformerFactoryWithOptions(turandot, informerResyncPeriod, turandotinformers.WithNamespace(namespace))
	}

	// Informers
	serviceInformer := self.turandotInformerFactory.Turandot().V1alpha1().Services()

	// Listers
	self.services = serviceInformer.Lister()

	// Processors

	processorPeriod := 5 * time.Second

	self.processors.Add(turandotresources.ServiceGVK, common.NewProcessor(
		"services",
		serviceInformer.Informer(),
		processorPeriod,
		func(name string, namespace string) (interface{}, error) {
			return self.getService(name, namespace)
		},
		func(object interface{}) (bool, error) {
			return self.processService(object.(*turandotresources.Service))
		},
	))

	return &self
}

func (self *Controller) Run(threadiness uint) error {
	defer utilruntime.HandleCrash()

	self.log.Info("starting informer factories")
	self.kubernetesInformerFactory.Start(self.stopChannel)
	self.turandotInformerFactory.Start(self.stopChannel)

	self.log.Info("waiting for processor informer caches to sync")
	utilruntime.HandleError(self.processors.WaitForCacheSync(self.stopChannel))

	self.log.Info("starting processors")
	self.processors.Start(threadiness, self.stopChannel)
	defer self.processors.ShutDown()

	self.log.Info("starting instantiator")
	go self.runInstantiator()
	defer self.stopInstantiator()

	<-self.stopChannel

	self.log.Info("shutting down")

	return nil
}
