package client

import (
	"fmt"

	spoolerpkg "github.com/tliron/kubernetes-registry-spooler/client"
	"github.com/tliron/kutil/version"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (self *Client) CreateInventorySpooler(inventory *resources.Inventory) (*core.Pod, error) {
	registry := "docker.io"
	appName := self.GetInventorySpoolerAppName(inventory.Name)
	instanceName := fmt.Sprintf("%s-%s", appName, inventory.Namespace)

	pod := &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:      appName,
			Namespace: inventory.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "spooler",
				"app.kubernetes.io/part-of":    self.PartOf,
				"app.kubernetes.io/managed-by": self.ManagedBy,
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "spooler",
					Image:           fmt.Sprintf("%s/%s", registry, self.InventorySpoolerImageName),
					ImagePullPolicy: core.PullAlways,
					VolumeMounts: []core.VolumeMount{
						{
							Name:      "secret",
							MountPath: "/secret",
							ReadOnly:  true,
						},
						{
							Name:      "spool",
							MountPath: "/spool",
						},
					},
					Env: []core.EnvVar{
						{
							Name:  "REGISTRY_SPOOLER_registry",
							Value: inventory.Spec.URL,
						},
						{
							Name:  "REGISTRY_SPOOLER_certificate",
							Value: "/secret/tls.crt",
						},
						{
							Name:  "REGISTRY_SPOOLER_verbose",
							Value: "2",
						},
					},
					LivenessProbe: &core.Probe{
						Handler: core.Handler{
							HTTPGet: &core.HTTPGetAction{
								Port: intstr.FromInt(8086),
								Path: "/live",
							},
						},
					},
					ReadinessProbe: &core.Probe{
						Handler: core.Handler{
							HTTPGet: &core.HTTPGetAction{
								Port: intstr.FromInt(8086),
								Path: "/ready",
							},
						},
					},
				},
			},
			Volumes: []core.Volume{
				{
					Name: "secret",
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: inventory.Spec.Secret,
						},
					},
				},
				{
					Name:         "spool",
					VolumeSource: self.CreateVolumeSource("1Gi"),
				},
			},
		},
	}

	ownerReferences := pod.GetOwnerReferences()
	ownerReferences = append(ownerReferences, *meta.NewControllerRef(inventory, inventory.GroupVersionKind()))
	pod.SetOwnerReferences(ownerReferences)

	return self.CreatePod(pod)
}

func (self *Client) WaitForInventorySpooler(namespace string, inventoryName string) (*core.Pod, error) {
	appName := self.GetInventorySpoolerAppName(inventoryName)
	return self.WaitForPod(namespace, appName)
}

func (self *Client) GetInventorySpoolerAppName(inventoryName string) string {
	return fmt.Sprintf("%s-inventory-%s-spooler", self.NamePrefix, inventoryName)
}

func (self *Client) Spooler(inventoryName string) *spoolerpkg.Client {
	appName := self.GetInventorySpoolerAppName(inventoryName)

	return spoolerpkg.NewClient(
		self.Kubernetes,
		self.REST,
		self.Config,
		self.Namespace,
		appName,
		"spooler",
		"/spool",
		"/secret/tls.crt",
	)
}
