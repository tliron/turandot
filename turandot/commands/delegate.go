package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(delegateCommand)
}

var delegateCommand = &cobra.Command{
	Use:   "delegate",
	Short: "Work with delegates",
}
