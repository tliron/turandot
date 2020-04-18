package client

import (
	"io"
	"os"
	"path/filepath"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func (self *Client) WriteToContainer(podName string, reader io.Reader, targetPath string) error {
	dir := filepath.Dir(targetPath)
	if err := self.Exec(podName, nil, nil, "mkdir", "--parents", dir); err == nil {
		return self.Exec(podName, reader, nil, "cp", "/dev/stdin", targetPath)
	} else {
		return err
	}
}

func (self *Client) Exec(podName string, stdin io.Reader, stdout io.Writer, command ...string) error {
	execOptions := core.PodExecOptions{
		Container: "operator",
		Command:   command,
		Stderr:    true,
		TTY:       false,
	}

	streamOptions := remotecommand.StreamOptions{
		Stderr: os.Stderr,
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

	request := self.rest.Post().Namespace(self.namespace).Resource("pods").Name(podName).SubResource("exec").VersionedParams(&execOptions, scheme.ParameterCodec)

	if executor, err := remotecommand.NewSPDYExecutor(self.config, "POST", request.URL()); err == nil {
		if err = executor.Stream(streamOptions); err == nil {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}
