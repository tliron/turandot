package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
	clientpkg "github.com/tliron/turandot/client"
)

func init() {
	templateCommand.AddCommand(templateDelistCommand)
	templateDelistCommand.Flags().StringVarP(&repository, "repository", "p", "default", "name of repository")
	templateDelistCommand.Flags().BoolVarP(&all, "all", "a", false, "delist all templates")
}

var templateDelistCommand = &cobra.Command{
	Use:   "delist [SERVICE TEMPLATE NAME]",
	Short: "Delist a service template from a repository",
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
	turandot := NewClient().Turandot()
	repository_, err := turandot.GetRepository(namespace, repository)
	util.FailOnError(err)
	spooler := turandot.Spooler(repository_)

	imageName := clientpkg.RepositoryImageNameForServiceTemplateName(serviceTemplateName)
	err = spooler.Delete(imageName)
	util.FailOnError(err)
}

func DelistAllTemplates() {
	turandot := NewClient().Turandot()
	repository_, err := turandot.GetRepository(namespace, repository)
	util.FailOnError(err)
	spoolerCommand, err := turandot.SpoolerCommand(repository_)
	util.FailOnError(err)
	images, err := spoolerCommand.List()
	util.FailOnError(err)
	spooler := turandot.Spooler(repository_)

	for _, image := range images {
		log.Infof("deleting template: %s", image)
		err := spooler.Delete(image)
		util.FailOnError(err)
	}
}
