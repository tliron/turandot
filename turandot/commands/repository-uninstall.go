package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	repositoryCommand.AddCommand(repositoryUninstallCommand)
	repositoryUninstallCommand.Flags().BoolVarP(&wait, "wait", "w", false, "wait for uninstallation to succeed")
}

var repositoryUninstallCommand = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the Turandot repository",
	Run: func(cmd *cobra.Command, args []string) {
		NewClient().Turandot().UninstallRepository(wait)
	},
}
