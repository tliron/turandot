package commands

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
	clientpkg "github.com/tliron/turandot/client"
	"github.com/tliron/turandot/tools"
)

func init() {
	templateCommand.AddCommand(templatePullCommand)
	templatePullCommand.Flags().StringVarP(&inventory, "inventory", "w", "default", "name of inventory")
}

var templatePullCommand = &cobra.Command{
	Use:   "pull [SERVICE TEMPLATE NAME] [LOCAL FILE PATH]",
	Short: "Pull a service template from an inventory to a local file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceTemplateName := args[0]
		path := args[1]
		PullServiceTemplate(serviceTemplateName, path)
	},
}

func PullServiceTemplate(serviceTemplateName string, path string) {
	file, err := os.Create(path)
	util.FailOnError(err)
	defer file.Close()
	imageName := clientpkg.InventoryImageNameForServiceTemplateName(serviceTemplateName)
	err = tools.PullLayerFromRegistry(imageName, file, NewClient().Turandot().Spooler(inventory))
	util.FailOnError(err)
}
