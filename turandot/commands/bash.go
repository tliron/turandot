package commands

import (
	puccinicommon "github.com/tliron/puccini/common"
)

func init() {
	rootCommand.AddCommand(puccinicommon.NewBashCompletionCommand(toolName, rootCommand))
}
