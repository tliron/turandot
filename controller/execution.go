package controller

import (
	contextpkg "context"
	"errors"
	"strings"

	"github.com/tliron/exturl"
	"github.com/tliron/kutil/util"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (self *Controller) processExecutions(context contextpkg.Context, executions parser.OrchestrationExecutions, service *resources.Service, urlContext *exturl.Context) (*resources.Service, error) {
	var err error

	for nodeTemplateName, nodeTemplateExecutions := range executions {
		arguments := map[string]string{
			"service":      service.Name,
			"nodeTemplate": nodeTemplateName,
			"mode":         service.Status.Mode,
			"state":        string(resources.ModeAchieved),
		}

	executions:
		for _, execution := range nodeTemplateExecutions {
			if execution.GetMode() != service.Status.Mode {
				continue
			}

			var err error
			switch execution_ := execution.(type) {
			case *parser.OrchestrationCloutExecution:
				if service, err = self.processCloutExecution(context, nodeTemplateName, execution_, service, urlContext); err != nil {
					arguments["state"] = string(resources.ModeFailed)
					arguments["message"] = err.Error()
					break executions
				}

			case *parser.OrchestrationContainerExecution:
				if service, err = self.processContainerExecution(context, nodeTemplateName, execution_, service, urlContext); err != nil {
					arguments["state"] = string(resources.ModeFailed)
					arguments["message"] = err.Error()
					break executions
				}

			case *parser.OrchestrationSSHExecution:
				if service, err = self.processSshExecution(context, nodeTemplateName, execution_, service, urlContext); err != nil {
					arguments["state"] = string(resources.ModeFailed)
					arguments["message"] = err.Error()
					break executions
				}
			}
		}

		if message, ok := arguments["message"]; ok {
			self.Log.Errorf("execution: %s", message)
		}

		if service, err = self.executeCloutUpdate(context, service, urlContext, "orchestration.states.set", arguments); err != nil {
			return service, err
		}
	}

	return service, nil
}

func (self *Controller) processCloutExecution(context contextpkg.Context, nodeTemplateName string, execution *parser.OrchestrationCloutExecution, service *resources.Service, urlContext *exturl.Context) (*resources.Service, error) {
	self.Log.Infof("executing scriptlet %q with arguments %q", execution.ScriptletName, execution.Arguments)

	return self.executeCloutUpdate(context, service, urlContext, execution.ScriptletName, execution.Arguments)
}

func (self *Controller) processContainerExecution(context contextpkg.Context, nodeTemplateName string, execution *parser.OrchestrationContainerExecution, service *resources.Service, urlContext *exturl.Context) (*resources.Service, error) {
	var selector string
	if execution.MatchLabels != nil {
		labels_ := labels.Set(execution.MatchLabels)
		selector = labels_.AsSelector().String()
	}
	// TODO: matchExpressions
	self.Log.Infof("pod selector: %s", selector)

	namespace := execution.Namespace
	if namespace == "" {
		namespace = service.Namespace
	}

	if pods, err := self.Kubernetes.CoreV1().Pods(namespace).List(self.Context, meta.ListOptions{LabelSelector: selector}); err == nil {
		if len(pods.Items) == 0 {
			return service, errors.New("pods not found")
		}

		for _, pod := range pods.Items {
			containerName := execution.ContainerName
			if containerName == "" {
				length := len(pod.Spec.Containers)
				if length == 1 {
					containerName = pod.Spec.Containers[0].Name
				} else if length > 1 {
					return service, errors.New("must specify \"container\" for pods that have more than one container")
				} else {
					return service, errors.New("pod has no containers")
				}
			}

			if execution.Artifacts != nil {
				for _, artifact := range execution.Artifacts {
					self.Log.Infof("copying artifact %q to pod %q container %q path %q", artifact.SourceURL, pod.Name, containerName, artifact.TargetPath)
					if url, err := urlContext.NewURL(artifact.SourceURL); err == nil {
						if reader, err := url.Open(context); err == nil {
							defer reader.Close()
							if err := self.Client.WriteToContainer(namespace, pod.Name, containerName, reader, artifact.TargetPath, artifact.Permissions); err != nil {
								return service, err
							}
						} else {
							return service, err
						}
					} else {
						return service, err
					}
				}
			}

			self.Log.Infof("executing %q on pod %q container %q", execution.Command, pod.Name, containerName)
			if url, err := urlContext.NewURL(service.Status.CloutPath); err == nil {
				if reader, err := url.Open(context); err == nil {
					defer reader.Close()

					var stdout strings.Builder
					if err := self.Client.Exec(namespace, pod.Name, containerName, reader, &stdout, execution.Command...); err == nil {
						yaml := stdout.String()
						if yaml != "" {
							return self.WriteServiceClout(yaml, service)
						} else {
							return service, nil
						}
					} else {
						return service, err
					}
				} else {
					return service, err
				}
			} else {
				return service, err
			}
		}

		return service, nil
	} else {
		return service, err
	}
}

func (self *Controller) processSshExecution(context contextpkg.Context, nodeTemplateName string, execution *parser.OrchestrationSSHExecution, service *resources.Service, urlContext *exturl.Context) (*resources.Service, error) {
	if execution.Host == "" {
		return service, errors.New("SSH execution did not specify host")
	}

	if execution.Artifacts != nil {
		for _, artifact := range execution.Artifacts {
			self.Log.Infof("copying artifact %q via SSH to %q path %q", artifact.SourceURL, execution.Host, artifact.TargetPath)
			if url, err := urlContext.NewURL(artifact.SourceURL); err == nil {
				if reader, err := url.Open(context); err == nil {
					defer reader.Close()
					if err := util.CopySSH(execution.Host, 22, execution.Username, execution.Key, reader, artifact.TargetPath, artifact.Permissions); err != nil {
						return service, err
					}
				} else {
					return service, err
				}
			} else {
				return service, err
			}
		}
	}

	if url, err := urlContext.NewURL(service.Status.CloutPath); err == nil {
		if reader, err := url.Open(context); err == nil {
			defer reader.Close()
			self.Log.Infof("executing %q via SSH to %q", execution.Command, execution.Host)
			if yaml, err := util.ExecSSH(execution.Host, 22, execution.Username, execution.Key, reader, execution.Command...); err == nil {
				if yaml != "" {
					return self.WriteServiceClout(yaml, service)
				} else {
					return service, nil
				}
			} else {
				return service, err
			}
		} else {
			return service, err
		}
	} else {
		return service, err
	}

}
