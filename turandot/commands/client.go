package commands

import (
	spoolerpkg "github.com/tliron/kubernetes-registry-spooler/client"
	kubernetesutil "github.com/tliron/kutil/kubernetes"
	"github.com/tliron/kutil/util"
	turandotpkg "github.com/tliron/turandot/apis/clientset/versioned"
	"github.com/tliron/turandot/controller"
	clientpkg "github.com/tliron/turandot/turandot/client"
	apiextensionspkg "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubernetespkg "k8s.io/client-go/kubernetes"
	restpkg "k8s.io/client-go/rest"
)

//
// Client
//

type Client struct {
	config     *restpkg.Config
	kubernetes kubernetespkg.Interface
	rest       restpkg.Interface
	namespace  string
}

func NewClient() *Client {
	config, err := kubernetesutil.NewConfigFromFlags(masterUrl, kubeconfigPath, context, log)
	util.FailOnError(err)

	kubernetes, err := kubernetespkg.NewForConfig(config)
	util.FailOnError(err)

	namespace_ := namespace
	if cluster {
		namespace_ = ""
	} else if namespace_ == "" {
		if namespace__, ok := kubernetesutil.GetConfiguredNamespace(kubeconfigPath, context); ok {
			namespace_ = namespace__
		}
		if namespace_ == "" {
			util.Fail("could not discover namespace and \"--namespace\" not provided")
		}
	}

	return &Client{
		config:     config,
		kubernetes: kubernetes,
		rest:       kubernetes.CoreV1().RESTClient(),
		namespace:  namespace_,
	}
}

func (self *Client) Turandot() *clientpkg.Client {
	apiExtensions, err := apiextensionspkg.NewForConfig(self.config)
	util.FailOnError(err)

	turandot, err := turandotpkg.NewForConfig(self.config)
	util.FailOnError(err)

	return clientpkg.NewClient(
		self.kubernetes,
		apiExtensions,
		turandot,
		self.rest,
		self.config,
		cluster,
		self.namespace,
		controller.NamePrefix,
		controller.PartOf,
		controller.ManagedBy,
		controller.OperatorImageName,
		controller.InventoryImageName,
		controller.InventorySpoolerImageName,
		controller.CacheDirectory,
		controller.SpoolDirectory,
	)
}

func (self *Client) Spooler() *spoolerpkg.Client {
	return spoolerpkg.NewClient(
		self.kubernetes,
		self.rest,
		self.config,
		self.namespace,
		controller.SpoolerAppName,
		controller.SpoolerContainerName,
		controller.SpoolDirectory,
	)
}
