package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	formatpkg "github.com/tliron/kutil/format"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
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
	// TODO: in cluster mode we must specify the namespace
	namespace := ""

	turandot := NewClient().Turandot()

	service, err := turandot.GetService(namespace, serviceName)
	util.FailOnError(err)

	if format != "" {
		formatpkg.Print(resources.ServiceToARD(service), format, terminal.Stdout, strict, pretty)
	} else {
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("Name"), terminal.ColorValue(service.Name))
		fmt.Fprintf(terminal.Stdout, "%s:\n", terminal.ColorTypeName("ServiceTemplate"))

		if (service.Spec.ServiceTemplate.Direct.URL != "") || (service.Spec.ServiceTemplate.Direct.Secret != "") {
			fmt.Fprintf(terminal.Stdout, "  %s:\n", terminal.ColorTypeName("Direct"))
			if service.Spec.ServiceTemplate.Direct.URL != "" {
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.ColorTypeName("URL"), terminal.ColorValue(service.Spec.ServiceTemplate.Direct.URL))
			}
			if service.Spec.ServiceTemplate.Direct.Secret != "" {
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.ColorTypeName("Secret"), terminal.ColorValue(service.Spec.ServiceTemplate.Direct.Secret))
			}
		}

		if (service.Spec.ServiceTemplate.Indirect.Repository != "") || (service.Spec.ServiceTemplate.Indirect.Name != "") {
			fmt.Fprintf(terminal.Stdout, "  %s:\n", terminal.ColorTypeName("Indirect"))
			if service.Spec.ServiceTemplate.Indirect.Repository != "" {
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.ColorTypeName("Repository"), terminal.ColorValue(service.Spec.ServiceTemplate.Indirect.Repository))
			}
			if service.Spec.ServiceTemplate.Indirect.Name != "" {
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.ColorTypeName("Name"), terminal.ColorValue(service.Spec.ServiceTemplate.Indirect.Name))
			}
		}

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
