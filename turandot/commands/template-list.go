package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	puccinicommon "github.com/tliron/puccini/common"
	"github.com/tliron/puccini/common/terminal"
	"github.com/tliron/turandot/client"
	"github.com/tliron/turandot/common"
)

func init() {
	templateCommand.AddCommand(templateListCommand)
	templateListCommand.PersistentFlags().BoolVarP(&bare, "bare", "b", false, "list bare names (not as a table)")
}

var templateListCommand = &cobra.Command{
	Use:   "list",
	Short: "List service templates registered in the inventory",
	Run: func(cmd *cobra.Command, args []string) {
		ListServiceTemplates()
	},
}

func ListServiceTemplates() {
	images, err := NewClient().Spooler().List()
	puccinicommon.FailOnError(err)
	if len(images) == 0 {
		return
	}

	if bare {
		for _, image := range images {
			if serviceTemplateName, ok := delegate.ServiceTemplateNameFromInventoryImageName(image); ok {
				fmt.Fprintln(terminal.Stdout, serviceTemplateName)
			}
		}
	} else {
		table := common.NewTable("Name", "Services")
		for _, image := range images {
			if serviceTemplateName, ok := delegate.ServiceTemplateNameFromInventoryImageName(image); ok {
				// TODO: get services
				services := []string{"TODO"}
				table.Add(serviceTemplateName, strings.Join(services, "\n"))
			}
		}
		table.Print()
	}
}
