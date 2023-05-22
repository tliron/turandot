package main

import (
	"github.com/tliron/kutil/util"

	_ "github.com/tliron/commonlog/simple"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // load all auth plugins
)

func main() {
	err := command.Execute()
	util.FailOnError(err)
	util.Exit(0)
}
