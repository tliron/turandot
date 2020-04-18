package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	puccinicommon "github.com/tliron/puccini/common"
	"github.com/tliron/puccini/common/terminal"
	"github.com/tliron/turandot/common"
)

func init() {
	delegateCommand.AddCommand(delegateListCommand)
	delegateListCommand.PersistentFlags().BoolVarP(&bare, "bare", "b", false, "list bare names (not as a table)")
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

	if bare {
		for _, delegate := range delegates {
			fmt.Fprintln(terminal.Stdout, delegate)
		}
	} else {
		// TODO fill table
		table := common.NewTable("Name", "Server", "Namespace")
		for _, delegate := range delegates {
			table.Add(delegate, "TODO", "TODO")
		}
		table.Print()
	}
}
