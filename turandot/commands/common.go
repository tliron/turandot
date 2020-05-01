package commands

import (
	"io"
	"os"

	"github.com/op/go-logging"
	spoolerpkg "github.com/tliron/kubernetes-registry-spooler/client"
	puccinicommon "github.com/tliron/puccini/common"
	"github.com/tliron/puccini/common/terminal"
	turandotpkg "github.com/tliron/turandot/apis/clientset/versioned"
	clientpkg "github.com/tliron/turandot/client"
	"github.com/tliron/turandot/common"
	apiextensionspkg "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubernetespkg "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const toolName = "turandot"

var log = logging.MustGetLogger(toolName)

var filePath string
var directoryPath string
var url string
var component string
var tail int
var follow bool
var bare bool
var all bool

type Client struct {
	config     *rest.Config
	kubernetes *kubernetespkg.Clientset
	namespace  string
}

func NewClient() *Client {
	config, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
	puccinicommon.FailOnError(err)

	kubernetes, err := kubernetespkg.NewForConfig(config)
	puccinicommon.FailOnError(err)

	namespace_ := namespace
	if cluster {
		namespace_ = ""
	} else if namespace_ == "" {
		if namespace__, ok := common.GetConfiguredNamespace(kubeconfigPath); ok {
			namespace_ = namespace__
		}
		if namespace_ == "" {
			puccinicommon.Fail("could not discover namespace and \"--namespace\" not provided")
		}
	}

	return &Client{
		config:     config,
		kubernetes: kubernetes,
		namespace:  namespace_,
	}
}

func (self *Client) Turandot() *clientpkg.Client {
	apiExtensions, err := apiextensionspkg.NewForConfig(self.config)
	puccinicommon.FailOnError(err)

	turandot, err := turandotpkg.NewForConfig(self.config)
	puccinicommon.FailOnError(err)

	return clientpkg.NewClient(
		self.kubernetes,
		apiExtensions,
		turandot,
		self.kubernetes.CoreV1().RESTClient(),
		self.config,
		cluster,
		self.namespace,
		"turandot",
		"turandot",
		"turandot",
		"tliron/turandot-operator",
		"library/registry",
		"tliron/kubernetes-registry-spooler",
		"/cache",
	)
}

func (self *Client) Spooler() *spoolerpkg.Client {
	return spoolerpkg.NewClient(
		self.kubernetes,
		self.kubernetes.CoreV1().RESTClient(),
		self.config,
		self.namespace,
		"turandot-inventory",
		"spooler",
		"/spool",
	)
}

func Logs(appNameSuffix string, containerName string) {
	// TODO: what happens if we follow more than one log?
	readers, err := NewClient().Turandot().Logs(appNameSuffix, containerName, tail, follow)
	puccinicommon.FailOnError(err)
	for _, reader := range readers {
		defer reader.Close()
	}
	for _, reader := range readers {
		io.Copy(terminal.Stdout, reader)
	}
}

func Shell(appNameSuffix string, containerName string) {
	err := NewClient().Turandot().Shell(appNameSuffix, containerName, os.Stdin, terminal.Stdout, terminal.Stderr)
	puccinicommon.FailOnError(err)
}
