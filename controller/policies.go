package controller

import (
	"errors"

	"github.com/tliron/puccini/ard"
)

func (self *Controller) processPolicies(policies interface{}) (OrchestrationPolicies, error) {
	if policies_, ok := NewOrchestrationPolicies(policies); ok {
		for nodeTemplateName, nodePolicies := range policies_ {
			self.log.Infof("provisioning policy for node template %s", nodeTemplateName)
			for _, policy := range nodePolicies {
				self.log.Infof("instantiable: %t", policy.Instantiable)
				self.log.Infof("substitutable: %t", policy.Substitutable)
				self.log.Infof("sites: %s", policy.Sites)
			}
		}
		return policies_, nil
	} else {
		self.log.Criticalf("%s", policies)
		return nil, errors.New("could not parse policies")
	}
}

//
// OrchestrationProvisioningPolicy
//

type OrchestrationProvisioningPolicy struct {
	Sites              []string
	Profile            bool
	Substitutable      bool
	Instantiable       bool
	Virtualizable      bool
	SubstitutionInputs map[string]interface{}
}

func NewOrchestrationProvisioningPolicy(data interface{}) (*OrchestrationProvisioningPolicy, bool) {
	properties := ard.NewNode(data)
	self := OrchestrationProvisioningPolicy{
		SubstitutionInputs: make(map[string]interface{}),
	}
	var ok bool
	if self.Substitutable, ok = properties.Get("substitutable").Boolean(true); !ok {
		return nil, false
	}
	if self.Instantiable, ok = properties.Get("instantiable").Boolean(true); !ok {
		return nil, false
	}
	if self.Virtualizable, ok = properties.Get("virtualizable").Boolean(true); !ok {
		return nil, false
	}
	if sites := properties.Get("sites"); sites != ard.NoNode {
		if sites_, ok := sites.List(true); ok {
			for _, site := range sites_ {
				if site_, ok := site.(string); ok {
					self.Sites = append(self.Sites, site_)
				} else {
					return nil, false
				}
			}
		} else {
			return nil, false
		}
	}
	if substitutionInputs := properties.Get("substitutionInputs"); substitutionInputs != ard.NoNode {
		if substitutionInputs_, ok := substitutionInputs.Map(true); ok {
			for name, input := range substitutionInputs_ {
				if name_, ok := name.(string); ok {
					self.SubstitutionInputs[name_] = input
				} else {
					return nil, false
				}
			}
		} else {
			return nil, false
		}
	}
	return &self, true
}

//
// OrchestrationPolicies
//

type OrchestrationPolicies map[string][]*OrchestrationProvisioningPolicy

func NewOrchestrationPolicies(data interface{}) (OrchestrationPolicies, bool) {
	if policies, ok := data.(ard.Map); ok {
		self := make(OrchestrationPolicies)
		for nodeTemplateName, nodePolicies := range policies {
			if nodeTemplateName_, ok := nodeTemplateName.(string); ok {
				if nodePolicies_, ok := nodePolicies.(ard.List); ok {
					var policies []*OrchestrationProvisioningPolicy
					for _, policy := range nodePolicies_ {
						policy_ := ard.NewNode(policy)
						if type_, ok := policy_.Get("type").String(false); ok {
							if properties, ok := policy_.Get("properties").Map(true); ok {
								switch type_ {
								case "provisioning":
									if policy__, ok := NewOrchestrationProvisioningPolicy(properties); ok {
										policies = append(policies, policy__)
									} else {
										return nil, false
									}
								}
							} else {
								return nil, false
							}
						} else {
							return nil, false
						}
					}
					if len(policies) > 0 {
						self[nodeTemplateName_] = policies
					}
				} else {
					return nil, false
				}
			} else {
				return nil, false
			}
		}
		return self, true
	}
	return nil, false
}
