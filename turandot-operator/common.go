package main

import (
	contextpkg "context"

	"github.com/op/go-logging"
)

const toolName = "turandot-operator"

var context = contextpkg.TODO()

var log = logging.MustGetLogger(toolName)
