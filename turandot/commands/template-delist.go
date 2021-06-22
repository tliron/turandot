package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	templateCommand.AddCommand(templateDelistCommand)
	templateDelistCommand.Flags().StringVarP(&registry, "registry", "r", "default", "name of registry")
	templateDelistCommand.Flags().BoolVarP(&all, "all", "a", false, "delist all templates")
}

var templateDelistCommand = &cobra.Command{
	Use:   "delist [[SERVICE TEMPLATE NAME]]",
	Short: "Delist a service template from a registry",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			serviceTemplateName := args[0]
			DelistServiceTemplate(serviceTemplateName)
		} else if all {
			DelistAllTemplates()
		} else {
			util.Fail("must provide service template name or specify \"--all\"")
		}
	},
}

func DelistServiceTemplate(serviceTemplateName string) {
	turandot := NewClient().Turandot()
	registry_, err := turandot.Reposure.RegistryClient().Get(namespace, registry)
	util.FailOnError(err)
	spooler := turandot.Reposure.SurrogateSpoolerClient(registry_)

	imageName := turandot.RegistryImageNameForServiceTemplateName(serviceTemplateName)
	err = spooler.DeleteImage(imageName)
	util.FailOnError(err)
}

func DelistAllTemplates() {
	turandot := NewClient().Turandot()
	registry_, err := turandot.Reposure.RegistryClient().Get(namespace, registry)
	util.FailOnError(err)
	command, err := turandot.Reposure.SurrogateCommandClient(registry_)
	util.FailOnError(err)
	imageNames, err := command.ListImages()
	util.FailOnError(err)
	spooler := turandot.Reposure.SurrogateSpoolerClient(registry_)

	for _, imageName := range imageNames {
		if serviceTemplateName, ok := turandot.ServiceTemplateNameForRegistryImageName(imageName); ok {
			log.Infof("deleting template: %s", serviceTemplateName)
			err := spooler.DeleteImage(imageName)
			util.FailOnError(err)
		}
	}
}
