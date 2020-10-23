package v1alpha1

import (
	"fmt"

	group "github.com/tliron/turandot/resources/turandot.puccini.cloud"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var RepositoryGVK = SchemeGroupVersion.WithKind(RepositoryKind)

const (
	RepositoryKind     = "Repository"
	RepositoryListKind = "RepositoryList"

	RepositorySingular  = "repository"
	RepositoryPlural    = "repositories"
	RepositoryShortName = "repo"
)

//
// Repository
//

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Repository struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositorySpec   `json:"spec"`
	Status RepositoryStatus `json:"status"`
}

type RepositorySpec struct {
	URL     string `json:"url"`
	Service string `json:"service"`
	Secret  string `json:"secret"`
}

type RepositoryStatus struct {
	SpoolerPod string `json:"spoolerPod"`
}

//
// RepositoryList
//

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RepositoryList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`

	Items []Repository `json:"items"`
}

//
// RepositoryCustomResourceDefinition
//

// See: assets/custom-resource-definitions.yaml

var RepositoryResourcesName = fmt.Sprintf("%s.%s", RepositoryPlural, group.GroupName)

var RepositoryCustomResourceDefinition = apiextensions.CustomResourceDefinition{
	ObjectMeta: meta.ObjectMeta{
		Name: RepositoryResourcesName,
	},
	Spec: apiextensions.CustomResourceDefinitionSpec{
		Group: group.GroupName,
		Names: apiextensions.CustomResourceDefinitionNames{
			Singular: RepositorySingular,
			Plural:   RepositoryPlural,
			Kind:     RepositoryKind,
			ListKind: RepositoryListKind,
			ShortNames: []string{
				RepositoryShortName,
			},
			Categories: []string{
				"all", // will appear in "kubectl get all"
			},
		},
		Scope: apiextensions.NamespaceScoped,
		Versions: []apiextensions.CustomResourceDefinitionVersion{
			{
				Name:    Version,
				Served:  true,
				Storage: true, // one and only one version must be marked with storage=true
				Subresources: &apiextensions.CustomResourceSubresources{ // requires CustomResourceSubresources feature gate enabled
					Status: &apiextensions.CustomResourceSubresourceStatus{},
				},
				Schema: &apiextensions.CustomResourceValidation{
					OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
						Type:     "object",
						Required: []string{"spec"},
						Properties: map[string]apiextensions.JSONSchemaProps{
							"spec": {
								Type: "object",
								Properties: map[string]apiextensions.JSONSchemaProps{
									"url": {
										Type: "string",
									},
									"service": {
										Type: "string",
									},
									"secret": {
										Type: "string",
									},
								},
							},
							"status": {
								Type: "object",
								Properties: map[string]apiextensions.JSONSchemaProps{
									"spoolerPod": {
										Type: "string",
									},
								},
							},
						},
					},
				},
			},
		},
	},
}
