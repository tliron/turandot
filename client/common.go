package client

import (
	"errors"
	"fmt"

	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanagerpkg "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	"github.com/tliron/kutil/kubernetes"
	"github.com/tliron/kutil/version"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	errorspkg "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (self *Client) GetSourceRegistryHost(registryHost string) (string, error) {
	if registryHost == "internal" {
		if registryHost, err := kubernetes.GetInternalRegistryHost(self.Context, self.Kubernetes); err == nil {
			return registryHost, nil
		} else {
			return "", fmt.Errorf("could not discover internal registry: %s", err.Error())
		}
	} else if registryHost != "" {
		return registryHost, nil
	} else {
		return "", errors.New("must provide \"--registry\"")
	}
}

func (self *Client) GetCertManager() error {
	if self.CertManager == nil {
		var err error
		if self.CertManager, err = certmanagerpkg.NewForConfig(self.Config); err == nil {
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}
}

func (self *Client) GetOperatorServiceAccount() (*core.ServiceAccount, error) {
	return self.Kubernetes.CoreV1().ServiceAccounts(self.Namespace).Get(self.Context, self.NamePrefix, meta.GetOptions{})
}

func (self *Client) CreateDeployment(deployment *apps.Deployment) (*apps.Deployment, error) {
	name := deployment.Name
	if deployment, err := self.Kubernetes.AppsV1().Deployments(self.Namespace).Create(self.Context, deployment, meta.CreateOptions{}); err == nil {
		return deployment, nil
	} else if errorspkg.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.AppsV1().Deployments(self.Namespace).Get(self.Context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) CreatePod(pod *core.Pod) (*core.Pod, error) {
	name := pod.Name
	if pod, err := self.Kubernetes.CoreV1().Pods(self.Namespace).Create(self.Context, pod, meta.CreateOptions{}); err == nil {
		return pod, nil
	} else if errorspkg.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.CoreV1().Pods(self.Namespace).Get(self.Context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) CreateService(service *core.Service) (*core.Service, error) {
	name := service.Name
	if service, err := self.Kubernetes.CoreV1().Services(self.Namespace).Create(self.Context, service, meta.CreateOptions{}); err == nil {
		return service, nil
	} else if errorspkg.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.CoreV1().Services(self.Namespace).Get(self.Context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) CreateCustomResourceDefinition(customResourceDefinition *apiextensions.CustomResourceDefinition) (*apiextensions.CustomResourceDefinition, error) {
	name := customResourceDefinition.Name
	if customResourceDefinition, err := self.APIExtensions.ApiextensionsV1().CustomResourceDefinitions().Create(self.Context, customResourceDefinition, meta.CreateOptions{}); err == nil {
		return customResourceDefinition, nil
	} else if errorspkg.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.APIExtensions.ApiextensionsV1().CustomResourceDefinitions().Get(self.Context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) CreateRole(role *rbac.Role) (*rbac.Role, error) {
	name := role.Name
	if role, err := self.Kubernetes.RbacV1().Roles(self.Namespace).Create(self.Context, role, meta.CreateOptions{}); err == nil {
		return role, err
	} else if errorspkg.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.RbacV1().Roles(self.Namespace).Get(self.Context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) CreateRoleBinding(roleBinding *rbac.RoleBinding) (*rbac.RoleBinding, error) {
	name := roleBinding.Name
	if roleBinding, err := self.Kubernetes.RbacV1().RoleBindings(self.Namespace).Create(self.Context, roleBinding, meta.CreateOptions{}); err == nil {
		return roleBinding, nil
	} else if errorspkg.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.RbacV1().RoleBindings(self.Namespace).Get(self.Context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) CreateClusterRoleBinding(clusterRoleBinding *rbac.ClusterRoleBinding) (*rbac.ClusterRoleBinding, error) {
	name := clusterRoleBinding.Name
	if clusterRoleBinding, err := self.Kubernetes.RbacV1().ClusterRoleBindings().Create(self.Context, clusterRoleBinding, meta.CreateOptions{}); err == nil {
		return clusterRoleBinding, nil
	} else if errorspkg.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.RbacV1().ClusterRoleBindings().Get(self.Context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) CreateCertificateIssuer(issuer *certmanager.Issuer) (*certmanager.Issuer, error) {
	name := issuer.Name
	if issuer, err := self.CertManager.CertmanagerV1().Issuers(self.Namespace).Create(self.Context, issuer, meta.CreateOptions{}); err == nil {
		return issuer, nil
	} else if errorspkg.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.CertManager.CertmanagerV1().Issuers(self.Namespace).Get(self.Context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) CreateCertificate(certificate *certmanager.Certificate) (*certmanager.Certificate, error) {
	name := certificate.Name
	if certificate, err := self.CertManager.CertmanagerV1().Certificates(self.Namespace).Create(self.Context, certificate, meta.CreateOptions{}); err == nil {
		return certificate, nil
	} else if errorspkg.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.CertManager.CertmanagerV1().Certificates(self.Namespace).Get(self.Context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) CreateNamespace(namespace *core.Namespace) (*core.Namespace, error) {
	name := namespace.Name
	if namespace, err := self.Kubernetes.CoreV1().Namespaces().Create(self.Context, namespace, meta.CreateOptions{}); err == nil {
		return namespace, nil
	} else if errorspkg.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.CoreV1().Namespaces().Get(self.Context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) CreateServiceAccount(serviceAccount *core.ServiceAccount) (*core.ServiceAccount, error) {
	name := serviceAccount.Name
	if serviceAccount, err := self.Kubernetes.CoreV1().ServiceAccounts(self.Namespace).Create(self.Context, serviceAccount, meta.CreateOptions{}); err == nil {
		return serviceAccount, nil
	} else if errorspkg.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.CoreV1().ServiceAccounts(self.Namespace).Get(self.Context, name, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) Labels(appName string, component string, namespace string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       appName,
		"app.kubernetes.io/instance":   fmt.Sprintf("%s-%s", appName, namespace),
		"app.kubernetes.io/version":    version.GitVersion,
		"app.kubernetes.io/component":  component,
		"app.kubernetes.io/part-of":    self.PartOf,
		"app.kubernetes.io/managed-by": self.ManagedBy,
	}
}

func (self *Client) VolumeSource(size string) core.VolumeSource {
	return core.VolumeSource{
		EmptyDir: &core.EmptyDirVolumeSource{},
	}

	// Since Kubernetes 1.19
	// Feature gate: GenericEphemeralVolumes
	// Previous versions will turn this into an EmptyDirVolumeSource
	// https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/
	// https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/1698-generic-ephemeral-volumes
	// import "k8s.io/apimachinery/pkg/api/resource"
	/*return core.VolumeSource{
		Ephemeral: &core.EphemeralVolumeSource{
			VolumeClaimTemplate: &core.PersistentVolumeClaimTemplate{
				Spec: core.PersistentVolumeClaimSpec{
					AccessModes: []core.PersistentVolumeAccessMode{
						core.ReadWriteOnce,
					},
					Resources: core.ResourceRequirements{
						Requests: core.ResourceList{
							core.ResourceStorage: resource.MustParse(size),
						},
					},
				},
			},
		},
	}*/
}
