package controller

import (
	"fmt"
	"io"

	namepkg "github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	gzip "github.com/klauspost/pgzip"
	urlpkg "github.com/tliron/kutil/url"
	reposure "github.com/tliron/reposure/resources/reposure.puccini.cloud/v1alpha1"
)

func (self *Controller) PublishOnRegistry(artifactName string, sourceUrl string, registry *reposure.Registry, urlContext *urlpkg.Context) (string, error) {
	if registryHost, err := self.Client.Reposure.RegistryClient().GetHost(registry); err == nil {
		if options, err := self.Client.Reposure.RegistryClient().GetRemoteOptions(registry); err == nil {
			if sourceUrl_, err := urlpkg.NewURL(sourceUrl, urlContext); err == nil {
				self.Log.Infof("publishing image %q at %q on %q", artifactName, sourceUrl_, registryHost)

				opener := func() (io.ReadCloser, error) {
					if reader, err := sourceUrl_.Open(); err == nil {
						return gzip.NewReader(reader)
					} else {
						return nil, err
					}
				}

				if contentTag, err := namepkg.NewTag("portable"); err == nil {
					tag := fmt.Sprintf("%s/%s", registryHost, artifactName)
					if tag_, err := namepkg.NewTag(tag); err == nil {
						if image, err := tarball.Image(opener, &contentTag); err == nil {
							if err := remote.Write(tag_, image, options...); err == nil {
								self.Log.Infof("published image %q at %q on %q", tag, sourceUrl_, registryHost)
								return tag, nil
							} else {
								return "", err
							}
						} else {
							return "", err
						}
					} else {
						return "", err
					}
				} else {
					return "", err
				}
			} else {
				return "", err
			}
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}
