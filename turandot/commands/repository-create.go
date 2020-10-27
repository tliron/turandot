package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

var serviceNamespace string
var service string
var port uint64
var provider string
var secret string

func init() {
	repositoryCommand.AddCommand(repositoryCreateCommand)
	repositoryCreateCommand.Flags().StringVarP(&url, "url", "u", "", "registry URL")
	repositoryCreateCommand.Flags().StringVarP(&serviceNamespace, "service-namespace", "", "", "registry service namespace name (defaults to repository namespace)")
	repositoryCreateCommand.Flags().StringVarP(&service, "service", "s", "", "registry service name")
	repositoryCreateCommand.Flags().Uint64VarP(&port, "port", "p", 5000, "registry service port")
	repositoryCreateCommand.Flags().StringVarP(&provider, "provider", "d", "", "registry provider (\"turandot\", \"minikube\", or \"openshift\")")
	repositoryCreateCommand.Flags().StringVarP(&secret, "secret", "t", "", "registry TLS secret name")
	repositoryCreateCommand.Flags().BoolVarP(&wait, "wait", "w", false, "wait for registry spooler to come up")
}

var repositoryCreateCommand = &cobra.Command{
	Use:   "create [REPOSITORY NAME]",
	Short: "Create a repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryName := args[0]

		if (url == "") && (service == "") && (provider == "") {
			failRepositoryCreate()
		}

		if url != "" {
			if (service != "") || (provider != "") {
				failRepositoryCreate()
			}
		} else if service != "" {
			if (url != "") || (provider != "") {
				failRepositoryCreate()
			}
		} else if provider != "" {
			if (url != "") || (service != "") {
				failRepositoryCreate()
			}

			switch provider {
			case "turandot":
				service = "turandot-repository"
				if secret == "" {
					secret = "turandot-repository"
				}

			case "minikube":
				// Note: The Docker container runtime always treats the registry at "127.0.0.1" as insecure
				// However CRI-O does not, thus the most compatible approach is to use the service
				// See: https://github.com/kubernetes/minikube/issues/6982
				serviceNamespace = "kube-system"
				service = "registry"
				// Insecure on port 80
				port = 80

			case "openshift":
			// TODO

			default:
				util.Fail("unsupported \"--provider\": must be \"turandot\", \"minikube\", or \"openshift\"")
			}
		}

		turandotClient := NewClient().Turandot()

		var err error
		if service != "" {
			_, err = turandotClient.CreateRepositoryIndirect(namespace, repositoryName, serviceNamespace, service, port, secret)
		} else {
			_, err = turandotClient.CreateRepositoryDirect(namespace, repositoryName, url, secret)
		}
		util.FailOnError(err)

		if wait {
			_, err = turandotClient.WaitForRepositorySpooler(namespace, repositoryName)
			util.FailOnError(err)
		}
	},
}

func failRepositoryCreate() {
	util.Fail("must specify only one of \"--url\", \"--service\", or \"--provider\"")
}
