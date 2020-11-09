package v1alpha1

import (
	"fmt"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/kubernetes"
	group "github.com/tliron/turandot/resources/turandot.puccini.cloud"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var RepositoryGVK = SchemeGroupVersion.WithKind(RepositoryKind)

type RepositoryType string

const (
	RepositoryKind     = "Repository"
	RepositoryListKind = "RepositoryList"

	RepositorySingular  = "repository"
	RepositoryPlural    = "repositories"
	RepositoryShortName = "repo"

	RepositoryTypeRegistry RepositoryType = "registry"
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
	Type             RepositoryType      `json:"type"`                       // Repository type
	Direct           *RepositoryDirect   `json:"direct,omitempty"`           // Direct reference to repository
	Indirect         *RepositoryIndirect `json:"indirect,omitempty"`         // Indirect reference to repository
	TLSSecret        string              `json:"tlsSecret,omitempty"`        // Name of TLS Secret required for connecting to the repository (optional)
	TLSSecretDataKey string              `json:"tlsSecretDataKey,omitempty"` // Name of key within the TLS Secret data required for connecting to the repository (optional)
	AuthSecret       string              `json:"authSecret,omitempty"`       // Name of authentication Secret required for connecting to the repository (optional)
}

type RepositoryDirect struct {
	Host string `json:"host"` // Repository host (either "host:port" or "host")
}

type RepositoryIndirect struct {
	Namespace string `json:"namespace,omitempty"` // Namespace for service resource (optional; defaults to same namespace as this repository)
	Service   string `json:"service"`             // Name of service resource
	Port      uint64 `json:"port"`                // TCP port to use with service
}

type RepositoryStatus struct {
	SpoolerPod string `json:"spoolerPod"` // Name of spooler pod resource (in the same namespace as this repository)
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
						Description: "Turandot repository",
						Type:        "object",
						Required:    []string{"spec"},
						Properties: map[string]apiextensions.JSONSchemaProps{
							"spec": {
								Type:     "object",
								Required: []string{"type"},
								Properties: map[string]apiextensions.JSONSchemaProps{
									"type": {
										Description: "Repository type",
										Type:        "string",
										Enum: []apiextensions.JSON{
											kubernetes.JSONString(RepositoryTypeRegistry),
										},
									},
									"direct": {
										Description: "Direct reference to repository",
										Type:        "object",
										Required:    []string{"host"},
										Properties: map[string]apiextensions.JSONSchemaProps{
											"host": {
												Description: "Repository host (either \"host:port\" or \"host\")",
												Type:        "string",
											},
										},
									},
									"indirect": {
										Description: "Indirect reference to repository",
										Type:        "object",
										Required:    []string{"service", "port"},
										Properties: map[string]apiextensions.JSONSchemaProps{
											"namespace": {
												Description: "Namespace for service resource (optional; defaults to same namespace as this repository)",
												Type:        "string",
											},
											"service": {
												Description: "Name of service resource",
												Type:        "string",
											},
											"port": {
												Description: "TCP port to use with service",
												Type:        "integer",
											},
										},
									},
									"tlsSecret": {
										Description: "Name of TLS Secret required for connecting to the repository (optional)",
										Type:        "string",
									},
									"tlsSecretDataKey": {
										Description: "Name of key within the TLS Secret data required for connecting to the repository (optional)",
										Type:        "string",
									},
									"authSecret": {
										Description: "Name of authentication Secret required for connecting to the repository (optional)",
										Type:        "string",
									},
								},
								OneOf: []apiextensions.JSONSchemaProps{
									{
										Required: []string{"direct"},
									},
									{
										Required: []string{"indirect"},
									},
								},
							},
							"status": {
								Type: "object",
								Properties: map[string]apiextensions.JSONSchemaProps{
									"spoolerPod": {
										Description: "Name of spooler pod resource (in the same namespace as this repository)",
										Type:        "string",
									},
								},
							},
						},
					},
				},
				AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
					{
						Name:     "Type",
						Type:     "string",
						JSONPath: ".spec.type",
					},
					{
						Name:     "SpoolerPod",
						Type:     "string",
						JSONPath: ".status.spoolerPod",
					},
				},
			},
		},
	},
}

func RepositoryToARD(repository *Repository) ard.StringMap {
	map_ := make(ard.StringMap)
	map_["Name"] = repository.Name
	map_["Type"] = repository.Spec.Type
	if repository.Spec.Direct != nil {
		map_["Direct"] = ard.StringMap{
			"Host": repository.Spec.Direct.Host,
		}
	} else if repository.Spec.Indirect != nil {
		map_["Indirect"] = ard.StringMap{
			"Namespace": repository.Spec.Indirect.Namespace,
			"Service":   repository.Spec.Indirect.Service,
			"Port":      repository.Spec.Indirect.Port,
		}
	}
	map_["TLSSecret"] = repository.Spec.TLSSecret
	map_["TLSSecretDataKey"] = repository.Spec.TLSSecretDataKey
	map_["AuthSecret"] = repository.Spec.AuthSecret
	map_["SpoolerPod"] = repository.Status.SpoolerPod
	return map_
}
