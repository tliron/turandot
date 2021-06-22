package commands

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/ard"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	"github.com/tliron/yamlkeys"
)

var template string
var inputs map[string]string
var inputsUrl string
var mode string

var inputValues = make(map[string]interface{})

func init() {
	serviceCommand.AddCommand(serviceDeployCommand)
	serviceDeployCommand.Flags().StringVarP(&registry, "registry", "r", "default", "name of registry")
	serviceDeployCommand.Flags().StringVarP(&template, "template", "t", "", "name of service template (must be registered in registry)")
	serviceDeployCommand.Flags().StringVarP(&filePath, "file", "f", "", "path to a local CSAR or TOSCA YAML file (will be uploaded)")
	serviceDeployCommand.Flags().StringVarP(&directoryPath, "directory", "d", "", "path to a local directory of TOSCA YAML files (will be uploaded)")
	serviceDeployCommand.Flags().StringVarP(&url, "url", "u", "", "URL to a CSAR or TOSCA YAML file (must be accessible from cluster)")
	serviceDeployCommand.Flags().StringToStringVarP(&inputs, "input", "i", nil, "specify an input (format is name=YAML)")
	serviceDeployCommand.Flags().StringVarP(&inputsUrl, "inputs", "s", "", "load inputs from a PATH or URL to YAML content")
	serviceDeployCommand.Flags().StringVarP(&mode, "mode", "e", "normal", "initial mode")
}

var serviceDeployCommand = &cobra.Command{
	Use:   "deploy [SERVICE NAME]",
	Short: "Deploy a service from a service template or from CSAR or TOSCA YAML content",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		DeployService(serviceName)
	},
}

func DeployService(serviceName string) {
	// TODO: in cluster mode we must specify the namespace
	namespace := ""

	ParseInputs()

	if template != "" {
		if (filePath != "") || (directoryPath != "") || (url != "") {
			deployFailOnlyOneOf()
		}

		turandot := NewClient().Turandot()
		registry_, err := turandot.Reposure.RegistryClient().Get(namespace, registry)
		util.FailOnError(err)
		_, err = turandot.CreateServiceFromTemplate(namespace, serviceName, registry_, template, inputValues, mode)
		util.FailOnError(err)
	} else if filePath != "" {
		if (template != "") || (directoryPath != "") || (url != "") {
			deployFailOnlyOneOf()
		}

		url_, err := urlpkg.NewValidFileURL(filePath, nil)
		util.FailOnError(err)
		createServiceFromContent(serviceName, url_)
	} else if directoryPath != "" {
		if (template != "") || (filePath != "") || (url != "") {
			deployFailOnlyOneOf()
		}

		// TODO pack directory into CSAR?
	} else if url != "" {
		if (template != "") || (filePath != "") || (directoryPath != "") {
			deployFailOnlyOneOf()
		}

		urlContext := urlpkg.NewContext()
		defer urlContext.Release()

		_, err := NewClient().Turandot().CreateServiceFromURL(namespace, serviceName, url, inputValues, mode, urlContext)
		util.FailOnError(err)
	} else {
		url_, err := urlpkg.ReadToInternalURLFromStdin("yaml")
		util.FailOnError(err)
		createServiceFromContent(serviceName, url_)
	}
}

func createServiceFromContent(serviceName string, url urlpkg.URL) {
	turandot := NewClient().Turandot()
	registry_, err := turandot.Reposure.RegistryClient().Get(namespace, registry)
	util.FailOnError(err)
	_, err = turandot.CreateServiceFromContent(namespace, serviceName, registry_, url, inputValues, mode)
	util.FailOnError(err)
}

func ParseInputs() {
	if inputsUrl != "" {
		log.Infof("load inputs from %q", inputsUrl)

		urlContext := urlpkg.NewContext()
		defer urlContext.Release()

		url, err := urlpkg.NewValidURL(inputsUrl, nil, urlContext)
		util.FailOnError(err)
		reader, err := url.Open()
		util.FailOnError(err)
		if closer, ok := reader.(io.Closer); ok {
			defer closer.Close()
		}
		data, err := yamlkeys.DecodeAll(reader)
		util.FailOnError(err)
		for _, data_ := range data {
			if map_, ok := data_.(ard.Map); ok {
				for key, value := range map_ {
					inputValues[yamlkeys.KeyString(key)] = value
				}
			} else {
				util.Failf("malformed inputs in %q", inputsUrl)
			}
		}
	}

	if inputs != nil {
		for name, input := range inputs {
			input_, _, err := ard.DecodeYAML(input, false)
			util.FailOnError(err)
			inputValues[name] = input_
		}
	}
}

func deployFailOnlyOneOf() {
	util.Fail("must provide only one of \"--template\", \"--file\", \"--directory\", or \"--url\"")
}
