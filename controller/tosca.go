package controller

import (
	"github.com/tliron/puccini/ard"
	"github.com/tliron/puccini/common/format"
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/common"
)

func (self *Controller) CompileServiceTemplate(serviceTemplateURL string, inputs map[string]string, cloutPath string, urlContext *urlpkg.Context) (string, error) {
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

	if file, err := format.OpenFileForWrite(cloutPath); err == nil {
		defer file.Close()
		if err := common.CompileTOSCA(serviceTemplateURL, inputs_, file, urlContext); err == nil {
			return common.GetFileHash(cloutPath)
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}
