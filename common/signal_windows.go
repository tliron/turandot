package common

import (
	"os"
)

var shutdownSignals = []os.Signal{os.Interrupt}
