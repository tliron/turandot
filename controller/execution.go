package controller

import (
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

// See:
//   https://github.com/cosiner/socker
//   https://github.com/pressly/sup

func (self *Controller) processExecutions(executions parser.OrchestrationExecutions, service *resources.Service, urlContext *urlpkg.Context) (*resources.Service, error) {
	for _, execution := range executions {
		self.Log.Infof("executing scriptlet %s on node template %s", execution.ScriptletName, execution.NodeTemplateName)

		arguments := make(map[string]string)
		if execution.Arguments != nil {
			for key, value := range execution.Arguments {
				arguments[key] = value
			}
		}
		arguments["nodeTemplate"] = execution.NodeTemplateName

		var err error
		if service, err = self.executeCloutUpdate(service, urlContext, execution.ScriptletName, arguments); err != nil {
			return service, err
		}
	}

	return service, nil
}
