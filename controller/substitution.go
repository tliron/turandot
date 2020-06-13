package controller

import (
	urlpkg "github.com/tliron/puccini/url"
)

func (self *Controller) Substitute(namespace string, nodeTemplateName string, inputs map[string]interface{}, mode string, site string, urlContext *urlpkg.Context) error {
	// hacky ;)
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

	if url, err := self.GetInventoryServiceTemplateURL(namespace, serviceTemplateName, urlContext); err == nil {
		if (site == "") || (site == self.Site) {
			// Local
			if _, err := self.Client.CreateService(namespace, serviceName, url, inputs, mode); err != nil {
				return err
			}
		} else {
			// Delegate
			self.Log.Infof("delegating %q to: %s", serviceTemplateName, site)
			if client, spooler, err := self.NewDelegate(site); err == nil {
				if err := client.Install(site, "docker.io", true); err != nil {
					return err
				}

				if err := client.CreateServiceFromContent(namespace, serviceName, spooler, url, inputs, mode, urlContext); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	} else {
		return err
	}

	return nil
}
