package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/kubernetes-sigs/reference-docs/gen-apidocs/generators/api"
)

//
// Generator
//

type Generator struct {
	entity      string
	excludes    []string
	includes    []*api.Definition
	annotations map[string]Annotation

	config      *api.Config
	definitions api.SortDefinitionsByName
	writer      io.StringWriter
}

func (self *Generator) Gather() {
	for _, definitions := range self.config.Definitions.ByKind {
		for _, definition := range definitions {
			isCapability := self.IsDefinitionCapability(definition)
			if self.entity == "data" {
				// Exclude capabilities
				if isCapability {
					continue
				}
			} else {
				// Only capabilities
				if !isCapability {
					continue
				}
			}

			// Exclusions
			if self.IsDefinitionExcluded(definition) {
				continue
			}

			// Exclude list resources
			if self.IsDefinitionList(definition) {
				continue
			}

			self.definitions = append(self.definitions, definition)
			// Stop of after first version (which is latest)
			break
		}
	}

	for _, definition := range self.includes {
		isCapability := self.IsDefinitionCapability(definition)
		if ((self.entity == "data") && !isCapability) || ((self.entity == "capability") && isCapability) {
			self.definitions = append(self.definitions, definition)
		}
	}

	for _, definition := range self.definitions {
		for _, derived := range self.GetDerivedDefinitions(definition) {
			self.definitions = append(self.definitions, derived)
		}
	}

	sort.Sort(self.definitions)
}

func (self *Generator) Generate() {
	self.Println("tosca_definitions_version: tosca_simple_yaml_1_3")

	self.Println()
	self.Println("# This file was automatically generated from data published at:")
	self.Printlnf("# %s", *sourceUrl)

	self.Println()
	self.Println("metadata:")
	self.Println()
	self.PrintText(2, "specification.name", self.config.SpecTitle, false)
	self.PrintText(2, "specification.version", self.config.SpecVersion[1:], false)

	self.Println()
	self.Println("imports:")
	self.Println()

	if self.entity == "capability" {
		self.Println("- data.yaml")
	} else {
		self.Println("- _data.yaml")
	}

	self.Println()
	self.Printlnf("%s_types:", self.entity)

	for _, definition := range self.definitions {
		self.Println()
		self.PrintDefinition(definition)
	}
}

func (self *Generator) PrintDefinition(definition *api.Definition) {
	name := self.GetDefinitionName(definition)
	self.Printlnf("  %s:", name)

	self.Println("    metadata:")
	if !self.IsDefinitionGenerated(definition) {
		self.PrintText(6, "specification.url", fmt.Sprintf(*referenceUrl, *api.KubernetesRelease, definition.LinkID()), false)
	}
	self.PrintText(6, "specification.version", self.config.SpecVersion[1:], false)

	if category := self.GetDefinitionCategory(definition); category != nil {
		self.PrintText(6, "specification.category", category.Name, false)
	}

	referredFrom := strings.Join(self.GetDefinitionReferredFrom(definition), ", ")
	self.PrintText(6, "specification.referredFrom", referredFrom, false)

	refersTo := strings.Join(self.GetDefinitionRefersTo(definition), ", ")
	self.PrintText(6, "specification.refersTo", refersTo, false)

	if self.IsDefinitionCapability(definition) {
		if name != "Metadata" {
			var apiVersion string
			if definition.GroupFullName != "core" {
				apiVersion = fmt.Sprintf("%s/%s", definition.GroupFullName, definition.Version)
			} else {
				apiVersion = definition.Version.String()
			}
			self.PrintText(6, "turandot.apiVersion", apiVersion, false)
			if name != definition.Name {
				self.PrintText(6, "turandot.kind", definition.Name, false)
			}
		}
	}

	if self.DoesDefinitionNeedMetadata(definition) {
		self.Println("      turandot.metadata: 'true'")
	}

	if annotation, ok := self.annotations[definition.Name]; ok {
		for _, metadata := range annotation.metadata {
			self.Printlnf("      %s", metadata)
		}
	}

	self.PrintText(4, "description", definition.Description(), true)

	if parent := self.GetDefinitionParent(definition); parent != nil {
		self.PrintText(4, "derived_from", self.GetDefinitionName(parent), true)
	}

	fields := self.GetDefinitionPropertyFields(definition)
	if len(fields) > 0 {
		self.Println("    properties:")
		for _, field := range fields {
			var refine []string
			if annotation, ok := self.annotations[definition.Name]; ok {
				if annotation.refine != nil {
					refine, _ = annotation.refine[field.Name]
				}
			}

			self.PrintField(field, refine, false)
		}
	}

	fields = self.GetDefinitionAttributeFields(definition)
	if len(fields) > 0 {
		self.Println("    attributes:")
		for _, field := range fields {
			self.PrintField(field, nil, true)
		}
	}
}

func (self *Generator) PrintField(field *api.Field, refine []string, isAttribute bool) {
	self.Printlnf("      %s:", field.Name)

	if isAttribute || (field.PatchStrategy != "") {
		self.Println("        metadata:")
		if isAttribute {
			self.PrintText(10, "puccini.information:turandot.mapping", field.Name, false)
		}
		if field.PatchStrategy != "" {
			self.PrintText(10, "turandot.patchStrategy", field.PatchStrategy, false)
			self.PrintText(10, "turandot.patchMergeKey", field.PatchMergeKey, false)
		}
	}

	self.PrintText(8, "description", field.Description, true)

	type_ := field.Type

	array := false
	if strings.HasSuffix(type_, " array") {
		type_ = type_[:len(field.Type)-6]
		array = true
	}

	if field.Definition != nil {
		type_ = self.GetDefinitionName(field.Definition)
	}

	switch type_ {
	case "object", "":
		type_ = "Any"

	case "number":
		type_ = "float"
	}

	printType := true
	printEntrySchema := true
	printRequiredFalse := !isAttribute
	for _, refine_ := range refine {
		if strings.HasPrefix(refine_, "type:") {
			printType = false
		} else if strings.HasPrefix(refine_, "entry_schema:") {
			printEntrySchema = false
		} else if strings.HasPrefix(refine_, "default:") {
			printRequiredFalse = false
		}
	}

	if printType {
		if array {
			self.Println("        type: list")
			if printEntrySchema {
				self.PrintText(8, "entry_schema", type_, false)
			}
		} else {
			self.PrintText(8, "type", type_, false)
		}
	}

	for _, refine_ := range refine {
		self.Printlnf("        %s", refine_)
	}

	if printRequiredFalse {
		self.Println("        required: false")
	}
}
