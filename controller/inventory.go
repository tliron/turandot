package controller

import (
	"fmt"
	"io"
	neturlpkg "net/url"
	"os"

	namepkg "github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	gzip "github.com/klauspost/pgzip"
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/client"
)

func (self *Controller) GetInventoryServiceTemplateURL(namespace string, serviceTemplateName string) (*urlpkg.DockerURL, error) {
	if ip, err := self.getFirstPodIp(namespace, "turandot-inventory"); err == nil {
		imageName := client.GetInventoryImageName(serviceTemplateName)
		url := fmt.Sprintf("docker://%s:5000/%s?format=csar", ip, imageName)
		if url_, err := neturlpkg.ParseRequestURI(url); err == nil {
			return urlpkg.NewDockerURL(url_), nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Controller) Push(imageName string, url string, ips []string) (string, error) {
	if url, err := urlpkg.NewURL(url); err == nil {
		defer url.Release()

		opener := func() (io.ReadCloser, error) {
			if reader, err := url.Open(); err == nil {
				return gzip.NewReader(reader)
			} else {
				return nil, err
			}
		}

		for _, ip := range ips {
			self.log.Infof("pushing image \"%s\" from \"%s\" to \"%s\"", imageName, url, ip)

			name := fmt.Sprintf("%s:5000/%s", ip, imageName)

			if contentTag, err := namepkg.NewTag("portable"); err == nil {
				if tag, err := namepkg.NewTag(name); err == nil {
					if image, err := tarball.Image(opener, &contentTag); err == nil {
						if err := remote.Write(tag, image); err == nil {
							self.log.Infof("pushed image \"%s\" from \"%s\" to \"%s\"", imageName, url, ip)
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

		return "", fmt.Errorf("did not push image: %s", imageName)
	} else {
		return "", err
	}
}

func PushTarballToRegistry(path string, name string) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		if image, err := tarball.ImageFromPath(path, &tag); err == nil {
			return remote.Write(tag, image)
		} else {
			return err
		}
	} else {
		return err
	}
}

func PushGzippedTarballToRegistry(path string, name string) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		opener := func() (io.ReadCloser, error) {
			if reader, err := os.Open(path); err == nil {
				return gzip.NewReader(reader)
			} else {
				return nil, err
			}
		}

		if image, err := tarball.Image(opener, &tag); err == nil {
			return remote.Write(tag, image)
		} else {
			return err
		}
	} else {
		return err
	}
}
