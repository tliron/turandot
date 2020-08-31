package client

import (
	"fmt"
	"strings"

	"github.com/tliron/kutil/kubernetes"
)

func (self *Client) ServiceClout(namespace string, serviceName string) (string, error) {
	if service, err := self.GetService(namespace, serviceName); err == nil {
		appName := fmt.Sprintf("%s-operator", self.NamePrefix)

		if podName, err := kubernetes.GetFirstPodName(self.Context, self.Kubernetes, self.Namespace, appName); err == nil {
			var builder strings.Builder
			if err := self.Exec(self.Namespace, podName, "operator", nil, &builder, "cat", service.Status.CloutPath); err == nil {
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
