package client

import (
	"time"

	"github.com/tliron/kutil/kubernetes"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	errorspkg "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	waitpkg "k8s.io/apimachinery/pkg/util/wait"
)

var timeout = 60 * time.Second

func (self *Client) WaitForPod(namespace string, appName string) (*core.Pod, error) {
	self.Log.Infof("waiting for a pod for app %q", appName)

	var pod *core.Pod
	err := waitpkg.PollImmediate(time.Second, timeout, func() (bool, error) {
		if pods, err := kubernetes.GetPods(self.Context, self.Kubernetes, namespace, appName); err == nil {
			for _, pod_ := range pods.Items {
				for _, containerStatus := range pod_.Status.ContainerStatuses {
					if containerStatus.Ready {
						self.Log.Infof("container %q ready for pod %q", containerStatus.Name, pod_.Name)
					} else {
						return false, nil
					}
				}

				for _, condition := range pod_.Status.Conditions {
					switch condition.Type {
					case core.ContainersReady:
						if condition.Status == core.ConditionTrue {
							pod = &pod_
							return true, nil
						}
					}
				}
			}
			return false, nil
		} else if errorspkg.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	})

	if (err == nil) && (pod != nil) {
		self.Log.Infof("a pod is available for app %q", appName)
		return pod, nil
	} else {
		return nil, err
	}
}

func (self *Client) WaitForDeployment(namespace string, appName string) (*apps.Deployment, error) {
	self.Log.Infof("waiting for a deployment for app %q", appName)

	var deployment *apps.Deployment
	err := waitpkg.PollImmediate(time.Second, timeout, func() (bool, error) {
		var err error
		if deployment, err = self.Kubernetes.AppsV1().Deployments(namespace).Get(self.Context, appName, meta.GetOptions{}); err == nil {
			for _, condition := range deployment.Status.Conditions {
				switch condition.Type {
				case apps.DeploymentAvailable:
					if condition.Status == core.ConditionTrue {
						return true, nil
					}
				case apps.DeploymentReplicaFailure:
					if condition.Status == core.ConditionTrue {
						self.Log.Infof("replica failure for a deployment for app %q", appName)
					}
				}
			}
			return false, nil
		} else {
			return false, err
		}
	})

	if err == nil {
		self.Log.Infof("a deployment is available for app %q", appName)
		//if err := self.waitForPods(appName, deployment); err == nil {
		return deployment, nil
		/*} else {
			return nil, err
		}*/
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
