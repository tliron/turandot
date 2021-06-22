package controller

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/format"
	"github.com/tliron/kutil/logging"
	problemspkg "github.com/tliron/kutil/problems"
	"github.com/tliron/kutil/terminal"
	urlpkg "github.com/tliron/kutil/url"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
	"github.com/tliron/puccini/tosca/compiler"
	"github.com/tliron/puccini/tosca/parser"
)

var pucciniLog = logging.GetLogger("turandot.puccini")

func CompileTOSCA(url string, inputs map[string]ard.Value, writer io.Writer, urlContext *urlpkg.Context) error {
	if url_, err := urlpkg.NewURL(url, urlContext); err == nil {
		if _, serviceTemplate, problems, err := parser.Parse(url_, terminal.NewStylist(false), nil, inputs); err == nil {
			if problems.Empty() {
				if clout, err := compiler.Compile(serviceTemplate, true); err == nil {
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

func ReadClout(reader io.Reader, urlContext *urlpkg.Context) (*cloutpkg.Clout, error) {
	if clout, err := cloutpkg.Read(reader, "yaml"); err == nil {
		var problems problemspkg.Problems
		if compiler.Resolve(clout, &problems, urlContext, false, "yaml", false, true, false); problems.Empty() {
			return clout, nil
		} else {
			return nil, errors.New(problems.ToString(true))
		}
	} else {
		return nil, err
	}
}

func WriteClout(clout *cloutpkg.Clout, writer io.Writer) error {
	return format.Write(clout, "yaml", terminal.Indent, false, writer)
}

func RequireCloutScriptlet(clout *cloutpkg.Clout, scriptletName string, arguments map[string]string, urlContext *urlpkg.Context) (string, error) {
	jsContext := js.NewContext(scriptletName, pucciniLog, arguments, false, "yaml", false, true, false, "", urlContext)
	var builder strings.Builder
	jsContext.Stdout = &builder
	if _, err := jsContext.Require(clout, scriptletName, nil); err == nil {
		return builder.String(), nil
	} else {
		return "", err
	}
}
