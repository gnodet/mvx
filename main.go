package main

import (
	"fmt"
	"os"

	"github.com/gnodet/mvx/cmd"
)

var (
	// Version information - will be set during build
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	// Set version info for commands to use
	cmd.SetVersionInfo(Version, Commit, Date)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
