package commands

import (
	contextpkg "context"

	"github.com/tliron/commonlog"
	"github.com/tliron/go-transcribe"
)

const toolName = "turandot"

var context = contextpkg.TODO()

var log = commonlog.GetLogger(toolName)

var filePath string
var directoryPath string
var url string
var component string
var tail int
var follow bool
var all bool
var site string
var wait bool
var registry string

func Transcriber() *transcribe.Transcriber {
	return &transcribe.Transcriber{
		Strict: strict,
		Pretty: pretty,
		Base64: base64,
	}
}
