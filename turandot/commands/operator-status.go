package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	operatorCommand.AddCommand(operatorStatusCommand)
}

var operatorStatusCommand = &cobra.Command{
	Use:   "status",
	Short: "Show the status of the Turandot operator",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
