package main

import (
	"fmt"
	"net/http"

	"github.com/heptiolabs/healthcheck"
	"github.com/tebeka/atexit"
	puccinicommon "github.com/tliron/puccini/common"
	puccinipkg "github.com/tliron/turandot/apis/clientset/versioned"
	"github.com/tliron/turandot/common"
	controllerpkg "github.com/tliron/turandot/controller"
	versionpkg "github.com/tliron/turandot/version"
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

	health := healthcheck.NewHandler()
	log.Info("starting health monitor")
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", healthPort), health)
		puccinicommon.FailOnError(err)
	}()

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

	dynamicClient, err := dynamic.NewForConfig(config)
	puccinicommon.FailOnError(err)

	pucciniClient, err := puccinipkg.NewForConfig(config)
	puccinicommon.FailOnError(err)

	// Controller

	controller := controllerpkg.NewController(
		toolName,
		site,
		cluster,
		namespace,
		dynamicClient,
		kubernetesClient,
		pucciniClient,
		config,
		cachePath,
		resyncPeriod,
		common.SetupSignalHandler(),
	)

	// Run

	err = controller.Run(concurrency)
	puccinicommon.FailOnError(err)
}
