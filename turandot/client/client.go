package client

import (
	turandotpkg "github.com/tliron/turandot/apis/clientset/versioned"
	clientpkg "github.com/tliron/turandot/client"
	apiextensionspkg "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubernetespkg "k8s.io/client-go/kubernetes"
	restpkg "k8s.io/client-go/rest"
)

//
// Client
//

type Client struct {
	*clientpkg.Client
}

func NewClient(kubernetes kubernetespkg.Interface, apiExtensions apiextensionspkg.Interface, turandot turandotpkg.Interface, rest restpkg.Interface, config *restpkg.Config, cluster bool, namespace string, namePrefix string, partOf string, managedBy string, operatorImageName string, inventoryImageName string, inventorySpoolerImageName string, cachePath string, spoolPath string) *Client {
	return &Client{
		Client: clientpkg.NewClient(
			"turandot.client",
			kubernetes,
			apiExtensions,
			turandot,
			rest,
			config,
			cluster,
			namespace,
			namePrefix,
			partOf,
			managedBy,
			operatorImageName,
			inventoryImageName,
			inventorySpoolerImageName,
			cachePath,
			spoolPath,
		),
	}
}
