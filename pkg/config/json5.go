package config

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// JSON5Preprocessor handles JSON5 syntax and converts it to valid JSON
type JSON5Preprocessor struct{}

// NewJSON5Preprocessor creates a new JSON5 preprocessor
func NewJSON5Preprocessor() *JSON5Preprocessor {
	return &JSON5Preprocessor{}
}

// Process converts JSON5 syntax to valid JSON
func (p *JSON5Preprocessor) Process(input []byte) ([]byte, error) {
	content := string(input)

	// Step 1: Remove comments
	content = p.removeComments(content)

	// Step 2: Quote unquoted keys
	content = p.quoteKeys(content)

	// Step 3: Handle trailing commas
	content = p.removeTrailingCommas(content)

	// Step 4: Validate the result is valid JSON
	var test interface{}
	if err := json.Unmarshal([]byte(content), &test); err != nil {
		return nil, fmt.Errorf("preprocessed JSON5 is not valid JSON: %w", err)
	}

	return []byte(content), nil
}

// removeComments removes both single-line (//) and multi-line (/* */) comments
func (p *JSON5Preprocessor) removeComments(content string) string {
	var result strings.Builder
	lines := strings.Split(content, "\n")

	inMultiLineComment := false

	for _, line := range lines {
		processedLine := p.processLineComments(line, &inMultiLineComment)
		if processedLine != "" || !inMultiLineComment {
			result.WriteString(processedLine)
			result.WriteString("\n")
		}
	}

	return result.String()
}

// processLineComments handles comment removal for a single line
func (p *JSON5Preprocessor) processLineComments(line string, inMultiLineComment *bool) string {
	var result strings.Builder
	inString := false
	escaped := false
	i := 0

	for i < len(line) {
		char := line[i]

		// Handle escape sequences in strings
		if escaped {
			result.WriteByte(char)
			escaped = false
			i++
			continue
		}

		// Handle string boundaries
		if char == '"' && !*inMultiLineComment {
			inString = !inString
			result.WriteByte(char)
			i++
			continue
		}

		// Handle escape character
		if char == '\\' && inString {
			escaped = true
			result.WriteByte(char)
			i++
			continue
		}

		// Skip processing if we're inside a string
		if inString {
			result.WriteByte(char)
			i++
			continue
		}

		// Handle multi-line comment start
		if !*inMultiLineComment && i < len(line)-1 && line[i:i+2] == "/*" {
			*inMultiLineComment = true
			i += 2
			continue
		}

		// Handle multi-line comment end
		if *inMultiLineComment && i < len(line)-1 && line[i:i+2] == "*/" {
			*inMultiLineComment = false
			i += 2
			continue
		}

		// Skip characters inside multi-line comments
		if *inMultiLineComment {
			i++
			continue
		}

		// Handle single-line comment
		if i < len(line)-1 && line[i:i+2] == "//" {
			// Rest of line is a comment
			break
		}

		// Regular character
		result.WriteByte(char)
		i++
	}

	return strings.TrimSpace(result.String())
}

// quoteKeys adds quotes around unquoted object keys
func (p *JSON5Preprocessor) quoteKeys(content string) string {
	// Regex to match unquoted keys: word characters followed by colon
	// This is a simplified approach - a full parser would be more robust
	keyRegex := regexp.MustCompile(`(\s*)([a-zA-Z_$][a-zA-Z0-9_$]*)\s*:`)

	return keyRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the key and surrounding whitespace
		parts := keyRegex.FindStringSubmatch(match)
		if len(parts) >= 3 {
			whitespace := parts[1]
			key := parts[2]

			// Don't quote if it's already quoted or if it's a reserved word in a string context
			if p.isReservedWord(key) {
				return match // Keep as-is for true, false, null
			}

			return fmt.Sprintf(`%s"%s":`, whitespace, key)
		}
		return match
	})
}

// isReservedWord checks if a word should not be quoted (true, false, null)
func (p *JSON5Preprocessor) isReservedWord(word string) bool {
	reserved := map[string]bool{
		"true":  true,
		"false": true,
		"null":  true,
	}
	return reserved[word]
}

// removeTrailingCommas removes trailing commas before closing brackets/braces
func (p *JSON5Preprocessor) removeTrailingCommas(content string) string {
	// Remove trailing commas before closing braces or brackets
	// This handles cases like: { "key": "value", } or [ "item", ]

	// Remove comma before closing brace
	braceRegex := regexp.MustCompile(`,(\s*})`)
	content = braceRegex.ReplaceAllString(content, `$1`)

	// Remove comma before closing bracket
	bracketRegex := regexp.MustCompile(`,(\s*])`)
	content = bracketRegex.ReplaceAllString(content, `$1`)

	return content
}

// ParseJSON5 parses JSON5 content and unmarshals it into the target
func ParseJSON5(data []byte, target interface{}) error {
	preprocessor := NewJSON5Preprocessor()

	// Preprocess JSON5 to valid JSON
	processedData, err := preprocessor.Process(data)
	if err != nil {
		return fmt.Errorf("failed to preprocess JSON5: %w", err)
	}

	// Parse as regular JSON
	if err := json.Unmarshal(processedData, target); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	return nil
}

// FormatAsJSON5 formats a configuration struct as JSON5 with proper formatting
func FormatAsJSON5(cfg *Config) (string, error) {
	// First marshal to JSON to get the structure
	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	// Add header comment
	result := "{\n"
	result += "  // mvx project configuration\n"
	result += "  // See https://github.com/gnodet/mvx for documentation\n\n"

	// Parse the JSON structure and format as JSON5
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Format each section
	sections := []string{"project", "tools", "environment", "commands"}
	first := true

	for _, section := range sections {
		if value, exists := data[section]; exists && value != nil {
			if !first {
				result += ",\n\n"
			}
			first = false

			switch section {
			case "project":
				result += "  // Project metadata\n"
				result += formatProjectSection(value)
			case "tools":
				result += "  // Tool versions and configurations\n"
				result += formatToolsSection(value)
			case "environment":
				result += "  // Environment variables\n"
				result += formatEnvironmentSection(value)
			case "commands":
				result += "  // Custom commands\n"
				result += formatCommandsSection(value)
			}
		}
	}

	result += "\n}\n"
	return result, nil
}

// formatProjectSection formats the project section
func formatProjectSection(value interface{}) string {
	project, ok := value.(map[string]interface{})
	if !ok {
		return ""
	}

	result := "  project: {\n"
	if name, exists := project["name"]; exists && name != nil {
		result += fmt.Sprintf("    name: %q,\n", name)
	}
	if desc, exists := project["description"]; exists && desc != nil {
		result += fmt.Sprintf("    description: %q\n", desc)
	}
	result += "  }"
	return result
}

// formatToolsSection formats the tools section
func formatToolsSection(value interface{}) string {
	tools, ok := value.(map[string]interface{})
	if !ok {
		return ""
	}

	result := "  tools: {\n"
	first := true

	for toolName, toolData := range tools {
		if !first {
			result += ",\n"
		}
		first = false

		if toolConfig, ok := toolData.(map[string]interface{}); ok {
			result += fmt.Sprintf("    %s: {\n", toolName)

			if version, exists := toolConfig["version"]; exists && version != nil {
				result += fmt.Sprintf("      version: %q", version)
			}

			if dist, exists := toolConfig["distribution"]; exists && dist != nil && dist != "" {
				result += fmt.Sprintf(",\n      distribution: %q", dist)
			}

			result += "\n    }"
		}
	}

	result += "\n  }"
	return result
}

// formatEnvironmentSection formats the environment section
func formatEnvironmentSection(value interface{}) string {
	env, ok := value.(map[string]interface{})
	if !ok {
		return ""
	}

	result := "  environment: {\n"
	first := true

	for key, val := range env {
		if !first {
			result += ",\n"
		}
		first = false
		result += fmt.Sprintf("    %s: %q", key, val)
	}

	result += "\n  }"
	return result
}

// formatCommandsSection formats the commands section
func formatCommandsSection(value interface{}) string {
	commands, ok := value.(map[string]interface{})
	if !ok {
		return ""
	}

	result := "  commands: {\n"
	first := true

	for cmdName, cmdData := range commands {
		if !first {
			result += ",\n\n"
		}
		first = false

		if cmdConfig, ok := cmdData.(map[string]interface{}); ok {
			// Quote command names that contain special characters
			quotedName := cmdName
			if strings.ContainsAny(cmdName, "-. ") {
				quotedName = fmt.Sprintf("%q", cmdName)
			}
			result += fmt.Sprintf("    %s: {\n", quotedName)

			if desc, exists := cmdConfig["description"]; exists && desc != nil && desc != "" {
				result += fmt.Sprintf("      description: %q,\n", desc)
			}

			if script, exists := cmdConfig["script"]; exists && script != nil && script != "" {
				result += fmt.Sprintf("      script: %q", script)
			}

			// Add other fields if they exist
			if wd, exists := cmdConfig["working_dir"]; exists && wd != nil && wd != "" {
				result += fmt.Sprintf(",\n      working_dir: %q", wd)
			}

			if override, exists := cmdConfig["override"]; exists && override != nil {
				if overrideBool, ok := override.(bool); ok && overrideBool {
					result += ",\n      override: true"
				}
			}

			result += "\n    }"
		}
	}

	result += "\n  }"
	return result
}
