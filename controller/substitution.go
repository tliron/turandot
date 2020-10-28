package controller

import (
	urlpkg "github.com/tliron/kutil/url"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func (self *Controller) Substitute(namespace string, nodeTemplateName string, inputs map[string]interface{}, mode string, site string, urlContext *urlpkg.Context) error {
	// hacky ;)
	repositoryName := "default"
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
		if repository, err := self.Client.GetRepository(namespace, repositoryName); err == nil {
			if _, err := self.Client.CreateServiceFromTemplate(namespace, serviceName, repository, serviceTemplateName, inputs, mode); err != nil {
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

			if err := remoteClient.InstallRepository("docker.io", true, true); err != nil {
				return err
			}

			var remoteRepository *resources.Repository
			if remoteRepository, err = remoteClient.CreateRepositoryIndirect(namespace, "default", "", "turandot-repository", 5000, "turandot-repository"); err != nil {
				return err
			}

			if url, err := remoteClient.GetRepositoryServiceTemplateURL(remoteRepository, serviceTemplateName); err == nil {
				if url_, err := urlpkg.NewURL(url, urlContext); err == nil {
					if _, err := remoteClient.CreateServiceFromContent(namespace, serviceName, remoteRepository, remoteClient.Spooler(remoteRepository), url_, inputs, mode); err != nil {
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
