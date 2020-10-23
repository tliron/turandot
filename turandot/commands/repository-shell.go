package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	repositoryCommand.AddCommand(repositoryShellCommand)
	repositoryShellCommand.PersistentFlags().StringVarP(&component, "component", "c", "registry", "sub-component (\"registry\" or \"spooler\")")
}

var repositoryShellCommand = &cobra.Command{
	Use:   "shell",
	Short: "Opens a shell to the Turandot repository",
	Run: func(cmd *cobra.Command, args []string) {
		Shell("repository", component)
	},
}
