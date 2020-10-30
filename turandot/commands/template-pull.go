package commands

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
	clientpkg "github.com/tliron/turandot/client"
)

func init() {
	templateCommand.AddCommand(templatePullCommand)
	templatePullCommand.Flags().StringVarP(&repository, "repository", "p", "default", "name of repository")
}

var templatePullCommand = &cobra.Command{
	Use:   "pull [SERVICE TEMPLATE NAME] [LOCAL FILE PATH]",
	Short: "Pull a service template from a repository to a local file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceTemplateName := args[0]
		path := args[1]
		PullServiceTemplate(serviceTemplateName, path)
	},
}

func PullServiceTemplate(serviceTemplateName string, path string) {
	turandot := NewClient().Turandot()
	repository_, err := turandot.GetRepository(namespace, repository)
	util.FailOnError(err)
	spoolerCommand, err := turandot.SpoolerCommand(repository_)
	util.FailOnError(err)

	file, err := os.Create(path)
	util.FailOnError(err)
	defer file.Close()

	artifactName := clientpkg.RepositoryArtifactNameForServiceTemplateName(serviceTemplateName)
	err = spoolerCommand.PullTarball(artifactName, file)
	util.FailOnError(err)
}
