package commands

import (
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
	terminal.Println(service.Status.Mode)
}

func SetMode(serviceName string, mode string) {
	// TODO: in cluster mode we must specify the namespace
	namespace := ""

	turandot := NewClient().Turandot()
	service, err := turandot.GetService(namespace, serviceName)
	util.FailOnError(err)
	_, err = turandot.UpdateServiceMode(service, mode)
	util.FailOnError(err)
}
