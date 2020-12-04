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
	Use:   "pull [SERVICE TEMPLATE NAME] [LOCAL FILE PATH]",
	Short: "Pull a service template from a registry to a local file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceTemplateName := args[0]
		path := args[1]
		PullServiceTemplate(serviceTemplateName, path)
	},
}

func PullServiceTemplate(serviceTemplateName string, path string) {
	turandot := NewClient().Turandot()
	registry_, err := turandot.Reposure.RegistryClient().Get(namespace, registry)
	util.FailOnError(err)
	command, err := turandot.Reposure.CommandClient(registry_)
	util.FailOnError(err)

	file, err := os.Create(path)
	util.FailOnError(err)
	defer file.Close()

	imageName := turandot.RegistryImageNameForServiceTemplateName(serviceTemplateName)
	err = command.PullTarball(imageName, file)
	util.FailOnError(err)
}
