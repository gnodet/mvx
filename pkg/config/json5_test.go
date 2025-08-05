package config

import (
	"strings"
	"testing"
)

func TestJSON5Preprocessor_RemoveComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "single line comments",
			input: `{
  // This is a comment
  "key": "value" // Inline comment
}`,
			expected: `{

"key": "value"
}`,
		},
		{
			name: "multi-line comments",
			input: `{
  /* Multi-line
     comment */
  "key": "value"
}`,
			expected: `{

"key": "value"
}`,
		},
		{
			name: "comments in strings should be preserved",
			input: `{
  "url": "https://example.com/path", // This comment should be removed
  "comment": "This // is not a comment"
}`,
			expected: `{
"url": "https://example.com/path",
"comment": "This // is not a comment"
}`,
		},
		{
			name: "mixed comments",
			input: `{
  // Single line comment
  /* Multi-line comment */
  "key": "value", // Inline comment
  /* Another
     multi-line */ "key2": "value2"
}`,
			expected: `{


"key": "value",
"key2": "value2"
}`,
		},
	}

	preprocessor := NewJSON5Preprocessor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := preprocessor.removeComments(tt.input)
			// Normalize whitespace for comparison
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tt.expected)
			if result != expected {
				t.Errorf("removeComments() = %q, want %q", result, expected)
			}
		})
	}
}

func TestJSON5Preprocessor_QuoteKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "unquoted keys",
			input: `{
  key: "value",
  anotherKey: "value2"
}`,
			expected: `{
  "key": "value",
  "anotherKey": "value2"
}`,
		},
		{
			name: "already quoted keys should remain unchanged",
			input: `{
  "key": "value",
  "anotherKey": "value2"
}`,
			expected: `{
  "key": "value",
  "anotherKey": "value2"
}`,
		},
		{
			name: "mixed quoted and unquoted keys",
			input: `{
  "quotedKey": "value",
  unquotedKey: "value2"
}`,
			expected: `{
  "quotedKey": "value",
  "unquotedKey": "value2"
}`,
		},
		{
			name: "reserved words should not be quoted",
			input: `{
  key: true,
  key2: false,
  key3: null
}`,
			expected: `{
  "key": true,
  "key2": false,
  "key3": null
}`,
		},
	}

	preprocessor := NewJSON5Preprocessor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := preprocessor.quoteKeys(tt.input)
			if result != tt.expected {
				t.Errorf("quoteKeys() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestJSON5Preprocessor_RemoveTrailingCommas(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "trailing comma in object",
			input: `{
  "key": "value",
  "key2": "value2",
}`,
			expected: `{
  "key": "value",
  "key2": "value2"
}`,
		},
		{
			name: "trailing comma in array",
			input: `{
  "array": ["item1", "item2",]
}`,
			expected: `{
  "array": ["item1", "item2"]
}`,
		},
		{
			name: "no trailing commas",
			input: `{
  "key": "value",
  "array": ["item1", "item2"]
}`,
			expected: `{
  "key": "value",
  "array": ["item1", "item2"]
}`,
		},
		{
			name: "nested objects with trailing commas",
			input: `{
  "nested": {
    "key": "value",
  },
  "array": [
    {"item": "value",},
  ],
}`,
			expected: `{
  "nested": {
    "key": "value"
  },
  "array": [
    {"item": "value"}
  ]
}`,
		},
	}

	preprocessor := NewJSON5Preprocessor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := preprocessor.removeTrailingCommas(tt.input)
			if result != tt.expected {
				t.Errorf("removeTrailingCommas() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestJSON5Preprocessor_Process(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "complete JSON5 example",
			input: `{
  // Configuration file
  project: {
    name: "test-project", // Project name
    version: "1.0.0",
  },
  
  /* Multi-line
     comment */
  tools: {
    java: {
      version: "21",
      distribution: "temurin", // Eclipse Temurin
    },
  },
}`,
			wantErr: false,
		},
		{
			name: "invalid JSON after preprocessing",
			input: `{
  key: "unclosed string
}`,
			wantErr: true,
		},
		{
			name: "empty object",
			input: `{
  // Just comments
}`,
			wantErr: false,
		},
	}

	preprocessor := NewJSON5Preprocessor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := preprocessor.Process([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseJSON5(t *testing.T) {
	input := `{
  // mvx configuration
  project: {
    name: "test-project", // Project name
    description: "A test project",
  },
  
  tools: {
    java: {
      version: "21",
      distribution: "temurin", // Eclipse Temurin
    },
  },
}`

	var config Config
	err := ParseJSON5([]byte(input), &config)
	if err != nil {
		t.Fatalf("ParseJSON5() error = %v", err)
	}

	// Verify the parsed configuration
	if config.Project.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", config.Project.Name)
	}

	if config.Project.Description != "A test project" {
		t.Errorf("Expected project description 'A test project', got '%s'", config.Project.Description)
	}

	javaTool, exists := config.Tools["java"]
	if !exists {
		t.Errorf("Expected java tool to be configured")
	} else {
		if javaTool.Version != "21" {
			t.Errorf("Expected Java version '21', got '%s'", javaTool.Version)
		}

		if javaTool.Distribution != "temurin" {
			t.Errorf("Expected Java distribution 'temurin', got '%s'", javaTool.Distribution)
		}
	}
}
