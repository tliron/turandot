package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	delegateCommand.AddCommand(delegateDeleteCommand)
	delegateDeleteCommand.Flags().BoolVarP(&all, "all", "a", false, "delete all delegates")
}

var delegateDeleteCommand = &cobra.Command{
	Use:   "delete [[DELEGATE NAME]]",
	Short: "Delete a delegate",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			delegateName := args[0]
			DeleteDelegate(delegateName)
		} else if all {
			DeleteAllDelegates()
		} else {
			util.Fail("must provide delegate name or specify \"--all\"")
		}
	},
}

func DeleteDelegate(delegateName string) {
	err := NewClient().Turandot().DeleteDelegate(delegateName)
	util.FailOnError(err)
}

func DeleteAllDelegates() {
	turandot := NewClient().Turandot()
	delegates, err := turandot.ListDelegates()
	util.FailOnError(err)
	for _, delegate := range delegates {
		log.Infof("deleting delegate: %s", delegate)
		err := turandot.DeleteDelegate(delegate)
		util.FailOnError(err)
	}
}
