package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	inventoryCommand.AddCommand(inventoryInstallCommand)
	inventoryInstallCommand.Flags().BoolVarP(&cluster, "cluster", "c", false, "cluster mode")
	inventoryInstallCommand.Flags().StringVarP(&registry, "registry", "g", "docker.io", "registry URL (use special value \"internal\" to discover internally deployed registry)")
	inventoryInstallCommand.Flags().BoolVarP(&wait, "wait", "w", false, "wait for installation to succeed")
}

var inventoryInstallCommand = &cobra.Command{
	Use:   "install",
	Short: "Install the Turandot inventory",
	Run: func(cmd *cobra.Command, args []string) {
		err := NewClient().Turandot().InstallInventory(registry, wait)
		util.FailOnError(err)
	},
}
