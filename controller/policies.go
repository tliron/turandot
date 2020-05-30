package controller

import (
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func (self *Controller) processPolicies(policies parser.OrchestrationPolicies, service *resources.Service, urlContext *urlpkg.Context) error {
	for nodeTemplateName, nodePolicies := range policies {
		self.Log.Infof("provisioning policy for node template %s", nodeTemplateName)
		for _, policy := range nodePolicies {
			self.Log.Infof("instantiable: %t", policy.Instantiable)
			self.Log.Infof("substitutable: %t", policy.Substitutable)
			self.Log.Infof("sites: %s", policy.Sites)

			// Substitutions
			if policy.Substitutable {
				for _, site := range policy.Sites {
					if err := self.Substitute(service.Namespace, nodeTemplateName, policy.SubstitutionInputs, site, urlContext); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
