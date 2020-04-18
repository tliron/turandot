package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(operatorCommand)
}

var operatorCommand = &cobra.Command{
	Use:   "operator",
	Short: "Control the Turandot operator",
}
