package v1alpha1

import (
	"fmt"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/kubernetes"
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
	ServiceShortName = "ts" // = Turandot (or TOSCA) Service

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
	ServiceTemplate ServiceTemplate   `json:"serviceTemplate"` // Service template used to instantiate this service (can be a direct or indirect reference)
	Inputs          map[string]string `json:"inputs"`          // TOSCA inputs to apply to the service template during instantiation
	Mode            string            `json:"mode"`            // Desired service mode
}

type ServiceTemplate struct {
	Direct   *ServiceTemplateDirect   `json:"direct,omitempty"`   // Direct reference to the service template used to instantiate this service
	Indirect *ServiceTemplateIndirect `json:"indirect,omitempty"` // Indirect reference to the service template used to instantiate this service
}

type ServiceTemplateDirect struct {
	URL              string `json:"url"`                        // Full URL of service template (CSAR or YAML file)
	TLSSecret        string `json:"tlsSecret,omitempty"`        // Name of TLS Secret required for connecting to the URL (optional)
	TLSSecretDataKey string `json:"tlsSecretDataKey,omitempty"` // Name of key within the TLS Secret data required for connecting to the URL (optional)
	AuthSecret       string `json:"authSecret,omitempty"`       // Name of authentication Secret required for connecting to the URL (optional)
}

type ServiceTemplateIndirect struct {
	Namespace  string `json:"namespace,omitempty"` // Namespace for Turandot repository resource (optional; defaults to same namespace as this service)
	Repository string `json:"repository"`          // Name of Turandot repository resource
	Name       string `json:"name"`                // Name of service template artifact in the repository (CSAR or YAML artifact)
}

type ServiceStatus struct {
	CloutPath string `json:"cloutPath"` // Path to instantiated service's Clout file (local to Turandot operator)
	CloutHash string `json:"cloutHash"` // Last known hash of service's Clout file

	ServiceTemplateURL string            `json:"serviceTemplateUrl"` // Full URL of service template (CSAR or YAML file)
	Inputs             map[string]string `json:"inputs"`             // TOSCA inputs that were applied to the service template when instantiatied
	Outputs            map[string]string `json:"outputs"`            // Last known TOSCA outputs

	InstantiationState ServiceInstantiationState       `json:"instantiationState"` // Current service instantiation state
	Mode               string                          `json:"mode"`               // Current service mode
	NodeStates         map[string]ServiceNodeModeState `json:"nodeStates"`         // Last known node states
}

type ServiceNodeModeState struct {
	Mode    string    `json:"mode"`    // Service mode
	State   ModeState `json:"state"`   // Node state for service mode
	Message string    `json:"message"` // Human-readable information regarding the node state (optional)
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
						Description: "Turandot service",
						Type:        "object",
						Required:    []string{"spec"},
						Properties: map[string]apiextensions.JSONSchemaProps{
							"spec": {
								Type:     "object",
								Required: []string{"serviceTemplate"},
								Properties: map[string]apiextensions.JSONSchemaProps{
									"serviceTemplate": {
										Description: "Service template used to instantiate this service (can be a direct or indirect reference)",
										Type:        "object",
										Properties: map[string]apiextensions.JSONSchemaProps{
											"direct": {
												Description: "Direct reference to the service template used to instantiate this service",
												Type:        "object",
												Required:    []string{"url"},
												Properties: map[string]apiextensions.JSONSchemaProps{
													"url": {
														Description: "Full URL of service template (CSAR or YAML file)",
														Type:        "string",
													},
													"tlsSecret": {
														Description: "Name of TLS Secret required for connecting to the URL (optional)",
														Type:        "string",
													},
													"tlsSecretDataKey": {
														Description: "Name of key within the TLS Secret data required for connecting to the repository (optional)",
														Type:        "string",
													},
													"authSecret": {
														Description: "Name of authentication Secret required for connecting to the URL (optional)",
														Type:        "string",
													},
												},
											},
											"indirect": {
												Description: "Indirect reference to the service template used to instantiate this service",
												Type:        "object",
												Required:    []string{"repository", "name"},
												Properties: map[string]apiextensions.JSONSchemaProps{
													"namespace": {
														Description: "Namespace for Turandot repository resource (optional; defaults to same namespace as this service)",
														Type:        "string",
													},
													"repository": {
														Description: "Name of Turandot repository resource",
														Type:        "string",
													},
													"name": {
														Description: "Name of service template artifact in the repository (CSAR or YAML artifact)",
														Type:        "string",
													},
												},
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
									"inputs": {
										Description: "TOSCA inputs to apply to the service template during instantiation",
										Type:        "object",
										Nullable:    true,
										AdditionalProperties: &apiextensions.JSONSchemaPropsOrBool{
											Schema: &apiextensions.JSONSchemaProps{
												Type: "string",
											},
										},
									},
									"mode": {
										Description: "Desired service mode",
										Type:        "string",
									},
								},
							},
							"status": {
								Type: "object",
								Properties: map[string]apiextensions.JSONSchemaProps{
									"cloutPath": {
										Description: "Path to instantiated service's Clout file (local to Turandot operator)",
										Type:        "string",
									},
									"cloutHash": {
										Description: "Last known hash of service's Clout file",
										Type:        "string",
									},
									"serviceTemplateUrl": {
										Description: "Full URL of service template (CSAR or YAML file)",
										Type:        "string",
									},
									"inputs": {
										Description: "TOSCA inputs that were applied to the service template when instantiatied",
										Type:        "object",
										Nullable:    true,
										AdditionalProperties: &apiextensions.JSONSchemaPropsOrBool{
											Schema: &apiextensions.JSONSchemaProps{
												Type: "string",
											},
										},
									},
									"outputs": {
										Description: "Last known TOSCA outputs",
										Type:        "object",
										Nullable:    true,
										AdditionalProperties: &apiextensions.JSONSchemaPropsOrBool{
											Schema: &apiextensions.JSONSchemaProps{
												Type: "string",
											},
										},
									},
									"instantiationState": {
										Description: "Current service instantiation state",
										Type:        "string",
										Enum: []apiextensions.JSON{
											kubernetes.JSONString(ServiceNotInstantiated),
											kubernetes.JSONString(ServiceInstantiating),
											kubernetes.JSONString(ServiceInstantiated),
										},
									},
									"mode": {
										Description: "Current service mode",
										Type:        "string",
									},
									"nodeStates": {
										Description: "Last known node states",
										Type:        "object",
										Nullable:    true,
										AdditionalProperties: &apiextensions.JSONSchemaPropsOrBool{
											Schema: &apiextensions.JSONSchemaProps{
												Type: "object",
												Properties: map[string]apiextensions.JSONSchemaProps{
													"mode": {
														Description: "Service mode",
														Type:        "string",
													},
													"state": {
														Description: "Node state for service mode",
														Type:        "string",
														Enum: []apiextensions.JSON{
															kubernetes.JSONString(ModeAccepted),
															kubernetes.JSONString(ModeRejected),
															kubernetes.JSONString(ModeAchieved),
															kubernetes.JSONString(ModeFailed),
														},
													},
													"message": {
														Description: "Human-readable information regarding the node state (optional)",
														Type:        "string",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
					{
						Name:     "ServiceTemplateUrl",
						Type:     "string",
						JSONPath: ".status.serviceTemplateUrl",
					},
					{
						Name:     "Mode",
						Type:     "string",
						JSONPath: ".status.mode",
					},
				},
			},
		},
	},
}

func ServiceToARD(service *Service) ard.StringMap {
	map_ := make(ard.StringMap)
	map_["Name"] = service.Name
	if service.Spec.ServiceTemplate.Direct != nil {
		map_["ServiceTemplate"] = ard.StringMap{
			"Direct": ard.StringMap{
				"URL":              service.Spec.ServiceTemplate.Direct.URL,
				"TLSSecret":        service.Spec.ServiceTemplate.Direct.TLSSecret,
				"TLSSecretDataKey": service.Spec.ServiceTemplate.Direct.TLSSecretDataKey,
				"AuthSecret":       service.Spec.ServiceTemplate.Direct.AuthSecret,
			},
		}
	} else if service.Spec.ServiceTemplate.Direct != nil {
		map_["ServiceTemplate"] = ard.StringMap{
			"Indirect": ard.StringMap{
				"Namespace":  service.Spec.ServiceTemplate.Indirect.Namespace,
				"Repository": service.Spec.ServiceTemplate.Indirect.Repository,
				"Name":       service.Spec.ServiceTemplate.Indirect.Name,
			},
		}
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
