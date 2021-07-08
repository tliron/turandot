package main

import (
	"github.com/kubernetes-sigs/reference-docs/gen-apidocs/generators/api"
)

type Annotation struct {
	entity       string
	rename       string
	noReferences bool
	metadata     []string
	refine       map[string][]string
	parent       string
	children     map[string]Derive
}

type Derive struct {
	fields []string
}

var excludes = []string{
	"APIGroupList",    // doesn't have ListMeta, but is a list
	"APIResourceList", // doesn't have ListMeta, but is a list
	"PodTemplate",
	"ListMeta",
	"Time",
	"MicroTime",
}

var includes = []*api.Definition{
	{
		Name: "PodTemplateOwner",
		Fields: api.Fields{
			&api.Field{
				Name:        "selector",
				Type:        "LabelSelector",
				Description: "More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors",
			},
			&api.Field{
				Name:        "template",
				Type:        "PodTemplateSpec",
				Description: "More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller#pod-template",
			},
		},
	},
}

var annotations = map[string]Annotation{
	// Naming conflicts with Simple Profile
	"Node": {
		rename: "KNode",
	},

	// Use the built-in TOSCA type
	"Time": {
		rename: "timestamp",
	},
	"MicroTime": {
		rename: "timestamp",
	},

	// Special resources
	"ObjectMeta": {
		rename:       "Metadata",
		entity:       "capability",
		noReferences: true,
		refine: map[string][]string{
			"labels": {
				"type: map",
				"entry_schema: string",
			},
		},
	},
	"PersistentVolumeClaim": {
		entity: "data",
	},

	// Refinements
	"APIServiceSpec": {
		refine: map[string][]string{
			"versionPriority": {
				"constraints:",
				"- greater_than: 0",
			},
		},
	},
	"AzureDiskVolumeSource": {
		refine: map[string][]string{
			"kind": {
				"constraints:",
				"- valid_values: [ Shared, Dedicated, Managed ]",
			},
		},
	},
	"ConfigMap": {
		refine: map[string][]string{
			"binaryData": {
				"type: map",
				"entry_schema: string # base64-encoded",
			},
			"data": {
				"type: map",
				"entry_schema: string",
			},
		},
	},
	"ConfigMapVolumeSource": {
		refine: map[string][]string{
			"defaultMode": {"type: FileMode # instead of integer"},
		},
	},
	"ContainerPort": {
		refine: map[string][]string{
			"containerPort": {"type: Port # instead of integer"},
			"hostIP":        {"type: IPAddress # instead of string"},
			"hostPort":      {"type: Port # instead of integer"},
			"name":          {"type: ServiceName # instead of string"},
			"protocol": {
				"constraints:",
				"- valid_values: [ UDP, TCP, SCTP ]",
			},
		},
	},
	"CronJobSpec": {
		refine: map[string][]string{
			"concurrencyPolicy": {
				"constraints:",
				"- valid_values: [ Allow, Forbid, Replace ]",
			},
			"failedJobsHistoryLimit":     {"type: Count # instead of integer"},
			"successfulJobsHistoryLimit": {"type: Count # instead of integer"},
		},
	},
	"CustomResourceConversion": {
		refine: map[string][]string{
			"strategy": {
				"constraints:",
				"- valid_values: [ None, Webhook ]",
			},
		},
	},
	"CustomResourceDefinitionSpec": {
		refine: map[string][]string{
			"scope": {
				"constraints:",
				"- valid_values: [ Cluster, Namespaced ]",
			},
		},
	},
	"DeleteOptions": {
		refine: map[string][]string{
			"propagationPolicy": {
				"constraints:",
				"- valid_values: [ Orphan, Background, Foreground ]",
			},
		},
	},
	"Deployment": {
		metadata: []string{
			"turandot.copy.1: metadata.labels->spec.selector.matchLabels",
			"turandot.copy.2: metadata.labels->spec.template.metadata.labels",
			"turandot.copy.3: metadata.annotations->spec.template.metadata.annotations",
		},
	},
	"DownwardAPIVolumeSource": {
		refine: map[string][]string{
			"defaultMode": {"type: FileMode # instead of integer"},
		},
	},
	"DownwardAPIVolumeFile": {
		refine: map[string][]string{
			"mode": {"type: FileMode # instead of integer"},
		},
	},
	"EmptyDirVolumeSource": {
		refine: map[string][]string{
			"medium": {
				"constraints:",
				"- valid_values: [ '', Memory ]",
				"default: ''",
			},
		},
	},
	"Endpoint": {
		refine: map[string][]string{
			"hostname": {"type: Hostname # instead of string"},
		},
	},
	"EndpointAddress": {
		refine: map[string][]string{
			"ip": {"type: IPAddress # instead of string"},
		},
	},
	"EndpointPort": {
		refine: map[string][]string{
			"port": {"type: Port # instead of integer"},
			"protocol": {
				"constraints:",
				"- valid_values: [ UDP, TCP, SCTP ]",
			},
		},
	},
	"EndpointSlice": {
		refine: map[string][]string{
			"addressType": {
				"constraints:",
				"- valid_values: [ IPv4, IPv6, FQDN ]",
			},
		},
	},
	"HorizontalPodAutoscalerSpec": {
		refine: map[string][]string{
			"targetCPUUtilizationPercentage": {"type: Factor # instead of integer"},
		},
	},
	"HorizontalPodAutoscalerStatus": {
		refine: map[string][]string{
			"currentCPUUtilizationPercentage": {"type: Factor # instead of integer"},
		},
	},
	"HostAlias": {
		refine: map[string][]string{
			"ip": {"type: IPAddress # instead of string"},
		},
	},
	"HPAScalingRules": {
		refine: map[string][]string{
			"stabilizationWindowSeconds": {
				"constraints:",
				"- in_range: [ 0, 3600 ]",
			},
		},
	},
	"IPBlock": {
		refine: map[string][]string{
			"cidr":   {"type: IPSubnet # instead of string"},
			"except": {"entry_schema: IPSubnet # instead of string"},
		},
	},
	"JSONSchemaProps": {
		refine: map[string][]string{
			"x-kubernetes-list-type": {
				"constraints:",
				"- valid_values: [ atomic, set, map ]",
			},
			"x-kubernetes-map-type": {
				"constraints:",
				"- valid_values: [ granular, atomic ]",
			},
		},
	},
	"KeyToPath": {
		refine: map[string][]string{
			"mode": {"type: FileMode # instead of integer"},
		},
	},
	"LabelSelectorRequirement": {
		refine: map[string][]string{
			"operator": {
				"constraints:",
				"- valid_values: [ In, NotIn, Exists, DoesNotExist ]",
			},
		},
	},
	"LoadBalancerIngress": {
		refine: map[string][]string{
			"ip": {"type: IPAddress # instead of string"},
		},
	},
	"ManagedFieldsEntry": {
		refine: map[string][]string{
			"operation": {
				"constraints:",
				"- valid_values: [ Apply, Update ]",
			},
		},
	},
	"MetricTarget": {
		refine: map[string][]string{
			"averageUtilization": {"type: Factor # instead of integer"},
		},
	},
	"MetricValueStatus": {
		refine: map[string][]string{
			"averageUtilization": {"type: Factor # instead of integer"},
		},
	},
	"MutatingWebhook": {
		refine: map[string][]string{
			"failurePolicy": {
				"constraints:",
				"- valid_values: [ Ignore, Fail ]",
			},
			"matchPolicy": {
				"constraints:",
				"- valid_values: [ Exact, Equivalent ]",
			},
			"reinvocationPolicy": {
				"constraints:",
				"- valid_values: [ Never, IfNeeded ]",
			},
			"sideEffects": {
				"constraints:",
				"- valid_values: [ None, NoneOnDryRun ]",
			},
			"timeoutSeconds": {
				"constraints:",
				"- in_range: [ 1, 30 ]",
			},
		},
	},
	"NodeSelectorRequirement": {
		refine: map[string][]string{
			"operator": {
				"constraints:",
				"- valid_values: [ In, NotIn, Exists, DoesNotExist, Gt, Lt ]",
			},
		},
	},
	"PodDNSConfig": {
		refine: map[string][]string{
			"nameservers": {"entry_schema: IPAddress # instead of string"},
		},
	},
	"PodID": {
		refine: map[string][]string{
			"ip": {"type: IPAddress # instead of string"},
		},
	},
	"PodSecurityContext": {
		refine: map[string][]string{
			"fsGroupChangePolicy": {
				"constraints:",
				"- valid_values: [ OnRootMismatch, Always ]",
			},
		},
	},
	"PodSpec": {
		refine: map[string][]string{
			"dnsPolicy": {
				"constraints:",
				"- valid_values: [ ClusterFirstWithHostNet, ClusterFirst, Default, None ]",
			},
		},
	},
	"PodStatus": {
		refine: map[string][]string{
			"hostIP": {"type: IPAddress # instead of string"},
			"phase": {
				"constraints:",
				"- valid_values: [ Pending, Running, Succeeded, Unknown ]",
			},
			"podIP": {"type: IPAddress # instead of string"},
		},
	},
	"PortStatus": {
		refine: map[string][]string{
			"protocol": {
				"constraints:",
				"- valid_values: [ TCP, UDP, SCTP ]",
			},
		},
	},
	"Probe": {
		refine: map[string][]string{
			"successThreshold": {
				"constraints:",
				"- greater_than: 0",
			},
		},
	},
	"ProjectedVolumeSource": {
		refine: map[string][]string{
			"defaultMode": {"type: FileMode # instead of integer"},
		},
	},
	"ReplicationController": {
		metadata: []string{
			"turandot.copy.1: metadata.labels->spec.selector.matchLabels", // not a LabelSelector, so we can't inherit PodTemplateOwner
			"turandot.copy.2: metadata.labels->spec.template.metadata.labels",
			"turandot.copy.3: metadata.annotations->spec.template.metadata.annotations",
		},
	},

	"RollingUpdateDaemonSet": {
		refine: map[string][]string{
			"maxSurge":       {"type: Amount # instead of integer|string"},
			"maxUnavailable": {"type: Amount # instead of integer|string"},
		},
	},
	"RollingUpdateDeployment": {
		refine: map[string][]string{
			"maxSurge":       {"type: Amount # instead of integer|string"},
			"maxUnavailable": {"type: Amount # instead of integer|string"},
		},
	},
	"RuleWithOperations": {
		refine: map[string][]string{
			"operations": {
				"constraints:",
				"- min_length: 1",
			},
			"scope": {
				"constraints:",
				"- valid_values: [ Cluster, Namespaced, '*' ]",
			},
		},
	},
	"ScopedResourceSelectorRequirement": {
		refine: map[string][]string{
			"operator": {
				"constraints:",
				"- valid_values: [ In, NotIn, Exists, DoesNotExist ]",
			},
		},
	},
	"SecretVolumeSource": {
		refine: map[string][]string{
			"defaultMode": {"type: FileMode # instead of integer"},
		},
	},
	"Service": {
		metadata: []string{"turandot.copy: metadata.labels->spec.selector"},
		children: map[string]Derive{
			"ClusterIP":    {},
			"ExternalName": {},
			"LoadBalancer": {},
			"NodePort":     {},
		},
	},
	"ServiceSpec": {
		refine: map[string][]string{
			"clusterIP":   {"type: IPAddress # instead of string"},
			"clusterIPs":  {"entry_schema: IPAddress # instead of string"},
			"externalIPs": {"entry_schema: IPAddress # instead of string"},
			"externalTrafficPolicy": {
				"constraints:",
				"- valid_values: [ Local, Cluster ]",
			},
			"healthCheckNodePort": {"type: Port # instead of integer"},
			"internalTrafficPolicy": {
				"constraints:",
				"- valid_values: [ Local, Cluster ]",
			},
			"ipFamilies": {
				"constraints:",
				"- valid_values: [ IPv4, IPv6 ]",
			},
			"ipFamilyPolicy": {
				"constraints:",
				"- valid_values: [ SingleStack, PreferDualStack, RequireDualStack ]",
			},
			"sessionAffinity": {
				"constraints:",
				"- valid_values: [ ClientIP, None ]",
			},
			"type": {
				"constraints:",
				"- valid_values: [ ExternalName, ClusterIP, NodePort, LoadBalancer ]",
			},
		},
		children: map[string]Derive{
			"ClusterIPSpec": {},
			"ExternalNameSpec": {fields: []string{
				"externalName",
			}},
			"LoadBalancerSpec": {fields: []string{
				"allocateLoadBalancerNodePorts",
				"loadBalancerClass",
				"loadBalancerIP",
				"loadBalancerSourceRanges",
			}},
			"NodePortSpec": {},
		},
	},
	"ServicePort": {
		refine: map[string][]string{
			"targetPort": {"type: PortOrServiceName # instead of Any"},
		},
	},
	"Toleration": {
		refine: map[string][]string{
			"effect": {
				"constraints:",
				"- valid_values: [ NoSchedule, PreferNoSchedule, NoExecute ]",
			},
			"operator": {
				"constraints:",
				"- valid_values: [ Exists, Equal ]",
			},
		},
	},
	"ValidatingWebhook": {
		refine: map[string][]string{
			"failurePolicy": {
				"constraints:",
				"- valid_values: [ Ignore, Fail ]",
			},
			"matchPolicy": {
				"constraints:",
				"- valid_values: [ Exact, Equivalent ]",
			},
			"sideEffects": {
				"constraints:",
				"- valid_values: [ None, NoneOnDryRun ]",
			},
			"timeoutSeconds": {
				"constraints:",
				"- in_range: [ 1, 30 ]",
			},
		},
	},

	// Derivations
	"ClusterIP": {
		refine: map[string][]string{
			"spec": {"type: ClusterIPSpec"},
		},
	},
	"ClusterIPSpec": {
		refine: map[string][]string{
			"type": {"default: ClusterIP"},
		},
	},
	"ExternalName": {
		refine: map[string][]string{
			"spec": {"type: ExternalNameSpec"},
		},
	},
	"ExternalNameSpec": {
		refine: map[string][]string{
			"type": {"default: ExternalName"},
		},
	},
	"LoadBalancer": {
		refine: map[string][]string{
			"spec": {"type: LoadBalancerSpec"},
		},
	},
	"LoadBalancerSpec": {
		refine: map[string][]string{
			"type": {"default: LoadBalancer"},
		},
	},
	"NodePort": {
		refine: map[string][]string{
			"spec": {"type: NodePortSpec"},
		},
	},
	"NodePortSpec": {
		refine: map[string][]string{
			"type": {"default: NodePort"},
		},
	},

	"DaemonSetSpec": {
		parent: "PodTemplateOwner",
	},
	"DeploymentSpec": {
		parent: "PodTemplateOwner",
	},
	"JobSpec": {
		parent: "PodTemplateOwner",
	},
	"PodTemplateOwner": {
		entity: "data",
		metadata: []string{
			"turandot.copy.1: metadata.labels->spec.selector.matchLabels",
			"turandot.copy.2: metadata.labels->spec.template.metadata.labels",
			"turandot.copy.3: metadata.annotations->spec.template.metadata.annotations",
		},
	},
	"ReplicaSetSpec": {
		parent: "PodTemplateOwner",
	},
	"StatefulSetSpec": {
		parent: "PodTemplateOwner",
	},
}
