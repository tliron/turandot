package commands

import (
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tliron/exturl"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

func init() {
	templateCommand.AddCommand(templateListCommand)
	templateListCommand.Flags().StringVarP(&registry, "registry", "r", "default", "name of registry")
}

var templateListCommand = &cobra.Command{
	Use:   "list",
	Short: "List service templates registered in a registry",
	Run: func(cmd *cobra.Command, args []string) {
		ListServiceTemplates()
	},
}

func ListServiceTemplates() {
	turandot := NewClient().Turandot()
	registry_, err := turandot.Reposure.RegistryClient().Get(namespace, registry)
	util.FailOnError(err)
	command, err := turandot.Reposure.SurrogateCommandClient(registry_)
	util.FailOnError(err)
	imageNames, err := command.ListImages()
	util.FailOnError(err)

	if len(imageNames) == 0 {
		return
	}
	sort.Strings(imageNames)

	switch format {
	case "":
		urlContext := exturl.NewContext()
		defer urlContext.Release()

		table := terminal.NewTable(maxWidth, "Name", "Services")
		for _, imageName := range imageNames {
			if serviceTemplateName, ok := turandot.ServiceTemplateNameForRegistryImageName(imageName); ok {
				services, err := turandot.ListServicesForImageName(registry, imageName, urlContext)
				util.FailOnError(err)
				sort.Strings(services)
				table.Add(serviceTemplateName, strings.Join(services, "\n"))
			}
		}
		table.Print()

	case "bare":
		for _, imageName := range imageNames {
			if serviceTemplateName, ok := turandot.ServiceTemplateNameForRegistryImageName(imageName); ok {
				terminal.Println(serviceTemplateName)
			}
		}

	default:
		urlContext := exturl.NewContext()
		defer urlContext.Release()

		list := make(ard.List, 0, len(imageNames))
		for _, imageName := range imageNames {
			if serviceTemplateName, ok := turandot.ServiceTemplateNameForRegistryImageName(imageName); ok {
				map_ := make(ard.StringMap)
				map_["Name"] = serviceTemplateName
				map_["Services"], err = turandot.ListServicesForImageName(registry, imageName, urlContext)
				util.FailOnError(err)
				list = append(list, map_)
			}
		}
		Transcriber().Print(list, os.Stdout, format)
	}
}
