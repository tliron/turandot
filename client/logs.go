package client

import (
	"fmt"
	"io"

	"github.com/tliron/turandot/common"
)

func (self *Client) Logs(appNameSuffix string, containerName string, tail int, follow bool) ([]io.ReadCloser, error) {
	appName := fmt.Sprintf("%s-%s", self.namePrefix, appNameSuffix)

	if podNames, err := self.getPodNames(appName); err == nil {
		readers := make([]io.ReadCloser, len(podNames))
		for index, podName := range podNames {
			if reader, err := common.Log(self.kubernetes, self.namespace, podName, containerName, tail, follow); err == nil {
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
