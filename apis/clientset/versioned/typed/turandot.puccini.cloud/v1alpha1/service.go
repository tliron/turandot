// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	json "encoding/json"
	"fmt"
	"time"

	turandotpuccinicloudv1alpha1 "github.com/tliron/turandot/apis/applyconfiguration/turandot.puccini.cloud/v1alpha1"
	scheme "github.com/tliron/turandot/apis/clientset/versioned/scheme"
	v1alpha1 "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ServicesGetter has a method to return a ServiceInterface.
// A group's client should implement this interface.
type ServicesGetter interface {
	Services(namespace string) ServiceInterface
}

// ServiceInterface has methods to work with Service resources.
type ServiceInterface interface {
	Create(ctx context.Context, service *v1alpha1.Service, opts v1.CreateOptions) (*v1alpha1.Service, error)
	Update(ctx context.Context, service *v1alpha1.Service, opts v1.UpdateOptions) (*v1alpha1.Service, error)
	UpdateStatus(ctx context.Context, service *v1alpha1.Service, opts v1.UpdateOptions) (*v1alpha1.Service, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Service, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ServiceList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Service, err error)
	Apply(ctx context.Context, service *turandotpuccinicloudv1alpha1.ServiceApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.Service, err error)
	ApplyStatus(ctx context.Context, service *turandotpuccinicloudv1alpha1.ServiceApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.Service, err error)
	ServiceExpansion
}

// services implements ServiceInterface
type services struct {
	client rest.Interface
	ns     string
}

// newServices returns a Services
func newServices(c *TurandotV1alpha1Client, namespace string) *services {
	return &services{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the service, and returns the corresponding service object, and an error if there is any.
func (c *services) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Service, err error) {
	result = &v1alpha1.Service{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("services").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Services that match those selectors.
func (c *services) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ServiceList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ServiceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("services").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested services.
func (c *services) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("services").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a service and creates it.  Returns the server's representation of the service, and an error, if there is any.
func (c *services) Create(ctx context.Context, service *v1alpha1.Service, opts v1.CreateOptions) (result *v1alpha1.Service, err error) {
	result = &v1alpha1.Service{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("services").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(service).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a service and updates it. Returns the server's representation of the service, and an error, if there is any.
func (c *services) Update(ctx context.Context, service *v1alpha1.Service, opts v1.UpdateOptions) (result *v1alpha1.Service, err error) {
	result = &v1alpha1.Service{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("services").
		Name(service.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(service).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *services) UpdateStatus(ctx context.Context, service *v1alpha1.Service, opts v1.UpdateOptions) (result *v1alpha1.Service, err error) {
	result = &v1alpha1.Service{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("services").
		Name(service.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(service).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the service and deletes it. Returns an error if one occurs.
func (c *services) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("services").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *services) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("services").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched service.
func (c *services) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Service, err error) {
	result = &v1alpha1.Service{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("services").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// Apply takes the given apply declarative configuration, applies it and returns the applied service.
func (c *services) Apply(ctx context.Context, service *turandotpuccinicloudv1alpha1.ServiceApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.Service, err error) {
	if service == nil {
		return nil, fmt.Errorf("service provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(service)
	if err != nil {
		return nil, err
	}
	name := service.Name
	if name == nil {
		return nil, fmt.Errorf("service.Name must be provided to Apply")
	}
	result = &v1alpha1.Service{}
	err = c.client.Patch(types.ApplyPatchType).
		Namespace(c.ns).
		Resource("services").
		Name(*name).
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *services) ApplyStatus(ctx context.Context, service *turandotpuccinicloudv1alpha1.ServiceApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.Service, err error) {
	if service == nil {
		return nil, fmt.Errorf("service provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(service)
	if err != nil {
		return nil, err
	}

	name := service.Name
	if name == nil {
		return nil, fmt.Errorf("service.Name must be provided to Apply")
	}

	result = &v1alpha1.Service{}
	err = c.client.Patch(types.ApplyPatchType).
		Namespace(c.ns).
		Resource("services").
		Name(*name).
		SubResource("status").
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
