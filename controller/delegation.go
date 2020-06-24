package controller

import (
	"fmt"
	"path/filepath"

	spoolerpkg "github.com/tliron/kubernetes-registry-spooler/client"
	turandotpkg "github.com/tliron/turandot/apis/clientset/versioned"
	clientpkg "github.com/tliron/turandot/client"
	"github.com/tliron/turandot/common"
	apiextensionspkg "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubernetespkg "k8s.io/client-go/kubernetes"
)

func (self *Controller) NewDelegate(name string) (*clientpkg.Client, *spoolerpkg.Client, error) {
	configPath := filepath.Join(self.CachePath, "delegates", fmt.Sprintf("%s.yaml", name))
	if config, err := common.NewConfig(configPath, ""); err == nil {
		namespace, _ := common.GetConfiguredNamespace(configPath, "")

		var kubernetes *kubernetespkg.Clientset
		kubernetes, err := kubernetespkg.NewForConfig(config)
		if err != nil {
			return nil, nil, err
		}

		var apiExtensions *apiextensionspkg.Clientset
		apiExtensions, err = apiextensionspkg.NewForConfig(config)
		if err != nil {
			return nil, nil, err
		}

		var turandot *turandotpkg.Clientset
		turandot, err = turandotpkg.NewForConfig(config)
		if err != nil {
			return nil, nil, err
		}

		rest := kubernetes.CoreV1().RESTClient()

		return clientpkg.NewClient(
				fmt.Sprintf("turandot.client.%s", name),
				kubernetes,
				apiExtensions,
				turandot,
				rest,
				config,
				false,
				namespace,
				NamePrefix,
				PartOf,
				ManagedBy,
				OperatorImageName,
				InventoryImageName,
				InventorySpoolerImageName,
				CacheDirectory,
				SpoolDirectory,
			),
			spoolerpkg.NewClient(
				kubernetes,
				rest,
				config,
				namespace,
				SpoolerAppName,
				SpoolerContainerName,
				SpoolDirectory,
			),
			nil
	} else {
		return nil, nil, err
	}
}
