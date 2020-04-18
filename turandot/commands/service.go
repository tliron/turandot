package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(serviceCommand)
}

var serviceCommand = &cobra.Command{
	Use:   "service",
	Short: "Work with services",
}
