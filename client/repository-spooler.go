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

func (self *Client) CreateRepositorySpooler(repository *resources.Repository) (*core.Pod, error) {
	var address string
	address, err := self.GetRepositoryAddress(repository)
	if err != nil {
		return nil, err
	}

	registry := "docker.io"
	appName := self.GetRepositorySpoolerAppName(repository.Name)
	instanceName := fmt.Sprintf("%s-%s", appName, repository.Namespace)

	pod := &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:      appName,
			Namespace: repository.Namespace,
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
					Image:           fmt.Sprintf("%s/%s", registry, self.RepositorySpoolerImageName),
					ImagePullPolicy: core.PullAlways,
					VolumeMounts: []core.VolumeMount{
						{
							Name:      "spool",
							MountPath: "/spool",
						},
					},
					Env: []core.EnvVar{
						{
							Name:  "REGISTRY_SPOOLER_registry",
							Value: address,
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
					Name:         "spool",
					VolumeSource: self.CreateVolumeSource("1Gi"),
				},
			},
		},
	}

	if repository.Spec.Secret != "" {
		pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      "secret",
			MountPath: "/secret",
			ReadOnly:  true,
		})

		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, core.EnvVar{
			Name:  "REGISTRY_SPOOLER_certificate",
			Value: "/secret/tls.crt",
		})

		pod.Spec.Volumes = append(pod.Spec.Volumes, core.Volume{
			Name: "secret",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: repository.Spec.Secret,
				},
			},
		})
	}

	ownerReferences := pod.GetOwnerReferences()
	ownerReferences = append(ownerReferences, *meta.NewControllerRef(repository, repository.GroupVersionKind()))
	pod.SetOwnerReferences(ownerReferences)

	return self.CreatePod(pod)
}

func (self *Client) WaitForRepositorySpooler(namespace string, repositoryName string) (*core.Pod, error) {
	appName := self.GetRepositorySpoolerAppName(repositoryName)
	return self.WaitForPod(namespace, appName)
}

func (self *Client) GetRepositorySpoolerAppName(repositoryName string) string {
	return fmt.Sprintf("%s-repository-%s-spooler", self.NamePrefix, repositoryName)
}

func (self *Client) Spooler(repository *resources.Repository) *spoolerpkg.Client {
	appName := self.GetRepositorySpoolerAppName(repository.Name)

	return spoolerpkg.NewClient(
		self.Kubernetes,
		self.REST,
		self.Config,
		self.Context,
		nil,
		self.Namespace,
		appName,
		"spooler",
		"/spool",
	)
}

func (self *Client) SpoolerCommand(repository *resources.Repository) (*spoolerpkg.CommandClient, error) {
	spooler := self.Spooler(repository)

	certificate := ""
	if repository.Spec.Secret != "" {
		certificate = "/secret/tls.crt"
	}

	if address, err := self.GetRepositoryAddress(repository); err == nil {
		return spoolerpkg.NewCommandClient(
			spooler,
			address,
			certificate,
		), nil
	} else {
		return nil, err
	}
}
