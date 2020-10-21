package v1alpha1

import (
	"fmt"

	"github.com/tliron/kutil/ard"
	group "github.com/tliron/turandot/resources/turandot.puccini.cloud"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ServiceGVK = SchemeGroupVersion.WithKind(ServiceKind)

type ServiceInstantiationState string

type ModeState string

const (
	ServiceKind     = "Service"
	ServiceListKind = "ServiceList"

	ServiceSingular  = "service"
	ServicePlural    = "services"
	ServiceShortName = "si" // = ServIce? Service Instance?

	ServiceNotInstantiated ServiceInstantiationState = "NotInstantiated"
	ServiceInstantiating   ServiceInstantiationState = "Instantiating"
	ServiceInstantiated    ServiceInstantiationState = "Instantiated"

	ModeAccepted ModeState = "Accepted"
	ModeRejected ModeState = "Rejected"
	ModeAchieved ModeState = "Achieved"
	ModeFailed   ModeState = "Failed"
)

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
	ServiceTemplate ServiceTemplate   `json:"serviceTemplate"`
	Inputs          map[string]string `json:"inputs"`
	Mode            string            `json:"mode"`
}

type ServiceTemplate struct {
	Direct   ServiceTemplateDirect   `json:"direct"`
	Indirect ServiceTemplateIndirect `json:"indirect"`
}

type ServiceTemplateDirect struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

type ServiceTemplateIndirect struct {
	Inventory string `json:"inventory"`
	Name      string `json:"name"`
}

type ServiceStatus struct {
	CloutPath string `json:"cloutPath"`
	CloutHash string `json:"cloutHash"`

	ServiceTemplateURL string            `json:"serviceTemplateUrl"`
	Inputs             map[string]string `json:"inputs"`
	Outputs            map[string]string `json:"outputs"`

	InstantiationState ServiceInstantiationState       `json:"instantiationState"`
	NodeStates         map[string]ServiceNodeModeState `json:"nodeStates"`
	Mode               string                          `json:"mode"`
}

type ServiceNodeModeState struct {
	Mode    string    `json:"mode"`
	State   ModeState `json:"state"`
	Message string    `json:"message"`
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

var one int64 = 1

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
								Required: []string{"serviceTemplate"},
								Properties: map[string]apiextensions.JSONSchemaProps{
									"serviceTemplate": {
										Type: "object",
										OneOf: []apiextensions.JSONSchemaProps{
											{
												Properties: map[string]apiextensions.JSONSchemaProps{
													"direct": {
														Properties: map[string]apiextensions.JSONSchemaProps{
															"url": {
																MinLength: &one,
															},
														},
													},
												},
											},
											{
												Properties: map[string]apiextensions.JSONSchemaProps{
													"indirect": {
														Properties: map[string]apiextensions.JSONSchemaProps{
															"inventory": {
																MinLength: &one,
															},
															"name": {
																MinLength: &one,
															},
														},
													},
												},
											},
										},
										Properties: map[string]apiextensions.JSONSchemaProps{
											"direct": {
												Type: "object",
												//Required: []string{"url"},
												Properties: map[string]apiextensions.JSONSchemaProps{
													"url": {
														Type: "string",
													},
													"secret": {
														Type: "string",
													},
												},
											},
											"indirect": {
												Type: "object",
												//Required: []string{"inventory", "name"},
												Properties: map[string]apiextensions.JSONSchemaProps{
													"inventory": {
														Type: "string",
													},
													"name": {
														Type: "string",
													},
												},
											},
										},
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
									"mode": {
										Type: "string",
									},
								},
							},
							"status": {
								Type: "object",
								Properties: map[string]apiextensions.JSONSchemaProps{
									"cloutPath": {
										Type: "string",
									},
									"cloutHash": {
										Type: "string",
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
									"instantiationState": {
										Type: "string",
										Enum: []apiextensions.JSON{
											{Raw: []byte(fmt.Sprintf("%q", ServiceNotInstantiated))},
											{Raw: []byte(fmt.Sprintf("%q", ServiceInstantiating))},
											{Raw: []byte(fmt.Sprintf("%q", ServiceInstantiated))},
										},
									},
									"nodeStates": {
										Type:     "object",
										Nullable: true,
										AdditionalProperties: &apiextensions.JSONSchemaPropsOrBool{
											Schema: &apiextensions.JSONSchemaProps{
												Type: "object",
												Properties: map[string]apiextensions.JSONSchemaProps{
													"mode": {
														Type: "string",
													},
													"state": {
														Type: "string",
														Enum: []apiextensions.JSON{
															{Raw: []byte(fmt.Sprintf("%q", ModeAccepted))},
															{Raw: []byte(fmt.Sprintf("%q", ModeRejected))},
															{Raw: []byte(fmt.Sprintf("%q", ModeAchieved))},
															{Raw: []byte(fmt.Sprintf("%q", ModeFailed))},
														},
													},
													"message": {
														Type: "string",
													},
												},
											},
										},
									},
									"mode": {
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

func ServiceToARD(service *Service) ard.StringMap {
	map_ := make(ard.StringMap)
	map_["Name"] = service.Name
	map_["ServiceTemplate"] = ard.StringMap{
		"Direct": ard.StringMap{
			"URL":    service.Spec.ServiceTemplate.Direct.URL,
			"Secret": service.Spec.ServiceTemplate.Direct.Secret,
		},
		"Indirect": ard.StringMap{
			"Inventory": service.Spec.ServiceTemplate.Indirect.Inventory,
			"Name":      service.Spec.ServiceTemplate.Indirect.Name,
		},
	}
	map_["Inputs"] = service.Spec.Inputs
	map_["Outputs"] = service.Status.Outputs
	map_["InstantiationState"] = service.Status.InstantiationState
	map_["CloutPath"] = service.Status.CloutPath
	map_["CloutHash"] = service.Status.CloutHash
	map_["Mode"] = service.Status.Mode
	nodeStates := make(ard.StringMap)
	if service.Status.NodeStates != nil {
		for node, nodeState := range service.Status.NodeStates {
			nodeStates[node] = ard.StringMap{
				"Mode":    nodeState.Mode,
				"State":   nodeState.State,
				"Message": nodeState.Message,
			}
		}
	}
	map_["NodeStates"] = nodeStates
	return map_
}
