package commands

import (
	"sort"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/ard"
	formatpkg "github.com/tliron/kutil/format"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
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
	util.FailOnError(err)
	if len(delegates) == 0 {
		return
	}
	sort.Strings(delegates)

	switch format {
	case "":
		// TODO fill table
		table := terminal.NewTable(maxWidth, "Name", "Server", "Namespace")
		for _, delegate := range delegates {
			table.Add(delegate, "TODO", "TODO")
		}
		table.Print()

	case "bare":
		for _, delegate := range delegates {
			terminal.Println(delegate)
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
