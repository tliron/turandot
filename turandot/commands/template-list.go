package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/ard"
	formatpkg "github.com/tliron/kutil/format"
	"github.com/tliron/kutil/terminal"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	clientpkg "github.com/tliron/turandot/client"
)

func init() {
	templateCommand.AddCommand(templateListCommand)
	templateListCommand.Flags().StringVarP(&inventory, "inventory", "w", "default", "name of inventory")
}

var templateListCommand = &cobra.Command{
	Use:   "list",
	Short: "List service templates registered in an inventory",
	Run: func(cmd *cobra.Command, args []string) {
		ListServiceTemplates()
	},
}

func ListServiceTemplates() {
	images, err := NewClient().Turandot().Spooler(inventory).List()
	util.FailOnError(err)
	if len(images) == 0 {
		return
	}
	sort.Strings(images)

	switch format {
	case "":
		client := NewClient().Turandot()
		urlContext := urlpkg.NewContext()
		defer urlContext.Release()

		table := terminal.NewTable(maxWidth, "Name", "Services")
		for _, image := range images {
			if serviceTemplateName, ok := clientpkg.ServiceTemplateNameForInventoryImageName(image); ok {
				services, err := client.ListServicesForImage(inventory, image, urlContext)
				util.FailOnError(err)
				sort.Strings(services)
				table.Add(serviceTemplateName, strings.Join(services, "\n"))
			}
		}
		table.Print()

	case "bare":
		for _, image := range images {
			if serviceTemplateName, ok := clientpkg.ServiceTemplateNameForInventoryImageName(image); ok {
				fmt.Fprintln(terminal.Stdout, serviceTemplateName)
			}
		}

	default:
		client := NewClient().Turandot()
		urlContext := urlpkg.NewContext()
		defer urlContext.Release()

		list := make(ard.List, 0, len(images))
		for _, image := range images {
			if serviceTemplateName, ok := clientpkg.ServiceTemplateNameForInventoryImageName(image); ok {
				map_ := make(ard.StringMap)
				map_["Name"] = serviceTemplateName
				map_["Services"], err = client.ListServicesForImage(inventory, image, urlContext)
				util.FailOnError(err)
				list = append(list, map_)
			}
		}
		formatpkg.Print(list, format, terminal.Stdout, strict, pretty)
	}
}
