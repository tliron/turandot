package controller

import (
	"fmt"

	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/turandot/controller/parser"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func (self *Controller) publishArtifactsToRegistry(artifacts parser.KubernetesArtifacts, service *resources.Service, urlContext *urlpkg.Context) (map[string]string, error) {
	if len(artifacts) > 0 {
		artifactMappings := make(map[string]string)

		for _, artifact := range artifacts {
			if registry, err := self.Client.Reposure.RegistryClient().Get(service.Namespace, artifact.Registry); err == nil {
				if err := self.Client.Reposure.RegistryClient().UpdateURLContext(registry, urlContext); err == nil {
					// Note: OpenShift registry permissions require the namespace as the repository
					imageName := fmt.Sprintf("%s/%s", service.Namespace, artifact.Name)

					if directClient, err := self.Client.Reposure.DirectClient(registry); err == nil {
						if url, err := urlpkg.NewURL(artifact.SourcePath, urlContext); err == nil {
							self.Log.Infof("publishing image %q at %q on %q", imageName, url.String(), directClient.Host)
							if name, err := directClient.PushGzippedTarballFromURL(url, imageName); err == nil {
								self.Log.Infof("published image %q at %q on %q", imageName, url.String(), directClient.Host)
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
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		/*
			if ips, err := kubernetes.GetServiceIPs(self.Context, self.Kubernetes, service.Namespace, "turandot-registry"); err == nil {
				for _, artifact := range artifacts {
					if name, err := self.PublishOnRegistry(artifact.Name, artifact.SourcePath, ips, urlContext); err == nil {
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
