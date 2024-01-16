package parser

import (
	"fmt"

	"github.com/tliron/go-ard"
	cloutpkg "github.com/tliron/puccini/clout"
)

func GetOutputs(clout *cloutpkg.Clout) (map[string]string, bool) {
	if tosca, ok := clout.Properties["tosca"]; ok {
		if outputs, ok := ard.With(tosca).Get("outputs").NilMeansZero().StringMap(); ok {
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
