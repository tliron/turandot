package version

import (
	"fmt"
	"runtime"
)

func Print() {
	// Why not use the version from runtime/debug.ReadBuildInfo? See:
	// https://github.com/golang/go/issues/29228
	if GitVersion != "" {
		fmt.Printf("version=%s\n", GitVersion)
	}
	if GitRevision != "" {
		fmt.Printf("revision=%s\n", GitRevision)
	}
	if Timestamp != "" {
		fmt.Printf("timestamp=%s\n", Timestamp)
	}
	fmt.Printf("arch=%s\n", runtime.GOARCH)
	fmt.Printf("os=%s\n", runtime.GOOS)
	fmt.Printf("compiler=%s\n", runtime.Compiler)
}
