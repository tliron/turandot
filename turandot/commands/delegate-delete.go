package commands

import (
	"github.com/spf13/cobra"
	puccinicommon "github.com/tliron/puccini/common"
)

func init() {
	delegateCommand.AddCommand(delegateDeleteCommand)
	delegateDeleteCommand.PersistentFlags().BoolVarP(&all, "all", "a", false, "delete all delegates")
}

var delegateDeleteCommand = &cobra.Command{
	Use:   "delete [DELEGATE NAME]",
	Short: "Delete a delegate",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			delegateName := args[0]
			DeleteDelegate(delegateName)
		} else if all {
			DeleteAllDelegates()
		}
	},
}

func DeleteDelegate(delegateName string) {
	err := NewClient().Turandot().DeleteDelegate(delegateName)
	puccinicommon.FailOnError(err)
}

func DeleteAllDelegates() {
	turandot := NewClient().Turandot()
	delegates, err := turandot.ListDelegates()
	puccinicommon.FailOnError(err)
	for _, delegate := range delegates {
		log.Infof("deleting delegate: %s", delegate)
		err := turandot.DeleteDelegate(delegate)
		puccinicommon.FailOnError(err)
	}
}
