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

const spoolPath = "/spool"

const spoolerContainerName = "spooler"

func (self *Client) CreateRepositorySpooler(repository *resources.Repository) (*core.Pod, error) {
	var address string
	address, err := self.GetRepositoryHost(repository)
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
					Name:            spoolerContainerName,
					Image:           fmt.Sprintf("%s/%s", registry, self.RepositorySpoolerImageName),
					ImagePullPolicy: core.PullAlways,
					VolumeMounts: []core.VolumeMount{
						{
							Name:      "spool",
							MountPath: spoolPath,
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
					VolumeSource: self.VolumeSource("1Gi"),
				},
			},
		},
	}

	if repository.Spec.TLSSecret != "" {
		pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      "tls",
			MountPath: tlsMountPath,
			ReadOnly:  true,
		})

		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, core.EnvVar{
			Name:  "REGISTRY_SPOOLER_certificate",
			Value: self.GetRepositoryCertificatePath(repository),
		})

		pod.Spec.Volumes = append(pod.Spec.Volumes, core.Volume{
			Name: "tls",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: repository.Spec.TLSSecret,
				},
			},
		})
	}

	if _, username, password, token, err := self.GetRepositoryAuth(repository); err == nil {
		if username != "" {
			pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, core.EnvVar{
				Name:  "REGISTRY_SPOOLER_username",
				Value: username,
			})
		}
		if password != "" {
			pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, core.EnvVar{
				Name:  "REGISTRY_SPOOLER_password",
				Value: password,
			})
		}
		if token != "" {
			pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, core.EnvVar{
				Name:  "REGISTRY_SPOOLER_token",
				Value: token,
			})
		}
	} else {
		return nil, err
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
	return fmt.Sprintf("%s-repository-spooler-%s", self.NamePrefix, repositoryName)
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
		spoolerContainerName,
		spoolPath,
	)
}

func (self *Client) SpoolerCommand(repository *resources.Repository) (*spoolerpkg.CommandClient, error) {
	spooler := self.Spooler(repository)

	var username string
	var password string
	var token string
	var err error
	if _, username, password, token, err = self.GetRepositoryAuth(repository); err != nil {
		return nil, err
	}

	if address, err := self.GetRepositoryHost(repository); err == nil {
		return spoolerpkg.NewCommandClient(
			spooler,
			address,
			self.GetRepositoryCertificatePath(repository),
			username,
			password,
			token,
		), nil
	} else {
		return nil, err
	}
}
