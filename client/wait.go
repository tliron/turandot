package delegate

import (
	"errors"
	"fmt"
	"time"

	"github.com/tliron/turandot/common"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	waitpkg "k8s.io/apimachinery/pkg/util/wait"
)

var timeout = 60 * time.Second

func (self *Client) waitForDeployment(appName string) (*apps.Deployment, error) {
	self.Log.Infof("waiting for deployment for %s", appName)

	var deployment *apps.Deployment
	err := waitpkg.PollImmediate(time.Second, timeout, func() (bool, error) {
		var err error
		if deployment, err = self.Kubernetes.AppsV1().Deployments(self.Namespace).Get(self.Context, appName, meta.GetOptions{}); err == nil {
			for _, condition := range deployment.Status.Conditions {
				switch condition.Type {
				case apps.DeploymentAvailable:
					if condition.Status == core.ConditionTrue {
						return true, nil
					}
				case apps.DeploymentReplicaFailure:
					if condition.Status == core.ConditionTrue {
						return false, fmt.Errorf("replica failure: %s", appName)
					}
				}
			}
			return false, nil
		} else {
			return false, err
		}
	})

	if err == nil {
		self.Log.Infof("deployment available for %s", appName)
		if err := self.waitForAPod(appName, deployment); err == nil {
			return deployment, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Client) waitForAPod(appName string, deployment *apps.Deployment) error {
	self.Log.Infof("waiting for a pod for %s", appName)

	return waitpkg.PollImmediate(time.Second, timeout, func() (bool, error) {
		if pods, err := common.GetPods(self.Context, self.Kubernetes, self.Namespace, appName); err == nil {
			for _, pod := range pods.Items {
				if self.isPodOwnedBy(&pod, deployment) {
					for _, condition := range pod.Status.Conditions {
						switch condition.Type {
						case core.ContainersReady:
							if condition.Status == core.ConditionTrue {
								self.Log.Infof("pod ready for %s: %s", appName, pod.Name)
								return true, nil
							}
						}
					}
				}
			}
			return false, nil
		} else {
			return false, err
		}
	})
}

func (self *Client) isPodOwnedBy(pod *core.Pod, deployment *apps.Deployment) bool {
	for _, owner := range pod.OwnerReferences {
		if (owner.APIVersion == "apps/v1") && (owner.Kind == "ReplicaSet") {
			if replicaSet, err := self.Kubernetes.AppsV1().ReplicaSets(self.Namespace).Get(self.Context, owner.Name, meta.GetOptions{}); err == nil {
				if self.isReplicaSetOwnedBy(replicaSet, deployment) {
					return true
				}
			}
		}
	}
	return false
}

func (self *Client) isReplicaSetOwnedBy(replicaSet *apps.ReplicaSet, deployment *apps.Deployment) bool {
	for _, owner := range replicaSet.OwnerReferences {
		if owner.UID == deployment.UID {
			return true
		}
	}
	return false
}

func (self *Client) getRegistry(registry string) (string, error) {
	if registry == "internal" {
		if registry, err := common.GetInternalRegistryURL(self.Kubernetes); err == nil {
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
