package commands

import (
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tliron/puccini/ard"
	"github.com/tliron/puccini/common"
	formatpkg "github.com/tliron/puccini/common/format"
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/yamlkeys"
)

var template string
var inputs []string
var inputsUrl string

//var remote bool

var inputValues = make(map[string]interface{})

func init() {
	serviceCommand.AddCommand(serviceDeployCommand)
	serviceDeployCommand.Flags().StringVarP(&template, "template", "t", "", "name of service template (must be registered in inventory)")
	serviceDeployCommand.Flags().StringVarP(&filePath, "file", "f", "", "path to a local CSAR or TOSCA YAML file (will be uploaded)")
	serviceDeployCommand.Flags().StringVarP(&directoryPath, "directory", "d", "", "path to a local directory of TOSCA YAML files (will be uploaded)")
	serviceDeployCommand.Flags().StringVarP(&url, "url", "u", "", "URL to a CSAR or TOSCA YAML file (must be accessible from cluster)")
	serviceDeployCommand.Flags().StringArrayVarP(&inputs, "input", "i", []string{}, "specify an input (name=YAML)")
	serviceDeployCommand.Flags().StringVarP(&inputsUrl, "inputs", "p", "", "load inputs from a PATH or URL to YAML content")
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
	ParseInputs()

	client := NewClient()

	if template != "" {
		if (filePath != "") || (directoryPath != "") || (url != "") {
			deployFailOnlyOneOf()
		}

		err := client.Turandot().DeployServiceFromTemplate(serviceName, template, inputValues)
		common.FailOnError(err)
	} else if filePath != "" {
		if (template != "") || (directoryPath != "") || (url != "") {
			deployFailOnlyOneOf()
		}

		var url_ urlpkg.URL
		var err error
		if filePath != "" {
			url_, err = urlpkg.NewValidFileURL(filePath)
		} else {
			url_, err = urlpkg.ReadToInternalURLFromStdin("yaml")
		}
		common.FailOnError(err)

		err = client.Turandot().DeployServiceFromContent(serviceName, client.Spooler(), url_, inputValues)
		common.FailOnError(err)
	} else if directoryPath != "" {
		if (template != "") || (filePath != "") || (url != "") {
			deployFailOnlyOneOf()
		}

		// TODO pack directory into CSAR?
	} else if url != "" {
		if (template != "") || (filePath != "") || (directoryPath != "") {
			deployFailOnlyOneOf()
		}

		err := client.Turandot().DeployServiceFromURL(serviceName, url, inputValues)
		common.FailOnError(err)
	} else {
		deployFailOnlyOneOf()
	}
}

func ParseInputs() {
	if inputsUrl != "" {
		log.Infof("load inputs from \"%s\"", inputsUrl)
		url, err := urlpkg.NewValidURL(inputsUrl, nil)
		common.FailOnError(err)
		reader, err := url.Open()
		common.FailOnError(err)
		if closer, ok := reader.(io.Closer); ok {
			defer closer.Close()
		}
		data, err := formatpkg.ReadYAML(reader)
		common.FailOnError(err)
		if map_, ok := data.(ard.Map); ok {
			for key, value := range map_ {
				inputValues[yamlkeys.KeyString(key)] = value
			}
		} else {
			common.Failf("malformed inputs in \"%s\"", inputsUrl)
		}
	}

	for _, input := range inputs {
		s := strings.SplitN(input, "=", 2)
		if len(s) != 2 {
			common.Failf("malformed input: %s", input)
		}
		value, err := formatpkg.DecodeYAML(s[1])
		common.FailOnError(err)
		inputValues[s[0]] = value
	}
}

func deployFailOnlyOneOf() {
	common.Fail("must provide only one of \"--template\", \"--file\", \"--directory\", or \"--url\"")
}
