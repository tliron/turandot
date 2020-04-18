package common

import (
	"bytes"
	templatepkg "text/template"

	"github.com/tliron/puccini/ard"
	"github.com/tliron/yamlkeys"
)

func DecodeYAMLTemplate(code string, data interface{}) (map[string]interface{}, error) {
	if template, err := templatepkg.New("").Parse(code); err == nil {
		var buffer bytes.Buffer
		if err := template.Execute(&buffer, data); err == nil {
			if value, err := yamlkeys.DecodeString(buffer.String()); err == nil {
				return ard.EnsureStringMaps(value), nil
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func DecodeAllYAML(code string) ([]ard.StringMap, error) {
	if values, err := yamlkeys.DecodeStringAll(code); err == nil {
		var r []ard.StringMap
		for _, value := range values {
			r = append(r, ard.EnsureStringMaps(value))
		}
		return r, nil
	} else {
		return nil, err
	}
}
