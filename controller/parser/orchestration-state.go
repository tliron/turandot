package parser

import (
	"github.com/tliron/kutil/util"
	"gopkg.in/yaml.v3"
)

//
// OrchestrationNodeState
//

type OrchestrationNodeState struct {
	Mode    string `yaml:"mode"`
	State   string `yaml:"state"`
	Message string `yaml:"message"`
}

//
// OrchestrationNodeStates
//

type OrchestrationNodeStates map[string]*OrchestrationNodeState

//
// OrchestrationStates
//

type OrchestrationStates map[string]OrchestrationNodeStates

func DecodeOrchestrationStates(code string) (OrchestrationStates, bool) {
	var self OrchestrationStates
	if err := yaml.Unmarshal(util.StringToBytes(code), &self); err == nil {
		return self, true
	} else {
		return nil, false
	}
}
