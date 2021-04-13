package controller

import (
	urlpkg "github.com/tliron/kutil/url"
	reposure "github.com/tliron/reposure/resources/reposure.puccini.cloud/v1alpha1"
)

func (self *Controller) Substitute(namespace string, nodeTemplateName string, inputs map[string]interface{}, mode string, site string, urlContext *urlpkg.Context) error {
	// hacky ;)
	registryName := "default"
	var serviceTemplateName string
	switch nodeTemplateName {
	case "central-pbx":
		serviceTemplateName = "asterisk-vnf"
	case "edge-pbx":
		serviceTemplateName = "asterisk-cnf"
	case "data-plane":
		serviceTemplateName = "simple-data-plane"
	}
	serviceName := serviceTemplateName

	if (site == "") || (site == self.Site) {
		// Local
		if registry, err := self.Client.Reposure.RegistryClient().Get(namespace, registryName); err == nil {
			if _, err := self.Client.CreateServiceFromTemplate(namespace, serviceName, registry, serviceTemplateName, inputs, mode); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Delegate
		self.Log.Infof("delegating %q to: %s", serviceTemplateName, site)
		if remoteClient, err := self.NewDelegate(site); err == nil {
			if err := remoteClient.InstallOperator(site, "docker.io", true); err != nil {
				return err
			}
			if err := remoteClient.Reposure.InstallOperator("docker.io", true); err != nil {
				return err
			}

			// hack: Minikube
			var remoteRegistry *reposure.Registry
			if remoteRegistry, err = remoteClient.Reposure.CreateRegistryIndirect(namespace, "default", "kube-system", "registry", 80, "", "", ""); err != nil {
				return err
			}

			if url, err := remoteClient.GetRegistryServiceTemplateURL(remoteRegistry, serviceTemplateName); err == nil {
				if url_, err := urlpkg.NewURL(url, urlContext); err == nil {
					if _, err := remoteClient.CreateServiceFromContent(namespace, serviceName, remoteRegistry, url_, inputs, mode); err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}
