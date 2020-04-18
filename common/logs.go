package common

import (
	"context"
	"io"

	core "k8s.io/api/core/v1"
	kubernetespkg "k8s.io/client-go/kubernetes"
)

func Log(kubernetes *kubernetespkg.Clientset, namespace string, podName string, containerName string, tail int, follow bool) (io.ReadCloser, error) {
	options := core.PodLogOptions{
		Container: containerName,
		Follow:    follow,
	}

	if tail >= 0 {
		tail_ := int64(tail)
		options.TailLines = &tail_
	}

	request := kubernetes.CoreV1().Pods(namespace).GetLogs(podName, &options)
	return request.Stream(context.TODO())
}
