package controller

import (
	"errors"

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
		self.Log.Criticalf("%s", policies)
		return nil, errors.New("could not parse policies")
	}
}
