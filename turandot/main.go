package main

import (
	"github.com/tebeka/atexit"
	"github.com/tliron/turandot/turandot/commands"
)

func main() {
	commands.Execute()
	atexit.Exit(0)
}
