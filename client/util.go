package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/tliron/turandot/common"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	errorspkg "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	waitpkg "k8s.io/apimachinery/pkg/util/wait"
)

var timeout = 60 * time.Second

func (self *Client) getPods(appName string) (*core.PodList, error) {
	labels_ := labels.Set(map[string]string{
		"app.kubernetes.io/name": appName,
	})
	selector := labels_.AsSelector().String()

	if pods, err := self.kubernetes.CoreV1().Pods(self.namespace).List(self.context, meta.ListOptions{LabelSelector: selector}); err == nil {
		if len(pods.Items) > 0 {
			return pods, nil
		} else {
			return nil, fmt.Errorf("no pods for app.kubernetes.io/name=\"%s\" in namespace \"%s\"", appName, self.namespace)
		}
	} else {
		return nil, err
	}
}

func (self *Client) getPodNames(appName string) ([]string, error) {
	if pods, err := self.getPods(appName); err == nil {
		names := make([]string, len(pods.Items))
		for index, pod := range pods.Items {
			names[index] = pod.Name
		}
		return names, nil
	} else {
		return nil, err
	}
}

func (self *Client) getFirstPodName(appName string) (string, error) {
	if names, err := self.getPodNames(appName); err == nil {
		return names[0], nil
	} else {
		return "", err
	}
}

func (self *Client) getPodIps(appName string) ([]string, error) {
	if pods, err := self.getPods(appName); err == nil {
		var ips []string
		for _, pod := range pods.Items {
			for _, ip := range pod.Status.PodIPs {
				ips = append(ips, ip.IP)
			}
		}
		if len(ips) > 0 {
			return ips, nil
		} else {
			return nil, fmt.Errorf("no IPs for pods for app.kubernetes.io/name=\"%s\" in namespace \"%s\"", appName, self.namespace)
		}
	} else {
		return nil, err
	}
}

func (self *Client) getFirstPodIp(appName string) (string, error) {
	if ips, err := self.getPodIps(appName); err == nil {
		return ips[0], nil
	} else {
		return "", err
	}
}

func (self *Client) createDeployment(deployment *apps.Deployment, appName string) (*apps.Deployment, error) {
	if deployment, err := self.kubernetes.AppsV1().Deployments(self.namespace).Create(self.context, deployment, meta.CreateOptions{}); err == nil {
		return deployment, nil
	} else if errorspkg.IsAlreadyExists(err) {
		return self.kubernetes.AppsV1().Deployments(self.namespace).Get(self.context, appName, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) waitForDeployment(appName string) (*apps.Deployment, error) {
	log.Infof("waiting for deployment for %s", appName)

	var deployment *apps.Deployment
	err := waitpkg.PollImmediate(time.Second, timeout, func() (bool, error) {
		var err error
		if deployment, err = self.kubernetes.AppsV1().Deployments(self.namespace).Get(self.context, appName, meta.GetOptions{}); err == nil {
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
		log.Infof("deployment available for %s", appName)
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
	log.Infof("waiting for a pod for %s", appName)

	return waitpkg.PollImmediate(time.Second, timeout, func() (bool, error) {
		if pods, err := self.getPods(appName); err == nil {
			for _, pod := range pods.Items {
				if self.isPodOwnedBy(&pod, deployment) {
					for _, condition := range pod.Status.Conditions {
						switch condition.Type {
						case core.ContainersReady:
							if condition.Status == core.ConditionTrue {
								log.Infof("pod ready for %s: %s", appName, pod.Name)
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
			if replicaSet, err := self.kubernetes.AppsV1().ReplicaSets(self.namespace).Get(self.context, owner.Name, meta.GetOptions{}); err == nil {
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
		if registry, err := common.GetInternalRegistryURL(self.kubernetes); err == nil {
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
