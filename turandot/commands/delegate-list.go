package commands

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tliron/puccini/ard"
	puccinicommon "github.com/tliron/puccini/common"
	formatpkg "github.com/tliron/puccini/common/format"
	"github.com/tliron/puccini/common/terminal"
	"github.com/tliron/turandot/common"
)

func init() {
	delegateCommand.AddCommand(delegateListCommand)
}

var delegateListCommand = &cobra.Command{
	Use:   "list",
	Short: "List delegates",
	Run: func(cmd *cobra.Command, args []string) {
		ListDelegates()
	},
}

func ListDelegates() {
	delegates, err := NewClient().Turandot().ListDelegates()
	puccinicommon.FailOnError(err)
	if len(delegates) == 0 {
		return
	}
	sort.Strings(delegates)

	switch format {
	case "":
		// TODO fill table
		table := common.NewTable(maxWidth, "Name", "Server", "Namespace")
		for _, delegate := range delegates {
			table.Add(delegate, "TODO", "TODO")
		}
		table.Print()

	case "bare":
		for _, delegate := range delegates {
			fmt.Fprintln(terminal.Stdout, delegate)
		}

	default:
		list := make(ard.List, len(delegates))
		for index, delegate := range delegates {
			map_ := make(ard.StringMap)
			map_["Name"] = delegate
			map_["Server"] = ""
			map_["Namespace"] = ""
			list[index] = map_
		}
		formatpkg.Print(list, format, terminal.Stdout, strict, pretty)
	}
}
