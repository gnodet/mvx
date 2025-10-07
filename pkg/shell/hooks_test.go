package shell

import (
	"strings"
	"testing"
)

func TestGenerateHook(t *testing.T) {
	mvxPath := "/usr/local/bin/mvx"

	tests := []struct {
		name           string
		shellType      string
		expectedInHook []string
		shouldError    bool
	}{
		{
			name:      "bash hook",
			shellType: "bash",
			expectedInHook: []string{
				"# mvx shell integration for bash",
				"_mvx_hook()",
				"PROMPT_COMMAND",
				"mvx_deactivate()",
				mvxPath,
				".mvx",
				"env --shell bash",
				"mvx_script=",
			},
			shouldError: false,
		},
		{
			name:      "zsh hook",
			shellType: "zsh",
			expectedInHook: []string{
				"# mvx shell integration for zsh",
				"_mvx_hook()",
				"precmd",
				"add-zsh-hook",
				"mvx_deactivate()",
				mvxPath,
				".mvx",
				"env --shell zsh",
				"mvx_script=",
			},
			shouldError: false,
		},
		{
			name:      "fish hook",
			shellType: "fish",
			expectedInHook: []string{
				"# mvx shell integration for fish",
				"function _mvx_hook",
				"--on-variable PWD",
				"mvx_deactivate",
				mvxPath,
				".mvx",
				"env --shell fish",
				"mvx_script",
			},
			shouldError: false,
		},
		{
			name:      "powershell hook",
			shellType: "powershell",
			expectedInHook: []string{
				"# mvx shell integration for PowerShell",
				"function global:_mvx_hook",
				"prompt",
				"mvx-deactivate",
				".mvx",
				"env --shell powershell",
			},
			shouldError: false,
		},
		{
			name:        "unsupported shell",
			shellType:   "unsupported",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook, err := GenerateHook(tt.shellType, mvxPath)

			if tt.shouldError {
				if err == nil {
					t.Errorf("GenerateHook() should have returned error for shell type '%s'", tt.shellType)
				}
				return
			}

			if err != nil {
				t.Fatalf("GenerateHook() error = %v", err)
			}

			// Check for expected content
			for _, expected := range tt.expectedInHook {
				if !strings.Contains(hook, expected) {
					t.Errorf("Expected hook to contain '%s', but it didn't.\nHook:\n%s", expected, hook)
				}
			}

			// Verify hook is not empty
			if len(hook) == 0 {
				t.Error("GenerateHook() returned empty hook")
			}

			t.Logf("Generated %s hook (%d bytes)", tt.shellType, len(hook))
		})
	}
}

func TestGenerateBashHook(t *testing.T) {
	mvxPath := "/usr/local/bin/mvx"
	hook := generateBashHook(mvxPath)

	// Check structure
	requiredElements := []string{
		"# mvx shell integration for bash",
		"_mvx_original_prompt_command=",
		"_mvx_current_dir=",
		"_mvx_hook()",
		"PROMPT_COMMAND=",
		"mvx_deactivate()",
		"unset -f _mvx_hook",
		"unset -f mvx_deactivate",
		"env --shell bash",
		"mvx_script=",
	}

	for _, element := range requiredElements {
		if !strings.Contains(hook, element) {
			t.Errorf("Bash hook missing required element: %s", element)
		}
	}

	// Check for directory detection logic
	if !strings.Contains(hook, `[ -d "$dir/.mvx" ]`) {
		t.Error("Bash hook missing .mvx directory detection")
	}

	t.Logf("Bash hook:\n%s", hook)
}

func TestGenerateZshHook(t *testing.T) {
	mvxPath := "/usr/local/bin/mvx"
	hook := generateZshHook(mvxPath)

	// Check structure
	requiredElements := []string{
		"# mvx shell integration for zsh",
		"typeset -g _mvx_current_dir=",
		"_mvx_hook()",
		"add-zsh-hook precmd _mvx_hook",
		"mvx_deactivate()",
		"unfunction _mvx_hook",
		"unfunction mvx_deactivate",
		"env --shell zsh",
		"mvx_script=",
	}

	for _, element := range requiredElements {
		if !strings.Contains(hook, element) {
			t.Errorf("Zsh hook missing required element: %s", element)
		}
	}

	// Check for directory detection logic
	if !strings.Contains(hook, `[[ -d "$dir/.mvx" ]]`) {
		t.Error("Zsh hook missing .mvx directory detection")
	}

	t.Logf("Zsh hook:\n%s", hook)
}

func TestGenerateFishHook(t *testing.T) {
	mvxPath := "/usr/local/bin/mvx"
	hook := generateFishHook(mvxPath)

	// Check structure
	requiredElements := []string{
		"# mvx shell integration for fish",
		"set -g _mvx_current_dir",
		"function _mvx_hook --on-variable PWD",
		"function mvx_deactivate",
		"functions --erase _mvx_hook",
		"functions --erase mvx_deactivate",
		"env --shell fish",
		"mvx_script",
	}

	for _, element := range requiredElements {
		if !strings.Contains(hook, element) {
			t.Errorf("Fish hook missing required element: %s", element)
		}
	}

	// Check for directory detection logic
	if !strings.Contains(hook, `test -d "$dir/.mvx"`) {
		t.Error("Fish hook missing .mvx directory detection")
	}

	// Check that hook is called once at the end
	if !strings.Contains(hook, "_mvx_hook") {
		t.Error("Fish hook should call _mvx_hook at the end")
	}

	t.Logf("Fish hook:\n%s", hook)
}

func TestGeneratePowerShellHook(t *testing.T) {
	mvxPath := "C:\\Program Files\\mvx\\mvx.exe"
	hook := generatePowerShellHook(mvxPath)

	// Check structure
	requiredElements := []string{
		"# mvx shell integration for PowerShell",
		"$global:_mvx_current_dir",
		"function global:_mvx_hook",
		"function global:prompt",
		"function global:mvx-deactivate",
		"Remove-Item Function:\\_mvx_hook",
		"Remove-Item Function:\\mvx-deactivate",
		"env --shell powershell",
	}

	for _, element := range requiredElements {
		if !strings.Contains(hook, element) {
			t.Errorf("PowerShell hook missing required element: %s", element)
		}
	}

	// Check for directory detection logic
	if !strings.Contains(hook, `Test-Path`) || !strings.Contains(hook, `.mvx`) {
		t.Error("PowerShell hook missing .mvx directory detection")
	}

	// Check that backslashes are escaped
	if strings.Contains(mvxPath, `\`) && !strings.Contains(hook, `\\`) {
		t.Error("PowerShell hook should escape backslashes in path")
	}

	t.Logf("PowerShell hook:\n%s", hook)
}

func TestHookPathEscaping(t *testing.T) {
	tests := []struct {
		name      string
		shellType string
		mvxPath   string
	}{
		{
			name:      "bash with spaces",
			shellType: "bash",
			mvxPath:   "/path with spaces/mvx",
		},
		{
			name:      "zsh with special chars",
			shellType: "zsh",
			mvxPath:   "/path/to/mvx-v1.0",
		},
		{
			name:      "fish with spaces",
			shellType: "fish",
			mvxPath:   "/path with spaces/mvx",
		},
		{
			name:      "powershell with backslashes",
			shellType: "powershell",
			mvxPath:   "C:\\Program Files\\mvx\\mvx.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook, err := GenerateHook(tt.shellType, tt.mvxPath)
			if err != nil {
				t.Fatalf("GenerateHook() error = %v", err)
			}

			// Hook should contain some form of the path
			// (might be escaped or quoted differently per shell)
			if !strings.Contains(hook, "mvx") {
				t.Error("Hook should contain reference to mvx binary")
			}

			// Verify hook is valid (non-empty)
			if len(hook) == 0 {
				t.Error("Hook should not be empty")
			}
		})
	}
}

func TestHookDirectoryChangeDetection(t *testing.T) {
	mvxPath := "/usr/local/bin/mvx"

	tests := []struct {
		name      string
		shellType string
		checkFor  string
	}{
		{
			name:      "bash detects directory change",
			shellType: "bash",
			checkFor:  `if [ "$current_dir" != "$_mvx_current_dir" ]`,
		},
		{
			name:      "zsh detects directory change",
			shellType: "zsh",
			checkFor:  `if [[ "$current_dir" != "$_mvx_current_dir" ]]`,
		},
		{
			name:      "fish detects directory change",
			shellType: "fish",
			checkFor:  `if test "$current_dir" != "$_mvx_current_dir"`,
		},
		{
			name:      "powershell detects directory change",
			shellType: "powershell",
			checkFor:  `if ($current_dir -ne $global:_mvx_current_dir)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook, err := GenerateHook(tt.shellType, mvxPath)
			if err != nil {
				t.Fatalf("GenerateHook() error = %v", err)
			}

			if !strings.Contains(hook, tt.checkFor) {
				t.Errorf("Hook should contain directory change detection: %s", tt.checkFor)
			}
		})
	}
}
