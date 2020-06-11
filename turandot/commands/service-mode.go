package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	puccinicommon "github.com/tliron/puccini/common"
	"github.com/tliron/puccini/common/terminal"
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
	service, err := NewClient().Turandot().GetService(serviceName)
	puccinicommon.FailOnError(err)
	fmt.Fprintln(terminal.Stdout, service.Status.Mode)
}

func SetMode(serviceName string, mode string) {
	client := NewClient().Turandot()
	service, err := client.GetService(serviceName)
	puccinicommon.FailOnError(err)
	if service.Spec.Mode != mode {
		service = service.DeepCopy()
		service.Spec.Mode = mode
		_, err = client.UpdateServiceSpec(service)
		puccinicommon.FailOnError(err)
	}
}
