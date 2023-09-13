package controller

import (
	contextpkg "context"
	"errors"
	"io"
	"strings"

	"github.com/tliron/commonlog"
	"github.com/tliron/exturl"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/problems"
	"github.com/tliron/kutil/terminal"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
	"github.com/tliron/puccini/tosca/parser"
)

var pucciniLog = commonlog.GetLogger("turandot.puccini")

var pucciniParser = parser.NewParser()

func CompileTOSCA(context contextpkg.Context, url string, inputs map[string]ard.Value, writer io.Writer, urlContext *exturl.Context) error {
	if url_, err := urlContext.NewURL(url); err == nil {
		// TODO: bases

		parserContext := pucciniParser.NewContext()
		parserContext.URL = url_
		parserContext.Stylist = terminal.NewStylist(false)
		parserContext.Inputs = inputs

		if serviceTemplate, err := parserContext.Parse(context); err == nil {
			problems := parserContext.GetProblems()
			if problems.Empty() {
				if clout, err := serviceTemplate.Compile(); err == nil {
					return WriteClout(clout, writer)
				} else {
					return err
				}
			} else {
				return errors.New(problems.ToString(true))
			}
		} else {
			return err
		}
	} else {
		return err
	}
}

func ReadClout(reader io.Reader, urlContext *exturl.Context) (*cloutpkg.Clout, error) {
	if clout, err := cloutpkg.Read(reader, "yaml"); err == nil {
		execContext := js.ExecContext{
			Clout:      clout,
			Problems:   problems.NewProblems(nil),
			URLContext: urlContext,
			Format:     "yaml",
		}

		if execContext.Resolve(); execContext.Problems.Empty() {
			return clout, nil
		} else {
			return nil, errors.New(execContext.Problems.ToString(true))
		}
	} else {
		return nil, err
	}
}

func WriteClout(clout *cloutpkg.Clout, writer io.Writer) error {
	return (&transcribe.Transcriber{Indent: "  "}).Write(clout, writer, "yaml")
}

func RequireCloutScriptlet(clout *cloutpkg.Clout, scriptletName string, arguments map[string]string, urlContext *exturl.Context) (string, error) {
	jsContext := js.NewContext(scriptletName, pucciniLog, arguments, false, "yaml", false, false, false, "", urlContext)
	var builder strings.Builder
	jsContext.Stdout = &builder
	if _, err := jsContext.Require(clout, scriptletName, nil); err == nil {
		return builder.String(), nil
	} else {
		return "", err
	}
}
