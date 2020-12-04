package controller

import (
	"fmt"
	"path/filepath"

	kubernetesutil "github.com/tliron/kutil/kubernetes"
	reposurepkg "github.com/tliron/reposure/apis/clientset/versioned"
	turandotpkg "github.com/tliron/turandot/apis/clientset/versioned"
	clientpkg "github.com/tliron/turandot/client"
	apiextensionspkg "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubernetespkg "k8s.io/client-go/kubernetes"
)

func (self *Controller) NewDelegate(name string) (*clientpkg.Client, error) {
	configPath := filepath.Join(self.CachePath, "delegates", fmt.Sprintf("%s.yaml", name))
	if config, err := kubernetesutil.NewConfig(configPath, ""); err == nil {
		namespace, _ := kubernetesutil.GetConfiguredNamespace(configPath, "")

		var kubernetes *kubernetespkg.Clientset
		kubernetes, err := kubernetespkg.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		var apiExtensions *apiextensionspkg.Clientset
		apiExtensions, err = apiextensionspkg.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		var turandot *turandotpkg.Clientset
		turandot, err = turandotpkg.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		var reposure *reposurepkg.Clientset
		reposure, err = reposurepkg.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		rest := kubernetes.CoreV1().RESTClient()

		return clientpkg.NewClient(
			fmt.Sprintf("turandot.client.%s", name),
			kubernetes,
			apiExtensions,
			turandot,
			reposure,
			rest,
			config,
			false,
			"",
			namespace,
			NamePrefix,
			PartOf,
			ManagedBy,
			OperatorImageName,
			CacheDirectory,
		), nil
	} else {
		return nil, err
	}
}
