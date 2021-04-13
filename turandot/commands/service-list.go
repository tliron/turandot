package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/ard"
	formatpkg "github.com/tliron/kutil/format"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func init() {
	serviceCommand.AddCommand(serviceListCommand)
}

var serviceListCommand = &cobra.Command{
	Use:   "list",
	Short: "List deployed services",
	Run: func(cmd *cobra.Command, args []string) {
		ListServices()
	},
}

func ListServices() {
	services, err := NewClient().Turandot().ListServices()
	util.FailOnError(err)
	if len(services.Items) == 0 {
		return
	}
	// TODO: sort services by name? they seem already sorted!

	switch format {
	case "":
		table := terminal.NewTable(maxWidth, "Name", "State", "Mode", "Inputs", "Outputs")
		for _, service := range services.Items {
			mode := fmt.Sprintf("%s\n", service.Status.Mode)
			if service.Status.NodeStates != nil {
				for node, nodeState := range service.Status.NodeStates {
					if nodeState.Mode == service.Status.Mode {
						mode += fmt.Sprintf("%s: %s\n", node, nodeState.State)
					}
				}
			}

			var inputs string
			if service.Spec.Inputs != nil {
				for _, name := range util.SortedMapStringStringKeys(service.Spec.Inputs) {
					input := service.Spec.Inputs[name]
					inputs += fmt.Sprintf("%s: %s\n", name, input)
				}
			}

			var outputs string
			if service.Status.Outputs != nil {
				for _, name := range util.SortedMapStringStringKeys(service.Status.Outputs) {
					output := service.Status.Outputs[name]
					outputs += fmt.Sprintf("%s: %s\n", name, output)
				}
			}

			table.Add(service.Name, string(service.Status.InstantiationState), mode, inputs, outputs)
		}
		table.Print()

	case "bare":
		for _, service := range services.Items {
			terminal.Println(service.Name)
		}

	default:
		list := make(ard.List, len(services.Items))
		for index, service := range services.Items {
			list[index] = resources.ServiceToARD(&service)
		}
		formatpkg.Print(list, format, terminal.Stdout, strict, pretty)
	}
}
