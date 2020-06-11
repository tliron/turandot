package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tliron/puccini/ard"
	puccinicommon "github.com/tliron/puccini/common"
	formatpkg "github.com/tliron/puccini/common/format"
	"github.com/tliron/puccini/common/terminal"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func init() {
	serviceCommand.AddCommand(serviceDescribeCommand)
}

var serviceDescribeCommand = &cobra.Command{
	Use:   "describe [SERVICE NAME]",
	Short: "Describe a deployed service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		DescribeService(args[0])
	},
}

func DescribeService(serviceName string) {
	service, err := NewClient().Turandot().GetService(serviceName)
	puccinicommon.FailOnError(err)

	if format != "" {
		formatpkg.Print(ServiceToARD(service), format, terminal.Stdout, strict, pretty)
	} else {
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("Name"), terminal.ColorValue(service.Name))
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("ServiceTemplateURL"), terminal.ColorValue(service.Spec.ServiceTemplateURL))

		if (service.Spec.Inputs != nil) && (len(service.Spec.Inputs) > 0) {
			fmt.Fprintf(terminal.Stdout, "%s:\n", terminal.ColorTypeName("Inputs"))
			for name, input := range service.Spec.Inputs {
				fmt.Fprintf(terminal.Stdout, "  %s: %s\n", terminal.ColorName(name), terminal.ColorValue(input))
			}
		}

		if (service.Status.Outputs != nil) && (len(service.Status.Outputs) > 0) {
			fmt.Fprintf(terminal.Stdout, "%s:\n", terminal.ColorTypeName("Outputs"))
			for name, output := range service.Status.Outputs {
				fmt.Fprintf(terminal.Stdout, "  %s: %s\n", terminal.ColorName(name), terminal.ColorValue(output))
			}
		}

		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("InstantiationState"), terminal.ColorValue(string(service.Status.InstantiationState)))
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("CloutPath"), terminal.ColorValue(service.Status.CloutPath))
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("CloutHash"), terminal.ColorValue(service.Status.CloutHash))
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("Mode"), terminal.ColorValue(service.Status.Mode))

		if service.Status.NodeStates != nil {
			fmt.Fprintf(terminal.Stdout, "%s:\n", terminal.ColorTypeName("NodeStates"))
			for node, nodeState := range service.Status.NodeStates {
				fmt.Fprintf(terminal.Stdout, "  %s:\n", terminal.ColorName(node))
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.ColorTypeName("Mode"), terminal.ColorValue(nodeState.Mode))
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.ColorTypeName("State"), terminal.ColorValue(string(nodeState.State)))
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.ColorTypeName("Message"), terminal.ColorValue(nodeState.Message))
			}
		}
	}
}

func ServiceToARD(service *resources.Service) ard.StringMap {
	map_ := make(ard.StringMap)
	map_["Name"] = service.Name
	map_["ServiceTemplateURL"] = service.Spec.ServiceTemplateURL
	map_["Inputs"] = service.Spec.Inputs
	map_["Outputs"] = service.Status.Outputs
	map_["InstantiationState"] = service.Status.InstantiationState
	map_["CloutPath"] = service.Status.CloutPath
	map_["CloutHash"] = service.Status.CloutHash
	map_["Mode"] = service.Status.Mode
	nodeStates := make(ard.StringMap)
	if service.Status.NodeStates != nil {
		for node, nodeState := range service.Status.NodeStates {
			nodeStates[node] = ard.StringMap{
				"Mode":    nodeState.Mode,
				"State":   nodeState.State,
				"Message": nodeState.Message,
			}
		}
	}
	map_["NodeStates"] = nodeStates
	return map_
}
