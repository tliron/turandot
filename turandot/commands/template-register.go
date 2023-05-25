package commands

import (
	contextpkg "context"

	"github.com/spf13/cobra"
	"github.com/tliron/exturl"
	"github.com/tliron/kutil/util"
)

func init() {
	templateCommand.AddCommand(templateRegisterCommand)
	templateRegisterCommand.Flags().StringVarP(&registry, "registry", "r", "default", "name of registry")
	templateRegisterCommand.Flags().StringVarP(&filePath, "file", "f", "", "path to a local CSAR or TOSCA YAML file (will be uploaded)")
	templateRegisterCommand.Flags().StringVarP(&directoryPath, "directory", "d", "", "path to a local directory of TOSCA YAML files (will be packed into a CSAR and uploaded)")
}

var templateRegisterCommand = &cobra.Command{
	Use:   "register [SERVICE TEMPLATE NAME]",
	Short: "Register a service template in a registry from CSAR or TOSCA YAML content",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceTemplateName := args[0]
		RegisterServiceTemplate(contextpkg.TODO(), serviceTemplateName)
	},
}

func RegisterServiceTemplate(context contextpkg.Context, serviceTemplateName string) {
	urlContext := exturl.NewContext()
	defer urlContext.Release()

	if filePath != "" {
		if directoryPath != "" {
			registerFailOnlyOneOf()
		}

		url, err := urlContext.NewValidFileURL(filePath)
		util.FailOnError(err)
		registerServiceTemplate(context, serviceTemplateName, url)
	} else if directoryPath != "" {
		if filePath != "" {
			registerFailOnlyOneOf()
		}

		// TODO pack directory into CSAR
	} else {
		url, err := urlContext.ReadToInternalURLFromStdin(context, "yaml")
		util.FailOnError(err)
		registerServiceTemplate(context, serviceTemplateName, url)
	}
}

func registerServiceTemplate(context contextpkg.Context, serviceTemplateName string, url exturl.URL) {
	turandot := NewClient().Turandot()
	registry_, err := turandot.Reposure.RegistryClient().Get(namespace, registry)
	util.FailOnError(err)
	spooler := turandot.Reposure.SurrogateSpoolerClient(registry_)

	imageName := turandot.RegistryImageNameForServiceTemplateName(serviceTemplateName)
	err = spooler.PushTarballFromURL(context, imageName, url)
	util.FailOnError(err)
}

func registerFailOnlyOneOf() {
	util.Fail("must provide only one of \"--file\" or \"--directory\"")
}
