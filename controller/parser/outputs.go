package parser

import (
	"github.com/tliron/puccini/ard"
)

func ParseOutputs(data interface{}) (map[string]string, bool) {
	if outputs, ok := ard.NewNode(data).Get("outputs").Map(false); ok {
		outputs_ := make(map[string]string)
		for name, output := range outputs {
			if name_, ok := name.(string); ok {
				if output_, ok := output.(string); ok {
					outputs_[name_] = output_
				} else {
					return nil, false
				}
			} else {
				return nil, false
			}
		}
		return outputs_, true
	} else {
		return nil, false
	}
}
