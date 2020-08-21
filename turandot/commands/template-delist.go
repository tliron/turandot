package commands

import (
	"github.com/spf13/cobra"
	puccinicommon "github.com/tliron/puccini/common"
	clientpkg "github.com/tliron/turandot/client"
)

func init() {
	templateCommand.AddCommand(templateDelistCommand)
	templateDelistCommand.Flags().BoolVarP(&all, "all", "a", false, "delist all templates")
}

var templateDelistCommand = &cobra.Command{
	Use:   "delist [SERVICE TEMPLATE NAME]",
	Short: "Delist a service template from the inventory",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			serviceTemplateName := args[0]
			DelistServiceTemplate(serviceTemplateName)
		} else if all {
			DelistAllTemplates()
		}
	},
}

func DelistServiceTemplate(serviceTemplateName string) {
	imageName := clientpkg.GetInventoryImageName(serviceTemplateName)
	err := NewClient().Spooler().Delete(imageName)
	puccinicommon.FailOnError(err)
}

func DelistAllTemplates() {
	spooler := NewClient().Spooler()
	images, err := spooler.List()
	puccinicommon.FailOnError(err)
	for _, image := range images {
		log.Infof("deleting template: %s", image)
		err := spooler.Delete(image)
		puccinicommon.FailOnError(err)
	}
}
