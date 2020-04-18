package controller

import (
	"fmt"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (self *Controller) getPods(namespace string, appName string) (*core.PodList, error) {
	labels_ := labels.Set(map[string]string{
		"app.kubernetes.io/name": appName,
	})
	selector := labels_.AsSelector().String()

	if pods, err := self.kubernetes.CoreV1().Pods(namespace).List(self.context, meta.ListOptions{LabelSelector: selector}); err == nil {
		if len(pods.Items) > 0 {
			return pods, nil
		} else {
			return nil, fmt.Errorf("no pods for app.kubernetes.io/name=\"%s\" in namespace \"%s\"", appName, namespace)
		}
	} else {
		return nil, err
	}
}

func (self *Controller) getPodIps(namespace string, appName string) ([]string, error) {
	if pods, err := self.getPods(namespace, appName); err == nil {
		var ips []string
		for _, pod := range pods.Items {
			for _, ip := range pod.Status.PodIPs {
				ips = append(ips, ip.IP)
			}
		}
		if len(ips) > 0 {
			return ips, nil
		} else {
			return nil, fmt.Errorf("no IPs for pods for app.kubernetes.io/name=\"%s\" in namespace \"%s\"", appName, namespace)
		}
	} else {
		return nil, err
	}
}

func (self *Controller) getFirstPodIp(namespace string, appName string) (string, error) {
	if ips, err := self.getPodIps(namespace, appName); err == nil {
		return ips[0], nil
	} else {
		return "", err
	}
}
