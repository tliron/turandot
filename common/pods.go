package common

import (
	contextpkg "context"
	"fmt"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kubernetespkg "k8s.io/client-go/kubernetes"
)

func GetPods(context contextpkg.Context, kubernetes kubernetespkg.Interface, namespace string, appName string) (*core.PodList, error) {
	labels_ := labels.Set(map[string]string{
		"app.kubernetes.io/name": appName,
	})
	selector := labels_.AsSelector().String()

	if pods, err := kubernetes.CoreV1().Pods(namespace).List(context, meta.ListOptions{LabelSelector: selector}); err == nil {
		if len(pods.Items) > 0 {
			return pods, nil
		} else {
			return nil, fmt.Errorf("no pods for app.kubernetes.io/name=\"%s\" in namespace \"%s\"", appName, namespace)
		}
	} else {
		return nil, err
	}
}

func GetPodNames(context contextpkg.Context, kubernetes kubernetespkg.Interface, namespace string, appName string) ([]string, error) {
	if pods, err := GetPods(context, kubernetes, namespace, appName); err == nil {
		names := make([]string, len(pods.Items))
		for index, pod := range pods.Items {
			names[index] = pod.Name
		}
		return names, nil
	} else {
		return nil, err
	}
}

func GetFirstPodName(context contextpkg.Context, kubernetes kubernetespkg.Interface, namespace string, appName string) (string, error) {
	if names, err := GetPodNames(context, kubernetes, namespace, appName); err == nil {
		return names[0], nil
	} else {
		return "", err
	}
}

func GetPodIPs(context contextpkg.Context, kubernetes kubernetespkg.Interface, namespace string, appName string) ([]string, error) {
	if pods, err := GetPods(context, kubernetes, namespace, appName); err == nil {
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

func GetFirstPodIP(context contextpkg.Context, kubernetes kubernetespkg.Interface, namespace string, appName string) (string, error) {
	if ips, err := GetPodIPs(context, kubernetes, namespace, appName); err == nil {
		return ips[0], nil
	} else {
		return "", err
	}
}
