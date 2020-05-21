package parser

import (
	"github.com/tliron/puccini/ard"
)

//
// OrchestrationScriptletOperation
//

type OrchestrationScriptletOperation struct {
	VertexID      string
	ScriptletName string
}

func NewOrchestrationScriptletOperation(value ard.Value) (*OrchestrationScriptletOperation, bool) {
	operation := ard.NewNode(value)
	if vertexId, ok := operation.Get("vertexId").String(false); ok {
		if scriptletName, ok := operation.Get("scriptletName").String(false); ok {
			return &OrchestrationScriptletOperation{vertexId, scriptletName}, true
		}
	}
	return nil, false
}

//
// OrchestrationOperations
//

type OrchestrationOperations []*OrchestrationScriptletOperation

func NewOrchestrationOperations(value ard.Value) (OrchestrationOperations, bool) {
	if operations, ok := ard.NewNode(value).Get("operations").List(false); ok {
		self := make(OrchestrationOperations, len(operations))
		for index, operation := range operations {
			if operation_, ok := NewOrchestrationScriptletOperation(operation); ok {
				self[index] = operation_
			} else {
				return nil, false
			}
		}
		return self, true
	} else {
		return nil, false
	}
}
