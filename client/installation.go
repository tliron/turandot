package client

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

	if self.cluster {
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
	if err := self.kubernetes.AppsV1().Deployments(self.namespace).Delete(self.context, fmt.Sprintf("%s-inventory", self.namePrefix), meta.DeleteOptions{}); err != nil {
		log.Warningf("%s", err)
	}
	if err := self.kubernetes.AppsV1().Deployments(self.namespace).Delete(self.context, fmt.Sprintf("%s-operator", self.namePrefix), meta.DeleteOptions{}); err != nil {
		log.Warningf("%s", err)
	}
	if self.cluster {
		if err := self.kubernetes.RbacV1().ClusterRoleBindings().Delete(self.context, self.namePrefix, meta.DeleteOptions{}); err != nil {
			log.Warningf("%s", err)
		}
	} else {
		if err := self.kubernetes.RbacV1().RoleBindings(self.namespace).Delete(self.context, self.namePrefix, meta.DeleteOptions{}); err != nil {
			log.Warningf("%s", err)
		}
		if err := self.kubernetes.RbacV1().Roles(self.namespace).Delete(self.context, self.namePrefix, meta.DeleteOptions{}); err != nil {
			log.Warningf("%s", err)
		}
	}
	if err := self.kubernetes.CoreV1().ServiceAccounts(self.namespace).Delete(self.context, self.namePrefix, meta.DeleteOptions{}); err != nil {
		log.Warningf("%s", err)
	}
	if err := self.apiExtensions.ApiextensionsV1().CustomResourceDefinitions().Delete(self.context, resources.ServiceCustomResourceDefinition.Name, meta.DeleteOptions{}); err != nil {
		log.Warningf("%s", err)
	}
}

func (self *Client) createCustomResourceDefinition() (*apiextensions.CustomResourceDefinition, error) {
	customResourceDefinition := &resources.ServiceCustomResourceDefinition

	if customResourceDefinition, err := self.apiExtensions.ApiextensionsV1().CustomResourceDefinitions().Create(self.context, customResourceDefinition, meta.CreateOptions{}); err == nil {
		return customResourceDefinition, nil
	} else if errors.IsAlreadyExists(err) {
		return self.apiExtensions.ApiextensionsV1().CustomResourceDefinitions().Get(self.context, resources.ServiceCustomResourceDefinition.Name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createNamespace() (*core.Namespace, error) {
	namespace := &core.Namespace{
		ObjectMeta: meta.ObjectMeta{
			Name: self.namespace,
		},
	}

	if namespace, err := self.kubernetes.CoreV1().Namespaces().Create(self.context, namespace, meta.CreateOptions{}); err == nil {
		return namespace, nil
	} else if errors.IsAlreadyExists(err) {
		return self.kubernetes.CoreV1().Namespaces().Get(self.context, self.namespace, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createServiceAccount() (*core.ServiceAccount, error) {
	serviceAccount := &core.ServiceAccount{
		ObjectMeta: meta.ObjectMeta{
			Name: self.namePrefix,
		},
	}

	if serviceAccount, err := self.kubernetes.CoreV1().ServiceAccounts(self.namespace).Create(self.context, serviceAccount, meta.CreateOptions{}); err == nil {
		return serviceAccount, nil
	} else if errors.IsAlreadyExists(err) {
		return self.kubernetes.CoreV1().ServiceAccounts(self.namespace).Get(self.context, self.namePrefix, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createRole() (*rbac.Role, error) {
	role := &rbac.Role{
		ObjectMeta: meta.ObjectMeta{
			Name: self.namePrefix,
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{rbac.APIGroupAll},
				Resources: []string{rbac.ResourceAll},
				Verbs:     []string{rbac.VerbAll},
			},
		},
	}

	if role, err := self.kubernetes.RbacV1().Roles(self.namespace).Create(self.context, role, meta.CreateOptions{}); err == nil {
		return role, err
	} else if errors.IsAlreadyExists(err) {
		return self.kubernetes.RbacV1().Roles(self.namespace).Get(self.context, self.namePrefix, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createRoleBinding(serviceAccount *core.ServiceAccount, role *rbac.Role) (*rbac.RoleBinding, error) {
	roleBinding := &rbac.RoleBinding{
		ObjectMeta: meta.ObjectMeta{
			Name: self.namePrefix,
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind, // serviceAccount.Kind is empty
				Name:      serviceAccount.Name,
				Namespace: self.namespace, // required
			},
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName, // role.GroupVersionKind().Group is empty
			Kind:     "Role",         // role.Kind is empty
			Name:     role.Name,
		},
	}

	if roleBinding, err := self.kubernetes.RbacV1().RoleBindings(self.namespace).Create(self.context, roleBinding, meta.CreateOptions{}); err == nil {
		return roleBinding, nil
	} else if errors.IsAlreadyExists(err) {
		return self.kubernetes.RbacV1().RoleBindings(self.namespace).Get(self.context, self.namePrefix, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createClusterRoleBinding(serviceAccount *core.ServiceAccount) (*rbac.ClusterRoleBinding, error) {
	clusterRoleBinding := &rbac.ClusterRoleBinding{
		ObjectMeta: meta.ObjectMeta{
			Name: self.namePrefix,
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind, // serviceAccount.Kind is empty
				Name:      serviceAccount.Name,
				Namespace: self.namespace, // required
			},
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
	}

	if clusterRoleBinding, err := self.kubernetes.RbacV1().ClusterRoleBindings().Create(self.context, clusterRoleBinding, meta.CreateOptions{}); err == nil {
		return clusterRoleBinding, nil
	} else if errors.IsAlreadyExists(err) {
		return self.kubernetes.RbacV1().ClusterRoleBindings().Get(self.context, self.namePrefix, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) createOperatorDeployment(site string, registry string, serviceAccount *core.ServiceAccount, replicas int32) (*apps.Deployment, error) {
	appName := fmt.Sprintf("%s-operator", self.namePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.namespace)

	deployment := &apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "operator",
				"app.kubernetes.io/part-of":    self.partOf,
				"app.kubernetes.io/managed-by": self.managedBy,
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
						"app.kubernetes.io/part-of":    self.partOf,
						"app.kubernetes.io/managed-by": self.managedBy,
					},
				},
				Spec: core.PodSpec{
					ServiceAccountName: serviceAccount.Name,
					Containers: []core.Container{
						{
							Name:            "operator",
							Image:           fmt.Sprintf("%s/%s", registry, self.operatorImageName),
							ImagePullPolicy: core.PullAlways,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "cache",
									MountPath: "/cache",
								},
							},
							Env: []core.EnvVar{
								{
									Name:  "TURANDOT_OPERATOR_site",
									Value: site,
								},
								{
									Name:  "TURANDOT_OPERATOR_cache",
									Value: "/cache",
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

	appName := fmt.Sprintf("%s-inventory", self.namePrefix)
	instanceName := fmt.Sprintf("%s-%s", appName, self.namespace)

	deployment := &apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       appName,
				"app.kubernetes.io/instance":   instanceName,
				"app.kubernetes.io/version":    version.GitVersion,
				"app.kubernetes.io/component":  "inventory",
				"app.kubernetes.io/part-of":    self.partOf,
				"app.kubernetes.io/managed-by": self.managedBy,
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
						"app.kubernetes.io/part-of":    self.partOf,
						"app.kubernetes.io/managed-by": self.managedBy,
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            "registry",
							Image:           fmt.Sprintf("%s/%s", registry, self.inventoryImageName),
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
							Image:           fmt.Sprintf("%s/%s", registry, self.inventorySpoolerImageName),
							ImagePullPolicy: core.PullAlways,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "spool",
									MountPath: "/spool",
								},
							},
							Env: []core.EnvVar{
								{
									Name:  "REGISTRY_SPOOLER_directory",
									Value: "/spool",
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
