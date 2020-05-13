package client

import (
	"fmt"
	"strings"

	"github.com/tliron/turandot/common"
)

func (self *Client) ServiceClout(serviceName string) (string, error) {
	if service, err := self.GetService(serviceName); err == nil {
		//return service.Status.CloutPath, nil
		appName := fmt.Sprintf("%s-operator", self.NamePrefix)

		if podName, err := common.GetFirstPodName(self.Context, self.Kubernetes, self.Namespace, appName); err == nil {
			var builder strings.Builder
			if err := self.Exec(podName, nil, &builder, "cat", service.Status.CloutPath); err == nil {
				return strings.TrimRight(builder.String(), "\n"), nil
			} else {
				return "", err
			}
		} else {
			return "", err
		}
	} else {
		return "", nil
	}
}
