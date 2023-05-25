package controller

import (
	contextpkg "context"
	"fmt"
	"time"

	"github.com/tliron/commonlog"
	kubernetesutil "github.com/tliron/kutil/kubernetes"
	reposurepkg "github.com/tliron/reposure/apis/clientset/versioned"
	turandotpkg "github.com/tliron/turandot/apis/clientset/versioned"
	turandotinformers "github.com/tliron/turandot/apis/informers/externalversions"
	turandotlisters "github.com/tliron/turandot/apis/listers/turandot.puccini.cloud/v1alpha1"
	clientpkg "github.com/tliron/turandot/client"
	turandotresources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	apiextensionspkg "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
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

	Config      *restpkg.Config
	Dynamic     *kubernetesutil.Dynamic
	Kubernetes  kubernetes.Interface
	Turandot    turandotpkg.Interface
	Client      *clientpkg.Client
	CachePath   string
	StopChannel <-chan struct{}

	Processors *kubernetesutil.Processors
	Events     record.EventRecorder

	KubernetesInformerFactory informers.SharedInformerFactory
	TurandotInformerFactory   turandotinformers.SharedInformerFactory

	Services turandotlisters.ServiceLister

	Context contextpkg.Context
	Log     commonlog.Logger
}

func NewController(context contextpkg.Context, toolName string, site string, clusterMode bool, clusterRole string, namespace string, dynamic dynamicpkg.Interface, kubernetes kubernetes.Interface, apiExtensions apiextensionspkg.Interface, turandot turandotpkg.Interface, reposure reposurepkg.Interface, config *restpkg.Config, cachePath string, informerResyncPeriod time.Duration, stopChannel <-chan struct{}) *Controller {
	if clusterMode {
		namespace = ""
	}

	log := commonlog.GetLoggerf("%s.controller", toolName)

	self := Controller{
		Site:        site,
		Config:      config,
		Dynamic:     kubernetesutil.NewDynamic(toolName, dynamic, kubernetes.Discovery(), namespace, context),
		Kubernetes:  kubernetes,
		Turandot:    turandot,
		CachePath:   cachePath,
		StopChannel: stopChannel,
		Processors:  kubernetesutil.NewProcessors(toolName),
		Events:      kubernetesutil.CreateEventRecorder(kubernetes, "Turandot", log),
		Context:     context,
		Log:         log,
	}

	self.Client = clientpkg.NewClient(
		kubernetes,
		apiExtensions,
		turandot,
		reposure,
		kubernetes.CoreV1().RESTClient(),
		config,
		context,
		clusterMode,
		clusterRole,
		namespace,
		NamePrefix,
		PartOf,
		ManagedBy,
		OperatorImageName,
		CacheDirectory,
		fmt.Sprintf("%s.client.%s", toolName, site),
	)

	if clusterMode {
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

	self.Processors.Add(turandotresources.ServiceGVK, kubernetesutil.NewProcessor(
		toolName,
		"services",
		serviceInformer.Informer(),
		processorPeriod,
		func(name string, namespace string) (any, error) {
			return self.Client.GetService(namespace, name)
		},
		func(object any) (bool, error) {
			return self.processService(context, object.(*turandotresources.Service))
		},
	))

	return &self
}

func (self *Controller) Run(concurrency uint, startup func()) error {
	defer utilruntime.HandleCrash()

	self.Log.Info("starting informer factories")
	self.KubernetesInformerFactory.Start(self.StopChannel)
	self.TurandotInformerFactory.Start(self.StopChannel)

	self.Log.Info("waiting for processor informer caches to sync")
	utilruntime.HandleError(self.Processors.WaitForCacheSync(self.StopChannel))

	self.Log.Infof("starting processors (concurrency=%d)", concurrency)
	self.Processors.Start(concurrency, self.StopChannel)
	defer self.Processors.ShutDown()

	if startup != nil {
		go startup()
	}

	<-self.StopChannel

	self.Log.Info("shutting down")

	return nil
}
