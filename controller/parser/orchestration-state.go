package parser

import (
	"github.com/tliron/puccini/ard"
)

//
// OrchestrationNodeState
//

type OrchestrationNodeState struct {
	Mode    string
	State   string
	Message string
}

func ParseOrchestrationNodeState(value ard.Value) (*OrchestrationNodeState, bool) {
	state := ard.NewNode(value)
	var self OrchestrationNodeState
	var ok bool
	if self.Mode, ok = state.Get("mode").String(false); !ok {
		return nil, false
	}
	if self.State, ok = state.Get("state").String(false); !ok {
		return nil, false
	}
	if self.Message, ok = state.Get("message").String(false); !ok {
		return nil, false
	}
	return &self, true
}

//
// OrchestrationNodeStates
//

type OrchestrationNodeStates map[string]*OrchestrationNodeState

func ParseOrchestrationNodeStates(value ard.Value) (OrchestrationNodeStates, bool) {
	if nodeStates, ok := value.(ard.Map); ok {
		self := make(OrchestrationNodeStates)
		for nodeTemplateName, state := range nodeStates {
			if nodeTemplateName_, ok := nodeTemplateName.(string); ok {
				if self[nodeTemplateName_], ok = ParseOrchestrationNodeState(state); !ok {
					return nil, false
				}
			} else {
				return nil, false
			}
		}
		return self, true
	} else {
		return nil, false
	}
}

//
// OrchestrationStates
//

type OrchestrationStates map[string]OrchestrationNodeStates

func ParseOrchestrationStates(value ard.Value) (OrchestrationStates, bool) {
	if serviceStates, ok := value.(ard.Map); ok {
		self := make(OrchestrationStates)
		for serviceName, nodeStates := range serviceStates {
			if serviceName_, ok := serviceName.(string); ok {
				if self[serviceName_], ok = ParseOrchestrationNodeStates(nodeStates); !ok {
					return nil, false
				}
			} else {
				return nil, false
			}
		}
		return self, true
	} else {
		return nil, false
	}
}
