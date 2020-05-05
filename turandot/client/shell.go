package client

import (
	"fmt"
	"io"

	"github.com/tliron/turandot/common"
)

func (self *Client) Shell(appNameSuffix string, containerName string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	appName := fmt.Sprintf("%s-%s", self.NamePrefix, appNameSuffix)

	if podName, err := common.GetFirstPodName(self.Context, self.Kubernetes, self.Namespace, appName); err == nil {
		return common.Shell(self.REST, self.Config, self.Namespace, podName, containerName, "bash", stdin, stdout, stderr)
	} else {
		return err
	}
}
