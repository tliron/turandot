package commands

import (
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/ard"
	formatpkg "github.com/tliron/kutil/format"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	"github.com/tliron/yamlkeys"
)

var template string
var inputs []string
var inputsUrl string
var mode string

var inputValues = make(map[string]interface{})

func init() {
	serviceCommand.AddCommand(serviceDeployCommand)
	serviceDeployCommand.Flags().StringVarP(&inventory, "inventory", "w", "default", "name of inventory")
	serviceDeployCommand.Flags().StringVarP(&template, "template", "t", "", "name of service template (must be registered in inventory)")
	serviceDeployCommand.Flags().StringVarP(&filePath, "file", "f", "", "path to a local CSAR or TOSCA YAML file (will be uploaded)")
	serviceDeployCommand.Flags().StringVarP(&directoryPath, "directory", "d", "", "path to a local directory of TOSCA YAML files (will be uploaded)")
	serviceDeployCommand.Flags().StringVarP(&url, "url", "u", "", "URL to a CSAR or TOSCA YAML file (must be accessible from cluster)")
	serviceDeployCommand.Flags().StringArrayVarP(&inputs, "input", "i", []string{}, "specify an input (name=YAML)")
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

	client := NewClient()

	urlContext := urlpkg.NewContext()
	defer urlContext.Release()

	if template != "" {
		if (filePath != "") || (directoryPath != "") || (url != "") {
			deployFailOnlyOneOf()
		}

		err := client.Turandot().CreateServiceFromTemplate(namespace, serviceName, inventory, template, inputValues, mode, urlContext)
		util.FailOnError(err)
	} else if filePath != "" {
		if (template != "") || (directoryPath != "") || (url != "") {
			deployFailOnlyOneOf()
		}

		var url_ urlpkg.URL
		var err error
		if filePath != "" {
			url_, err = urlpkg.NewValidFileURL(filePath, nil)
		} else {
			url_, err = urlpkg.ReadToInternalURLFromStdin("yaml")
		}
		util.FailOnError(err)

		turandot := client.Turandot()
		spooler := turandot.Spooler(inventory)
		err = turandot.CreateServiceFromContent(namespace, serviceName, inventory, spooler, url_, inputValues, mode, urlContext)
		util.FailOnError(err)
	} else if directoryPath != "" {
		if (template != "") || (filePath != "") || (url != "") {
			deployFailOnlyOneOf()
		}

		// TODO pack directory into CSAR?
	} else if url != "" {
		if (template != "") || (filePath != "") || (directoryPath != "") {
			deployFailOnlyOneOf()
		}

		err := client.Turandot().CreateServiceFromURL(namespace, serviceName, url, inputValues, mode, urlContext)
		util.FailOnError(err)
	} else {
		deployFailOnlyOneOf()
	}
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
		data, err := formatpkg.ReadAllYAML(reader)
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

	for _, input := range inputs {
		s := strings.SplitN(input, "=", 2)
		if len(s) != 2 {
			util.Failf("malformed input: %s", input)
		}
		value, err := formatpkg.DecodeYAML(s[1])
		util.FailOnError(err)
		inputValues[s[0]] = value
	}
}

func deployFailOnlyOneOf() {
	util.Fail("must provide only one of \"--template\", \"--file\", \"--directory\", or \"--url\"")
}
