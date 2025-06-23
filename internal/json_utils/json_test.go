package jsonutils

import (
	"testing"
)

func TestIsJsonFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid JSON objects
		{
			name:     "valid empty object",
			input:    "{}",
			expected: true,
		},
		{
			name:     "valid object with content",
			input:    `{"key": "value"}`,
			expected: true,
		},
		{
			name:     "valid object with whitespace",
			input:    "  {  }  ",
			expected: true,
		},
		{
			name:     "valid object with nested content",
			input:    `{"key": {"nested": "value"}}`,
			expected: true,
		},

		// Valid JSON arrays
		{
			name:     "valid empty array",
			input:    "[]",
			expected: true,
		},
		{
			name:     "valid array with content",
			input:    `["item1", "item2"]`,
			expected: true,
		},
		{
			name:     "valid array with whitespace",
			input:    "  [  ]  ",
			expected: true,
		},
		{
			name:     "valid array with nested content",
			input:    `[{"key": "value"}, [1, 2, 3]]`,
			expected: true,
		},

		// Invalid cases
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: false,
		},
		{
			name:     "single character",
			input:    "a",
			expected: false,
		},
		{
			name:     "mismatched brackets - object start array end",
			input:    "{]",
			expected: false,
		},
		{
			name:     "mismatched brackets - array start object end",
			input:    "[}",
			expected: false,
		},
		{
			name:     "object missing closing brace",
			input:    "{",
			expected: false,
		},
		{
			name:     "object missing opening brace",
			input:    "}",
			expected: false,
		},
		{
			name:     "array missing closing bracket",
			input:    "[",
			expected: false,
		},
		{
			name:     "array missing opening bracket",
			input:    "]",
			expected: false,
		},
		{
			name:     "plain text",
			input:    "hello world",
			expected: false,
		},
		{
			name:     "number",
			input:    "123",
			expected: false,
		},
		{
			name:     "string literal",
			input:    `"hello"`,
			expected: false,
		},
		{
			name:     "boolean",
			input:    "true",
			expected: false,
		},
		{
			name:     "null",
			input:    "null",
			expected: false,
		},
		{
			name:     "object with content but no quotes",
			input:    "{key: value}",
			expected: true, // This is technically valid format (starts with {, ends with })
		},
		{
			name:     "array with content but no quotes",
			input:    "[item1, item2]",
			expected: true, // This is technically valid format (starts with [, ends with ])
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsJsonFormat(tt.input)
			if result != tt.expected {
				t.Errorf("IsJsonFormat(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Benchmark test for performance
func BenchmarkIsJsonFormat(b *testing.B) {
	testCases := []string{
		"{}",
		`{"key": "value"}`,
		"[]",
		`["item1", "item2"]`,
		"",
		"hello world",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range testCases {
			IsJsonFormat(input)
		}
	}
}
