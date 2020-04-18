package common

import (
	"context"
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetInternalRegistryURL(kubernetesClient *kubernetes.Clientset) (string, error) {
	if service, err := kubernetesClient.CoreV1().Services("kube-system").Get(context.TODO(), "registry", meta.GetOptions{}); err == nil {
		return fmt.Sprintf("%s:80", service.Spec.ClusterIP), nil
	} else {
		return "", err
	}
}
