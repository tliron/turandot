package main

import (
	"strings"

	"github.com/go-openapi/spec"
)

//
// Node
//

type Node struct {
	Name          string
	Group         string
	Version       string
	Kind          string
	OriginalGroup string
	OriginalKind  string
	Schema        spec.Schema
	Generator     *Generator
}

func (self *Generator) NewNode(name string, schema spec.Schema) *Node {
	var group, version, kind, originalGroup, originalKind string
	if extension, ok := schema.Extensions[extensionName]; ok {
		group, version, kind = parseExtension(extension)
		originalGroup = group
		originalKind = kind
	} else {
		group, version, kind = parseGVK(name)
		originalGroup = group
		originalKind = kind

		// Remove default prefix
		if strings.HasPrefix(group, defaultGroupPrefix) {
			group = group[len(defaultGroupPrefix):]
		}

		// Remove "core"
		if group == coreGroup {
			group = ""
		}
	}

	if group == "" {
		group = self.Configuration.Groups.Default
	}

	if rename, ok := self.Configuration.Rename[name]; ok {
		kind = rename
	}

	return &Node{
		Name:          name,
		Group:         group,
		Version:       version,
		Kind:          kind,
		OriginalGroup: originalGroup,
		OriginalKind:  originalKind,
		Schema:        schema,
		Generator:     self,
	}
}

func (self *Node) GetOverride() *Type {
	return self.Generator.Configuration.Override[self.Kind]
}

func (self *Node) GetParent() *Node {
	if override := self.GetOverride(); override != nil {
		if override.DerivedFrom != "" {
			return self.Generator.Nodes.Find(override.DerivedFrom)
		}
	}
	return nil
}

func (self *Node) GetRemovedFieldNames() []string {
	var names []string

	// Fields given to derived types
	if override := self.GetOverride(); override != nil {
		for _, derive := range override.Derive {
			for _, field := range derive {
				names = append(names, field)
			}
		}
	}

	return names
}

func (self *Node) GetAddedFields() map[string]spec.Schema {
	schemas := make(map[string]spec.Schema)

	// Fields taken from parent type
	parent := self.GetParent()
	if parent != nil {
		if override := parent.GetOverride(); override != nil {
			if names, ok := override.Derive[self.Kind]; ok {
				for _, name := range names {
					for name_, schema := range parent.Schema.Properties {
						if name_ == name {
							schemas[name] = schema
						}
					}
				}
			}
		}
	}

	// An override of an inherited field would not have a schema
	if nodeOverride := self.GetOverride(); nodeOverride != nil {
		for name, override := range nodeOverride.Fields {
			hasProperty := false
			for name_ := range self.Schema.Properties {
				if name_ == name {
					hasProperty = true
					break
				}
			}

			if !hasProperty {
				schemas[name] = spec.Schema{}

				if override.Type == "" {
					// Use type from parent
					if parent != nil {
						for name_, schema := range parent.Schema.Properties {
							if name_ == name {
								override.Type = self.Generator.GetTypeName(schema)
							}
						}
					}
				}

				if override.Required == nil {
					override.Required = &True
				}
			}
		}
	}

	return schemas
}

func (self *Node) IsCapability() bool {
	if override := self.GetOverride(); override != nil {
		switch override.Entity {
		case "capability":
			return true
		case "data":
			return false
		}
	}

	if parent := self.GetParent(); parent != nil {
		// Same entity as parent
		return parent.IsCapability()
	}

	// Another possible heuristic:
	// It has the "x-kubernetes-group-version-kind" Swagger extension

	count := 0
	for name := range self.Schema.Properties {
		switch name {
		case "apiVersion", "kind", "metadata":
			count++
		}
	}
	return count == 3 // has all 3 fields
}

func (self *Node) IsExcluded() bool {
	for _, name := range self.Generator.Configuration.Exclude {
		if name == self.OriginalKind {
			return true
		}
	}
	for group_ := range self.Generator.Configuration.Groups.Imports {
		if group_ == self.OriginalGroup {
			return true
		}
	}
	return false
}

func (self *Node) IsNewest() bool {
	for _, node := range self.Generator.Nodes {
		if (node.Kind == self.Kind) && (node.Group == self.Group) {
			if isVersionNewer(node.Version, self.Version) {
				return false
			}
		}
	}
	return true
}

const apiMachineryGroup = "k8s.io.apimachinery.pkg.apis.meta"

func (self *Node) IsList() bool {
	return self.HasProperty("metadata", apiMachineryGroup, "v1", "ListMeta")
}

func (self *Node) IsObject() bool {
	return self.HasProperty("metadata", apiMachineryGroup, "v1", "ObjectMeta")
}

func (self *Node) SplitFields(isCapability bool) (map[string]spec.Schema, map[string]spec.Schema) {
	properties := make(map[string]spec.Schema)
	attributes := make(map[string]spec.Schema)

	removedNames := self.GetRemovedFieldNames()
	for name, schema := range self.Schema.Properties {
		isRemoved := false
		for _, name_ := range removedNames {
			if name_ == name {
				isRemoved = true
				break
			}
		}
		if isRemoved {
			continue
		}

		if isGVK(schema, apiMachineryGroup, "v1", "ObjectMeta") {
			// This special field should not be set explicitly
			continue
		}

		if isCapability {
			switch name {
			case "apiVersion", "kind":
				// Implicit fields
				continue
			}

			if name == "status" {
				attributes[name] = schema
				continue
			}
		}

		properties[name] = schema
	}

	for name, schema := range self.GetAddedFields() {
		properties[name] = schema
	}

	return properties, attributes
}

func (self *Node) HasProperty(name string, group string, version string, kind string) bool {
	if schema, ok := self.Schema.Properties[name]; ok {
		return isGVK(schema, group, version, kind)
	} else {
		return false
	}
}
