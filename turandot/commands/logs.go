package commands

import (
	"fmt"
	"io"

	"github.com/tliron/kutil/kubernetes"
	terminalutil "github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
	"github.com/tliron/turandot/controller"
)

func Logs(appNameSuffix string, containerName string) {
	// TODO: what happens if we follow more than one log?
	readers, err := NewClient().Logs(appNameSuffix, containerName, tail, follow)
	util.FailOnError(err)
	for _, reader := range readers {
		defer reader.Close()
	}
	for _, reader := range readers {
		io.Copy(terminalutil.Stdout, reader)
	}
}

func (self *Client) Logs(appNameSuffix string, containerName string, tail int, follow bool) ([]io.ReadCloser, error) {
	appName := fmt.Sprintf("%s-%s", controller.NamePrefix, appNameSuffix)

	if podNames, err := kubernetes.GetPodNames(self.Context, self.Kubernetes, self.Namespace, appName); err == nil {
		readers := make([]io.ReadCloser, len(podNames))
		for index, podName := range podNames {
			if reader, err := kubernetes.Log(self.Context, self.Kubernetes, self.Namespace, podName, containerName, tail, follow); err == nil {
				readers[index] = reader
			} else {
				for i := 0; i < index; i++ {
					readers[i].Close()
				}
				return nil, err
			}
		}
		return readers, nil
	} else {
		return nil, err
	}
}
