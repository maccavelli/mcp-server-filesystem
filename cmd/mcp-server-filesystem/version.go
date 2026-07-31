package main

import (
	"fmt"
	"os"
	"strings"
)

// RawVersion is the build-time version string (may include a leading "v").
var RawVersion = "v4.3.4"

// Version is RawVersion with any leading "v" stripped.
var Version = strings.TrimPrefix(RawVersion, "v")

func printVersion() {
	fmt.Fprintf(os.Stderr, "mcp-server-filesystem version %s\n", Version)
}
