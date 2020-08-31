package commands

import (
	"io"
	"os"

	"github.com/op/go-logging"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

const toolName = "turandot"

var log = logging.MustGetLogger(toolName)

var filePath string
var directoryPath string
var url string
var component string
var tail int
var follow bool
var all bool
var wait bool

func Logs(appNameSuffix string, containerName string) {
	// TODO: what happens if we follow more than one log?
	readers, err := NewClient().Turandot().Logs(appNameSuffix, containerName, tail, follow)
	util.FailOnError(err)
	for _, reader := range readers {
		defer reader.Close()
	}
	for _, reader := range readers {
		io.Copy(terminal.Stdout, reader)
	}
}

func Shell(appNameSuffix string, containerName string) {
	err := NewClient().Turandot().Shell(appNameSuffix, containerName, os.Stdin, terminal.Stdout, terminal.Stderr)
	util.FailOnError(err)
}
