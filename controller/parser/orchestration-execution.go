package parser

import (
	"github.com/tliron/puccini/ard"
)

//
// OrchestrationExecution
//

type OrchestrationExecution interface {
	GetMode() string
}

//
// OrchestrationCloutExecution
//

type OrchestrationCloutExecution struct {
	Mode          string
	ScriptletName string
	Arguments     map[string]string
}

func ParseOrchestrationCloutExecution(value ard.Value) (*OrchestrationCloutExecution, bool) {
	execution := ard.NewNode(value)

	if mode, ok := execution.Get("mode").String(false); ok {
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
				Mode:          mode,
				ScriptletName: scriptletName,
				Arguments:     arguments,
			}, true
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}
}

// OrchestrationExecution interface
func (self *OrchestrationCloutExecution) GetMode() string {
	return self.Mode
}

//
// OrchestrationContainerExecution
//

type OrchestrationContainerExecution struct {
	Mode             string
	Command          []string // len > 0
	Namespace        string   // can be emtpy
	MatchLabels      map[string]string
	MatchExpressions interface{}
	ContainerName    string // can be emtpy
	Artifacts        []string
}

func ParseOrchestrationContainerExecution(value ard.Value) (*OrchestrationContainerExecution, bool) {
	execution := ard.NewNode(value)

	if mode, ok := execution.Get("mode").String(false); ok {
		if command, ok := execution.Get("command").List(false); ok {
			namespace, _ := execution.Get("namespace").String(false)
			containerName, _ := execution.Get("container").String(false)

			command_ := make([]string, 0, len(command))
			for _, value := range command {
				if value_, ok := value.(string); ok {
					command_ = append(command_, value_)
				}
			}
			if len(command_) == 0 {
				return nil, false
			}

			matchLabels := make(map[string]string)
			if matchLabels_, ok := execution.Get("selector").Get("matchLabels").Map(false); ok {
				for key, value := range matchLabels_ {
					if key_, ok := key.(string); ok {
						if value_, ok := value.(string); ok {
							matchLabels[key_] = value_
						}
					}
				}
			}
			if len(matchLabels) == 0 {
				matchLabels = nil
			}

			// TODO: matchExpressions

			var artifacts []string
			if artifacts_, ok := execution.Get("artifacts").List(false); ok {
				for _, value := range artifacts_ {
					if value_, ok := value.(string); ok {
						artifacts = append(artifacts, value_)
					}
				}
			}

			return &OrchestrationContainerExecution{
				Mode:          mode,
				Command:       command_,
				Namespace:     namespace,
				MatchLabels:   matchLabels,
				ContainerName: containerName,
				Artifacts:     artifacts,
			}, true
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}
}

// OrchestrationExecution interface
func (self *OrchestrationContainerExecution) GetMode() string {
	return self.Mode
}

//
// OrchestrationExecutions
//

type OrchestrationExecutions map[string][]OrchestrationExecution

func ParseOrchestrationExecutions(value ard.Value) (OrchestrationExecutions, bool) {
	if executions, ok := ard.NewNode(value).Get("executions").List(false); ok {
		self := make(OrchestrationExecutions)

		for _, execution := range executions {
			if nodeTemplateName, ok := ard.NewNode(execution).Get("nodeTemplate").String(false); ok {
				nodeTemplateExecutions, _ := self[nodeTemplateName]

				if type_, ok := ard.NewNode(execution).Get("type").String(false); ok {
					switch type_ {
					case "clout":
						if execution_, ok := ParseOrchestrationCloutExecution(execution); ok {
							nodeTemplateExecutions = append(nodeTemplateExecutions, execution_)
						} else {
							return nil, false
						}

					case "container":
						if execution_, ok := ParseOrchestrationContainerExecution(execution); ok {
							nodeTemplateExecutions = append(nodeTemplateExecutions, execution_)
						} else {
							return nil, false
						}
					}
				} else {
					return nil, false
				}

				if len(nodeTemplateExecutions) > 0 {
					self[nodeTemplateName] = nodeTemplateExecutions
				}
			} else {
				return nil, false
			}
		}

		if len(self) == 0 {
			self = nil
		}

		return self, true
	} else {
		return nil, false
	}
}
