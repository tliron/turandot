package client

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tliron/turandot/common"
	"k8s.io/client-go/tools/clientcmd"
)

func (self *Client) SetDelegate(name string, kubeconfigPath string, context string, namespace string) error {
	sourceConfig, err := common.NewConfigForContext(kubeconfigPath, context)
	if err != nil {
		return err
	}

	config, err := common.NewSelfContainedConfig(sourceConfig, namespace)
	if err != nil {
		return err
	}

	configBytes, err := clientcmd.Write(*config)
	if err != nil {
		return err
	}

	appName := fmt.Sprintf("%s-%s", self.NamePrefix, "operator")
	configPath := self.getDelegateConfigPath(name)
	if podNames, err := common.GetPodNames(self.Context, self.Kubernetes, self.Namespace, appName); err == nil {
		for _, podName := range podNames {
			//os.Stdout.Write(configBytes)
			self.Log.Infof("setting delegate %q in operator pod: %s/%s", name, self.Namespace, podName)
			if err := self.WriteToContainer(self.Namespace, podName, "operator", bytes.NewReader(configBytes), configPath, nil); err != nil {
				return err
			}
		}
	} else {
		return err
	}

	return nil
}

func (self *Client) DeleteDelegate(name string) error {
	appName := fmt.Sprintf("%s-%s", self.NamePrefix, "operator")
	configPath := self.getDelegateConfigPath(name)
	if podNames, err := common.GetPodNames(self.Context, self.Kubernetes, self.Namespace, appName); err == nil {
		for _, podName := range podNames {
			if err := self.Exec(self.Namespace, podName, "operator", nil, nil, "rm", "--force", configPath); err != nil {
				return err
			}
		}
	} else {
		return err
	}

	return nil
}

func (self *Client) ListDelegates() ([]string, error) {
	appName := fmt.Sprintf("%s-%s", self.NamePrefix, "operator")
	if podName, err := common.GetFirstPodName(self.Context, self.Kubernetes, self.Namespace, appName); err == nil {
		var buffer bytes.Buffer
		if err := self.Exec(self.Namespace, podName, "operator", nil, &buffer, "find", filepath.Join(self.CachePath, "delegates"), "-type", "f", "-printf", "%f\n"); err == nil {
			var names []string
			for _, filename := range strings.Split(strings.TrimRight(buffer.String(), "\n"), "\n") {
				names = append(names, strings.TrimSuffix(filename, ".yaml"))
			}
			return names, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Client) getDelegateConfigPath(name string) string {
	return filepath.Join(self.CachePath, "delegates", fmt.Sprintf("%s.yaml", name))
}
