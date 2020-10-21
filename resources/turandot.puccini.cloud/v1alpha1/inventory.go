package v1alpha1

import (
	"fmt"

	group "github.com/tliron/turandot/resources/turandot.puccini.cloud"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var InventoryGVK = SchemeGroupVersion.WithKind(InventoryKind)

const (
	InventoryKind     = "Inventory"
	InventoryListKind = "InventoryList"

	InventorySingular  = "inventory"
	InventoryPlural    = "inventories"
	InventoryShortName = "inv"
)

//
// Inventory
//

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Inventory struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   InventorySpec   `json:"spec"`
	Status InventoryStatus `json:"status"`
}

type InventorySpec struct {
	URL     string `json:"url"`
	Service string `json:"service"`
	Secret  string `json:"secret"`
}

type InventoryStatus struct {
	SpoolerPod string `json:"spoolerPod"`
}

//
// InventoryList
//

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type InventoryList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`

	Items []Inventory `json:"items"`
}

//
// InventoryCustomResourceDefinition
//

// See: assets/custom-resource-definitions.yaml

var InventoryResourcesName = fmt.Sprintf("%s.%s", InventoryPlural, group.GroupName)

var InventoryCustomResourceDefinition = apiextensions.CustomResourceDefinition{
	ObjectMeta: meta.ObjectMeta{
		Name: InventoryResourcesName,
	},
	Spec: apiextensions.CustomResourceDefinitionSpec{
		Group: group.GroupName,
		Names: apiextensions.CustomResourceDefinitionNames{
			Singular: InventorySingular,
			Plural:   InventoryPlural,
			Kind:     InventoryKind,
			ListKind: InventoryListKind,
			ShortNames: []string{
				InventoryShortName,
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
