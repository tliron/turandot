package v1alpha1

import (
	"fmt"

	group "github.com/tliron/turandot/resources/turandot.puccini.cloud"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceStatusString string

const (
	ServiceKind     = "Service"
	ServiceListKind = "ServiceList"

	ServiceSingular  = "service"
	ServicePlural    = "services"
	ServiceShortName = "si" // = ServIce? Service Instance?

	ServiceStatusNotInstantiated ServiceStatusString = "NotInstantiated"
	ServiceStatusInstantiating   ServiceStatusString = "Instantiating"
	ServiceStatusInstantiated    ServiceStatusString = "Instantiated"
)

var ServiceGVK = SchemeGroupVersion.WithKind(ServiceKind)

//
// Service
//

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Service struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec"`
	Status ServiceStatus `json:"status"`
}

type ServiceSpec struct {
	ServiceTemplateURL string            `json:"serviceTemplateUrl"`
	Inputs             map[string]string `json:"inputs"`
}

type ServiceStatus struct {
	Status             ServiceStatusString `json:"status"`
	ServiceTemplateURL string              `json:"serviceTemplateUrl"`
	Inputs             map[string]string   `json:"inputs"`
	Outputs            map[string]string   `json:"outputs"`
	CloutPath          string              `json:"cloutPath"`
	CloutHash          string              `json:"cloutHash"`
}

//
// ServiceList
//

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ServiceList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`

	Items []Service `json:"items"`
}

//
// ServiceCustomResourceDefinition
//

// See: assets/custom-resource-definitions.yaml

var ServiceResourcesName = fmt.Sprintf("%s.%s", ServicePlural, group.GroupName)

var ServiceCustomResourceDefinition = apiextensions.CustomResourceDefinition{
	ObjectMeta: meta.ObjectMeta{
		Name: ServiceResourcesName,
	},
	Spec: apiextensions.CustomResourceDefinitionSpec{
		Group: group.GroupName,
		Names: apiextensions.CustomResourceDefinitionNames{
			Singular: ServiceSingular,
			Plural:   ServicePlural,
			Kind:     ServiceKind,
			ListKind: ServiceListKind,
			ShortNames: []string{
				ServiceShortName,
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
								Type:     "object",
								Required: []string{"serviceTemplateUrl"},
								Properties: map[string]apiextensions.JSONSchemaProps{
									"serviceTemplateUrl": {
										Type: "string",
									},
									"inputs": {
										Type:     "object",
										Nullable: true,
										AdditionalProperties: &apiextensions.JSONSchemaPropsOrBool{
											Schema: &apiextensions.JSONSchemaProps{
												Type: "string",
											},
										},
									},
								},
							},
							"status": {
								Type: "object",
								Properties: map[string]apiextensions.JSONSchemaProps{
									"status": {
										Type: "string",
										Enum: []apiextensions.JSON{
											{Raw: []byte(fmt.Sprintf("\"%s\"", ServiceStatusNotInstantiated))},
											{Raw: []byte(fmt.Sprintf("\"%s\"", ServiceStatusInstantiating))},
											{Raw: []byte(fmt.Sprintf("\"%s\"", ServiceStatusInstantiated))},
										},
									},
									"serviceTemplateUrl": {
										Type: "string",
									},
									"inputs": {
										Type:     "object",
										Nullable: true,
										AdditionalProperties: &apiextensions.JSONSchemaPropsOrBool{
											Schema: &apiextensions.JSONSchemaProps{
												Type: "string",
											},
										},
									},
									"outputs": {
										Type:     "object",
										Nullable: true,
										AdditionalProperties: &apiextensions.JSONSchemaPropsOrBool{
											Schema: &apiextensions.JSONSchemaProps{
												Type: "string",
											},
										},
									},
									"cloutPath": {
										Type: "string",
									},
									"cloutHash": {
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
