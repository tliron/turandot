package commands

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	templateCommand.AddCommand(templatePullCommand)
	templatePullCommand.Flags().StringVarP(&registry, "registry", "r", "default", "name of registry")
}

var templatePullCommand = &cobra.Command{
	Use:   "pull [SERVICE TEMPLATE NAME]",
	Short: "Pull a service template from as a CSAR a registry to stdout",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceTemplateName := args[0]
		PullServiceTemplate(serviceTemplateName)
	},
}

func PullServiceTemplate(serviceTemplateName string) {
	turandot := NewClient().Turandot()
	registry_, err := turandot.Reposure.RegistryClient().Get(namespace, registry)
	util.FailOnError(err)
	command, err := turandot.Reposure.SurrogateCommandClient(registry_)
	util.FailOnError(err)

	imageName := turandot.RegistryImageNameForServiceTemplateName(serviceTemplateName)
	err = command.PullLayer(imageName, os.Stdout)
	util.FailOnError(err)
}
