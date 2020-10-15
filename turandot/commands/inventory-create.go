package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

var service string

func init() {
	inventoryCommand.AddCommand(inventoryCreateCommand)
	inventoryCreateCommand.Flags().StringVarP(&url, "url", "u", "", "registry URL")
	inventoryCreateCommand.Flags().StringVarP(&service, "service", "s", "", "registry service name")
	// TODO:
	// --turandot
	// --minikube
	// --openshift
}

var inventoryCreateCommand = &cobra.Command{
	Use:   "create [INVENTORY NAME]",
	Short: "Create an inventory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if ((url == "") && (service == "")) || ((url != "") && (service != "")) {
			util.Fail("must provide either \"--url\" or \"--service\"")
		}

		inventoryName := args[0]

		_, err := NewClient().Turandot().CreateInventory(namespace, inventoryName, url, service)
		util.FailOnError(err)
	},
}
