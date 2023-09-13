package commands

import (
	"os"

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
		Transcriber().Print(resources.ServiceToARD(service), os.Stdout, format)
	} else {
		terminal.Printf("%s: %s\n", terminal.DefaultStylist.TypeName("Name"), terminal.DefaultStylist.Value(service.Name))
		terminal.Printf("%s:\n", terminal.DefaultStylist.TypeName("ServiceTemplate"))

		if service.Spec.ServiceTemplate.Direct != nil {
			terminal.Printf("  %s:\n", terminal.DefaultStylist.TypeName("Direct"))
			if service.Spec.ServiceTemplate.Direct.URL != "" {
				terminal.Printf("    %s: %s\n", terminal.DefaultStylist.TypeName("URL"), terminal.DefaultStylist.Value(service.Spec.ServiceTemplate.Direct.URL))
			}
			if service.Spec.ServiceTemplate.Direct.TLSSecret != "" {
				terminal.Printf("    %s: %s\n", terminal.DefaultStylist.TypeName("TLSSecret"), terminal.DefaultStylist.Value(service.Spec.ServiceTemplate.Direct.TLSSecret))
			}
			if service.Spec.ServiceTemplate.Direct.TLSSecretDataKey != "" {
				terminal.Printf("    %s: %s\n", terminal.DefaultStylist.TypeName("TLSSecretDataKey"), terminal.DefaultStylist.Value(service.Spec.ServiceTemplate.Direct.TLSSecretDataKey))
			}
			if service.Spec.ServiceTemplate.Direct.AuthSecret != "" {
				terminal.Printf("    %s: %s\n", terminal.DefaultStylist.TypeName("AuthSecret"), terminal.DefaultStylist.Value(service.Spec.ServiceTemplate.Direct.AuthSecret))
			}
		}

		if service.Spec.ServiceTemplate.Indirect != nil {
			terminal.Printf("  %s:\n", terminal.DefaultStylist.TypeName("Indirect"))
			if service.Spec.ServiceTemplate.Indirect.Namespace != "" {
				terminal.Printf("    %s: %s\n", terminal.DefaultStylist.TypeName("Namespace"), terminal.DefaultStylist.Value(service.Spec.ServiceTemplate.Indirect.Namespace))
			}
			if service.Spec.ServiceTemplate.Indirect.Registry != "" {
				terminal.Printf("    %s: %s\n", terminal.DefaultStylist.TypeName("Registry"), terminal.DefaultStylist.Value(service.Spec.ServiceTemplate.Indirect.Registry))
			}
			if service.Spec.ServiceTemplate.Indirect.Name != "" {
				terminal.Printf("    %s: %s\n", terminal.DefaultStylist.TypeName("Name"), terminal.DefaultStylist.Value(service.Spec.ServiceTemplate.Indirect.Name))
			}
		}

		if (service.Spec.Inputs != nil) && (len(service.Spec.Inputs) > 0) {
			terminal.Printf("%s:\n", terminal.DefaultStylist.TypeName("Inputs"))
			for name, input := range service.Spec.Inputs {
				terminal.Printf("  %s: %s\n", terminal.DefaultStylist.Name(name), terminal.DefaultStylist.Value(input))
			}
		}

		if (service.Status.Outputs != nil) && (len(service.Status.Outputs) > 0) {
			terminal.Printf("%s:\n", terminal.DefaultStylist.TypeName("Outputs"))
			for name, output := range service.Status.Outputs {
				terminal.Printf("  %s: %s\n", terminal.DefaultStylist.Name(name), terminal.DefaultStylist.Value(output))
			}
		}

		terminal.Printf("%s: %s\n", terminal.DefaultStylist.TypeName("InstantiationState"), terminal.DefaultStylist.Value(string(service.Status.InstantiationState)))
		terminal.Printf("%s: %s\n", terminal.DefaultStylist.TypeName("CloutPath"), terminal.DefaultStylist.Value(service.Status.CloutPath))
		terminal.Printf("%s: %s\n", terminal.DefaultStylist.TypeName("CloutHash"), terminal.DefaultStylist.Value(service.Status.CloutHash))
		terminal.Printf("%s: %s\n", terminal.DefaultStylist.TypeName("Mode"), terminal.DefaultStylist.Value(service.Status.Mode))

		if service.Status.NodeStates != nil {
			terminal.Printf("%s:\n", terminal.DefaultStylist.TypeName("NodeStates"))
			for node, nodeState := range service.Status.NodeStates {
				terminal.Printf("  %s:\n", terminal.DefaultStylist.Name(node))
				terminal.Printf("    %s: %s\n", terminal.DefaultStylist.TypeName("Mode"), terminal.DefaultStylist.Value(nodeState.Mode))
				terminal.Printf("    %s: %s\n", terminal.DefaultStylist.TypeName("State"), terminal.DefaultStylist.Value(string(nodeState.State)))
				if nodeState.Message != "" {
					terminal.Printf("    %s: %s\n", terminal.DefaultStylist.TypeName("Message"), terminal.DefaultStylist.Value(nodeState.Message))
				}
			}
		}
	}
}
