package delegate

import (
	"fmt"

	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	"github.com/tliron/turandot/version"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (self *Client) Install(site string, registry string, wait bool) error {
	var err error

	if registry, err = self.getRegistry(registry); err != nil {
		return err
	}

	if _, err = self.createCustomResourceDefinition(); err != nil {
		return err
	}

	if _, err = self.createNamespace(); err != nil {
		return err
	}

	var serviceAccount *core.ServiceAccount
	if serviceAccount, err = self.createServiceAccount(); err != nil {
		return err
	}

	if self.Cluster {
		if _, err = self.createClusterRoleBinding(serviceAccount); err != nil {
			return err
		}
	} else {
		var role *rbac.Role
		if role, err = self.createRole(); err != nil {
			return err
		}
		if _, err = self.createRoleBinding(serviceAccount, role); err != nil {
			return err
		}
	}

	var operatorDeployment *apps.Deployment
	if operatorDeployment, err = self.createOperatorDeployment(site, registry, serviceAccount, 1); err != nil {
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
		if _, err := self.waitForDeployment(operatorDeployment.Name); err != nil {
			return err
		}
		if _, err := self.waitForDeployment(inventoryDeployment.Name); err != nil {
			return err
		}
	}

	return nil
}

func (self *Client) Uninstall() {
	if err := self.Kubernetes.CoreV1().Services(self.Namespace).Delete(self.Context, fmt.Sprintf("%s-inventory", self.NamePrefix), meta.DeleteOptions{}); err != nil {
		self.Log.Warningf("%s", err)
	}
	if err := self.Kubernetes.AppsV1().Deployments(self.Namespace).Delete(self.Context, fmt.Sprintf("%s-inventory", self.NamePrefix), meta.DeleteOptions{}); err != nil {
		self.Log.Warningf("%s", err)
	}
	if err := self.Kubernetes.AppsV1().Deployments(self.Namespace).Delete(self.Context, fmt.Sprintf("%s-operator", self.NamePrefix), meta.DeleteOptions{}); err != nil {
		self.Log.Warningf("%s", err)
	}
	if self.Cluster {
		if err := self.Kubernetes.RbacV1().ClusterRoleBindings().Delete(self.Context, self.NamePrefix, meta.DeleteOptions{}); err != nil {
			self.Log.Warningf("%s", err)
		}
	} else {
		if err := self.Kubernetes.RbacV1().RoleBindings(self.Namespace).Delete(self.Context, self.NamePrefix, meta.DeleteOptions{}); err != nil {
			self.Log.Warningf("%s", err)
		}
		if err := self.Kubernetes.RbacV1().Roles(self.Namespace).Delete(self.Context, self.NamePrefix, meta.DeleteOptions{}); err != nil {
			self.Log.Warningf("%s", err)
		}
	}
	if err := self.Kubernetes.CoreV1().ServiceAccounts(self.Namespace).Delete(self.Context, self.NamePrefix, meta.DeleteOptions{}); err != nil {
		self.Log.Warningf("%s", err)
	}
	if err := self.APIExtensions.ApiextensionsV1().CustomResourceDefinitions().Delete(self.Context, resources.ServiceCustomResourceDefinition.Name, meta.DeleteOptions{}); err != nil {
		self.Log.Warningf("%s", err)
	}
}

func (self *Client) createCustomResourceDefinition() (*apiextensions.CustomResourceDefinition, error) {
	customResourceDefinition := &resources.ServiceCustomResourceDefinition

	if customResourceDefinition, err := self.APIExtensions.ApiextensionsV1().CustomResourceDefinitions().Create(self.Context, customResourceDefinition, meta.CreateOptions{}); err == nil {
		return customResourceDefinition, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.APIExtensions.ApiextensionsV1().CustomResourceDefinitions().Get(self.Context, resources.ServiceCustomResourceDefinition.Name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createNamespace() (*core.Namespace, error) {
	namespace := &core.Namespace{
		ObjectMeta: meta.ObjectMeta{
			Name: self.Namespace,
		},
	}

	if namespace, err := self.Kubernetes.CoreV1().Namespaces().Create(self.Context, namespace, meta.CreateOptions{}); err == nil {
		return namespace, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.CoreV1().Namespaces().Get(self.Context, self.Namespace, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createServiceAccount() (*core.ServiceAccount, error) {
	serviceAccount := &core.ServiceAccount{
		ObjectMeta: meta.ObjectMeta{
			Name: self.NamePrefix,
		},
	}

	if serviceAccount, err := self.Kubernetes.CoreV1().ServiceAccounts(self.Namespace).Create(self.Context, serviceAccount, meta.CreateOptions{}); err == nil {
		return serviceAccount, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.CoreV1().ServiceAccounts(self.Namespace).Get(self.Context, self.NamePrefix, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createRole() (*rbac.Role, error) {
	role := &rbac.Role{
		ObjectMeta: meta.ObjectMeta{
			Name: self.NamePrefix,
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{rbac.APIGroupAll},
				Resources: []string{rbac.ResourceAll},
				Verbs:     []string{rbac.VerbAll},
			},
		},
	}

	if role, err := self.Kubernetes.RbacV1().Roles(self.Namespace).Create(self.Context, role, meta.CreateOptions{}); err == nil {
		return role, err
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.RbacV1().Roles(self.Namespace).Get(self.Context, self.NamePrefix, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createRoleBinding(serviceAccount *core.ServiceAccount, role *rbac.Role) (*rbac.RoleBinding, error) {
	roleBinding := &rbac.RoleBinding{
		ObjectMeta: meta.ObjectMeta{
			Name: self.NamePrefix,
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind, // serviceAccount.Kind is empty
				Name:      serviceAccount.Name,
				Namespace: self.Namespace, // required
			},
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName, // role.GroupVersionKind().Group is empty
			Kind:     "Role",         // role.Kind is empty
			Name:     role.Name,
		},
	}

	if roleBinding, err := self.Kubernetes.RbacV1().RoleBindings(self.Namespace).Create(self.Context, roleBinding, meta.CreateOptions{}); err == nil {
		return roleBinding, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.RbacV1().RoleBindings(self.Namespace).Get(self.Context, self.NamePrefix, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createClusterRoleBinding(serviceAccount *core.ServiceAccount) (*rbac.ClusterRoleBinding, error) {
	clusterRoleBinding := &rbac.ClusterRoleBinding{
		ObjectMeta: meta.ObjectMeta{
			Name: self.NamePrefix,
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind, // serviceAccount.Kind is empty
				Name:      serviceAccount.Name,
				Namespace: self.Namespace, // required
			},
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
	}

	if clusterRoleBinding, err := self.Kubernetes.RbacV1().ClusterRoleBindings().Create(self.Context, clusterRoleBinding, meta.CreateOptions{}); err == nil {
		return clusterRoleBinding, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.RbacV1().ClusterRoleBindings().Get(self.Context, self.NamePrefix, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createOperatorDeployment(site string, registry string, serviceAccount *core.ServiceAccount, replicas int32) (*apps.Deployment, error) {
	appName := fmt.Sprintf("%s-operator", self.NamePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.Namespace)

	deployment := &apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "operator",
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
					"app.kubernetes.io/component": "operator",
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":       appName,
						"app.kubernetes.io/instance":   instanceName,
						"app.kubernetes.io/version":    version.GitVersion,
						"app.kubernetes.io/component":  "operator",
						"app.kubernetes.io/part-of":    self.PartOf,
						"app.kubernetes.io/managed-by": self.ManagedBy,
					},
				},
				Spec: core.PodSpec{
					ServiceAccountName: serviceAccount.Name,
					Containers: []core.Container{
						{
							Name:            "operator",
							Image:           fmt.Sprintf("%s/%s", registry, self.OperatorImageName),
							ImagePullPolicy: core.PullAlways,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "cache",
									MountPath: self.CachePath,
								},
							},
							Env: []core.EnvVar{
								{
									Name:  "TURANDOT_OPERATOR_site",
									Value: site,
								},
								{
									Name:  "TURANDOT_OPERATOR_cache",
									Value: self.CachePath,
								},
								{
									Name:  "TURANDOT_OPERATOR_concurrency",
									Value: "3",
								},
								{
									Name:  "TURANDOT_OPERATOR_verbose",
									Value: "1",
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
							Name: "cache",
							VolumeSource: core.VolumeSource{
								EmptyDir: &core.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	return self.createDeployment(deployment, appName)
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
							Env: []core.EnvVar{
								{
									// necessary!
									Name:  "REGISTRY_STORAGE_DELETE_ENABLED",
									Value: "true",
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
							Name: "spool",
							VolumeSource: core.VolumeSource{
								EmptyDir: &core.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	return self.createDeployment(deployment, appName)
}

func (self *Client) createDeployment(deployment *apps.Deployment, appName string) (*apps.Deployment, error) {
	if deployment, err := self.Kubernetes.AppsV1().Deployments(self.Namespace).Create(self.Context, deployment, meta.CreateOptions{}); err == nil {
		return deployment, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.AppsV1().Deployments(self.Namespace).Get(self.Context, appName, meta.GetOptions{})
	} else {
		return nil, err
	}
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
