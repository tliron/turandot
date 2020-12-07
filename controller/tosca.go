package controller

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/format"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
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
		if err := CompileTOSCA(serviceTemplateURL, inputs_, file, urlContext); err == nil {
			return util.GetFileHash(cloutPath)
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}
