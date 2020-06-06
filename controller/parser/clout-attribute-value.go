package parser

import (
	"encoding/json"

	puccinicommon "github.com/tliron/puccini/common"
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

func (self CloutAttributeValues) StringMap() map[string]string {
	map_ := make(map[string]string)
	for vertexId, list := range self {
		if bytes, err := json.Marshal(list); err == nil {
			map_[vertexId] = puccinicommon.BytesToString(bytes)
		}
	}
	return map_
}
