package config

import (
	"reflect"
	"testing"
)

func TestParseJSON5_BasicFeatures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "simple JSON",
			input: `{"key": "value"}`,
			expected: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name: "comments and unquoted keys",
			input: `{
				// This is a comment
				name: "test-project",
				version: "1.0.0" // Another comment
			}`,
			expected: map[string]interface{}{
				"name":    "test-project",
				"version": "1.0.0",
			},
		},
		{
			name: "trailing commas",
			input: `{
				"key1": "value1",
				"key2": "value2",
			}`,
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]interface{}
			err := ParseJSON5([]byte(tt.input), &result)
			if err != nil {
				t.Fatalf("ParseJSON5() error = %v", err)
			}

			// Compare the results
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseJSON5() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseJSON5_ErrorHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid JSON5",
			input: `{ invalid json5 }`,
		},
		{
			name:  "unclosed brace",
			input: `{ "key": "value"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]interface{}
			err := ParseJSON5([]byte(tt.input), &result)
			if err == nil {
				t.Error("ParseJSON5() expected error, got nil")
			}
		})
	}
}
