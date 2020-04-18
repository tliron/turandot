package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(templateCommand)
}

var templateCommand = &cobra.Command{
	Use:   "template",
	Short: "Work with service templates in the inventory",
}
