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

var websiteCmd = &cobra.Command{
	Use:   "website",
	Short: "Website tasks (build, dev, serve)",
}

var websiteBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the website (npm run build)",
	RunE:  func(cmd *cobra.Command, args []string) error { return runWebsiteNpm("build") },
}

var websiteDevCmd = &cobra.Command{
	Use:   "dev",
	Short: "Run website dev server (npm run start-with-snippets)",
	RunE:  func(cmd *cobra.Command, args []string) error { return runWebsiteNpm("start-with-snippets") },
}

var websiteServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve built website (npm run serve)",
	RunE:  func(cmd *cobra.Command, args []string) error { return runWebsiteNpm("serve") },
}

func init() {
	websiteCmd.AddCommand(websiteBuildCmd)
	websiteCmd.AddCommand(websiteDevCmd)
	websiteCmd.AddCommand(websiteServeCmd)
	rootCmd.AddCommand(websiteCmd)
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
