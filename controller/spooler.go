package controller

import (
	"fmt"

	"github.com/tliron/kutil/version"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (self *Controller) CreateSpooler(inventory *resources.Inventory) (*core.Pod, error) {
	registry := "docker.io"
	appName := fmt.Sprintf("%s-inventory-%s-spooler", self.Client.NamePrefix, inventory.Name)
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
				"app.kubernetes.io/part-of":    self.Client.PartOf,
				"app.kubernetes.io/managed-by": self.Client.ManagedBy,
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "spooler",
					Image:           fmt.Sprintf("%s/%s", registry, self.Client.InventorySpoolerImageName),
					ImagePullPolicy: core.PullAlways,
					VolumeMounts: []core.VolumeMount{
						{
							Name:      "spool",
							MountPath: self.Client.SpoolPath,
						},
					},
					Env: []core.EnvVar{
						{
							Name:  "REGISTRY_SPOOLER_directory",
							Value: self.Client.SpoolPath,
						},
						{
							Name:  "REGISTRY_SPOOLER_registry",
							Value: inventory.Spec.URL,
						},
						{
							Name:  "REGISTRY_SPOOLER_verbose",
							Value: "2",
						},
					},
					// TODO: next version of API?
					// See: https://github.com/kubernetes/enhancements/blob/master/keps/sig-apps/sidecarcontainers.md
					//      https://banzaicloud.com/blog/k8s-sidecars/
					// Lifecycle: &core.Lifecycle{Type: "sidecar"},
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
					Name:         "spool",
					VolumeSource: self.Client.CreateVolumeSource("1Gi"),
				},
			},
		},
	}

	ownerReferences := pod.GetOwnerReferences()
	ownerReferences = append(ownerReferences, *meta.NewControllerRef(inventory, inventory.GroupVersionKind()))
	pod.SetOwnerReferences(ownerReferences)

	return self.Client.CreatePod(pod)
}
