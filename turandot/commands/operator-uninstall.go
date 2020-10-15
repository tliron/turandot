package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	operatorCommand.AddCommand(operatorUninstallCommand)
	operatorUninstallCommand.Flags().BoolVarP(&wait, "wait", "w", false, "wait for uninstallation to succeed")
}

var operatorUninstallCommand = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the Turandot operator",
	Run: func(cmd *cobra.Command, args []string) {
		NewClient().Turandot().UninstallOperator(wait)
	},
}
