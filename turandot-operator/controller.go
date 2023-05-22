package main

import (
	"fmt"
	"net/http"

	"github.com/heptiolabs/healthcheck"
	"github.com/tliron/kutil/kubernetes"
	"github.com/tliron/kutil/util"
	versionpkg "github.com/tliron/kutil/version"
	reposurepkg "github.com/tliron/reposure/apis/clientset/versioned"
	turandotpkg "github.com/tliron/turandot/apis/clientset/versioned"
	controllerpkg "github.com/tliron/turandot/controller"
	apiextensionspkg "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	kubernetespkg "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func Controller() {
	if version {
		versionpkg.Print()
		util.Exit(0)
		return
	}

	log.Noticef("%s version=%s revision=%s site=%s", toolName, versionpkg.GitVersion, versionpkg.GitRevision, site)

	// Config

	config, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
	util.FailOnError(err)

	if clusterMode {
		namespace = ""
	} else if namespace == "" {
		if namespace_, ok := kubernetes.GetConfiguredNamespace(kubeconfigPath, kubeconfigContext); ok {
			namespace = namespace_
		}
		if namespace == "" {
			namespace = kubernetes.GetServiceAccountNamespace()
		}
		if namespace == "" {
			util.Fail("could not discover namespace and namespace not provided")
		}
	}

	// Clients

	dynamicClient, err := dynamic.NewForConfig(config)
	util.FailOnError(err)

	kubernetesClient, err := kubernetespkg.NewForConfig(config)
	util.FailOnError(err)

	apiExtensionsClient, err := apiextensionspkg.NewForConfig(config)
	util.FailOnError(err)

	turandotClient, err := turandotpkg.NewForConfig(config)
	util.FailOnError(err)

	reposureClient, err := reposurepkg.NewForConfig(config)
	util.FailOnError(err)

	// Controller

	controller := controllerpkg.NewController(
		context,
		toolName,
		site,
		clusterMode,
		clusterRole,
		namespace,
		dynamicClient,
		kubernetesClient,
		apiExtensionsClient,
		turandotClient,
		reposureClient,
		config,
		cachePath,
		resyncPeriod,
		util.SetupSignalHandler(),
	)

	// Run

	err = controller.Run(concurrency, func() {
		log.Info("starting health monitor")
		health := healthcheck.NewHandler()
		err := http.ListenAndServe(fmt.Sprintf(":%d", healthPort), health)
		util.FailOnError(err)
	})
	util.FailOnError(err)
}
