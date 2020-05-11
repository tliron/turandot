package controller

import (
	"fmt"

	"github.com/tliron/turandot/controller/parser"
)

func (self *Controller) processPolicies(policies interface{}) (parser.OrchestrationPolicies, error) {
	if policies_, ok := parser.NewOrchestrationPolicies(policies); ok {
		for nodeTemplateName, nodePolicies := range policies_ {
			self.Log.Infof("provisioning policy for node template %s", nodeTemplateName)
			for _, policy := range nodePolicies {
				self.Log.Infof("instantiable: %t", policy.Instantiable)
				self.Log.Infof("substitutable: %t", policy.Substitutable)
				self.Log.Infof("sites: %s", policy.Sites)
			}
		}
		return policies_, nil
	} else {
		return nil, fmt.Errorf("could not parse policies: %s", policies)
	}
}
