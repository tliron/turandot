package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(inventoryCommand)
}

var inventoryCommand = &cobra.Command{
	Use:   "inventory",
	Short: "Control the Turandot inventory",
}
