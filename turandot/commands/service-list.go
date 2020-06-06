package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	puccinicommon "github.com/tliron/puccini/common"
	"github.com/tliron/puccini/common/terminal"
	"github.com/tliron/turandot/common"
)

func init() {
	serviceCommand.AddCommand(serviceListCommand)
	serviceListCommand.PersistentFlags().BoolVarP(&bare, "bare", "b", false, "list bare names (not as a table)")
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
	puccinicommon.FailOnError(err)
	if len(services.Items) == 0 {
		return
	}
	// TODO: sort services by name? they seem already sorted!

	if bare {
		for _, service := range services.Items {
			fmt.Fprintln(terminal.Stdout, service.Name)
		}
	} else {
		table := common.NewTable(maxWidth, "Name", "Status", "ServiceTemplateURL", "Inputs", "Outputs")
		for _, service := range services.Items {
			var inputs string
			if service.Spec.Inputs != nil {
				for _, name := range common.SortedMapStringStringKeys(service.Spec.Inputs) {
					input := service.Spec.Inputs[name]
					inputs += fmt.Sprintf("%s: %s\n", name, input)
				}
			}

			var outputs string
			if service.Status.Outputs != nil {
				for _, name := range common.SortedMapStringStringKeys(service.Status.Outputs) {
					output := service.Status.Outputs[name]
					outputs += fmt.Sprintf("%s: %s\n", name, output)
				}
			}

			table.Add(service.Name, string(service.Status.Status), service.Spec.ServiceTemplateURL, inputs, outputs)
		}
		table.Print()
	}
}
