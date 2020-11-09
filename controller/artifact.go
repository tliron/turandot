package controller

import (
	"fmt"

	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func (self *Controller) publishArtifactsToRepository(artifacts parser.KubernetesArtifacts, service *resources.Service, urlContext *urlpkg.Context) (map[string]string, error) {
	if len(artifacts) > 0 {
		artifactMappings := make(map[string]string)

		for _, artifact := range artifacts {
			if repository, err := self.Client.GetRepository(service.Namespace, artifact.Repository); err == nil {
				if err := self.Client.UpdateRepositoryURLContext(repository, urlContext); err == nil {
					// Note: OpenShift registry permissions require the namespace as the tag category
					artifactName := fmt.Sprintf("%s/%s", service.Namespace, artifact.Name)

					if name, err := self.PublishOnRepository(artifactName, artifact.SourcePath, repository, urlContext); err == nil {
						artifactMappings[artifact.SourcePath] = name
					} else {
						return nil, err
					}
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		/*
			if ips, err := kubernetes.GetServiceIPs(self.Context, self.Kubernetes, service.Namespace, "turandot-repository"); err == nil {
				for _, artifact := range artifacts {
					if name, err := self.PublishOnRepository(artifact.Name, artifact.SourcePath, ips, urlContext); err == nil {
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
