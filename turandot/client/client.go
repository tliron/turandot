package client

import (
	reposurepkg "github.com/tliron/reposure/apis/clientset/versioned"
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

func NewClient(kubernetes kubernetespkg.Interface, apiExtensions apiextensionspkg.Interface, turandot turandotpkg.Interface, reposure reposurepkg.Interface, rest restpkg.Interface, config *restpkg.Config, clusterMode bool, clusterRole string, namespace string, namePrefix string, partOf string, managedBy string, operatorImageName string, cachePath string) *Client {
	return &Client{
		Client: clientpkg.NewClient(
			"turandot.client",
			kubernetes,
			apiExtensions,
			turandot,
			reposure,
			rest,
			config,
			clusterMode,
			clusterRole,
			namespace,
			namePrefix,
			partOf,
			managedBy,
			operatorImageName,
			cachePath,
		),
	}
}
