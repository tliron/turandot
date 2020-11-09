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

func (self *Controller) PublishOnRepository(artifactName string, sourceUrl string, repository *resources.Repository, urlContext *urlpkg.Context) (string, error) {
	if repositoryHost, err := self.Client.GetRepositoryHost(repository); err == nil {
		if options, err := self.Client.GetRepositoryRemoteOptions(repository); err == nil {
			if sourceUrl_, err := urlpkg.NewURL(sourceUrl, urlContext); err == nil {
				self.Log.Infof("publishing image %q at %q on %q", artifactName, sourceUrl_, repositoryHost)

				opener := func() (io.ReadCloser, error) {
					if reader, err := sourceUrl_.Open(); err == nil {
						return gzip.NewReader(reader)
					} else {
						return nil, err
					}
				}

				if contentTag, err := namepkg.NewTag("portable"); err == nil {
					tag := fmt.Sprintf("%s/%s", repositoryHost, artifactName)
					if tag_, err := namepkg.NewTag(tag); err == nil {
						if image, err := tarball.Image(opener, &contentTag); err == nil {
							if err := remote.Write(tag_, image, options...); err == nil {
								self.Log.Infof("published image %q at %q on %q", tag, sourceUrl_, repositoryHost)
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

func (self *Controller) UpdateRepositorySpoolerPod(repository *resources.Repository, spoolerPod string) (*resources.Repository, error) {
	self.Log.Infof("updating spooler pod to %q for repository: %s/%s", spoolerPod, repository.Namespace, repository.Name)

	for {
		repository = repository.DeepCopy()
		repository.Status.SpoolerPod = spoolerPod

		service_, err, retry := self.updateRepositoryStatus(repository)
		if retry {
			repository = service_
		} else {
			return service_, err
		}
	}
}
func (self *Controller) updateRepositoryStatus(repository *resources.Repository) (*resources.Repository, error, bool) {
	if repository_, err := self.Client.UpdateRepositoryStatus(repository); err == nil {
		return repository_, nil, false
	} else if errors.IsConflict(err) {
		self.Log.Warningf("retrying status update for repository: %s/%s", repository.Namespace, repository.Name)
		if repository_, err := self.Client.GetRepository(repository.Namespace, repository.Name); err == nil {
			return repository_, nil, true
		} else {
			return repository, err, false
		}
	} else {
		return repository, err, false
	}
}

func (self *Controller) processRepository(repository *resources.Repository) (bool, error) {
	// Create spooler
	if pod, err := self.Client.CreateRepositorySpooler(repository); err == nil {
		if _, err := self.UpdateRepositorySpoolerPod(repository, pod.Name); err == nil {
			return true, nil
		} else {
			return false, err
		}
	} else {
		return false, err
	}
}
