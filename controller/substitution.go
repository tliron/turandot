package controller

func (self *Controller) Substitute(namespace string, nodeTemplateName string, inputs map[string]interface{}, site string) error {
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

	if url, err := self.GetInventoryServiceTemplateURL(namespace, serviceTemplateName); err == nil {
		defer url.Release()

		if (site == "") || (site == self.Site) {
			// Local
			if _, err := self.CreateService(namespace, serviceName, url, inputs); err != nil {
				return err
			}
		} else {
			// Delegate
			if client, spooler, err := self.NewDelegate(site); err == nil {
				if err := client.Install(site, "docker.io", true); err != nil {
					return err
				}

				if err := client.DeployServiceFromContent(serviceName, spooler, url, inputs); err != nil {
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
