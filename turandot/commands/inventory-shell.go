package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	inventoryCommand.AddCommand(inventoryShellCommand)
	inventoryShellCommand.PersistentFlags().StringVarP(&component, "component", "p", "registry", "sub-component (\"registry\" or \"spooler\")")
}

var inventoryShellCommand = &cobra.Command{
	Use:   "shell",
	Short: "Opens a shell to the Turandot inventory",
	Run: func(cmd *cobra.Command, args []string) {
		Shell("inventory", component)
	},
}
