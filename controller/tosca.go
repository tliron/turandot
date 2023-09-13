package controller

import (
	contextpkg "context"

	"github.com/tliron/exturl"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/util"
)

func (self *Controller) CompileServiceTemplate(context contextpkg.Context, serviceTemplateURL string, inputs map[string]string, cloutPath string, urlContext *exturl.Context) (string, error) {
	self.Log.Infof("compiling TOSCA service template: %s", serviceTemplateURL)
	self.Log.Infof("inputs: %s", inputs)

	// Decode inputs
	inputs_ := make(map[string]ard.Value)
	for key, input := range inputs {
		var err error
		if inputs_[key], _, err = ard.DecodeYAML(util.StringToBytes(input), false); err != nil {
			return "", err
		}
	}

	if file, err := transcribe.OpenFileForWrite(cloutPath); err == nil {
		defer file.Close()
		if err := CompileTOSCA(context, serviceTemplateURL, inputs_, file, urlContext); err == nil {
			return util.GetFileHash(cloutPath)
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}
