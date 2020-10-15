package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	inventoryCommand.AddCommand(inventoryUninstallCommand)
	inventoryUninstallCommand.Flags().BoolVarP(&wait, "wait", "w", false, "wait for uninstallation to succeed")
}

var inventoryUninstallCommand = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the Turandot inventory",
	Run: func(cmd *cobra.Command, args []string) {
		NewClient().Turandot().UninstallInventory(wait)
	},
}
