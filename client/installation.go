package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/tliron/kutil/kubernetes"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	errorspkg "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	waitpkg "k8s.io/apimachinery/pkg/util/wait"
)

func (self *Client) GetRegistry(registry string) (string, error) {
	if registry == "internal" {
		if registry, err := kubernetes.GetInternalRegistryURL(self.Kubernetes); err == nil {
			return registry, nil
		} else {
			return "", fmt.Errorf("could not discover internal registry: %s", err.Error())
		}
	}

	if registry != "" {
		return registry, nil
	} else {
		return "", errors.New("must provide \"--registry\"")
	}
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

func (self *Client) WaitForDeletion(name string, condition func() bool) {
	err := waitpkg.PollImmediate(time.Second, timeout, func() (bool, error) {
		self.Log.Infof("waiting for %s to delete", name)
		return !condition(), nil
	})
	if err != nil {
		self.Log.Warningf("error while waiting for %s to delete: %s", name, err.Error())
	}
}

func (self *Client) CreateVolumeSource(size string) core.VolumeSource {
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
						core.ReadWriteMany,
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

func (self *Client) createNamespace() (*core.Namespace, error) {
	namespace := &core.Namespace{
		ObjectMeta: meta.ObjectMeta{
			Name: self.Namespace,
		},
	}

	if namespace, err := self.Kubernetes.CoreV1().Namespaces().Create(self.Context, namespace, meta.CreateOptions{}); err == nil {
		return namespace, nil
	} else if errorspkg.IsAlreadyExists(err) {
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
	} else if errorspkg.IsAlreadyExists(err) {
		self.Log.Infof("%s", err.Error())
		return self.Kubernetes.CoreV1().ServiceAccounts(self.Namespace).Get(self.Context, self.NamePrefix, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) getServiceAccount() (*core.ServiceAccount, error) {
	return self.Kubernetes.CoreV1().ServiceAccounts(self.Namespace).Get(self.Context, self.NamePrefix, meta.GetOptions{})
}
