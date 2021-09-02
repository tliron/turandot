package main

import (
	"sort"

	"github.com/go-openapi/spec"
)

//
// Nodes
//

type Nodes []*Node

var True = true

func (self *Generator) ReadNodes() Nodes {
	var nodes Nodes

	for name, schema := range self.OpenAPI.Spec().Definitions {
		node := self.NewNode(name, schema)
		if !node.IsExcluded() && !node.IsList() {
			nodes = append(nodes, node)
		}
	}

	// Add nodes
	for name, add := range self.Configuration.Add {
		// We'll reuse the override codepath
		self.Configuration.Override[name] = add

		var schema spec.Schema
		schema.Description = add.Description

		// Properties
		schema.Properties = make(spec.SchemaProperties)
		for name_ := range add.Fields {
			var property spec.Schema
			property.Type = spec.StringOrArray{".." + name_}
			schema.Properties[name_] = property
			schema.Required = append(schema.Required, name_)
		}

		nodes = append(nodes, self.NewNode(".."+name, schema))
	}

	// Add derived nodes
	parentNames := make(map[string]string)
	for parentName, override := range self.Configuration.Override {
		for childName := range override.Derive {
			parentNames[childName] = parentName
			var schema spec.Schema
			nodes = append(nodes, self.NewNode(".."+childName, schema))
		}
	}

	// Set parents for derived nodes
	for childName, parentName := range parentNames {
		override, ok := self.Configuration.Override[childName]
		if !ok {
			override = new(Type)
			self.Configuration.Override[childName] = override
		}
		override.DerivedFrom = parentName
	}

	return nodes
}

func (self *Generator) SplitNodes() (Nodes, Nodes) {
	var capabilityNodes Nodes
	var dataNodes Nodes

	for _, node := range self.Nodes {
		if !node.IsNewest() {
			continue
		}

		if node.IsCapability() {
			capabilityNodes = append(capabilityNodes, node)
		} else {
			dataNodes = append(dataNodes, node)
		}
	}

	sort.Sort(capabilityNodes)
	sort.Sort(dataNodes)

	return capabilityNodes, dataNodes
}

func (self Nodes) Find(name string) *Node {
	for _, node := range self {
		if node.Kind == name {
			return node
		}
	}
	return nil
}

// sort.Interface
func (self Nodes) Len() int {
	return len(self)
}

// sort.Interface
func (self Nodes) Less(i int, j int) bool {
	return self[i].Kind < self[j].Kind
}

// sort.Interface
func (self Nodes) Swap(i int, j int) {
	self[i], self[j] = self[j], self[i]
}
