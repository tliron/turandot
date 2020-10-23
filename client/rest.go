package client

import (
	"io"

	kubernetesutil "github.com/tliron/kutil/kubernetes"
)

func (self *Client) WriteToContainer(namespace string, podName string, containerName string, reader io.Reader, targetPath string, permissions *int64) error {
	return kubernetesutil.WriteToContainer(self.REST, self.Config, namespace, podName, containerName, reader, targetPath, permissions)
}

func (self *Client) Exec(namespace string, podName string, containerName string, stdin io.Reader, stdout io.Writer, command ...string) error {
	return kubernetesutil.Exec(self.REST, self.Config, namespace, podName, containerName, stdin, stdout, nil, false, command...)
}
