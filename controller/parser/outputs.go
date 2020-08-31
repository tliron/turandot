package parser

import (
	"fmt"

	"github.com/tliron/kutil/ard"
	cloutpkg "github.com/tliron/puccini/clout"
)

func GetOutputs(clout *cloutpkg.Clout) (map[string]string, bool) {
	if tosca, ok := clout.Properties["tosca"]; ok {
		if outputs, ok := ard.NewNode(tosca).Get("outputs").StringMap(true); ok {
			outputs_ := make(map[string]string)
			for name, output := range outputs {
				outputs_[name] = fmt.Sprintf("%v", output)
			}
			return outputs_, true
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}
}
