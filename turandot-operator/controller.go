package main

import (
	"fmt"
	"net/http"

	"github.com/heptiolabs/healthcheck"
	"github.com/tebeka/atexit"
	puccinicommon "github.com/tliron/puccini/common"
	turandotpkg "github.com/tliron/turandot/apis/clientset/versioned"
	"github.com/tliron/turandot/common"
	controllerpkg "github.com/tliron/turandot/controller"
	versionpkg "github.com/tliron/turandot/version"
	apiextensionspkg "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	kubernetespkg "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	// Load all auth plugins:
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func Controller() {
	if version {
		versionpkg.Print()
		atexit.Exit(0)
		return
	}

	log.Infof("%s version=%s revision=%s site=%s", toolName, versionpkg.GitVersion, versionpkg.GitRevision, site)

	// Config

	config, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
	puccinicommon.FailOnError(err)

	if cluster {
		namespace = ""
	} else if namespace == "" {
		if namespace_, ok := common.GetConfiguredNamespace(kubeconfigPath); ok {
			namespace = namespace_
		}
		if namespace == "" {
			namespace = common.GetServiceAccountNamespace()
		}
		if namespace == "" {
			log.Fatal("could not discover namespace and namespace not provided")
		}
	}

	// Clients

	kubernetesClient, err := kubernetespkg.NewForConfig(config)
	puccinicommon.FailOnError(err)

	apiExtensionsClient, err := apiextensionspkg.NewForConfig(config)
	puccinicommon.FailOnError(err)

	dynamicClient, err := dynamic.NewForConfig(config)
	puccinicommon.FailOnError(err)

	turandotClient, err := turandotpkg.NewForConfig(config)
	puccinicommon.FailOnError(err)

	// Controller

	controller := controllerpkg.NewController(
		toolName,
		site,
		cluster,
		namespace,
		dynamicClient,
		kubernetesClient,
		apiExtensionsClient,
		turandotClient,
		config,
		cachePath,
		resyncPeriod,
		common.SetupSignalHandler(),
	)

	// Run

	err = controller.Run(concurrency, func() {
		log.Info("starting health monitor")
		health := healthcheck.NewHandler()
		err := http.ListenAndServe(fmt.Sprintf(":%d", healthPort), health)
		puccinicommon.FailOnError(err)
	})
	puccinicommon.FailOnError(err)
}
