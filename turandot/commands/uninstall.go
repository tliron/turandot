package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(uninstallCommand)
}

var uninstallCommand = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Turandot",
	Run: func(cmd *cobra.Command, args []string) {
		NewClient().Turandot().Uninstall()
	},
}
