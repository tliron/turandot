package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

func init() {
	serviceCommand.AddCommand(serviceCloutCommand)
}

var serviceCloutCommand = &cobra.Command{
	Use:   "clout [SERVICE NAME]",
	Short: "Get a deployed service's Clout",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		Clout(serviceName)
	},
}

func Clout(serviceName string) {
	// TODO: in cluster mode we must specify the namespace
	namespace := ""

	clout, err := NewClient().Turandot().GetServiceClout(namespace, serviceName)
	util.FailOnError(err)

	terminal.Println(clout)
}
