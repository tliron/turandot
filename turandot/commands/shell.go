package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/tliron/kutil/kubernetes"
	"github.com/tliron/kutil/util"
	"github.com/tliron/turandot/controller"
)

func Shell(appNameSuffix string, containerName string) {
	util.ToRawTerminal(func() error {
		return NewClient().Shell(appNameSuffix, containerName, os.Stdin, os.Stdout, os.Stderr)
	})
}

func (self *Client) Shell(appNameSuffix string, containerName string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	appName := fmt.Sprintf("%s-%s", controller.NamePrefix, appNameSuffix)

	if podName, err := kubernetes.GetFirstPodName(self.Context, self.Kubernetes, self.Namespace, appName); err == nil {
		return kubernetes.Exec(self.REST, self.Config, self.Namespace, podName, containerName, stdin, stdout, stderr, true, "sh")
	} else {
		return err
	}
}
