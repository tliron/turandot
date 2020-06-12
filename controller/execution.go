package controller

import (
	"errors"
	"strings"

	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// See:
//   https://github.com/cosiner/socker
//   https://github.com/pressly/sup

func (self *Controller) processExecutions(executions parser.OrchestrationExecutions, service *resources.Service, urlContext *urlpkg.Context) (*resources.Service, error) {
	for nodeTemplateName, nodeTemplateExecutions := range executions {
		for _, execution := range nodeTemplateExecutions {
			if execution.GetMode() != service.Status.Mode {
				continue
			}

			var err error
			switch execution_ := execution.(type) {
			case *parser.OrchestrationCloutExecution:
				if service, err = self.processCloutExecution(nodeTemplateName, execution_, service, urlContext); err != nil {
					return service, err
				}

			case *parser.OrchestrationContainerExecution:
				if service, err = self.processContainerExecution(nodeTemplateName, execution_, service, urlContext); err != nil {
					return service, err
				}
			}
		}
	}

	return service, nil
}

func (self *Controller) processCloutExecution(nodeTemplateName string, execution *parser.OrchestrationCloutExecution, service *resources.Service, urlContext *urlpkg.Context) (*resources.Service, error) {
	self.Log.Infof("executing scriptlet %q with arguments %q", execution.ScriptletName, execution.Arguments)

	return self.executeCloutUpdate(service, urlContext, execution.ScriptletName, execution.Arguments)
}

func (self *Controller) processContainerExecution(nodeTemplateName string, execution *parser.OrchestrationContainerExecution, service *resources.Service, urlContext *urlpkg.Context) (*resources.Service, error) {
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
					if url, err := urlpkg.NewURL(artifact.SourceURL, urlContext); err == nil {
						if reader, err := url.Open(); err == nil {
							defer reader.Close()
							if err := self.Client.WriteToContainer(pod.Name, containerName, reader, artifact.TargetPath, artifact.Permissions); err != nil {
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
			if url, err := urlpkg.NewURL(service.Status.CloutPath, urlContext); err == nil {
				if reader, err := url.Open(); err == nil {
					defer reader.Close()

					var stdout strings.Builder
					if err := self.Client.Exec(pod.Name, containerName, reader, &stdout, execution.Command...); err == nil {
						yaml := stdout.String()
						if yaml != "" {
							return self.UpdateClout(yaml, service)
						}
					} else {
						return service, nil
					}
				} else {
					return service, err
				}
			} else {
				return service, err
			}
		}
	} else {
		return service, err
	}

	return service, nil
}
