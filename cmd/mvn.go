package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/tools"
	"github.com/spf13/cobra"
)

// mvnCmd runs Maven with the environment provisioned by mvx
var mvnCmd = &cobra.Command{
	Use:                "mvn [args...]",
	Short:              "Run Apache Maven with mvx-managed environment",
	DisableFlagParsing: true, // We handle parsing manually to support both mvx and Maven flags
	Long: `Run Apache Maven using the version configured in .mvx/config.

All arguments are passed directly to Maven without interpretation, allowing
Maven flags like -V, -X, -P, etc. to work naturally.

Examples:
  mvx mvn clean install
  mvx mvn -V
  mvx mvn -X clean compile
  mvx mvn -Plicense-check -N
  mvx mvn org.eclipse.tycho:tycho-versions-plugin:0.25.0:set-version ...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse mvx global flags from os.Args and extract Maven arguments
		mavenArgs, err := parseHybridArgs()
		if err != nil {
			return err
		}

		// Handle backward compatibility: remove '--' separator if present and warn
		if len(mavenArgs) > 0 && mavenArgs[0] == "--" {
			printWarning("The '--' separator is no longer needed with 'mvx mvn'. You can remove it for cleaner syntax.")
			printWarning("  Before: mvx mvn -- %s", strings.Join(mavenArgs[1:], " "))
			printWarning("  After:  mvx mvn %s", strings.Join(mavenArgs[1:], " "))
			mavenArgs = mavenArgs[1:] // Remove the '--' separator
		}

		projectRoot, err := findProjectRoot()
		if err != nil {
			return fmt.Errorf("failed to find project root: %w", err)
		}

		cfg, err := config.LoadConfig(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		mgr, err := tools.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create tool manager: %w", err)
		}

		// Ensure Maven is configured
		toolCfg, ok := cfg.Tools["maven"]
		if !ok {
			return fmt.Errorf("no Maven tool configured. Add tools.maven to .mvx/config and run 'mvx setup'")
		}

		// Only ensure Maven and Java are installed (Maven needs Java)
		requiredTools := []string{"maven"}
		if _, hasJava := cfg.Tools["java"]; hasJava {
			requiredTools = append(requiredTools, "java")
		}

		if err := mgr.InstallSpecificTools(cfg, requiredTools); err != nil {
			return fmt.Errorf("failed to install required tools: %w", err)
		}
		envMap, err := mgr.SetupEnvironment(cfg)
		if err != nil {
			return err
		}

		// Create environment map starting with existing environment
		envVars := make(map[string]string)
		for _, envVar := range os.Environ() {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}

		// Override with mvx-managed environment variables (these take precedence)
		for k, v := range envMap {
			envVars[k] = v
		}

		// Convert back to slice format
		var env []string
		for k, v := range envVars {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}

		mvnTool, _ := mgr.GetTool("maven")

		// Resolve Maven version to handle any overrides
		resolvedVersion, err := mgr.ResolveVersion("maven", toolCfg)
		if err != nil {
			return fmt.Errorf("failed to resolve Maven version: %w", err)
		}

		// Create resolved config for Maven
		resolvedToolCfg := toolCfg
		resolvedToolCfg.Version = resolvedVersion

		bin, err := mvnTool.GetBinPath(resolvedVersion, resolvedToolCfg)
		if err != nil {
			return err
		}
		mvnExe := filepath.Join(bin, "mvn")
		if isWindows() {
			mvnExe += ".cmd"
		}

		c := exec.Command(mvnExe, mavenArgs...)
		c.Dir = projectRoot
		c.Env = env
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		return c.Run()
	},
}

// parseHybridArgs parses os.Args to extract mvx global flags and Maven arguments
// This allows commands like: mvx --verbose mvn -V clean install
func parseHybridArgs() ([]string, error) {
	args := os.Args[1:] // Remove program name

	// Find the "mvn" command position
	mvnIndex := -1
	for i, arg := range args {
		if arg == "mvn" {
			mvnIndex = i
			break
		}
	}

	if mvnIndex == -1 {
		// This shouldn't happen since we're in the mvn command, but handle gracefully
		return args, nil
	}

	// Extract mvx flags (everything before "mvn")
	mvxFlags := args[:mvnIndex]

	// Extract Maven arguments (everything after "mvn")
	mavenArgs := args[mvnIndex+1:]

	// Parse mvx global flags and set global variables
	for i := 0; i < len(mvxFlags); i++ {
		flag := mvxFlags[i]
		switch flag {
		case "--verbose", "-v":
			verbose = true
		case "--quiet", "-q":
			quiet = true
		case "--help", "-h":
			// Let Cobra handle help
			continue
		default:
			// Unknown mvx flag - this could be an error or we could ignore it
			printWarning("Unknown mvx flag: %s (will be ignored)", flag)
		}
	}

	return mavenArgs, nil
}

func init() { rootCmd.AddCommand(mvnCmd) }
