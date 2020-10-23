package commands

import (
	"github.com/spf13/cobra"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	clientpkg "github.com/tliron/turandot/client"
	"github.com/tliron/turandot/tools"
)

func init() {
	templateCommand.AddCommand(templateRegisterCommand)
	templateRegisterCommand.Flags().StringVarP(&repository, "repository", "p", "default", "name of repository")
	templateRegisterCommand.Flags().StringVarP(&filePath, "file", "f", "", "path to a local CSAR or TOSCA YAML file (will be uploaded)")
	templateRegisterCommand.Flags().StringVarP(&directoryPath, "directory", "d", "", "path to a local directory of TOSCA YAML files (will be uploaded)")
}

var templateRegisterCommand = &cobra.Command{
	Use:   "register [SERVICE TEMPLATE NAME]",
	Short: "Register a service template in a repository from CSAR or TOSCA YAML content",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceTemplateName := args[0]
		RegisterServiceTemplate(serviceTemplateName)
	},
}

func RegisterServiceTemplate(serviceTemplateName string) {
	if filePath != "" {
		if (directoryPath != "") || (url != "") {
			registerFailOnlyOneOf()
		}

		var url urlpkg.URL
		var err error
		if filePath != "" {
			url, err = urlpkg.NewValidFileURL(filePath, nil)
		} else {
			url, err = urlpkg.ReadToInternalURLFromStdin("yaml")
		}
		util.FailOnError(err)

		turandot := NewClient().Turandot()
		repository_, err := turandot.GetRepository(namespace, repository)
		util.FailOnError(err)
		spooler := turandot.Spooler(repository_)

		imageName := clientpkg.RepositoryImageNameForServiceTemplateName(serviceTemplateName)
		err = tools.PublishOnRegistry(imageName, url, spooler)
		util.FailOnError(err)
	} else if directoryPath != "" {
		if (filePath != "") || (url != "") {
			registerFailOnlyOneOf()
		}

		// TODO pack directory into CSAR
	} else {
		registerFailOnlyOneOf()
	}
}

func registerFailOnlyOneOf() {
	util.Fail("must provide only one of \"--file\" or \"--directory\"")
}
