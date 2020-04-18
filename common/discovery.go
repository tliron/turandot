package common

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	discoverypkg "k8s.io/client-go/discovery"
)

func FindResourceForKind(discovery_ discoverypkg.DiscoveryInterface, gvk schema.GroupVersionKind, supportedVerbs ...string) (schema.GroupVersionResource, error) {
	if gvrs, err := FindResourcesForKind(discovery_, gvk, supportedVerbs...); err == nil {
		count := len(gvrs)
		if count == 1 {
			return gvrs[0], nil
		} else if count == 0 {
			return schema.GroupVersionResource{}, fmt.Errorf("%s resources not found for: %s", supportedVerbs, gvk.String())
		} else {
			return schema.GroupVersionResource{}, fmt.Errorf("too many %s resources found for: %s", supportedVerbs, gvk.String())
		}
	} else {
		return schema.GroupVersionResource{}, err
	}
}

func FindResourcesForKind(discovery discoverypkg.DiscoveryInterface, gvk schema.GroupVersionKind, supportedVerbs ...string) ([]schema.GroupVersionResource, error) {
	gv := gvk.GroupVersion()
	groupVersion := gv.String()
	kind := gvk.Kind

	if _, resourceLists, err := discovery.ServerGroupsAndResources(); err == nil {
		var gvrs []schema.GroupVersionResource

		for _, resourceList := range resourceLists {
			if resourceList.GroupVersion == groupVersion {
				for _, resource := range resourceList.APIResources {
					if resource.Kind == kind {
						var matchedVerbs []string
						for _, verb := range supportedVerbs {
							for _, verb_ := range resource.Verbs {
								if verb == verb_ {
									matchedVerbs = append(matchedVerbs, verb)
									break
								}
							}
						}

						if len(matchedVerbs) == len(supportedVerbs) {
							gvr := gv.WithResource(resource.Name)
							gvrs = append(gvrs, gvr)
						}
					}
				}
			}
		}

		return gvrs, nil
	} else {
		return nil, err
	}
}
