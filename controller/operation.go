package controller

import (
	"fmt"

	cloutpkg "github.com/tliron/puccini/clout"
	urlpkg "github.com/tliron/puccini/url"
	"github.com/tliron/turandot/common"
	"github.com/tliron/turandot/controller/parser"
)

// See:
//   https://github.com/cosiner/socker
//   https://github.com/pressly/sup

func (self *Controller) processOperations(operations interface{}, clout *cloutpkg.Clout, urlContext *urlpkg.Context) error {
	if operations, ok := parser.NewOrchestrationOperations(operations); ok {
		for _, operation := range operations {
			self.Log.Infof("executing scriptlet %s on vertex %s", operation.ScriptletName, operation.VertexID)
			if _, err := common.ExecScriptlet(clout, operation.ScriptletName, urlContext); err != nil {
				return err
			}
		}
		return nil
	} else {
		return fmt.Errorf("could not parse operations: %v", operations)
	}
}
