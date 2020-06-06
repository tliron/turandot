package controller

import (
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func (self *Controller) processPolicies(policies parser.OrchestrationPolicies, service *resources.Service, urlContext *urlpkg.Context) error {
	for nodeTemplateName, nodePolicies := range policies {
		self.Log.Infof("processing policies for node template %s", nodeTemplateName)
		for _, policy := range nodePolicies {
			switch policy_ := policy.(type) {
			case *parser.OrchestrationProvisioningPolicy:
				self.Log.Infof("instantiable: %t", policy_.Instantiable)
				self.Log.Infof("substitutable: %t", policy_.Substitutable)
				self.Log.Infof("sites: %s", policy_.Sites)

				// Substitutions
				if policy_.Substitutable {
					for _, site := range policy_.Sites {
						if err := self.Substitute(service.Namespace, nodeTemplateName, policy_.SubstitutionInputs, site, urlContext); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}
