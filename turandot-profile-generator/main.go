package main

import (
	"github.com/tliron/kutil/util"
)

func main() {
	err := command.Execute()
	util.FailOnError(err)
}
