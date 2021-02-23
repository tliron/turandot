package client

import (
	contextpkg "context"
	"fmt"

	certmanagerpkg "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	"github.com/tliron/kutil/logging"
	reposurepkg "github.com/tliron/reposure/apis/clientset/versioned"
	reposureclient "github.com/tliron/reposure/client/admin"
	reposurecontroller "github.com/tliron/reposure/controller"
	turandotpkg "github.com/tliron/turandot/apis/clientset/versioned"
	apiextensionspkg "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubernetespkg "k8s.io/client-go/kubernetes"
	restpkg "k8s.io/client-go/rest"
)

//
// Client
//

type Client struct {
	Kubernetes    kubernetespkg.Interface
	APIExtensions apiextensionspkg.Interface
	Turandot      turandotpkg.Interface
	REST          restpkg.Interface
	CertManager   certmanagerpkg.Interface
	Config        *restpkg.Config
	Reposure      *reposureclient.Client

	ClusterMode       bool
	ClusterRole       string
	Namespace         string
	NamePrefix        string
	PartOf            string
	ManagedBy         string
	OperatorImageName string
	CachePath         string

	Context contextpkg.Context
	Log     logging.Logger
}

func NewClient(kubernetes kubernetespkg.Interface, apiExtensions apiextensionspkg.Interface, turandot turandotpkg.Interface, reposure reposurepkg.Interface, rest restpkg.Interface, config *restpkg.Config, context contextpkg.Context, clusterMode bool, clusterRole string, namespace string, namePrefix string, partOf string, managedBy string, operatorImageName string, cachePath string, logName string) *Client {
	reposure_ := reposureclient.NewClient(
		kubernetes,
		apiExtensions,
		reposure,
		rest,
		config,
		context,
		clusterMode,
		clusterRole,
		namespace,
		reposurecontroller.NamePrefix,
		reposurecontroller.PartOf,
		reposurecontroller.ManagedBy,
		reposurecontroller.OperatorImageReference,
		reposurecontroller.SurrogateImageReference,
		reposurecontroller.SimpleImageReference,
		fmt.Sprintf("%s.reposure", logName),
	)

	return &Client{
		Kubernetes:        kubernetes,
		APIExtensions:     apiExtensions,
		Turandot:          turandot,
		REST:              rest,
		Config:            config,
		Reposure:          reposure_,
		ClusterMode:       clusterMode,
		ClusterRole:       clusterRole,
		Namespace:         namespace,
		NamePrefix:        namePrefix,
		PartOf:            partOf,
		ManagedBy:         managedBy,
		OperatorImageName: operatorImageName,
		CachePath:         cachePath,
		Context:           context,
		Log:               logging.GetLogger(logName),
	}
}
