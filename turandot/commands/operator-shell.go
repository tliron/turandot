package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	operatorCommand.AddCommand(operatorShellCommand)
}

var operatorShellCommand = &cobra.Command{
	Use:   "shell",
	Short: "Opens a shell to the Turandot operator",
	Run: func(cmd *cobra.Command, args []string) {
		Shell("operator", "operator")
	},
}
