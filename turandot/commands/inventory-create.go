package commands

import (
	contextpkg "context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var service string
var provider string
var secret string

func init() {
	inventoryCommand.AddCommand(inventoryCreateCommand)
	inventoryCreateCommand.Flags().StringVarP(&url, "url", "u", "", "registry URL")
	inventoryCreateCommand.Flags().StringVarP(&service, "service", "s", "", "registry service name")
	inventoryCreateCommand.Flags().StringVarP(&provider, "provider", "p", "", "registry provider (\"turandot\", \"minikube\", or \"openshift\")")
	inventoryCreateCommand.Flags().StringVarP(&secret, "secret", "t", "", "registry TLS secret name")
	inventoryCreateCommand.Flags().BoolVarP(&wait, "wait", "w", false, "wait for inventory spooler to come up")
}

var inventoryCreateCommand = &cobra.Command{
	Use:   "create [INVENTORY NAME]",
	Short: "Create an inventory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inventoryName := args[0]

		var client *Client

		if (url == "") && (service == "") && (provider == "") {
			failInventoryCreate()
		}

		if url != "" {
			if (service != "") || (provider != "") {
				failInventoryCreate()
			}
		} else if service != "" {
			if (url != "") || (provider != "") {
				failInventoryCreate()
			}
		} else if provider != "" {
			if (url != "") || (service != "") {
				failInventoryCreate()
			}

			switch provider {
			case "turandot":
				client = NewClient()
				service_, err := client.kubernetes.CoreV1().Services(namespace).Get(contextpkg.TODO(), "turandot-inventory", meta.GetOptions{})
				util.FailOnError(err)
				url = fmt.Sprintf("%s:5000", service_.Spec.ClusterIP)
				if secret == "" {
					secret = "turandot-inventory"
				}

			case "minikube":
				// Note: The Docker container runtime always treats the registry at "127.0.0.1" as insecure
				// However CRI-O does not, thus the most compatible approach is to use the service
				// See: https://github.com/kubernetes/minikube/issues/6982
				client = NewClient()
				service_, err := client.kubernetes.CoreV1().Services("kube-system").Get(contextpkg.TODO(), "registry", meta.GetOptions{})
				util.FailOnError(err)
				// Insecure on port 80
				url = fmt.Sprintf("%s:80", service_.Spec.ClusterIP)

			case "openshift":

			default:
				util.Fail("unsupported \"--provider\": must be \"turandot\", \"minikube\", or \"openshift\"")
			}
		}

		if client == nil {
			client = NewClient()
		}
		turandotClient := client.Turandot()

		_, err := turandotClient.CreateInventory(namespace, inventoryName, url, service, secret)
		util.FailOnError(err)
		if wait {
			_, err = turandotClient.WaitForInventorySpooler(namespace, inventoryName)
			util.FailOnError(err)
		}
	},
}

func failInventoryCreate() {
	util.Fail("must specify only one of \"--url\", \"--service\", or \"--provider\"")
}
