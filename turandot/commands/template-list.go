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
	templateListCommand.Flags().StringVarP(&repository, "repository", "p", "default", "name of repository")
}

var templateListCommand = &cobra.Command{
	Use:   "list",
	Short: "List service templates registered in a repository",
	Run: func(cmd *cobra.Command, args []string) {
		ListServiceTemplates()
	},
}

func ListServiceTemplates() {
	turandot := NewClient().Turandot()
	repository_, err := turandot.GetRepository(namespace, repository)
	util.FailOnError(err)
	spoolerCommand, err := turandot.SpoolerCommand(repository_)
	util.FailOnError(err)
	images, err := spoolerCommand.List()
	util.FailOnError(err)

	if len(images) == 0 {
		return
	}
	sort.Strings(images)

	switch format {
	case "":
		urlContext := urlpkg.NewContext()
		defer urlContext.Release()

		table := terminal.NewTable(maxWidth, "Name", "Services")
		for _, image := range images {
			if serviceTemplateName, ok := clientpkg.ServiceTemplateNameForRepositoryImageName(image); ok {
				services, err := turandot.ListServicesForImage(repository, image, urlContext)
				util.FailOnError(err)
				sort.Strings(services)
				table.Add(serviceTemplateName, strings.Join(services, "\n"))
			}
		}
		table.Print()

	case "bare":
		for _, image := range images {
			if serviceTemplateName, ok := clientpkg.ServiceTemplateNameForRepositoryImageName(image); ok {
				fmt.Fprintln(terminal.Stdout, serviceTemplateName)
			}
		}

	default:
		urlContext := urlpkg.NewContext()
		defer urlContext.Release()

		list := make(ard.List, 0, len(images))
		for _, image := range images {
			if serviceTemplateName, ok := clientpkg.ServiceTemplateNameForRepositoryImageName(image); ok {
				map_ := make(ard.StringMap)
				map_["Name"] = serviceTemplateName
				map_["Services"], err = turandot.ListServicesForImage(repository, image, urlContext)
				util.FailOnError(err)
				list = append(list, map_)
			}
		}
		formatpkg.Print(list, format, terminal.Stdout, strict, pretty)
	}
}
