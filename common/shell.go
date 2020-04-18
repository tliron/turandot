package common

import (
	"io"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	restpkg "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func Shell(rest restpkg.Interface, config *restpkg.Config, namespace string, podName string, containerName string, command string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	execOptions := core.PodExecOptions{
		Container: containerName,
		Command:   []string{command},
		TTY:       true,
	}

	streamOptions := remotecommand.StreamOptions{
		Tty: true, // seems unnecessary
	}

	if stdin != nil {
		execOptions.Stdin = true
		streamOptions.Stdin = stdin
	}

	if stdout != nil {
		execOptions.Stdout = true
		streamOptions.Stdout = stdout
	}

	if stderr != nil {
		execOptions.Stderr = true
		streamOptions.Stderr = stderr
	}

	request := rest.Post().Namespace(namespace).Resource("pods").Name(podName).SubResource("exec").VersionedParams(&execOptions, scheme.ParameterCodec)

	if executor, err := remotecommand.NewSPDYExecutor(config, "POST", request.URL()); err == nil {
		return executor.Stream(streamOptions)
	} else {
		return err
	}
}
