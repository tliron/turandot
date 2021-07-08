package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/kubernetes-sigs/reference-docs/gen-apidocs/generators/api"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

var referenceUrl = flag.String("reference-url", "https://kubernetes.io/docs/reference/generated/kubernetes-api/v%s/#%s", "Reference URL format string.")
var sourceUrl = flag.String("source-url", "", "Source URL.")
var output = flag.String("output", ".", "Output path.")

func main() {
	flag.Parse()
	config := api.NewConfig()

	terminal.Printf("generating: %s\n", *output)

	dataTosca, err := os.Create(filepath.Join(*output, "data.yaml"))
	util.FailOnError(err)
	util.OnExit(func() {
		if err := dataTosca.Close(); err != nil {
			terminal.Eprintln(err)
		}
	})

	generator := Generator{
		entity:      "data",
		excludes:    excludes,
		includes:    includes,
		annotations: annotations,
		config:      config,
		writer:      dataTosca,
	}
	generator.Gather()
	generator.Generate()

	capabilitiesTosca, err := os.Create(filepath.Join(*output, "capabilities.yaml"))
	util.FailOnError(err)
	util.OnExit(func() {
		if err := capabilitiesTosca.Close(); err != nil {
			terminal.Eprintln(err)
		}
	})

	generator = Generator{
		entity:      "capability",
		excludes:    excludes,
		includes:    includes,
		annotations: annotations,
		config:      config,
		writer:      capabilitiesTosca,
	}
	generator.Gather()
	generator.Generate()
}
