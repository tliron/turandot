package commands

import (
	"github.com/spf13/cobra"
	urlpkg "github.com/tliron/kutil/url"
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
		registry_, err := turandot.Reposure.RegistryClient().Get(namespace, registry)
		util.FailOnError(err)
		spooler := turandot.Reposure.SpoolerClient(registry_)

		imageName := turandot.RegistryImageNameForServiceTemplateName(serviceTemplateName)
		err = spooler.PushTarballFromURL(imageName, url)
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
