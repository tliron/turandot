package main

import (
	"github.com/kubernetes-sigs/reference-docs/gen-apidocs/generators/api"
)

func (self *Generator) IsDefinitionExcluded(definition *api.Definition) bool {
	for _, name := range self.excludes {
		if name == definition.Name {
			return true
		}
	}
	return false
}

func (self *Generator) IsDefinitionList(definition *api.Definition) bool {
	for _, field := range definition.Fields {
		if field.Type == "ListMeta" {
			return true
		}
	}
	return false
}

func (self *Generator) IsDefinitionCapability(definition *api.Definition) bool {
	if annotation, ok := self.annotations[definition.Name]; ok {
		switch annotation.entity {
		case "capability":
			return true
		case "data":
			return false
		}
	}

	//return self.GetDefinitionCategory(definition) != nil
	count := 0
	for _, field := range definition.Fields {
		switch field.Name {
		case "apiVersion", "kind", "metadata":
			count++
		}
	}
	return count == 3
}

func (self *Generator) IsDefinitionGenerated(definition *api.Definition) bool {
	for _, definition_ := range self.includes {
		if definition_ == definition {
			return true
		}
	}

	for _, definition_ := range self.definitions {
		for _, derived := range self.GetDerivedDefinitions(definition_) {
			if derived.Name == definition.Name {
				return true
			}
		}
	}

	return false
}

func (self *Generator) DoesDefinitionNeedMetadata(definition *api.Definition) bool {
	for _, field := range definition.Fields {
		if field.Type == "ObjectMeta" {
			return true
		}
	}
	return false
}

func (self *Generator) GetDefinitionName(definition *api.Definition) string {
	if annotation, ok := self.annotations[definition.Name]; ok {
		if annotation.rename != "" {
			return annotation.rename
		}
	}

	count := 0
	for _, d := range self.definitions {
		if d.Name == definition.Name {
			count++
		}
	}
	if count > 1 {
		// Not unique
		return definition.GroupFullName + "." + definition.Name
	} else {
		return definition.Name
	}
}

func (self *Generator) GetDefinitionParent(definition *api.Definition) *api.Definition {
	if annotation, ok := self.annotations[definition.Name]; ok {
		if annotation.parent != "" {
			for _, definition_ := range self.definitions {
				if definition_.Name == annotation.parent {
					return definition_
				}
			}
		}
	}

	for _, definition_ := range self.definitions {
		for _, derived := range self.GetDerivedDefinitions(definition_) {
			if derived.Name == definition.Name {
				return definition_
			}
		}
	}

	return nil
}

func (self *Generator) GetDefinitionCategory(definition *api.Definition) *api.ResourceCategory {
	for _, category := range self.config.ResourceCategories {
		for _, resource := range category.Resources {
			if resource.Definition == definition {
				return &category
			}
		}
	}
	return nil
}

func (self *Generator) GetDefinitionPropertyFields(definition *api.Definition) []*api.Field {
	var derive []string
	if annotation, ok := self.annotations[definition.Name]; ok {
		for _, child := range annotation.children {
			derive = append(derive, child.fields...)
		}
	}

	isCapability := self.IsDefinitionCapability(definition)

	var fields []*api.Field
	for _, field := range definition.Fields {
		skip := false

		// Skip fields that are supposed to be in derived definitions
		for _, derive_ := range derive {
			if derive_ == field.Name {
				skip = true
			}
		}

		// Skip parent fields
		if !skip && !self.IsDefinitionGenerated(definition) {
			if parent := self.GetDefinitionParent(definition); parent != nil {
				for _, field_ := range parent.Fields {
					if field_.Name == field.Name {
						skip = true
						break
					}
				}
			}
		}

		if !skip {
			if isCapability && (field.Name == "status") {
				// Skip status subresource
				skip = true
			} else if field.Type == "ObjectMeta" {
				// See: DoesDefinitionNeedMetadata
				skip = true
			} else if self.IsDefinitionCapability(definition) {
				// Skip generated fields and metadata
				switch field.Name {
				case "apiVersion", "kind", "metadata":
					skip = true
				}
			}
		}

		if !skip {
			fields = append(fields, field)
		}
	}
	return fields
}

func (self *Generator) GetDefinitionAttributeFields(definition *api.Definition) []*api.Field {
	var fields []*api.Field
	if self.IsDefinitionCapability(definition) {
		for _, field := range definition.Fields {
			if field.Name == "status" {
				fields = append(fields, field)
			}
		}
	}
	return fields
}

func (self *Generator) GetDefinitionReferredFrom(definition *api.Definition) []string {
	if annotation, ok := self.annotations[definition.Name]; ok {
		if annotation.noReferences {
			return nil
		}
	}

	var referredFrom []string
	for _, appearsIn := range definition.AppearsIn {
		if !self.IsDefinitionExcluded(appearsIn) && !self.IsDefinitionList(appearsIn) {
			referredFrom = append(referredFrom, self.GetDefinitionName(appearsIn))
		}
	}
	return referredFrom
}

func (self *Generator) GetDefinitionRefersTo(definition *api.Definition) []string {
	if annotation, ok := self.annotations[definition.Name]; ok {
		if annotation.noReferences {
			return nil
		}
	}

	var refersTo []string
	for _, inline := range definition.Inline {
		if !self.IsDefinitionExcluded(inline) && !self.IsDefinitionList(inline) {
			refersTo = append(refersTo, self.GetDefinitionName(inline))
		}
	}
	return refersTo
}

func (self *Generator) GetDerivedDefinitions(definition *api.Definition) []*api.Definition {
	var definitions []*api.Definition

	if annotation, ok := self.annotations[definition.Name]; ok {
		for name := range annotation.children {
			definitions = append(definitions, self.GetDerivedDefinition(definition, name))
		}
	}

	return definitions
}

func (self *Generator) GetDerivedDefinition(definition *api.Definition, name string) *api.Definition {
	var fields []*api.Field

	// Refined fields
	if annotation, ok := self.annotations[name]; ok {
		if annotation.refine != nil {
			for name := range annotation.refine {
				for _, field := range definition.Fields {
					if field.Name == name {
						fields = append(fields, &api.Field{
							Name: name,
							Type: field.Type,
						})
					}
				}
			}
		}
	}

	child := self.annotations[definition.Name].children[name]
	for _, field := range child.fields {
		for _, field_ := range definition.Fields {
			if field_.Name == field {
				fields = append(fields, field_)
			}
		}
	}

	return &api.Definition{
		Name:   name,
		Fields: fields,
	}
}
