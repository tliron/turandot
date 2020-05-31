package parser

import (
	"github.com/tliron/puccini/ard"
)

//
// OrchestrationCloutExecution
//

type OrchestrationCloutExecution struct {
	NodeTemplateName string
	ScriptletName    string
	Arguments        map[string]string
}

func ParseOrchestrationCloutExecution(value ard.Value) (*OrchestrationCloutExecution, bool) {
	execution := ard.NewNode(value)
	if nodeTemplateName, ok := execution.Get("nodeTemplate").String(false); ok {
		if scriptletName, ok := execution.Get("scriptlet").String(false); ok {
			arguments := make(map[string]string)
			if arguments_, ok := execution.Get("arguments").Map(false); ok {
				for key, value := range arguments_ {
					if key_, ok := key.(string); ok {
						if value_, ok := value.(string); ok {
							arguments[key_] = value_
						}
					}
				}
			}
			if len(arguments) == 0 {
				arguments = nil
			}

			return &OrchestrationCloutExecution{
				NodeTemplateName: nodeTemplateName,
				ScriptletName:    scriptletName,
				Arguments:        arguments,
			}, true
		}
	}
	return nil, false
}

//
// OrchestrationExecutions
//

type OrchestrationExecutions []*OrchestrationCloutExecution

func ParseOrchestrationExecutions(value ard.Value) (OrchestrationExecutions, bool) {
	if executions, ok := ard.NewNode(value).Get("executions").List(false); ok {
		var self OrchestrationExecutions
		for _, execution := range executions {
			if type_, ok := ard.NewNode(execution).Get("type").String(false); ok {
				switch type_ {
				case "clout":
					if execution_, ok := ParseOrchestrationCloutExecution(execution); ok {
						self = append(self, execution_)
					} else {
						return nil, false
					}
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
