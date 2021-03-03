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
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.StyleTypeName("Name"), terminal.StyleValue(service.Name))
		fmt.Fprintf(terminal.Stdout, "%s:\n", terminal.StyleTypeName("ServiceTemplate"))

		if service.Spec.ServiceTemplate.Direct != nil {
			fmt.Fprintf(terminal.Stdout, "  %s:\n", terminal.StyleTypeName("Direct"))
			if service.Spec.ServiceTemplate.Direct.URL != "" {
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.StyleTypeName("URL"), terminal.StyleValue(service.Spec.ServiceTemplate.Direct.URL))
			}
			if service.Spec.ServiceTemplate.Direct.TLSSecret != "" {
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.StyleTypeName("TLSSecret"), terminal.StyleValue(service.Spec.ServiceTemplate.Direct.TLSSecret))
			}
			if service.Spec.ServiceTemplate.Direct.TLSSecretDataKey != "" {
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.StyleTypeName("TLSSecretDataKey"), terminal.StyleValue(service.Spec.ServiceTemplate.Direct.TLSSecretDataKey))
			}
			if service.Spec.ServiceTemplate.Direct.AuthSecret != "" {
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.StyleTypeName("AuthSecret"), terminal.StyleValue(service.Spec.ServiceTemplate.Direct.AuthSecret))
			}
		}

		if service.Spec.ServiceTemplate.Indirect != nil {
			fmt.Fprintf(terminal.Stdout, "  %s:\n", terminal.StyleTypeName("Indirect"))
			if service.Spec.ServiceTemplate.Indirect.Namespace != "" {
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.StyleTypeName("Namespace"), terminal.StyleValue(service.Spec.ServiceTemplate.Indirect.Namespace))
			}
			if service.Spec.ServiceTemplate.Indirect.Registry != "" {
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.StyleTypeName("Registry"), terminal.StyleValue(service.Spec.ServiceTemplate.Indirect.Registry))
			}
			if service.Spec.ServiceTemplate.Indirect.Name != "" {
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.StyleTypeName("Name"), terminal.StyleValue(service.Spec.ServiceTemplate.Indirect.Name))
			}
		}

		if (service.Spec.Inputs != nil) && (len(service.Spec.Inputs) > 0) {
			fmt.Fprintf(terminal.Stdout, "%s:\n", terminal.StyleTypeName("Inputs"))
			for name, input := range service.Spec.Inputs {
				fmt.Fprintf(terminal.Stdout, "  %s: %s\n", terminal.StyleName(name), terminal.StyleValue(input))
			}
		}

		if (service.Status.Outputs != nil) && (len(service.Status.Outputs) > 0) {
			fmt.Fprintf(terminal.Stdout, "%s:\n", terminal.StyleTypeName("Outputs"))
			for name, output := range service.Status.Outputs {
				fmt.Fprintf(terminal.Stdout, "  %s: %s\n", terminal.StyleName(name), terminal.StyleValue(output))
			}
		}

		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.StyleTypeName("InstantiationState"), terminal.StyleValue(string(service.Status.InstantiationState)))
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.StyleTypeName("CloutPath"), terminal.StyleValue(service.Status.CloutPath))
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.StyleTypeName("CloutHash"), terminal.StyleValue(service.Status.CloutHash))
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.StyleTypeName("Mode"), terminal.StyleValue(service.Status.Mode))

		if service.Status.NodeStates != nil {
			fmt.Fprintf(terminal.Stdout, "%s:\n", terminal.StyleTypeName("NodeStates"))
			for node, nodeState := range service.Status.NodeStates {
				fmt.Fprintf(terminal.Stdout, "  %s:\n", terminal.StyleName(node))
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.StyleTypeName("Mode"), terminal.StyleValue(nodeState.Mode))
				fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.StyleTypeName("State"), terminal.StyleValue(string(nodeState.State)))
				if nodeState.Message != "" {
					fmt.Fprintf(terminal.Stdout, "    %s: %s\n", terminal.StyleTypeName("Message"), terminal.StyleValue(nodeState.Message))
				}
			}
		}
	}
}
