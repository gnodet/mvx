package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnodet/mvx/pkg/shell"
	"github.com/spf13/cobra"
)

// activateCmd represents the activate command
var activateCmd = &cobra.Command{
	Use:   "activate [shell]",
	Short: "Generate shell integration code for automatic environment activation",
	Long: `Generate shell integration code that automatically activates mvx when entering
directories with .mvx configuration.

This command outputs shell-specific code that should be evaluated in your shell's
configuration file (e.g., ~/.bashrc, ~/.zshrc, ~/.config/fish/config.fish).

When activated, mvx will:
  - Detect when you enter a directory with .mvx configuration
  - Automatically update PATH with mvx-managed tools
  - Set up environment variables from your configuration
  - Cache the environment to avoid repeated setup

Supported shells:
  - bash
  - zsh
  - fish
  - powershell

Examples:
  # Bash - add to ~/.bashrc
  eval "$(mvx activate bash)"

  # Zsh - add to ~/.zshrc
  eval "$(mvx activate zsh)"

  # Fish - add to ~/.config/fish/config.fish
  mvx activate fish | source

  # PowerShell - add to $PROFILE
  Invoke-Expression (mvx activate powershell | Out-String)

After adding to your shell configuration, restart your shell or source the file:
  source ~/.bashrc    # bash
  source ~/.zshrc     # zsh
  source ~/.config/fish/config.fish  # fish

Configuration:
  MVX_VERBOSE=true        # Enable verbose output to see what mvx is doing`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		shellType := args[0]

		// Validate shell type
		validShells := []string{"bash", "zsh", "fish", "powershell"}
		isValid := false
		for _, s := range validShells {
			if s == shellType {
				isValid = true
				break
			}
		}

		if !isValid {
			printError("Unsupported shell: %s", shellType)
			printError("Supported shells: bash, zsh, fish, powershell")
			os.Exit(1)
		}

		// Get the path to the mvx binary
		mvxPath, err := getMvxBinaryPath()
		if err != nil {
			printError("Failed to determine mvx binary path: %v", err)
			os.Exit(1)
		}

		// Generate shell hook
		hook, err := shell.GenerateHook(shellType, mvxPath)
		if err != nil {
			printError("Failed to generate shell hook: %v", err)
			os.Exit(1)
		}

		// Output the hook (this will be evaluated by the shell)
		fmt.Print(hook)
	},
}

// deactivateCmd represents the deactivate command
var deactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate mvx shell integration",
	Long: `Deactivate mvx shell integration for the current shell session.

This command is automatically available after running 'mvx activate' and
removes mvx-managed tools from PATH and unsets mvx environment variables.

Note: This only affects the current shell session. To permanently disable
mvx activation, remove the 'eval "$(mvx activate ...)"' line from your
shell configuration file.

Examples:
  mvx deactivate    # Remove mvx from current session`,

	Run: func(cmd *cobra.Command, args []string) {
		// This command is primarily handled by the shell hook itself
		// When called directly, we just provide information
		printInfo("To deactivate mvx in your current shell session:")
		printInfo("")
		printInfo("  Bash/Zsh:")
		printInfo("    Run: mvx_deactivate")
		printInfo("")
		printInfo("  Fish:")
		printInfo("    Run: mvx_deactivate")
		printInfo("")
		printInfo("  PowerShell:")
		printInfo("    Run: mvx-deactivate")
		printInfo("")
		printInfo("To permanently disable mvx activation, remove the activation")
		printInfo("line from your shell configuration file:")
		printInfo("  - Bash: ~/.bashrc")
		printInfo("  - Zsh: ~/.zshrc")
		printInfo("  - Fish: ~/.config/fish/config.fish")
		printInfo("  - PowerShell: $PROFILE")
	},
}

// getMvxBinaryPath returns the path to the mvx binary
func getMvxBinaryPath() (string, error) {
	// Try to get the path from the current executable
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	return exePath, nil
}
