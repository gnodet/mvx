package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/executor"
	"github.com/gnodet/mvx/pkg/tools"
	"github.com/spf13/cobra"
)

var websiteCmd = &cobra.Command{
	Use:   "website",
	Short: "Website tasks (build, dev, serve)",
}

var websiteBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the website",
	RunE:  func(cmd *cobra.Command, args []string) error { return runWebsiteCommand("build", args) },
}

var websiteServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the website",
	RunE:  func(cmd *cobra.Command, args []string) error { return runWebsiteCommand("serve", args) },
}

func init() {
	websiteCmd.AddCommand(websiteBuildCmd)
	websiteCmd.AddCommand(websiteServeCmd)
	rootCmd.AddCommand(websiteCmd)
}

// runWebsiteCommand runs a website subcommand, checking for overrides first
func runWebsiteCommand(subcommand string, args []string) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Try to load configuration to check for overrides
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		// No config found, fall back to default behavior
		return runWebsiteNpm(subcommand)
	}

	// Create tool manager and executor
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	exec := executor.NewExecutor(cfg, manager, projectRoot)

	// Check for override command (e.g., "website build", "website serve")
	overrideCommandName := fmt.Sprintf("website %s", subcommand)
	if cmdConfig, exists := cfg.Commands[overrideCommandName]; exists && cmdConfig.Override {
		printInfo("🔨 Running overridden website %s command", subcommand)
		if cmdConfig.Description != "" {
			printInfo("   %s", cmdConfig.Description)
		}
		return exec.ExecuteCommand(overrideCommandName, args)
	}

	// No override found, use default behavior
	return runWebsiteNpm(subcommand)
}

func runWebsiteNpm(subcmd string) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return err
	}

	mgr, err := tools.NewManager()
	if err != nil {
		return err
	}

	// Ensure Node is configured
	if _, ok := cfg.Tools["node"]; !ok {
		return fmt.Errorf("no Node.js tool configured. Add tools.node to .mvx/config and run 'mvx setup'")
	}

	if err := mgr.InstallTools(cfg); err != nil {
		return err
	}
	envMap, err := mgr.SetupEnvironment(cfg)
	if err != nil {
		return err
	}
	var env []string
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	env = append(env, os.Environ()...)

	workDir := filepath.Join(projectRoot, "website")
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return fmt.Errorf("website directory not found: %s", workDir)
	}

	npm := "npm"
	if isWindows() {
		npm = "npm.cmd"
	}

	// Install deps for build/dev
	if subcmd == "build" || subcmd == "start-with-snippets" || subcmd == "dev" {
		ci := exec.Command(npm, "install")
		ci.Dir = workDir
		ci.Env = env
		ci.Stdout = os.Stdout
		ci.Stderr = os.Stderr
		ci.Stdin = os.Stdin
		if err := ci.Run(); err != nil {
			return err
		}
	}

	c := exec.Command(npm, "run", subcmd)
	c.Dir = workDir
	c.Env = env
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}
