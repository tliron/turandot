package commands

import (
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
	Config     *restpkg.Config
	Kubernetes kubernetespkg.Interface
	REST       restpkg.Interface
	Namespace  string
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
		Config:     config,
		Kubernetes: kubernetes,
		REST:       kubernetes.CoreV1().RESTClient(),
		Namespace:  namespace_,
	}
}

func (self *Client) Turandot() *clientpkg.Client {
	apiExtensions, err := apiextensionspkg.NewForConfig(self.Config)
	util.FailOnError(err)

	turandot, err := turandotpkg.NewForConfig(self.Config)
	util.FailOnError(err)

	return clientpkg.NewClient(
		self.Kubernetes,
		apiExtensions,
		turandot,
		self.REST,
		self.Config,
		cluster,
		self.Namespace,
		controller.NamePrefix,
		controller.PartOf,
		controller.ManagedBy,
		controller.OperatorImageName,
		controller.RepositoryImageName,
		controller.RepositorySpoolerImageName,
		controller.CacheDirectory,
	)
}
