package delegate

import (
	contextpkg "context"

	"github.com/op/go-logging"
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
	Config        *restpkg.Config

	Cluster                   bool
	Namespace                 string
	NamePrefix                string
	PartOf                    string
	ManagedBy                 string
	OperatorImageName         string
	InventoryImageName        string
	InventorySpoolerImageName string
	CachePath                 string
	SpoolPath                 string

	Context contextpkg.Context
	Log     *logging.Logger
}

func NewClient(loggerName string, kubernetes kubernetespkg.Interface, apiExtensions apiextensionspkg.Interface, turandot turandotpkg.Interface, rest restpkg.Interface, config *restpkg.Config, cluster bool, namespace string, namePrefix string, partOf string, managedBy string, operatorImageName string, inventoryImageName string, inventorySpoolerImageName string, cachePath string, spoolPath string) *Client {
	return &Client{
		Kubernetes:                kubernetes,
		APIExtensions:             apiExtensions,
		Turandot:                  turandot,
		REST:                      rest,
		Config:                    config,
		Cluster:                   cluster,
		Namespace:                 namespace,
		NamePrefix:                namePrefix,
		PartOf:                    partOf,
		ManagedBy:                 managedBy,
		OperatorImageName:         operatorImageName,
		InventoryImageName:        inventoryImageName,
		InventorySpoolerImageName: inventorySpoolerImageName,
		CachePath:                 cachePath,
		SpoolPath:                 spoolPath,
		Context:                   contextpkg.TODO(),
		Log:                       logging.MustGetLogger(loggerName),
	}
}
