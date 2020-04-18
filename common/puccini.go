package common

import (
	"fmt"
	"strings"

	"github.com/op/go-logging"
	"github.com/tliron/puccini/ard"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
	"github.com/tliron/puccini/common/format"
	problemspkg "github.com/tliron/puccini/common/problems"
	"github.com/tliron/puccini/common/terminal"
	"github.com/tliron/puccini/tosca/compiler"
	"github.com/tliron/puccini/tosca/parser"
	urlpkg "github.com/tliron/puccini/url"
)

var pucciniLog = logging.MustGetLogger("turandot.puccini")

func CompileTOSCA(url string, cloutPath string, inputs map[string]ard.Value) error {
	if url_, err := urlpkg.NewURL(url); err == nil {
		defer url_.Release()
		if serviceTemplate, problems, err := parser.Parse(url_, nil, inputs); err == nil {
			if problems.Empty() {
				if clout, err := compiler.Compile(serviceTemplate, true); err == nil {
					return UpdateClout(clout, cloutPath)
				} else {
					return err
				}
			} else {
				return fmt.Errorf("%s", problems)
			}
		} else if (problems != nil) && !problems.Empty() {
			return fmt.Errorf("%s\n%s", err.Error(), problems)
		} else {
			return err
		}
	} else {
		return err
	}
}

func ReadClout(cloutPath string) (*cloutpkg.Clout, error) {
	if url_, err := urlpkg.NewURL(cloutPath); err == nil {
		defer url_.Release()
		if reader, err := url_.Open(); err == nil {
			defer reader.Close()
			if clout, err := cloutpkg.Read(reader, url_.Format()); err == nil {
				var problems problemspkg.Problems
				if compiler.Resolve(clout, &problems, "yaml", false, true, false); problems.Empty() {
					return clout, nil
				} else {
					return nil, fmt.Errorf("%s", problems)
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func UpdateClout(clout *cloutpkg.Clout, cloutPath string) error {
	if file, err := format.OpenFileForWrite(cloutPath); err == nil {
		defer file.Close()
		return format.Write(clout, "yaml", terminal.Indent, false, file)
	} else {
		return err
	}
}

func ExecScriptlet(clout *cloutpkg.Clout, scriptletName string) (string, error) {
	jsContext := js.NewContext(scriptletName, pucciniLog, false, "yaml", false, true, false, "")
	var builder strings.Builder
	jsContext.Stdout = &builder
	if err := jsContext.Exec(clout, scriptletName, nil); err == nil {
		return builder.String(), nil
	} else {
		return "", nil
	}
}
