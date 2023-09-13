// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

// ServiceTemplateDirectApplyConfiguration represents an declarative configuration of the ServiceTemplateDirect type for use
// with apply.
type ServiceTemplateDirectApplyConfiguration struct {
	URL              *string `json:"url,omitempty"`
	TLSSecret        *string `json:"tlsSecret,omitempty"`
	TLSSecretDataKey *string `json:"tlsSecretDataKey,omitempty"`
	AuthSecret       *string `json:"authSecret,omitempty"`
}

// ServiceTemplateDirectApplyConfiguration constructs an declarative configuration of the ServiceTemplateDirect type for use with
// apply.
func ServiceTemplateDirect() *ServiceTemplateDirectApplyConfiguration {
	return &ServiceTemplateDirectApplyConfiguration{}
}

// WithURL sets the URL field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the URL field is set to the value of the last call.
func (b *ServiceTemplateDirectApplyConfiguration) WithURL(value string) *ServiceTemplateDirectApplyConfiguration {
	b.URL = &value
	return b
}

// WithTLSSecret sets the TLSSecret field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLSSecret field is set to the value of the last call.
func (b *ServiceTemplateDirectApplyConfiguration) WithTLSSecret(value string) *ServiceTemplateDirectApplyConfiguration {
	b.TLSSecret = &value
	return b
}

// WithTLSSecretDataKey sets the TLSSecretDataKey field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLSSecretDataKey field is set to the value of the last call.
func (b *ServiceTemplateDirectApplyConfiguration) WithTLSSecretDataKey(value string) *ServiceTemplateDirectApplyConfiguration {
	b.TLSSecretDataKey = &value
	return b
}

// WithAuthSecret sets the AuthSecret field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the AuthSecret field is set to the value of the last call.
func (b *ServiceTemplateDirectApplyConfiguration) WithAuthSecret(value string) *ServiceTemplateDirectApplyConfiguration {
	b.AuthSecret = &value
	return b
}