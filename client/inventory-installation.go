package client

import (
	"fmt"

	"github.com/tliron/kutil/kubernetes"
	"github.com/tliron/kutil/version"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (self *Client) InstallInventory(registry string, wait bool) error {
	var err error

	if registry, err = self.GetRegistry(registry); err != nil {
		return err
	}

	if _, err = self.createNamespace(); err != nil {
		return err
	}

	var serviceAccount *core.ServiceAccount
	if serviceAccount, err = self.createServiceAccount(); err != nil {
		return err
	}

	if _, err = self.createInventoryTlsSecret(); err != nil {
		return err
	}

	var inventoryDeployment *apps.Deployment
	if inventoryDeployment, err = self.createInventoryDeployment(registry, serviceAccount, 1); err != nil {
		return err
	}

	if _, err = self.createInventoryService(); err != nil {
		return err
	}

	if wait {
		if _, err := self.waitForDeployment(inventoryDeployment.Name); err != nil {
			return err
		}
	}

	return nil
}

func (self *Client) UninstallInventory(wait bool) {
	var gracePeriodSeconds int64 = 0
	deleteOptions := meta.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	// Inventory service
	if err := self.Kubernetes.CoreV1().Services(self.Namespace).Delete(self.Context, fmt.Sprintf("%s-inventory", self.NamePrefix), deleteOptions); err != nil {
		self.Log.Warningf("%s", err)
	}

	// Inventory deployment
	if err := self.Kubernetes.AppsV1().Deployments(self.Namespace).Delete(self.Context, fmt.Sprintf("%s-inventory", self.NamePrefix), deleteOptions); err != nil {
		self.Log.Warningf("%s", err)
	}

	// Inventory secret
	if err := self.Kubernetes.CoreV1().Secrets(self.Namespace).Delete(self.Context, fmt.Sprintf("%s-inventory", self.NamePrefix), deleteOptions); err != nil {
		self.Log.Warningf("%s", err)
	}

	// Service account
	if err := self.Kubernetes.CoreV1().ServiceAccounts(self.Namespace).Delete(self.Context, self.NamePrefix, deleteOptions); err != nil {
		self.Log.Warningf("%s", err)
	}

	if wait {
		self.WaitForDeletion("inventory service", func() bool {
			_, err := self.Kubernetes.CoreV1().Services(self.Namespace).Get(self.Context, fmt.Sprintf("%s-inventory", self.NamePrefix), meta.GetOptions{})
			return err == nil
		})
		self.WaitForDeletion("inventory deployment", func() bool {
			_, err := self.Kubernetes.AppsV1().Deployments(self.Namespace).Get(self.Context, fmt.Sprintf("%s-inventory", self.NamePrefix), meta.GetOptions{})
			return err == nil
		})
		self.WaitForDeletion("service account", func() bool {
			_, err := self.Kubernetes.CoreV1().ServiceAccounts(self.Namespace).Get(self.Context, self.NamePrefix, meta.GetOptions{})
			return err == nil
		})
	}
}

func (self *Client) createInventoryConfigMap() (*core.ConfigMap, error) {
	appName := fmt.Sprintf("%s-inventory", self.NamePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.Namespace)

	configMap := &core.ConfigMap{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "inventory",
				"app.kubernetes.io/part-of":    self.PartOf,
				"app.kubernetes.io/managed-by": self.ManagedBy,
			},
		},
	}

	if configMap, err := self.Kubernetes.CoreV1().ConfigMaps(self.Namespace).Create(self.Context, configMap, meta.CreateOptions{}); err == nil {
		return configMap, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.CoreV1().ConfigMaps(self.Namespace).Get(self.Context, appName, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createInventoryImagePullSecret(server string, username string, password string) (*core.Secret, error) {
	// See: https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod
	//      https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
	//      https://docs.docker.com/engine/reference/commandline/cli/#configjson-properties

	appName := fmt.Sprintf("%s-inventory", self.NamePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.Namespace)

	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "inventory",
				"app.kubernetes.io/part-of":    self.PartOf,
				"app.kubernetes.io/managed-by": self.ManagedBy,
			},
		},
	}

	if err := kubernetes.SetSecretDockerConfigJson(secret, server, username, password); err != nil {
		return nil, err
	}

	if secret, err := self.Kubernetes.CoreV1().Secrets(self.Namespace).Create(self.Context, secret, meta.CreateOptions{}); err == nil {
		return secret, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.CoreV1().Secrets(self.Namespace).Get(self.Context, appName, meta.GetOptions{})
	} else {
		return nil, err
	}
}

// See: https://nip.io/
//      https://cert-manager.io/docs/

func (self *Client) createInventoryTlsSecret() (*core.Secret, error) {
	appName := fmt.Sprintf("%s-inventory", self.NamePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.Namespace)

	var crt []byte
	var key []byte

	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "inventory",
				"app.kubernetes.io/part-of":    self.PartOf,
				"app.kubernetes.io/managed-by": self.ManagedBy,
			},
		},
		Type: core.SecretTypeTLS,
		Data: map[string][]byte{
			core.TLSCertKey:       crt,
			core.TLSPrivateKeyKey: key,
		},
	}

	if secret, err := self.Kubernetes.CoreV1().Secrets(self.Namespace).Create(self.Context, secret, meta.CreateOptions{}); err == nil {
		return secret, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.CoreV1().Secrets(self.Namespace).Get(self.Context, appName, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createInventoryDeployment(registry string, serviceAccount *core.ServiceAccount, replicas int32) (*apps.Deployment, error) {
	// https://hub.docker.com/_/registry
	// https://github.com/ContainerSolutions/trow
	// https://github.com/google/go-containerregistry

	appName := fmt.Sprintf("%s-inventory", self.NamePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.Namespace)

	deployment := &apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "inventory",
				"app.kubernetes.io/part-of":    self.PartOf,
				"app.kubernetes.io/managed-by": self.ManagedBy,
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: &replicas,
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":      appName,
					"app.kubernetes.io/instance":  instanceName,
					"app.kubernetes.io/version":   version.GitVersion,
					"app.kubernetes.io/component": "inventory",
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":       appName,
						"app.kubernetes.io/instance":   instanceName,
						"app.kubernetes.io/version":    version.GitVersion,
						"app.kubernetes.io/component":  "inventory",
						"app.kubernetes.io/part-of":    self.PartOf,
						"app.kubernetes.io/managed-by": self.ManagedBy,
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            "registry",
							Image:           fmt.Sprintf("%s/%s", registry, self.InventoryImageName),
							ImagePullPolicy: core.PullAlways,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "secret",
									MountPath: "/secret",
									ReadOnly:  true,
								},
								{
									Name:      "registry",
									MountPath: "/var/lib/registry",
								},
							},
							Env: []core.EnvVar{
								{
									// necessary!
									Name:  "REGISTRY_STORAGE_DELETE_ENABLED",
									Value: "true",
								},
								{
									Name:  "_REGISTRY_HTTP_TLS_CERTIFICATE",
									Value: "/secret/tls.crt",
								},
								{
									Name:  "_REGISTRY_HTTP_TLS_KEY",
									Value: "/secret/tls.key",
								},
							},
							LivenessProbe: &core.Probe{
								Handler: core.Handler{
									HTTPGet: &core.HTTPGetAction{
										Port: intstr.FromInt(5000),
									},
								},
							},
							ReadinessProbe: &core.Probe{
								Handler: core.Handler{
									HTTPGet: &core.HTTPGetAction{
										Port: intstr.FromInt(5000),
									},
								},
							},
						},
						{
							Name:            "spooler",
							Image:           fmt.Sprintf("%s/%s", registry, self.InventorySpoolerImageName),
							ImagePullPolicy: core.PullAlways,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "spool",
									MountPath: self.SpoolPath,
								},
							},
							Env: []core.EnvVar{
								{
									Name:  "REGISTRY_SPOOLER_directory",
									Value: self.SpoolPath,
								},
								{
									Name:  "REGISTRY_SPOOLER_registry",
									Value: "localhost:5000",
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
							Name: "secret",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: appName,
								},
							},
						},
						{
							Name:         "registry",
							VolumeSource: self.CreateVolumeSource("1Gi"),
						},
						{
							Name:         "spool",
							VolumeSource: self.CreateVolumeSource("1Gi"),
						},
					},
				},
			},
		},
	}

	return self.CreateDeployment(deployment)
}

func (self *Client) createInventoryService() (*core.Service, error) {
	appName := fmt.Sprintf("%s-inventory", self.NamePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.Namespace)

	service := &core.Service{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "inventory",
				"app.kubernetes.io/part-of":    self.PartOf,
				"app.kubernetes.io/managed-by": self.ManagedBy,
			},
		},
		Spec: core.ServiceSpec{
			Type: core.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app.kubernetes.io/name":      appName,
				"app.kubernetes.io/instance":  instanceName,
				"app.kubernetes.io/version":   version.GitVersion,
				"app.kubernetes.io/component": "inventory",
			},
			Ports: []core.ServicePort{
				{
					Name:       "registry",
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(5000),
					Port:       5000,
				},
			},
		},
	}

	if service, err := self.Kubernetes.CoreV1().Services(self.Namespace).Create(self.Context, service, meta.CreateOptions{}); err == nil {
		return service, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.CoreV1().Services(self.Namespace).Get(self.Context, appName, meta.GetOptions{})
	} else {
		return nil, err
	}
}
