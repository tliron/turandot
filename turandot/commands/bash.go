package commands

import (
	"github.com/tliron/puccini/common"
)

func init() {
	rootCommand.AddCommand(common.NewBashCompletionCommand(toolName, rootCommand))
}
