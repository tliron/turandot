package controller

import (
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func (self *Controller) publishArtifactsToInventory(artifacts parser.KubernetesArtifacts, service *resources.Service, urlContext *urlpkg.Context) (map[string]string, error) {
	if len(artifacts) > 0 {
		artifactMappings := make(map[string]string)

		if inventoryUrl, err := self.Client.GetInventoryURL(service.Namespace, "default"); err == nil {
			for _, artifact := range artifacts {
				if name, err := self.PublishOnInventory(artifact.Name, artifact.SourcePath, inventoryUrl, urlContext); err == nil {
					artifactMappings[artifact.SourcePath] = name
				} else {
					return nil, err
				}
			}
		} else {
			return nil, err
		}

		/*
			if ips, err := kubernetes.GetServiceIPs(self.Context, self.Kubernetes, service.Namespace, "turandot-inventory"); err == nil {
				for _, artifact := range artifacts {
					if name, err := self.PublishOnInventory(artifact.Name, artifact.SourcePath, ips, urlContext); err == nil {
						artifactMappings[artifact.SourcePath] = name
					} else {
						return nil, err
					}
				}
			}
		*/

		if len(artifactMappings) == 0 {
			return nil, nil
		} else {
			return artifactMappings, nil
		}
	} else {
		return nil, nil
	}
}
