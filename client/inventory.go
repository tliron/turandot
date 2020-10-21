package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	neturlpkg "net/url"
	"strings"

	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const serviceTemplateCategory = "service-templates"

var serviceTemplateImageNamePrefix = fmt.Sprintf("%s/", serviceTemplateCategory)

func (self *Client) GetInventory(namespace string, inventoryName string) (*resources.Inventory, error) {
	// Default to same namespace as operator
	if namespace == "" {
		namespace = self.Namespace
	}

	if inventory, err := self.Turandot.TurandotV1alpha1().Inventories(namespace).Get(self.Context, inventoryName, meta.GetOptions{}); err == nil {
		// When retrieved from cache the GVK may be empty
		if inventory.Kind == "" {
			inventory = inventory.DeepCopy()
			inventory.APIVersion, inventory.Kind = resources.InventoryGVK.ToAPIVersionAndKind()
		}
		return inventory, nil
	} else {
		return nil, err
	}
}

func (self *Client) ListInventories() (*resources.InventoryList, error) {
	// TODO: all inventories in cluster mode
	return self.Turandot.TurandotV1alpha1().Inventories(self.Namespace).List(self.Context, meta.ListOptions{})
}

func (self *Client) CreateInventory(namespace string, inventoryName string, url string, serviceName string, secretName string) (*resources.Inventory, error) {
	// Default to same namespace as operator
	if namespace == "" {
		namespace = self.Namespace
	}

	inventory := &resources.Inventory{
		ObjectMeta: meta.ObjectMeta{
			Name:      inventoryName,
			Namespace: namespace,
		},
		Spec: resources.InventorySpec{
			URL:     url,
			Service: serviceName,
			Secret:  secretName,
		},
	}

	if inventory, err := self.Turandot.TurandotV1alpha1().Inventories(namespace).Create(self.Context, inventory, meta.CreateOptions{}); err == nil {
		return inventory, nil
	} else if errors.IsAlreadyExists(err) {
		return self.Turandot.TurandotV1alpha1().Inventories(namespace).Get(self.Context, inventoryName, meta.GetOptions{})
	} else {
		return nil, err
	}
}

func (self *Client) UpdateInventoryStatus(inventory *resources.Inventory) (*resources.Inventory, error) {
	if inventory_, err := self.Turandot.TurandotV1alpha1().Inventories(inventory.Namespace).UpdateStatus(self.Context, inventory, meta.UpdateOptions{}); err == nil {
		// When retrieved from cache the GVK may be empty
		if inventory_.Kind == "" {
			inventory_ = inventory_.DeepCopy()
			inventory_.APIVersion, inventory_.Kind = resources.InventoryGVK.ToAPIVersionAndKind()
		}
		return inventory_, nil
	} else {
		return inventory, err
	}
}

func (self *Client) DeleteInventory(namespace string, inventoryName string) error {
	// Default to same namespace as operator
	if namespace == "" {
		namespace = self.Namespace
	}

	return self.Turandot.TurandotV1alpha1().Inventories(namespace).Delete(self.Context, inventoryName, meta.DeleteOptions{})
}

func (self *Client) GetInventoryURL(namespace string, inventoryName string) (string, error) {
	if inventory, err := self.GetInventory(namespace, inventoryName); err == nil {
		return inventory.Spec.URL, nil
	} else {
		return "", err
	}
}

func (self *Client) GetInventoryCertificate(namespace string, inventoryName string) (*x509.Certificate, error) {
	if inventory, err := self.GetInventory(namespace, inventoryName); err == nil {
		return self.GetSecretCertificate(namespace, inventory.Spec.Secret)
	} else {
		return nil, err
	}
}

func (self *Client) GetInventoryHTTPRoundTripper(namespace string, inventoryName string) (http.RoundTripper, error) {
	if certificate, err := self.GetInventoryCertificate(namespace, inventoryName); err == nil {
		certPool := x509.NewCertPool()
		certPool.AddCert(certificate)
		return util.NewForceHTTPSRoundTripper(&http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		}), nil
	} else {
		return nil, err
	}
}

func (self *Client) GetInventoryURLForCSAR(namespace string, inventoryName string, imageName string, urlContext *urlpkg.Context) (*urlpkg.DockerURL, error) {
	if url, err := self.GetInventoryURL(namespace, inventoryName); err == nil {
		url := fmt.Sprintf("docker://%s/%s?format=csar", url, imageName)
		if url_, err := neturlpkg.ParseRequestURI(url); err == nil {
			return urlpkg.NewDockerURL(url_, urlContext), nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	/*
		// Default to same namespace as operator
		if namespace == "" {
			namespace = self.Namespace
		}

			appName := fmt.Sprintf("%s-inventory", self.NamePrefix)
			if ip, err := kubernetes.GetFirstServiceIP(self.Context, self.Kubernetes, namespace, appName); err == nil {
				url := fmt.Sprintf("docker://%s:5000/%s?format=csar", ip, imageName)
				if url_, err := neturlpkg.ParseRequestURI(url); err == nil {
					return urlpkg.NewDockerURL(url_, urlContext), nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
	*/
}

func (self *Client) GetInventoryServiceTemplateURL(namespace string, inventoryName string, serviceTemplateName string, urlContext *urlpkg.Context) (*urlpkg.DockerURL, error) {
	return self.GetInventoryURLForCSAR(namespace, inventoryName, InventoryImageNameForServiceTemplateName(serviceTemplateName), urlContext)
}

func InventoryImageNameForServiceTemplateName(serviceTemplateName string) string {
	return fmt.Sprintf("%s%s", serviceTemplateImageNamePrefix, serviceTemplateName)
}

func ServiceTemplateNameForInventoryImageName(imageName string) (string, bool) {
	if strings.HasPrefix(imageName, serviceTemplateImageNamePrefix) {
		return imageName[len(serviceTemplateImageNamePrefix):], true
	} else {
		return "", false
	}
}
