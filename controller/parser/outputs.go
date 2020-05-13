package parser

import (
	"github.com/tliron/puccini/ard"
	cloutpkg "github.com/tliron/puccini/clout"
)

func GetOutputs(clout *cloutpkg.Clout) (map[string]string, bool) {
	if tosca, ok := clout.Properties["tosca"]; ok {
		if outputs, ok := ard.NewNode(tosca).Get("outputs").StringMap(true); ok {
			outputs_ := make(map[string]string)
			for name, output := range outputs {
				if output_, ok := output.(string); ok {
					outputs_[name] = output_
				} else {
					return nil, false
				}
			}
			return outputs_, true
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}
}
