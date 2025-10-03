package util

import (
	"fmt"
	"os"
)

// IsVerbose returns true if verbose logging is enabled
func IsVerbose() bool {
	return os.Getenv("MVX_VERBOSE") == "true"
}

// LogVerbose prints verbose log messages
func LogVerbose(format string, args ...interface{}) {
	if IsVerbose() {
		fmt.Printf("[VERBOSE] "+format+"\n", args...)
	}
}
