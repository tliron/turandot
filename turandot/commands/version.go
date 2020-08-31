package commands

import (
	"github.com/tliron/kutil/cobra"
)

func init() {
	rootCommand.AddCommand(cobra.NewVersionCommand(toolName))
}
