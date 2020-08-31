package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

func init() {
	serviceCommand.AddCommand(serviceModeCommand)
}

var serviceModeCommand = &cobra.Command{
	Use:   "mode [SERVICE NAME] [[MODE]]",
	Short: "Get or set a deployed service's mode",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 2 {
			SetMode(args[0], args[1])
		} else {
			GetMode(args[0])
		}
	},
}

func GetMode(serviceName string) {
	// TODO: in cluster mode we must specify the namespace
	namespace := ""

	service, err := NewClient().Turandot().GetService(namespace, serviceName)
	util.FailOnError(err)
	fmt.Fprintln(terminal.Stdout, service.Status.Mode)
}

func SetMode(serviceName string, mode string) {
	// TODO: in cluster mode we must specify the namespace
	namespace := ""

	client := NewClient().Turandot()
	service, err := client.GetService(namespace, serviceName)
	util.FailOnError(err)
	_, err = client.UpdateServiceMode(service, mode)
	util.FailOnError(err)
}
