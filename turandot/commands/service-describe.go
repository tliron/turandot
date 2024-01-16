package commands

import (
	"github.com/spf13/cobra"
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
		Transcriber().Write(resources.ServiceToARD(service))
	} else {
		terminal.Printf("%s: %s\n", terminal.StdoutStylist.TypeName("Name"), terminal.StdoutStylist.Value(service.Name))
		terminal.Printf("%s:\n", terminal.StdoutStylist.TypeName("ServiceTemplate"))

		if service.Spec.ServiceTemplate.Direct != nil {
			terminal.Printf("  %s:\n", terminal.StdoutStylist.TypeName("Direct"))
			if service.Spec.ServiceTemplate.Direct.URL != "" {
				terminal.Printf("    %s: %s\n", terminal.StdoutStylist.TypeName("URL"), terminal.StdoutStylist.Value(service.Spec.ServiceTemplate.Direct.URL))
			}
			if service.Spec.ServiceTemplate.Direct.TLSSecret != "" {
				terminal.Printf("    %s: %s\n", terminal.StdoutStylist.TypeName("TLSSecret"), terminal.StdoutStylist.Value(service.Spec.ServiceTemplate.Direct.TLSSecret))
			}
			if service.Spec.ServiceTemplate.Direct.TLSSecretDataKey != "" {
				terminal.Printf("    %s: %s\n", terminal.StdoutStylist.TypeName("TLSSecretDataKey"), terminal.StdoutStylist.Value(service.Spec.ServiceTemplate.Direct.TLSSecretDataKey))
			}
			if service.Spec.ServiceTemplate.Direct.AuthSecret != "" {
				terminal.Printf("    %s: %s\n", terminal.StdoutStylist.TypeName("AuthSecret"), terminal.StdoutStylist.Value(service.Spec.ServiceTemplate.Direct.AuthSecret))
			}
		}

		if service.Spec.ServiceTemplate.Indirect != nil {
			terminal.Printf("  %s:\n", terminal.StdoutStylist.TypeName("Indirect"))
			if service.Spec.ServiceTemplate.Indirect.Namespace != "" {
				terminal.Printf("    %s: %s\n", terminal.StdoutStylist.TypeName("Namespace"), terminal.StdoutStylist.Value(service.Spec.ServiceTemplate.Indirect.Namespace))
			}
			if service.Spec.ServiceTemplate.Indirect.Registry != "" {
				terminal.Printf("    %s: %s\n", terminal.StdoutStylist.TypeName("Registry"), terminal.StdoutStylist.Value(service.Spec.ServiceTemplate.Indirect.Registry))
			}
			if service.Spec.ServiceTemplate.Indirect.Name != "" {
				terminal.Printf("    %s: %s\n", terminal.StdoutStylist.TypeName("Name"), terminal.StdoutStylist.Value(service.Spec.ServiceTemplate.Indirect.Name))
			}
		}

		if (service.Spec.Inputs != nil) && (len(service.Spec.Inputs) > 0) {
			terminal.Printf("%s:\n", terminal.StdoutStylist.TypeName("Inputs"))
			for name, input := range service.Spec.Inputs {
				terminal.Printf("  %s: %s\n", terminal.StdoutStylist.Name(name), terminal.StdoutStylist.Value(input))
			}
		}

		if (service.Status.Outputs != nil) && (len(service.Status.Outputs) > 0) {
			terminal.Printf("%s:\n", terminal.StdoutStylist.TypeName("Outputs"))
			for name, output := range service.Status.Outputs {
				terminal.Printf("  %s: %s\n", terminal.StdoutStylist.Name(name), terminal.StdoutStylist.Value(output))
			}
		}

		terminal.Printf("%s: %s\n", terminal.StdoutStylist.TypeName("InstantiationState"), terminal.StdoutStylist.Value(string(service.Status.InstantiationState)))
		terminal.Printf("%s: %s\n", terminal.StdoutStylist.TypeName("CloutPath"), terminal.StdoutStylist.Value(service.Status.CloutPath))
		terminal.Printf("%s: %s\n", terminal.StdoutStylist.TypeName("CloutHash"), terminal.StdoutStylist.Value(service.Status.CloutHash))
		terminal.Printf("%s: %s\n", terminal.StdoutStylist.TypeName("Mode"), terminal.StdoutStylist.Value(service.Status.Mode))

		if service.Status.NodeStates != nil {
			terminal.Printf("%s:\n", terminal.StdoutStylist.TypeName("NodeStates"))
			for node, nodeState := range service.Status.NodeStates {
				terminal.Printf("  %s:\n", terminal.StdoutStylist.Name(node))
				terminal.Printf("    %s: %s\n", terminal.StdoutStylist.TypeName("Mode"), terminal.StdoutStylist.Value(nodeState.Mode))
				terminal.Printf("    %s: %s\n", terminal.StdoutStylist.TypeName("State"), terminal.StdoutStylist.Value(string(nodeState.State)))
				if nodeState.Message != "" {
					terminal.Printf("    %s: %s\n", terminal.StdoutStylist.TypeName("Message"), terminal.StdoutStylist.Value(nodeState.Message))
				}
			}
		}
	}
}
