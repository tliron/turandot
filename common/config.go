package common

import (
	"io/ioutil"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// See: clientcmd.BuildConfigFromFlags

func NewConfig(configPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: configPath,
		},
		&clientcmd.ConfigOverrides{},
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

func GetConfiguredNamespace(configPath string) (string, bool) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: configPath,
		},
		&clientcmd.ConfigOverrides{},
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
