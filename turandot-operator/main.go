package main

import (
	"github.com/tebeka/atexit"
	"github.com/tliron/kutil/util"

	_ "github.com/tliron/kutil/logging/simple"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // load all auth plugins
)

func main() {
	err := command.Execute()
	util.FailOnError(err)
	atexit.Exit(0)
}
