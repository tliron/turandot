package client

import (
	"fmt"

	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanagermeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/tliron/kutil/version"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (self *Client) InstallRepository(registry string, wait bool) error {
	var err error

	secure := false
	if err = self.EnsureCertManager(); err == nil {
		secure = true
	} else {
		self.Log.Warningf("%s", err)
	}

	if registry, err = self.GetRegistry(registry); err != nil {
		return err
	}

	var serviceAccount *core.ServiceAccount
	if serviceAccount, err = self.getServiceAccount(); err != nil {
		return err
	}

	var repositoryDeployment *apps.Deployment
	if repositoryDeployment, err = self.createRepositoryDeployment(registry, serviceAccount, 1, secure); err != nil {
		return err
	}

	var service *core.Service
	if service, err = self.createRepositoryService(); err != nil {
		return err
	}

	if secure {
		var issuer *certmanager.Issuer
		if issuer, err = self.createRepositoryCertificateIssuer(); err != nil {
			return err
		}

		if _, err = self.createRepositoryCertificate(issuer, service); err != nil {
			return err
		}
	}

	if wait {
		if _, err := self.WaitForDeployment(self.Namespace, repositoryDeployment.Name); err != nil {
			return err
		}
	}

	return nil
}

func (self *Client) UninstallRepository(wait bool) {
	var gracePeriodSeconds int64 = 0
	deleteOptions := meta.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	name := fmt.Sprintf("%s-repository", self.NamePrefix)

	// Service
	if err := self.Kubernetes.CoreV1().Services(self.Namespace).Delete(self.Context, name, deleteOptions); err != nil {
		self.Log.Warningf("%s", err)
	}

	// Deployment
	if err := self.Kubernetes.AppsV1().Deployments(self.Namespace).Delete(self.Context, name, deleteOptions); err != nil {
		self.Log.Warningf("%s", err)
	}

	secure := false
	if err := self.EnsureCertManager(); err == nil {
		secure = true

		// Certificate
		if err := self.CertManager.CertmanagerV1().Certificates(self.Namespace).Delete(self.Context, name, deleteOptions); err != nil {
			self.Log.Warningf("%s", err)
		}

		// Issuer
		if err := self.CertManager.CertmanagerV1().Issuers(self.Namespace).Delete(self.Context, name, deleteOptions); err != nil {
			self.Log.Warningf("%s", err)
		}

		// Secret (deleting the Certificate will not delete the Secret!)
		if err := self.Kubernetes.CoreV1().Secrets(self.Namespace).Delete(self.Context, name, deleteOptions); err != nil {
			self.Log.Warningf("%s", err)
		}
	} else {
		self.Log.Warningf("%s", err)
	}

	if wait {
		getOptions := meta.GetOptions{}
		self.WaitForDeletion("repository service", func() bool {
			_, err := self.Kubernetes.CoreV1().Services(self.Namespace).Get(self.Context, name, getOptions)
			return err == nil
		})
		self.WaitForDeletion("repository deployment", func() bool {
			_, err := self.Kubernetes.AppsV1().Deployments(self.Namespace).Get(self.Context, name, getOptions)
			return err == nil
		})
		if secure {
			self.WaitForDeletion("repository certificate", func() bool {
				_, err := self.CertManager.CertmanagerV1().Certificates(self.Namespace).Get(self.Context, name, getOptions)
				return err == nil
			})
			self.WaitForDeletion("repository issuer", func() bool {
				_, err := self.CertManager.CertmanagerV1().Issuers(self.Namespace).Get(self.Context, name, getOptions)
				return err == nil
			})
			self.WaitForDeletion("repository secret", func() bool {
				_, err := self.Kubernetes.CoreV1().Secrets(self.Namespace).Get(self.Context, name, getOptions)
				return err == nil
			})
		}
	}
}

func (self *Client) createRepositoryDeployment(registry string, serviceAccount *core.ServiceAccount, replicas int32, secure bool) (*apps.Deployment, error) {
	// https://hub.docker.com/_/registry
	// https://github.com/ContainerSolutions/trow
	// https://github.com/google/go-containerregistry

	appName := fmt.Sprintf("%s-repository", self.NamePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.Namespace)

	deployment := &apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "repository",
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
					"app.kubernetes.io/component": "repository",
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":       appName,
						"app.kubernetes.io/instance":   instanceName,
						"app.kubernetes.io/version":    version.GitVersion,
						"app.kubernetes.io/component":  "repository",
						"app.kubernetes.io/part-of":    self.PartOf,
						"app.kubernetes.io/managed-by": self.ManagedBy,
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            "registry",
							Image:           fmt.Sprintf("%s/%s", registry, self.RepositoryImageName),
							ImagePullPolicy: core.PullAlways,
							VolumeMounts: []core.VolumeMount{
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
							},
							// Note: Probes skip certificate validation for HTTPS
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
					},
					Volumes: []core.Volume{
						{
							Name:         "registry",
							VolumeSource: self.CreateVolumeSource("1Gi"),
						},
					},
				},
			},
		},
	}

	if secure {
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      "secret",
			MountPath: "/secret",
			ReadOnly:  true,
		})

		deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env,
			core.EnvVar{
				Name:  "REGISTRY_HTTP_TLS_CERTIFICATE",
				Value: "/secret/tls.crt",
			},
			core.EnvVar{
				Name:  "REGISTRY_HTTP_TLS_KEY",
				Value: "/secret/tls.key",
			},
		)

		deployment.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.HTTPGet.Scheme = "HTTPS"
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.HTTPGet.Scheme = "HTTPS"

		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, core.Volume{
			Name: "secret",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: appName,
				},
			},
		})
	}

	return self.CreateDeployment(deployment)
}

func (self *Client) createRepositoryService() (*core.Service, error) {
	appName := fmt.Sprintf("%s-repository", self.NamePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.Namespace)

	service := &core.Service{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "repository",
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
				"app.kubernetes.io/component": "repository",
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

func (self *Client) createRepositoryCertificateIssuer() (*certmanager.Issuer, error) {
	appName := fmt.Sprintf("%s-repository", self.NamePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.Namespace)

	issuer := &certmanager.Issuer{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "certificate-issuer",
				"app.kubernetes.io/part-of":    self.PartOf,
				"app.kubernetes.io/managed-by": self.ManagedBy,
			},
		},
		Spec: certmanager.IssuerSpec{
			IssuerConfig: certmanager.IssuerConfig{
				SelfSigned: &certmanager.SelfSignedIssuer{},
			},
		},
	}

	if issuer, err := self.CertManager.CertmanagerV1().Issuers(self.Namespace).Create(self.Context, issuer, meta.CreateOptions{}); err == nil {
		return issuer, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.CertManager.CertmanagerV1().Issuers(self.Namespace).Get(self.Context, appName, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createRepositoryCertificate(issuer *certmanager.Issuer, service *core.Service) (*certmanager.Certificate, error) {
	appName := fmt.Sprintf("%s-repository", self.NamePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.Namespace)

	ipAddress := service.Spec.ClusterIP

	certificate := &certmanager.Certificate{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "certificate",
				"app.kubernetes.io/part-of":    self.PartOf,
				"app.kubernetes.io/managed-by": self.ManagedBy,
			},
		},
		Spec: certmanager.CertificateSpec{
			SecretName:  appName,
			IPAddresses: []string{ipAddress},
			URIs:        []string{"https://turandot.puccini.cloud"},
			IssuerRef: certmanagermeta.ObjectReference{
				Name: issuer.Name,
			},
		},
	}

	if certificate, err := self.CertManager.CertmanagerV1().Certificates(self.Namespace).Create(self.Context, certificate, meta.CreateOptions{}); err == nil {
		return certificate, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.CertManager.CertmanagerV1().Certificates(self.Namespace).Get(self.Context, appName, meta.GetOptions{})
	} else {
		return nil, err
	}
}

/*
func (self *Client) createRepositoryConfigMap() (*core.ConfigMap, error) {
	appName := fmt.Sprintf("%s-repository", self.NamePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.Namespace)

	configMap := &core.ConfigMap{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "repository",
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

func (self *Client) createRepositoryImagePullSecret(server string, username string, password string) (*core.Secret, error) {
	// See: https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod
	//      https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
	//      https://docs.docker.com/engine/reference/commandline/cli/#configjson-properties

	appName := fmt.Sprintf("%s-repository", self.NamePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.Namespace)

	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "repository",
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

func (self *Client) createRepositoryTlsSecret() (*core.Secret, error) {
	appName := fmt.Sprintf("%s-repository", self.NamePrefix)
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
				"app.kubernetes.io/component":  "repository",
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
*/
