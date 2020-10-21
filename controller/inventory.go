package controller

import (
	"fmt"
	"io"

	namepkg "github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	gzip "github.com/klauspost/pgzip"
	urlpkg "github.com/tliron/kutil/url"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (self *Controller) PublishOnInventory(imageName string, sourceUrl string, inventoryUrl string, urlContext *urlpkg.Context) (string, error) {
	if sourceUrl_, err := urlpkg.NewURL(sourceUrl, urlContext); err == nil {
		opener := func() (io.ReadCloser, error) {
			if reader, err := sourceUrl_.Open(); err == nil {
				return gzip.NewReader(reader)
			} else {
				return nil, err
			}
		}

		self.Log.Infof("publishing image %q at %q on %q", imageName, sourceUrl_, inventoryUrl)

		name := fmt.Sprintf("%s/%s", inventoryUrl, imageName)

		if contentTag, err := namepkg.NewTag("portable"); err == nil {
			if tag, err := namepkg.NewTag(name); err == nil {
				if image, err := tarball.Image(opener, &contentTag); err == nil {
					httpRoundTripper := urlContext.GetHTTPRoundTripper()
					if httpRoundTripper != nil {
						err = remote.Write(tag, image, remote.WithTransport(httpRoundTripper))
					} else {
						err = remote.Write(tag, image)
					}

					if err == nil {
						self.Log.Infof("published image %q at %q on %q", imageName, sourceUrl_, inventoryUrl)
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
	} else {
		return "", err
	}
}

func (self *Controller) UpdateInventorySpoolerPod(inventory *resources.Inventory, spoolerPod string) (*resources.Inventory, error) {
	self.Log.Infof("updating spooler pod to %q for inventory: %s/%s", spoolerPod, inventory.Namespace, inventory.Name)

	for {
		inventory = inventory.DeepCopy()
		inventory.Status.SpoolerPod = spoolerPod

		service_, err, retry := self.updateInventoryStatus(inventory)
		if retry {
			inventory = service_
		} else {
			return service_, err
		}
	}
}
func (self *Controller) updateInventoryStatus(inventory *resources.Inventory) (*resources.Inventory, error, bool) {
	if inventory_, err := self.Client.UpdateInventoryStatus(inventory); err == nil {
		return inventory_, nil, false
	} else if errors.IsConflict(err) {
		self.Log.Warningf("retrying status update for inventory: %s/%s", inventory.Namespace, inventory.Name)
		if inventory_, err := self.Client.GetInventory(inventory.Namespace, inventory.Name); err == nil {
			return inventory_, nil, true
		} else {
			return inventory, err, false
		}
	} else {
		return inventory, err, false
	}
}

func (self *Controller) processInventory(inventory *resources.Inventory) (bool, error) {
	// Create spooler
	if pod, err := self.Client.CreateInventorySpooler(inventory); err == nil {
		if _, err := self.UpdateInventorySpoolerPod(inventory, pod.Name); err == nil {
			return true, nil
		} else {
			return false, err
		}
	} else {
		return false, err
	}
}
