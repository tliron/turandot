package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	group "github.com/tliron/turandot/resources/turandot.puccini.cloud"
)

// Group version used to register these objects
// Note: Generator *requires* it to be called "SchemeGroupVersion"
var SchemeGroupVersion = schema.GroupVersion{Group: group.GroupName, Version: Version}

// Takes an unqualified kind and returns a group-qualified GroupKind
// Note: Generator *requires* it to be called "Kind"
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Takes an unqualified resource and returns a group-qualified GroupResource
// Note: Generator *requires* it to be called "Resource"
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Registers this API group and version to a scheme
// Note: Generator *requires* it to be called "AddToScheme"
var AddToScheme = schemeBuilder.AddToScheme

var schemeBuilder = runtime.NewSchemeBuilder(func(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&Service{},
		&ServiceList{},
		&Repository{},
		&RepositoryList{},
	)
	meta.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
})
