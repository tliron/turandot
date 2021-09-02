package main

import (
	"strings"

	"github.com/go-openapi/spec"
)

const extensionName = "x-kubernetes-group-version-kind"
const defaultGroupPrefix = "k8s.io.api."
const coreGroup = "core"

func (self *Generator) GetTypeName(schema spec.Schema) string {
	name, group, _, kind := getGVK(schema)
	if rename, ok := self.Configuration.Rename[name]; ok {
		return rename
	} else if import_, ok := self.Configuration.Groups.Imports[group]; ok {
		return import_.NamespacePrefix + ":" + kind
	} else if kind != "" {
		return kind
	} else {
		return name
	}
}

func parseExtension(extension interface{}) (string, string, string) {
	extension_ := extension.([]interface{})
	map_ := extension_[0].(map[string]interface{})
	group := fixGroup(map_["group"].(string))
	version := map_["version"].(string)
	kind := map_["kind"].(string)
	return group, version, kind
}

func parseGVK(name string) (string, string, string) {
	split := strings.Split(name, ".")
	length := len(split)
	if length < 2 {
		return "", "", ""
	}
	group := fixGroup(strings.Join(split[0:length-2], "."))
	version := split[length-2]
	kind := split[length-1]
	return group, version, kind
}

func getGVK(schema spec.Schema) (string, string, string, string) {
	if len(schema.Type) > 0 {
		return schema.Type[0], "", "", ""
	} else {
		ref := schema.Ref.GetPointer().String() // example: #/definitions/k8s.io.api.core.v1.LocalObjectReference
		split := strings.Split(ref, "/")
		name := split[len(split)-1]
		group, version, kind := parseGVK(name)
		return name, group, version, kind
	}
}

func isGVK(schema spec.Schema, group string, version string, kind string) bool {
	_, group_, version_, kind_ := getGVK(schema)
	return (group_ == group) && (version_ == version) && (kind_ == kind)
}

func fixGroup(group string) string {
	// In the main Kubernetes API the standard prefix is reversed for some reason
	if strings.HasPrefix(group, "io.k8s.") {
		group = "k8s.io." + group[len("io.k8s."):]
	}
	return group
}
