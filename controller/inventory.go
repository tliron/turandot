package controller

import (
	"fmt"
	"io"
	neturlpkg "net/url"

	namepkg "github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	gzip "github.com/klauspost/pgzip"
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/client"
	"github.com/tliron/turandot/common"
)

func (self *Controller) GetInventoryServiceTemplateURL(namespace string, serviceTemplateName string, urlContext *urlpkg.Context) (*urlpkg.DockerURL, error) {
	if ip, err := common.GetFirstServiceIP(self.Context, self.Kubernetes, namespace, "turandot-inventory"); err == nil {
		imageName := delegate.GetInventoryImageName(serviceTemplateName)
		url := fmt.Sprintf("docker://%s:5000/%s?format=csar", ip, imageName)
		if url_, err := neturlpkg.ParseRequestURI(url); err == nil {
			return urlpkg.NewDockerURL(url_, urlContext), nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Controller) PublishOnInventory(imageName string, url string, ips []string, urlContext *urlpkg.Context) (string, error) {
	if url, err := urlpkg.NewURL(url, urlContext); err == nil {
		opener := func() (io.ReadCloser, error) {
			if reader, err := url.Open(); err == nil {
				return gzip.NewReader(reader)
			} else {
				return nil, err
			}
		}

		for _, ip := range ips {
			self.Log.Infof("publishing image %q at %q on %q", imageName, url, ip)

			name := fmt.Sprintf("%s:5000/%s", ip, imageName)

			if contentTag, err := namepkg.NewTag("portable"); err == nil {
				if tag, err := namepkg.NewTag(name); err == nil {
					if image, err := tarball.Image(opener, &contentTag); err == nil {
						if err := remote.Write(tag, image); err == nil {
							self.Log.Infof("published image %q at %q on %q", imageName, url, ip)
							return name, nil
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

		return "", fmt.Errorf("did not publish image: %s", imageName)
	} else {
		return "", err
	}
}
