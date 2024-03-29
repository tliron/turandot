// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

// ServiceSpecApplyConfiguration represents an declarative configuration of the ServiceSpec type for use
// with apply.
type ServiceSpecApplyConfiguration struct {
	ServiceTemplate *ServiceTemplateApplyConfiguration `json:"serviceTemplate,omitempty"`
	Inputs          map[string]string                  `json:"inputs,omitempty"`
	Mode            *string                            `json:"mode,omitempty"`
}

// ServiceSpecApplyConfiguration constructs an declarative configuration of the ServiceSpec type for use with
// apply.
func ServiceSpec() *ServiceSpecApplyConfiguration {
	return &ServiceSpecApplyConfiguration{}
}

// WithServiceTemplate sets the ServiceTemplate field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ServiceTemplate field is set to the value of the last call.
func (b *ServiceSpecApplyConfiguration) WithServiceTemplate(value *ServiceTemplateApplyConfiguration) *ServiceSpecApplyConfiguration {
	b.ServiceTemplate = value
	return b
}

// WithInputs puts the entries into the Inputs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Inputs field,
// overwriting an existing map entries in Inputs field with the same key.
func (b *ServiceSpecApplyConfiguration) WithInputs(entries map[string]string) *ServiceSpecApplyConfiguration {
	if b.Inputs == nil && len(entries) > 0 {
		b.Inputs = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.Inputs[k] = v
	}
	return b
}

// WithMode sets the Mode field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Mode field is set to the value of the last call.
func (b *ServiceSpecApplyConfiguration) WithMode(value string) *ServiceSpecApplyConfiguration {
	b.Mode = &value
	return b
}
