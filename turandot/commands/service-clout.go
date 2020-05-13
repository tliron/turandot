package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	puccinicommon "github.com/tliron/puccini/common"
	"github.com/tliron/puccini/common/terminal"
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
	clout, err := NewClient().Turandot().ServiceClout(serviceName)
	puccinicommon.FailOnError(err)

	fmt.Fprintln(terminal.Stdout, clout)
}
