package controller

import (
	"github.com/tliron/puccini/ard"
	"github.com/tliron/puccini/common/format"
	"github.com/tliron/turandot/common"
)

func (self *Controller) CompileServiceTemplate(serviceTemplateURL string, inputs map[string]string, cloutPath string) (string, error) {
	self.Log.Infof("compiling TOSCA service template: %s", serviceTemplateURL)
	self.Log.Infof("inputs: %s", inputs)

	// Decode inputs
	inputs_ := make(map[string]ard.Value)
	for key, input := range inputs {
		var err error
		if inputs_[key], err = format.DecodeYAML(input); err != nil {
			return "", err
		}
	}

	if err := common.CompileTOSCA(serviceTemplateURL, cloutPath, inputs_); err == nil {
		return common.GetFileHash(cloutPath)
	} else {
		return "", err
	}
}
