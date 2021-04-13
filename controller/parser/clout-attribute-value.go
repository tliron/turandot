package parser

import (
	"github.com/tliron/kutil/format"
)

//
// CloutAttributeValue
//

type CloutAttributeValue struct {
	CapabilityName string      `json:"capability"`
	AttributeName  string      `json:"attribute"`
	Value          interface{} `json:"value"`
}

//
// CloutAttributeValueList
//

type CloutAttributeValueList []*CloutAttributeValue

//
// CloutAttributeValues
//

type CloutAttributeValues map[string]CloutAttributeValueList

func NewCloutAttributeValues() CloutAttributeValues {
	return make(CloutAttributeValues)
}

func (self CloutAttributeValues) Set(vertexId string, capabilityName string, attributeName string, value interface{}) {
	self[vertexId] = append(self[vertexId], &CloutAttributeValue{
		CapabilityName: capabilityName,
		AttributeName:  attributeName,
		Value:          value,
	})
}

func (self CloutAttributeValues) JSON() map[string]string {
	map_ := make(map[string]string)
	for vertexId, list := range self {
		if value, err := format.EncodeJSON(list, ""); err == nil {
			map_[vertexId] = value
		}
	}
	return map_
}
