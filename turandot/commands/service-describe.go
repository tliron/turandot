package commands

import (
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
		terminal.Printf("%s: %s\n", terminal.Stylize.TypeName("Name"), terminal.Stylize.Value(service.Name))
		terminal.Printf("%s:\n", terminal.Stylize.TypeName("ServiceTemplate"))

		if service.Spec.ServiceTemplate.Direct != nil {
			terminal.Printf("  %s:\n", terminal.Stylize.TypeName("Direct"))
			if service.Spec.ServiceTemplate.Direct.URL != "" {
				terminal.Printf("    %s: %s\n", terminal.Stylize.TypeName("URL"), terminal.Stylize.Value(service.Spec.ServiceTemplate.Direct.URL))
			}
			if service.Spec.ServiceTemplate.Direct.TLSSecret != "" {
				terminal.Printf("    %s: %s\n", terminal.Stylize.TypeName("TLSSecret"), terminal.Stylize.Value(service.Spec.ServiceTemplate.Direct.TLSSecret))
			}
			if service.Spec.ServiceTemplate.Direct.TLSSecretDataKey != "" {
				terminal.Printf("    %s: %s\n", terminal.Stylize.TypeName("TLSSecretDataKey"), terminal.Stylize.Value(service.Spec.ServiceTemplate.Direct.TLSSecretDataKey))
			}
			if service.Spec.ServiceTemplate.Direct.AuthSecret != "" {
				terminal.Printf("    %s: %s\n", terminal.Stylize.TypeName("AuthSecret"), terminal.Stylize.Value(service.Spec.ServiceTemplate.Direct.AuthSecret))
			}
		}

		if service.Spec.ServiceTemplate.Indirect != nil {
			terminal.Printf("  %s:\n", terminal.Stylize.TypeName("Indirect"))
			if service.Spec.ServiceTemplate.Indirect.Namespace != "" {
				terminal.Printf("    %s: %s\n", terminal.Stylize.TypeName("Namespace"), terminal.Stylize.Value(service.Spec.ServiceTemplate.Indirect.Namespace))
			}
			if service.Spec.ServiceTemplate.Indirect.Registry != "" {
				terminal.Printf("    %s: %s\n", terminal.Stylize.TypeName("Registry"), terminal.Stylize.Value(service.Spec.ServiceTemplate.Indirect.Registry))
			}
			if service.Spec.ServiceTemplate.Indirect.Name != "" {
				terminal.Printf("    %s: %s\n", terminal.Stylize.TypeName("Name"), terminal.Stylize.Value(service.Spec.ServiceTemplate.Indirect.Name))
			}
		}

		if (service.Spec.Inputs != nil) && (len(service.Spec.Inputs) > 0) {
			terminal.Printf("%s:\n", terminal.Stylize.TypeName("Inputs"))
			for name, input := range service.Spec.Inputs {
				terminal.Printf("  %s: %s\n", terminal.Stylize.Name(name), terminal.Stylize.Value(input))
			}
		}

		if (service.Status.Outputs != nil) && (len(service.Status.Outputs) > 0) {
			terminal.Printf("%s:\n", terminal.Stylize.TypeName("Outputs"))
			for name, output := range service.Status.Outputs {
				terminal.Printf("  %s: %s\n", terminal.Stylize.Name(name), terminal.Stylize.Value(output))
			}
		}

		terminal.Printf("%s: %s\n", terminal.Stylize.TypeName("InstantiationState"), terminal.Stylize.Value(string(service.Status.InstantiationState)))
		terminal.Printf("%s: %s\n", terminal.Stylize.TypeName("CloutPath"), terminal.Stylize.Value(service.Status.CloutPath))
		terminal.Printf("%s: %s\n", terminal.Stylize.TypeName("CloutHash"), terminal.Stylize.Value(service.Status.CloutHash))
		terminal.Printf("%s: %s\n", terminal.Stylize.TypeName("Mode"), terminal.Stylize.Value(service.Status.Mode))

		if service.Status.NodeStates != nil {
			terminal.Printf("%s:\n", terminal.Stylize.TypeName("NodeStates"))
			for node, nodeState := range service.Status.NodeStates {
				terminal.Printf("  %s:\n", terminal.Stylize.Name(node))
				terminal.Printf("    %s: %s\n", terminal.Stylize.TypeName("Mode"), terminal.Stylize.Value(nodeState.Mode))
				terminal.Printf("    %s: %s\n", terminal.Stylize.TypeName("State"), terminal.Stylize.Value(string(nodeState.State)))
				if nodeState.Message != "" {
					terminal.Printf("    %s: %s\n", terminal.Stylize.TypeName("Message"), terminal.Stylize.Value(nodeState.Message))
				}
			}
		}
	}
}
