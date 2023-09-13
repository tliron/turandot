package parser

import (
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
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

	if mode, ok := execution.Get("mode").String(); ok {
		if scriptletName, ok := execution.Get("scriptlet").String(); ok {
			arguments := make(map[string]string)
			if arguments_, ok := execution.Get("arguments").Map(); ok {
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
// OrchestrationArtifact
//

type OrchestrationArtifact struct {
	SourceURL   string
	TargetPath  string
	Permissions *int64
}

func ParseOrchestrationArtifact(value ard.Value) (*OrchestrationArtifact, bool) {
	artifact := ard.NewNode(value)
	if sourceUrl, ok := artifact.Get("sourceUrl").String(); ok {
		if targetPath, ok := artifact.Get("targetPath").String(); ok {
			self := OrchestrationArtifact{
				SourceURL:  sourceUrl,
				TargetPath: targetPath,
			}
			if permissions, ok := artifact.Get("permissions").Integer(); ok {
				self.Permissions = &permissions
			}
			return &self, true
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}
}

//
// OrchestrationArtifacts
//

type OrchestrationArtifacts []*OrchestrationArtifact

func ParseOrchestrationArtifacts(value ard.List) (OrchestrationArtifacts, bool) {
	self := make(OrchestrationArtifacts, len(value))
	for index, artifact := range value {
		if artifact_, ok := ParseOrchestrationArtifact(artifact); ok {
			self[index] = artifact_
		} else {
			return nil, false
		}
	}
	return self, true
}

//
// OrchestrationContainerExecution
//

type OrchestrationContainerExecution struct {
	Mode             string
	Command          []string // len > 0
	Namespace        string   // can be emtpy
	MatchLabels      map[string]string
	MatchExpressions any
	ContainerName    string // can be emtpy
	Artifacts        OrchestrationArtifacts
}

func ParseOrchestrationContainerExecution(value ard.Value) (*OrchestrationContainerExecution, bool) {
	execution := ard.NewNode(value)

	if mode, ok := execution.Get("mode").String(); ok {
		if command, ok := execution.Get("command").List(); ok {
			namespace, _ := execution.Get("namespace").String()
			containerName, _ := execution.Get("container").String()

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
			if matchLabels_, ok := execution.Get("selector").Get("matchLabels").Map(); ok {
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

			var artifacts OrchestrationArtifacts
			if artifacts_, ok := execution.Get("artifacts").List(); ok {
				if artifacts, ok = ParseOrchestrationArtifacts(artifacts_); !ok {
					return nil, false
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
// OrchestrationSSHExecution
//

type OrchestrationSSHExecution struct {
	Mode      string
	Command   []string // len > 0
	Host      string
	Username  string
	Key       string
	Artifacts OrchestrationArtifacts
}

func ParseOrchestrationSSHExecution(value ard.Value) (*OrchestrationSSHExecution, bool) {
	execution := ard.NewNode(value)

	if mode, ok := execution.Get("mode").String(); ok {
		if command, ok := execution.Get("command").List(); ok {
			if host, ok := execution.Get("host").String(); ok {
				if username, ok := execution.Get("username").String(); ok {
					if key, ok := execution.Get("key").String(); ok {

						command_ := make([]string, 0, len(command))
						for _, value := range command {
							if value_, ok := value.(string); ok {
								command_ = append(command_, value_)
							}
						}
						if len(command_) == 0 {
							return nil, false
						}

						var artifacts OrchestrationArtifacts
						if artifacts_, ok := execution.Get("artifacts").List(); ok {
							if artifacts, ok = ParseOrchestrationArtifacts(artifacts_); !ok {
								return nil, false
							}
						}

						return &OrchestrationSSHExecution{
							Mode:      mode,
							Command:   command_,
							Host:      host,
							Username:  username,
							Key:       key,
							Artifacts: artifacts,
						}, true
					} else {
						return nil, false
					}
				} else {
					return nil, false
				}
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}
}

// OrchestrationExecution interface
func (self *OrchestrationSSHExecution) GetMode() string {
	return self.Mode
}

//
// OrchestrationExecutions
//

type OrchestrationExecutions map[string][]OrchestrationExecution

func DecodeOrchestrationExecutions(code string) (OrchestrationExecutions, bool) {
	if value, _, err := ard.DecodeYAML(util.StringToBytes(code), false); err == nil {
		if executions, ok := ard.NewNode(value).Get("executions").List(); ok {
			self := make(OrchestrationExecutions)

			for _, execution := range executions {
				if nodeTemplateName, ok := ard.NewNode(execution).Get("nodeTemplate").String(); ok {
					nodeTemplateExecutions, _ := self[nodeTemplateName]

					if type_, ok := ard.NewNode(execution).Get("type").String(); ok {
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

						case "ssh":
							if execution_, ok := ParseOrchestrationSSHExecution(execution); ok {
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
	} else {
		return nil, false
	}
}
