package commands

import (
	contextpkg "context"

	"github.com/tliron/kutil/logging"
)

const toolName = "turandot"

var context = contextpkg.TODO()

var log = logging.GetLogger(toolName)

var filePath string
var directoryPath string
var url string
var component string
var sourceRegistry string
var tail int
var follow bool
var all bool
var site string
var wait bool
var registry string
