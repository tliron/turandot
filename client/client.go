package client

import (
	contextpkg "context"

	turandotpkg "github.com/tliron/turandot/apis/clientset/versioned"
	apiextensionspkg "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubernetespkg "k8s.io/client-go/kubernetes"
	restpkg "k8s.io/client-go/rest"
)

type Client struct {
	kubernetes                *kubernetespkg.Clientset
	apiExtensions             *apiextensionspkg.Clientset
	turandot                  *turandotpkg.Clientset
	rest                      restpkg.Interface
	config                    *restpkg.Config
	cluster                   bool
	namespace                 string
	namePrefix                string
	partOf                    string
	managedBy                 string
	operatorImageName         string
	inventoryImageName        string
	inventorySpoolerImageName string
	cachePath                 string
	context                   contextpkg.Context
}

func NewClient(kubernetes *kubernetespkg.Clientset, apiExtensions *apiextensionspkg.Clientset, turandot *turandotpkg.Clientset, rest restpkg.Interface, config *restpkg.Config, cluster bool, namespace string, namePrefix string, partOf string, managedBy string, operatorImageName string, inventoryImageName string, inventorySpoolerImageName string, cachePath string) *Client {
	return &Client{
		kubernetes:                kubernetes,
		apiExtensions:             apiExtensions,
		turandot:                  turandot,
		rest:                      rest,
		config:                    config,
		cluster:                   cluster,
		namespace:                 namespace,
		namePrefix:                namePrefix,
		partOf:                    partOf,
		managedBy:                 managedBy,
		operatorImageName:         operatorImageName,
		inventoryImageName:        inventoryImageName,
		inventorySpoolerImageName: inventorySpoolerImageName,
		cachePath:                 cachePath,
		context:                   contextpkg.TODO(),
	}
}
