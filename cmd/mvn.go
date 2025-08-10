package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/tools"
	"github.com/spf13/cobra"
)

// mvnCmd runs Maven with the environment provisioned by mvx
var mvnCmd = &cobra.Command{
	Use:   "mvn [args...]",
	Short: "Run Apache Maven with mvx-managed environment",
	Long: `Run Apache Maven using the version configured in .mvx/config.

Examples:
  mvx mvn clean install
  mvx mvn -Plicense-check -N
  mvx mvn org.eclipse.tycho:tycho-versions-plugin:0.25.0:set-version ...`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// Ensure tools installed and get environment
		if err := mgr.InstallTools(cfg); err != nil {
			return fmt.Errorf("failed to install tools: %w", err)
		}
		envMap, err := mgr.SetupEnvironment(cfg)
		if err != nil {
			return err
		}
		var env []string
		for k, v := range envMap { env = append(env, fmt.Sprintf("%s=%s", k, v)) }
		// Preserve existing environment variables
		env = append(env, os.Environ()...)

		mvnTool, _ := mgr.GetTool("maven")
		bin, err := mvnTool.GetBinPath(toolCfg.Version, toolCfg)
		if err != nil { return err }
		mvnExe := filepath.Join(bin, "mvn")
		if isWindows() { mvnExe += ".cmd" }

		c := exec.Command(mvnExe, args...)
		c.Dir = projectRoot
		c.Env = env
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		return c.Run()
	},
}

func init() { rootCmd.AddCommand(mvnCmd) }

