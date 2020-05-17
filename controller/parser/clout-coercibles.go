package parser

import (
	"github.com/tliron/puccini/ard"
)

func ToCloutCoercible(value ard.Value) ard.StringMap {
	coercible := make(ard.StringMap)

	switch value_ := value.(type) {
	case ard.List:
		list := make(ard.List, len(value_))
		for index, element := range value_ {
			list[index] = ToCloutCoercible(element)
		}
		coercible["$list"] = list

	case ard.StringMap:
		var list ard.List
		for k, v := range value_ {
			element := ToCloutCoercible(v)
			element["$key"] = ToCloutCoercible(k)
			list = append(list, element)
		}
		coercible["$map"] = list

	default:
		coercible["$value"] = value
	}

	return coercible
}

func CloutCoerciblesMerged(a ard.StringMap, b ard.StringMap) bool {
	if bValue, ok := b["$value"]; ok {
		if aValue, ok := a["$value"]; ok {
			return ard.Equals(aValue, bValue)
		}
	}

	if bList, ok := b["$list"]; ok {
		if aList, ok := a["$list"]; ok {
			return ard.Equals(aList, bList)
		}
	}

	if bMap, ok := b["$map"]; ok {
		if aMap, ok := a["$map"]; ok {
			return ard.Equals(aMap, bMap)
		}
	}

	return false
}

func MergeCloutCoercibles(a ard.StringMap, b ard.StringMap) {
	if value, ok := b["$value"]; ok {
		a["$value"] = value
		delete(a, "$list")
		delete(a, "$map")
		delete(a, "$functionCall")
		return
	}

	if list, ok := b["$list"]; ok {
		a["$list"] = list
		delete(a, "$value")
		delete(a, "$map")
		delete(a, "$functionCall")
		return
	}

	if map_, ok := b["$map"]; ok {
		a["$map"] = map_
		delete(a, "$value")
		delete(a, "$list")
		delete(a, "$functionCall")
		return
	}
}
