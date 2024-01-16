package parser

import (
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
	"gopkg.in/yaml.v3"
)

//
// OrchestrationProvisioningPolicy
//

type OrchestrationProvisioningPolicy struct {
	Sites              []string
	Profile            bool
	Substitutable      bool
	Instantiable       bool
	Virtualizable      bool
	SubstitutionInputs map[string]any
}

func ParseOrchestrationProvisioningPolicy(value ard.Value) (*OrchestrationProvisioningPolicy, bool) {
	properties := ard.With(value)
	self := OrchestrationProvisioningPolicy{
		SubstitutionInputs: make(map[string]any),
	}
	var ok bool
	if self.Substitutable, ok = properties.Get("substitutable").NilMeansZero().Boolean(); !ok {
		return nil, false
	}
	if self.Instantiable, ok = properties.Get("instantiable").NilMeansZero().Boolean(); !ok {
		return nil, false
	}
	if self.Virtualizable, ok = properties.Get("virtualizable").NilMeansZero().Boolean(); !ok {
		return nil, false
	}
	if sites := properties.Get("sites"); sites != ard.NoNode {
		if sites_, ok := sites.NilMeansZero().List(); ok {
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
		if substitutionInputs_, ok := substitutionInputs.NilMeansZero().Map(); ok {
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

type OrchestrationPolicies map[string][]any

func DecodeOrchestrationPolicies(code string) (OrchestrationPolicies, bool) {
	var policies ard.StringMap
	if err := yaml.Unmarshal(util.StringToBytes(code), &policies); err == nil {
		self := make(OrchestrationPolicies)
		for nodeTemplateName, nodePolicies := range policies {
			if nodePolicies_, ok := nodePolicies.(ard.List); ok {
				var policies []any
				for _, policy := range nodePolicies_ {
					policy_ := ard.With(policy)
					if type_, ok := policy_.Get("type").String(); ok {
						if properties, ok := policy_.Get("properties").NilMeansZero().Map(); ok {
							switch type_ {
							case "provisioning":
								if policy__, ok := ParseOrchestrationProvisioningPolicy(properties); ok {
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
					self[nodeTemplateName] = policies
				}
			} else {
				return nil, false
			}
		}
		return self, true
	} else {
		return nil, false
	}
}
