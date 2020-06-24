package common

import (
	"io/ioutil"

	"github.com/op/go-logging"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// See: clientcmd.BuildConfigFromFlags
func NewConfigFromFlags(masterUrl string, configPath string, context string, log *logging.Logger) (*rest.Config, error) {
	if configPath == "" && masterUrl == "" {
		if config, err := rest.InClusterConfig(); err == nil {
			return config, nil
		} else {
			log.Warningf("could not create InClusterConfig: %s", err)
		}
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: configPath,
		},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
			ClusterInfo: api.Cluster{
				Server: masterUrl,
			},
		},
	).ClientConfig()
}

func NewConfig(configPath string, context string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: configPath,
		},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		},
	).ClientConfig()
}

func NewConfigForContext(configPath string, context string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: configPath,
		},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		},
	).ClientConfig()
}

func GetConfiguredNamespace(configPath string, context string) (string, bool) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: configPath,
		},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		},
	)
	namespace, _, _ := clientConfig.Namespace()
	return namespace, namespace != ""
}

func NewSelfContainedConfig(restConfig *rest.Config, namespace string) (*api.Config, error) {
	var err error

	caData := restConfig.CAData
	if (caData == nil) && (restConfig.CAFile != "") {
		if caData, err = ioutil.ReadFile(restConfig.CAFile); err != nil {
			return nil, err
		}
	}

	ccData := restConfig.CertData
	if (ccData == nil) && (restConfig.CertFile != "") {
		if ccData, err = ioutil.ReadFile(restConfig.CertFile); err != nil {
			return nil, err
		}
	}

	ckData := restConfig.KeyData
	if (ckData == nil) && (restConfig.KeyFile != "") {
		if ckData, err = ioutil.ReadFile(restConfig.KeyFile); err != nil {
			return nil, err
		}
	}

	config := api.NewConfig()
	config.CurrentContext = "default"
	config.Contexts["default"] = api.NewContext()
	config.Contexts["default"].Cluster = "default"
	config.Contexts["default"].AuthInfo = "default"
	config.Contexts["default"].Namespace = namespace
	config.Clusters["default"] = api.NewCluster()
	config.Clusters["default"].Server = restConfig.Host
	config.Clusters["default"].CertificateAuthorityData = caData
	config.AuthInfos["default"] = api.NewAuthInfo()
	config.AuthInfos["default"].ClientCertificateData = ccData
	config.AuthInfos["default"].ClientKeyData = ckData

	return config, nil
}
