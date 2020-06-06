package common

import (
	contextpkg "context"
	"fmt"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kubernetespkg "k8s.io/client-go/kubernetes"
)

func GetServices(context contextpkg.Context, kubernetes kubernetespkg.Interface, namespace string, appName string) (*core.ServiceList, error) {
	labels_ := labels.Set(map[string]string{
		"app.kubernetes.io/name": appName,
	})
	selector := labels_.AsSelector().String()

	if services, err := kubernetes.CoreV1().Services(namespace).List(context, meta.ListOptions{LabelSelector: selector}); err == nil {
		if len(services.Items) > 0 {
			return services, nil
		} else {
			return nil, fmt.Errorf("no services for app.kubernetes.io/name=%q in namespace %q", appName, namespace)
		}
	} else {
		return nil, err
	}
}

func GetServiceIPs(context contextpkg.Context, kubernetes kubernetespkg.Interface, namespace string, appName string) ([]string, error) {
	if services, err := GetServices(context, kubernetes, namespace, appName); err == nil {
		var ips []string
		for _, service := range services.Items {
			ips = append(ips, service.Spec.ClusterIP)
		}
		if len(ips) > 0 {
			return ips, nil
		} else {
			return nil, fmt.Errorf("no IPs for services for app.kubernetes.io/name=%q in namespace %q", appName, namespace)
		}
	} else {
		return nil, err
	}
}

func GetFirstServiceIP(context contextpkg.Context, kubernetes kubernetespkg.Interface, namespace string, appName string) (string, error) {
	if ips, err := GetServiceIPs(context, kubernetes, namespace, appName); err == nil {
		return ips[0], nil
	} else {
		return "", err
	}
}
