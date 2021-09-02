package main

//
// Configuration
//

type Configuration struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	OpenAPI      string            `yaml:"open-api"`
	ReferenceURL string            `yaml:"reference-url"`
	OutputDir    string            `yaml:"output-dir"`
	Groups       Groups            `yaml:"groups"`
	Exclude      []string          `yaml:"exclude"`
	Rename       map[string]string `yaml:"rename"`
	Add          map[string]*Type  `yaml:"add"`
	Override     map[string]*Type  `yaml:"override"`
}

//
// Groups
//

type Groups struct {
	Default string            `yaml:"default"`
	Imports map[string]Import `yaml:"imports"`
}

//
// Import
//

type Import struct {
	NamespacePrefix string `yaml:"namespace_prefix"`
	File            string `yaml:"file"`
}

//
// Type
//

type Type struct {
	Entity      string              `yaml:"entity"`
	Metadata    []map[string]string `yaml:"metadata"`
	Description string              `yaml:"description"`
	Fields      map[string]*Field   `yaml:"fields"`
	DerivedFrom string              `yaml:"derived_from"`
	Derive      map[string][]string `yaml:"derive"`
}

//
// Field
//

type Field struct {
	Type        string   `yaml:"type"`
	EntrySchema string   `yaml:"entry_schema"`
	Description string   `yaml:"description"`
	Default     string   `yaml:"default"`
	Required    *bool    `yaml:"required"`
	Constraints []string `yaml:"constraints"`
}
