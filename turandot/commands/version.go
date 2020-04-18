package commands

import (
	"github.com/tliron/turandot/version"
)

func init() {
	rootCommand.AddCommand(version.NewCommand(toolName))
}
