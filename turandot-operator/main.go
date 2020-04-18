package main

import (
	"github.com/tebeka/atexit"
)

func main() {
	command.Execute()
	atexit.Exit(0)
}
