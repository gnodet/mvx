package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Display version information for mvx including version number, 
commit hash, build date, and runtime information.`,
	Run: func(cmd *cobra.Command, args []string) {
		showVersion()
	},
}

func showVersion() {
	fmt.Printf("mvx version %s\n", version)
	
	if verbose {
		fmt.Printf("Commit:      %s\n", commit)
		fmt.Printf("Built:       %s\n", date)
		fmt.Printf("Go version:  %s\n", runtime.Version())
		fmt.Printf("OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
	}
}
