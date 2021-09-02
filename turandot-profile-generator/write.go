package main

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/go-openapi/spec"
)

func (self *Generator) WriteHeader(entity string) {
	self.WriteKeyValue(0, "tosca_definitions_version", "tosca_simple_yaml_1_3", false)

	self.Writeln()
	self.Writeln("# This file was automatically generated from data published at:")
	self.Writelnf("# %s", self.Configuration.OpenAPI)

	self.Writeln()
	self.WriteKey(0, "metadata")
	self.Writeln()
	self.WriteKeyValue(1, "specification.name", self.Configuration.Name, false)
	self.WriteKeyValue(1, "specification.version", self.Configuration.Version, false)
	self.WriteKeyValue(1, "specification.url", self.Configuration.ReferenceURL, false)

	self.Writeln()
	self.WriteKey(0, "imports")
	self.Writeln()

	switch entity {
	case "capability":
		self.Writeln("- data.yaml")

	case "data":
		path := filepath.Join(filepath.Dir(self.ConfigurationPath), self.Configuration.OutputDir, "_data.yaml")
		if _, err := os.Stat(path); err == nil {
			self.Writeln("- _data.yaml")
		}
	}

	for _, import_ := range self.Configuration.Groups.Imports {
		self.Writeln("- namespace_prefix: ", import_.NamespacePrefix)
		self.Writeln("  file: ", import_.File)
	}
}

func (self *Generator) WriteType(node *Node, isCapability bool) {
	self.WriteKey(1, node.Kind)

	self.WriteKey(2, "metadata")
	self.WriteKeyValue(3, "specification.name", self.Configuration.Name, false)
	self.WriteKeyValue(3, "specification.version", self.Configuration.Version, false)
	if isCapability && (node.Kind != "Metadata") {
		apiVersion := node.Version
		if node.Group != "" {
			apiVersion = node.Group + "/" + apiVersion
		}
		self.WriteKeyValue(3, "turandot.apiVersion", apiVersion, false)
		if node.OriginalKind != node.Kind {
			self.WriteKeyValue(3, "turandot.kind", node.OriginalKind, false)
		}
		if node.IsObject() {
			self.WriteKeyValue(3, "turandot.metadata", "'true'", false)
		}
	}

	var derivedFrom string
	if override := node.GetOverride(); override != nil {
		for _, metadata := range override.Metadata {
			for key, value := range metadata {
				self.WriteKeyValue(3, key, value, false)
				break // only one key
			}
		}
		derivedFrom = override.DerivedFrom
	}

	self.WriteKeyValue(2, "description", node.Schema.Description, true)

	if derivedFrom != "" {
		self.WriteKeyValue(2, "derived_from", derivedFrom, false)
	}

	properties, attributes := node.SplitFields(isCapability)
	self.WriteFields("properties", node, properties)
	self.WriteFields("attributes", node, attributes)
}

func (self *Generator) WriteFields(type_ string, node *Node, fields map[string]spec.Schema) {
	if len(fields) > 0 {
		self.WriteKey(2, type_)

		names := make([]string, 0, len(fields))
		for name := range fields {
			names = append(names, name)
		}
		sort.Strings(names)

		nodeOverride := node.GetOverride()
		for _, name := range names {
			schema := fields[name]

			required := type_ == "attributes" // attributes don't have "required" keyword
			if !required {
				for _, name_ := range node.Schema.Required {
					if name_ == name {
						required = true
						break
					}
				}
			}

			var override *Field
			if nodeOverride != nil {
				override, _ = nodeOverride.Fields[name]
			}

			self.WriteField(name, schema, required, override)
		}
	}
}

func (self *Generator) WriteField(name string, schema spec.Schema, required bool, override *Field) {
	var description string
	var type_ string
	var entrySchema string
	var default_ string

	if override != nil {
		description = override.Description
		type_ = override.Type
		entrySchema = override.EntrySchema
		default_ = override.Default
		if override.Required != nil {
			required = *override.Required
		}
	}

	if description == "" {
		description = schema.Description
	}

	if (entrySchema == "") && (len(schema.Type) > 0) {
		switch schema.Type[0] {
		case "array":
			if (schema.Items != nil) && (schema.Items.Schema != nil) {
				type_ = "list"
				entrySchema = self.GetTypeName(*schema.Items.Schema)
			}

		case "object":
			if (schema.AdditionalProperties != nil) && (schema.AdditionalProperties.Schema != nil) {
				type_ = "map"
				entrySchema = self.GetTypeName(*schema.AdditionalProperties.Schema)
			}
		}
	}

	if type_ == "" {
		type_ = self.GetTypeName(schema)
	}

	self.WriteKey(3, name)
	self.WriteKeyValue(4, "description", description, true)
	self.WriteKeyValue(4, "type", type_, false)
	self.WriteKeyValue(4, "entry_schema", entrySchema, false)
	self.WriteKeyValue(4, "default", default_, false)

	if !required {
		// The TOSCA default is required=true
		self.WriteKeyValue(4, "required", "false", false)
	}

	if (override != nil) && (len(override.Constraints) > 0) {
		self.WriteKey(4, "constraints")
		for _, constraint := range override.Constraints {
			self.Writeln(indentation(5), "- ", constraint)
		}
	}
}
