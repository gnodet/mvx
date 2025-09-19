package cmd

import (
	"os"
	"reflect"
	"testing"
)

func TestParseHybridArgs(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name            string
		args            []string
		expectedMaven   []string
		expectedVerbose bool
		expectedQuiet   bool
	}{
		{
			name:            "Maven flags only",
			args:            []string{"mvx", "mvn", "-V"},
			expectedMaven:   []string{"-V"},
			expectedVerbose: false,
			expectedQuiet:   false,
		},
		{
			name:            "mvx verbose + Maven flags",
			args:            []string{"mvx", "--verbose", "mvn", "-V"},
			expectedMaven:   []string{"-V"},
			expectedVerbose: true,
			expectedQuiet:   false,
		},
		{
			name:            "mvx quiet + Maven flags",
			args:            []string{"mvx", "--quiet", "mvn", "-X", "clean"},
			expectedMaven:   []string{"-X", "clean"},
			expectedVerbose: false,
			expectedQuiet:   true,
		},
		{
			name:            "mvx short flags + Maven flags",
			args:            []string{"mvx", "-v", "mvn", "-Pprofile", "install"},
			expectedMaven:   []string{"-Pprofile", "install"},
			expectedVerbose: true,
			expectedQuiet:   false,
		},
		{
			name:            "Complex Maven command",
			args:            []string{"mvx", "mvn", "-X", "-Dmaven.test.skip=true", "clean", "install"},
			expectedMaven:   []string{"-X", "-Dmaven.test.skip=true", "clean", "install"},
			expectedVerbose: false,
			expectedQuiet:   false,
		},
		{
			name:            "Both mvx flags + complex Maven",
			args:            []string{"mvx", "--verbose", "--quiet", "mvn", "-X", "-Pproduction", "package"},
			expectedMaven:   []string{"-X", "-Pproduction", "package"},
			expectedVerbose: true,
			expectedQuiet:   true,
		},
		{
			name:            "No Maven args",
			args:            []string{"mvx", "mvn"},
			expectedMaven:   []string{},
			expectedVerbose: false,
			expectedQuiet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global variables
			verbose = false
			quiet = false

			// Set os.Args for the test
			os.Args = tt.args

			// Call parseHybridArgs
			mavenArgs, err := parseHybridArgs()
			if err != nil {
				t.Fatalf("parseHybridArgs() error = %v", err)
			}

			// Check Maven arguments
			if !reflect.DeepEqual(mavenArgs, tt.expectedMaven) {
				t.Errorf("parseHybridArgs() mavenArgs = %v, want %v", mavenArgs, tt.expectedMaven)
			}

			// Check verbose flag
			if verbose != tt.expectedVerbose {
				t.Errorf("parseHybridArgs() verbose = %v, want %v", verbose, tt.expectedVerbose)
			}

			// Check quiet flag
			if quiet != tt.expectedQuiet {
				t.Errorf("parseHybridArgs() quiet = %v, want %v", quiet, tt.expectedQuiet)
			}
		})
	}
}

func TestParseHybridArgsEdgeCases(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	t.Run("No mvn command", func(t *testing.T) {
		verbose = false
		quiet = false
		os.Args = []string{"mvx", "--verbose", "build"}

		mavenArgs, err := parseHybridArgs()
		if err != nil {
			t.Fatalf("parseHybridArgs() error = %v", err)
		}

		// Should return all args when no mvn command found
		expected := []string{"--verbose", "build"}
		if !reflect.DeepEqual(mavenArgs, expected) {
			t.Errorf("parseHybridArgs() mavenArgs = %v, want %v", mavenArgs, expected)
		}
	})

	t.Run("Unknown mvx flag", func(t *testing.T) {
		verbose = false
		quiet = false
		os.Args = []string{"mvx", "--unknown-flag", "mvn", "-V"}

		mavenArgs, err := parseHybridArgs()
		if err != nil {
			t.Fatalf("parseHybridArgs() error = %v", err)
		}

		// Should still work and return Maven args
		expected := []string{"-V"}
		if !reflect.DeepEqual(mavenArgs, expected) {
			t.Errorf("parseHybridArgs() mavenArgs = %v, want %v", mavenArgs, expected)
		}
	})
}
