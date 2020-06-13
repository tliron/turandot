package delegate

import (
	"io"
	"path/filepath"
	"strconv"

	"github.com/tliron/puccini/common/terminal"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func (self *Client) WriteToContainer(namespace string, podName string, containerName string, reader io.Reader, targetPath string, permissions *int64) error {
	dir := filepath.Dir(targetPath)
	if err := self.Exec(namespace, podName, containerName, nil, nil, "mkdir", "--parents", dir); err == nil {
		if err := self.Exec(namespace, podName, containerName, reader, nil, "cp", "/dev/stdin", targetPath); err == nil {
			if permissions != nil {
				return self.Exec(namespace, podName, containerName, nil, nil, "chmod", strconv.FormatInt(*permissions, 8), targetPath)
			} else {
				return nil
			}
		} else {
			return err
		}
	} else {
		return err
	}
}

func (self *Client) Exec(namespace string, podName string, containerName string, stdin io.Reader, stdout io.Writer, command ...string) error {
	execOptions := core.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stderr:    true,
		TTY:       false,
	}

	streamOptions := remotecommand.StreamOptions{
		Stderr: terminal.Stderr,
		Tty:    false,
	}

	if stdin != nil {
		execOptions.Stdin = true
		streamOptions.Stdin = stdin
	}

	if stdout != nil {
		execOptions.Stdout = true
		streamOptions.Stdout = stdout
	}

	request := self.REST.Post().Namespace(namespace).Resource("pods").Name(podName).SubResource("exec").VersionedParams(&execOptions, scheme.ParameterCodec)

	if executor, err := remotecommand.NewSPDYExecutor(self.Config, "POST", request.URL()); err == nil {
		if err = executor.Stream(streamOptions); err == nil {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}
