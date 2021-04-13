package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

func init() {
	serviceCommand.AddCommand(serviceOutputCommand)
}

var serviceOutputCommand = &cobra.Command{
	Use:   "output [SERVICE NAME] [OUTPUT NAME]",
	Short: "Get a deployed service's output",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		outputName := args[1]
		ServiceOutput(serviceName, outputName)
	},
}

func ServiceOutput(serviceName string, outputName string) {
	// TODO: in cluster mode we must specify the namespace
	namespace := ""

	service, err := NewClient().Turandot().GetService(namespace, serviceName)
	util.FailOnError(err)

	if service.Status.Outputs != nil {
		if output, ok := service.Status.Outputs[outputName]; ok {
			// TODO: unpack the YAML
			// TODO: support output in various formats
			terminal.Println(output)
			return
		}
	}

	util.Failf("output %q not found in service %q", outputName, serviceName)
}
