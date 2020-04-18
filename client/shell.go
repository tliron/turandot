package client

import (
	"fmt"
	"io"

	"github.com/tliron/turandot/common"
)

func (self *Client) Shell(appNameSuffix string, containerName string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	appName := fmt.Sprintf("%s-%s", self.namePrefix, appNameSuffix)

	if podName, err := self.getFirstPodName(appName); err == nil {
		return common.Shell(self.rest, self.config, self.namespace, podName, containerName, "bash", stdin, stdout, stderr)
	} else {
		return err
	}
}
