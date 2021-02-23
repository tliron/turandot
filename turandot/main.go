package main

import (
	"github.com/tebeka/atexit"
	"github.com/tliron/turandot/turandot/commands"

	_ "github.com/tliron/kutil/logging/simple"
)

func main() {
	commands.Execute()
	atexit.Exit(0)
}
