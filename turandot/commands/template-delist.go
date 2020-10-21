package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
	clientpkg "github.com/tliron/turandot/client"
)

func init() {
	templateCommand.AddCommand(templateDelistCommand)
	templateDelistCommand.Flags().StringVarP(&inventory, "inventory", "w", "default", "name of inventory")
	templateDelistCommand.Flags().BoolVarP(&all, "all", "a", false, "delist all templates")
}

var templateDelistCommand = &cobra.Command{
	Use:   "delist [SERVICE TEMPLATE NAME]",
	Short: "Delist a service template from an inventory",
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
	imageName := clientpkg.InventoryImageNameForServiceTemplateName(serviceTemplateName)
	err := NewClient().Turandot().Spooler(inventory).Delete(imageName)
	util.FailOnError(err)
}

func DelistAllTemplates() {
	spooler := NewClient().Turandot().Spooler(inventory)
	images, err := spooler.List()
	util.FailOnError(err)
	for _, image := range images {
		log.Infof("deleting template: %s", image)
		err := spooler.Delete(image)
		util.FailOnError(err)
	}
}
