package commands

import (
	"io"
	"os"

	"github.com/op/go-logging"
	terminalutil "github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
	"golang.org/x/crypto/ssh/terminal"
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
var site string
var registry string
var wait bool

func Logs(appNameSuffix string, containerName string) {
	// TODO: what happens if we follow more than one log?
	readers, err := NewClient().Turandot().Logs(appNameSuffix, containerName, tail, follow)
	util.FailOnError(err)
	for _, reader := range readers {
		defer reader.Close()
	}
	for _, reader := range readers {
		io.Copy(terminalutil.Stdout, reader)
	}
}

func Shell(appNameSuffix string, containerName string) {
	// We need stdout to be in raw"" mode
	fd := int(os.Stdout.Fd())
	state, err := terminal.MakeRaw(fd)
	util.FailOnError(err)
	defer terminal.Restore(fd, state)
	err = NewClient().Turandot().Shell(appNameSuffix, containerName, os.Stdin, os.Stdout, os.Stderr)
	util.FailOnError(err)
}
