package controller

import (
	contextpkg "context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/tliron/commonlog"
	"github.com/tliron/exturl"
	"github.com/tliron/go-ard"
	problemspkg "github.com/tliron/kutil/problems"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/transcribe"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
	"github.com/tliron/puccini/tosca/parser"
)

var pucciniLog = commonlog.GetLogger("turandot.puccini")

var parserContext = parser.NewContext()

func CompileTOSCA(context contextpkg.Context, url string, inputs map[string]ard.Value, writer io.Writer, urlContext *exturl.Context) error {
	if url_, err := urlContext.NewURL(url); err == nil {
		// TODO: origins
		if _, serviceTemplate, problems, err := parserContext.Parse(context, url_, nil, terminal.NewStylist(false), nil, inputs); err == nil {
			if problems.Empty() {
				if clout, err := serviceTemplate.Compile(); err == nil {
					return WriteClout(clout, writer)
				} else {
					return err
				}
			} else {
				return errors.New(problems.ToString(true))
			}
		} else if (problems != nil) && !problems.Empty() {
			return fmt.Errorf("%s\n%s", err.Error(), problems.ToString(true))
		} else {
			return err
		}
	} else {
		return err
	}
}

func ReadClout(reader io.Reader, urlContext *exturl.Context) (*cloutpkg.Clout, error) {
	if clout, err := cloutpkg.Read(reader, "yaml"); err == nil {
		var problems problemspkg.Problems
		if js.Resolve(clout, &problems, urlContext, false, "yaml", false, false); problems.Empty() {
			return clout, nil
		} else {
			return nil, errors.New(problems.ToString(true))
		}
	} else {
		return nil, err
	}
}

func WriteClout(clout *cloutpkg.Clout, writer io.Writer) error {
	return transcribe.Write(clout, "yaml", terminal.Indent, false, writer)
}

func RequireCloutScriptlet(clout *cloutpkg.Clout, scriptletName string, arguments map[string]string, urlContext *exturl.Context) (string, error) {
	jsContext := js.NewContext(scriptletName, pucciniLog, arguments, false, "yaml", false, false, "", urlContext)
	var builder strings.Builder
	jsContext.Stdout = &builder
	if _, err := jsContext.Require(clout, scriptletName, nil); err == nil {
		return builder.String(), nil
	} else {
		return "", err
	}
}
