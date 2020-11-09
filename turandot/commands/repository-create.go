package commands

import (
	contextpkg "context"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var host string
var serviceNamespace string
var service string
var port uint64
var tlsSecret string
var tlsSecretDataKey string
var provider string

func init() {
	repositoryCommand.AddCommand(repositoryCreateCommand)
	repositoryCreateCommand.Flags().StringVarP(&host, "host", "", "", "registry host (\"host\" or \"host:port\")")
	repositoryCreateCommand.Flags().StringVarP(&serviceNamespace, "service-namespace", "", "", "registry service namespace name (defaults to repository namespace)")
	repositoryCreateCommand.Flags().StringVarP(&service, "service", "", "", "registry service name")
	repositoryCreateCommand.Flags().Uint64VarP(&port, "port", "", 5000, "registry service port")
	repositoryCreateCommand.Flags().StringVarP(&tlsSecret, "tls-secret", "", "", "registry TLS secret name")
	repositoryCreateCommand.Flags().StringVarP(&tlsSecretDataKey, "tls-secret-data-key", "", "", "registry TLS secret data key name")
	repositoryCreateCommand.Flags().StringVarP(&provider, "provider", "", "", "registry provider (\"turandot\", \"minikube\", or \"openshift\")")
	repositoryCreateCommand.Flags().BoolVarP(&wait, "wait", "w", false, "wait for registry spooler to come up")
}

var repositoryCreateCommand = &cobra.Command{
	Use:   "create [REPOSITORY NAME]",
	Short: "Create a repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryName := args[0]
		var authSecret string

		if (host == "") && (service == "") && (provider == "") {
			failRepositoryCreate()
		}

		client := NewClient()

		if host != "" {
			if (service != "") || (provider != "") {
				failRepositoryCreate()
			}
		} else if service != "" {
			if (host != "") || (provider != "") {
				failRepositoryCreate()
			}
		} else if provider != "" {
			if (host != "") || (service != "") {
				failRepositoryCreate()
			}

			switch provider {
			case "turandot":
				service = "turandot-repository"
				if tlsSecret == "" {
					tlsSecret = "turandot-repository"
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
				host = "image-registry.openshift-image-registry.svc:5000"
				if (tlsSecret == "") || (authSecret == "") {
					// We will use the "builder" service account's service-ca certificate and auth token
					serviceAccount, err := client.Kubernetes.CoreV1().ServiceAccounts(client.Namespace).Get(contextpkg.TODO(), "builder", meta.GetOptions{})
					util.FailOnError(err)
					for _, secretName := range serviceAccount.Secrets {
						secret, err := client.Kubernetes.CoreV1().Secrets(client.Namespace).Get(contextpkg.TODO(), secretName.Name, meta.GetOptions{})
						util.FailOnError(err)
						if secret.Type == core.SecretTypeServiceAccountToken {
							if tlsSecret == "" {
								tlsSecret = secret.Name
							}
							if tlsSecretDataKey == "" {
								tlsSecretDataKey = "service-ca.crt"
							}
							if authSecret == "" {
								authSecret = secret.Name
							}
							break
						}
					}
				}

			default:
				util.Fail("unsupported \"--provider\": must be \"turandot\", \"minikube\", or \"openshift\"")
			}
		}

		turandotClient := client.Turandot()

		var err error
		if service != "" {
			_, err = turandotClient.CreateRepositoryIndirect(namespace, repositoryName, serviceNamespace, service, port, tlsSecret, tlsSecretDataKey, authSecret)
		} else {
			_, err = turandotClient.CreateRepositoryDirect(namespace, repositoryName, host, tlsSecret, tlsSecretDataKey, authSecret)
		}
		util.FailOnError(err)

		if wait {
			_, err = turandotClient.WaitForRepositorySpooler(namespace, repositoryName)
			util.FailOnError(err)
		}
	},
}

func failRepositoryCreate() {
	util.Fail("must specify only one of \"--address\", \"--service\", or \"--provider\"")
}
