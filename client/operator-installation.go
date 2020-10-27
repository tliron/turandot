package client

import (
	"fmt"

	"github.com/tliron/kutil/version"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (self *Client) InstallOperator(site string, registry string, wait bool) error {
	var err error

	if registry, err = self.GetRegistry(registry); err != nil {
		return err
	}

	if _, err = self.createServiceCustomResourceDefinition(); err != nil {
		return err
	}

	if _, err = self.createRepositoryCustomResourceDefinition(); err != nil {
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
		if _, err = self.createAdminClusterRoleBinding(serviceAccount); err != nil {
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
		if _, err = self.createViewClusterRoleBinding(serviceAccount); err != nil {
			return err
		}
	}

	var operatorDeployment *apps.Deployment
	if operatorDeployment, err = self.createOperatorDeployment(site, registry, serviceAccount, 1); err != nil {
		return err
	}

	if wait {
		if _, err := self.WaitForDeployment(self.Namespace, operatorDeployment.Name); err != nil {
			return err
		}
	}

	return nil
}

func (self *Client) UninstallOperator(wait bool) {
	var gracePeriodSeconds int64 = 0
	deleteOptions := meta.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	name := fmt.Sprintf("%s-operator", self.NamePrefix)

	// Deployment
	if err := self.Kubernetes.AppsV1().Deployments(self.Namespace).Delete(self.Context, name, deleteOptions); err != nil {
		self.Log.Warningf("%s", err)
	}

	// Cluster role binding
	if err := self.Kubernetes.RbacV1().ClusterRoleBindings().Delete(self.Context, self.NamePrefix, deleteOptions); err != nil {
		self.Log.Warningf("%s", err)
	}

	if !self.Cluster {
		// Role binding
		if err := self.Kubernetes.RbacV1().RoleBindings(self.Namespace).Delete(self.Context, self.NamePrefix, deleteOptions); err != nil {
			self.Log.Warningf("%s", err)
		}

		// Role
		if err := self.Kubernetes.RbacV1().Roles(self.Namespace).Delete(self.Context, self.NamePrefix, deleteOptions); err != nil {
			self.Log.Warningf("%s", err)
		}
	}

	// Service account
	if err := self.Kubernetes.CoreV1().ServiceAccounts(self.Namespace).Delete(self.Context, self.NamePrefix, deleteOptions); err != nil {
		self.Log.Warningf("%s", err)
	}

	// Service custom resource definition
	if err := self.APIExtensions.ApiextensionsV1().CustomResourceDefinitions().Delete(self.Context, resources.ServiceCustomResourceDefinition.Name, deleteOptions); err != nil {
		self.Log.Warningf("%s", err)
	}

	// Repository custom resource definition
	if err := self.APIExtensions.ApiextensionsV1().CustomResourceDefinitions().Delete(self.Context, resources.RepositoryCustomResourceDefinition.Name, deleteOptions); err != nil {
		self.Log.Warningf("%s", err)
	}

	if wait {
		getOptions := meta.GetOptions{}
		self.WaitForDeletion("operator deployment", func() bool {
			_, err := self.Kubernetes.AppsV1().Deployments(self.Namespace).Get(self.Context, name, getOptions)
			return err == nil
		})
		self.WaitForDeletion("cluster role binding", func() bool {
			_, err := self.Kubernetes.RbacV1().ClusterRoleBindings().Get(self.Context, self.NamePrefix, getOptions)
			return err == nil
		})
		if !self.Cluster {
			self.WaitForDeletion("role binding", func() bool {
				_, err := self.Kubernetes.RbacV1().RoleBindings(self.Namespace).Get(self.Context, self.NamePrefix, getOptions)
				return err == nil
			})
			self.WaitForDeletion("role", func() bool {
				_, err := self.Kubernetes.RbacV1().Roles(self.Namespace).Get(self.Context, self.NamePrefix, getOptions)
				return err == nil
			})
		}
		self.WaitForDeletion("service account", func() bool {
			_, err := self.Kubernetes.CoreV1().ServiceAccounts(self.Namespace).Get(self.Context, self.NamePrefix, getOptions)
			return err == nil
		})
		self.WaitForDeletion("service custom resource definition", func() bool {
			_, err := self.APIExtensions.ApiextensionsV1().CustomResourceDefinitions().Get(self.Context, resources.ServiceCustomResourceDefinition.Name, getOptions)
			return err == nil
		})
		self.WaitForDeletion("repository custom resource definition", func() bool {
			_, err := self.APIExtensions.ApiextensionsV1().CustomResourceDefinitions().Get(self.Context, resources.RepositoryCustomResourceDefinition.Name, getOptions)
			return err == nil
		})
	}
}

func (self *Client) createServiceCustomResourceDefinition() (*apiextensions.CustomResourceDefinition, error) {
	return self.createCustomResourceDefinition(&resources.ServiceCustomResourceDefinition)
}

func (self *Client) createRepositoryCustomResourceDefinition() (*apiextensions.CustomResourceDefinition, error) {
	return self.createCustomResourceDefinition(&resources.RepositoryCustomResourceDefinition)
}

func (self *Client) createCustomResourceDefinition(customResourceDefinition *apiextensions.CustomResourceDefinition) (*apiextensions.CustomResourceDefinition, error) {
	if customResourceDefinition, err := self.APIExtensions.ApiextensionsV1().CustomResourceDefinitions().Create(self.Context, customResourceDefinition, meta.CreateOptions{}); err == nil {
		return customResourceDefinition, nil
	} else if errors.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.APIExtensions.ApiextensionsV1().CustomResourceDefinitions().Get(self.Context, resources.ServiceCustomResourceDefinition.Name, meta.GetOptions{})
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

func (self *Client) createViewClusterRoleBinding(serviceAccount *core.ServiceAccount) (*rbac.ClusterRoleBinding, error) {
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
			Name:     "view",
		},
	}

	return self.createClusterRoleBinding(clusterRoleBinding)
}

func (self *Client) createAdminClusterRoleBinding(serviceAccount *core.ServiceAccount) (*rbac.ClusterRoleBinding, error) {
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

	return self.createClusterRoleBinding(clusterRoleBinding)
}

func (self *Client) createClusterRoleBinding(clusterRoleBinding *rbac.ClusterRoleBinding) (*rbac.ClusterRoleBinding, error) {
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
							Name:         "cache",
							VolumeSource: self.CreateVolumeSource("1Gi"),
						},
					},
				},
			},
		},
	}

	return self.CreateDeployment(deployment)
}
