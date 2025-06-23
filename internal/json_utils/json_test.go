package jsonutils

import (
	"strings"
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

// Benchmark comparing IsJsonFormat vs string.Contains("payload")
func BenchmarkIsJsonFormatVsPayloadCheck(b *testing.B) {
	testCases := []struct {
		name    string
		message string
	}{
		{
			name:    "valid_json_with_payload",
			message: `{"payload": {"data": "value"}, "timestamp": 1234567890}`,
		},
		{
			name:    "valid_json_without_payload",
			message: `{"data": "value", "timestamp": 1234567890}`,
		},
		{
			name:    "valid_json_array_with_payload",
			message: `[{"payload": "data"}, {"other": "value"}]`,
		},
		{
			name:    "valid_json_array_without_payload",
			message: `[{"data": "value"}, {"other": "value"}]`,
		},
		{
			name:    "invalid_json_with_payload_text",
			message: `payload: some data here`,
		},
		{
			name:    "invalid_json_without_payload",
			message: `some random text without payload`,
		},
		{
			name:    "empty_string",
			message: "",
		},
		{
			name:    "whitespace_only",
			message: "   ",
		},
		{
			name:    "large_valid_json_with_payload",
			message: `{"payload": {"data": "value", "nested": {"deep": {"level": 5, "items": [1, 2, 3, 4, 5]}}}, "metadata": {"timestamp": 1234567890, "version": "1.0.0", "source": "polymarket"}}`,
		},
		{
			name:    "large_valid_json_without_payload",
			message: `{"data": {"value": "test", "nested": {"deep": {"level": 5, "items": [1, 2, 3, 4, 5]}}}, "metadata": {"timestamp": 1234567890, "version": "1.0.0", "source": "polymarket"}}`,
		},
	}

	b.Run("IsJsonFormat", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, tc := range testCases {
				IsJsonFormat(tc.message)
			}
		}
	})

	b.Run("StringContainsPayload", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, tc := range testCases {
				_ = strings.Contains(tc.message, "payload")
			}
		}
	})
}

// Benchmark for specific use case: checking if message should be processed
func BenchmarkMessageProcessingCheck(b *testing.B) {
	testCases := []struct {
		name    string
		message string
	}{
		{
			name:    "polymarket_json_with_payload",
			message: `{"payload": {"market_id": "123", "price": 0.65}, "timestamp": 1234567890}`,
		},
		{
			name:    "polymarket_json_without_payload",
			message: `{"market_id": "123", "price": 0.65, "timestamp": 1234567890}`,
		},
		{
			name:    "non_json_with_payload_text",
			message: `payload: market update for id 123`,
		},
		{
			name:    "non_json_without_payload",
			message: `market update for id 123`,
		},
		{
			name:    "empty_message",
			message: "",
		},
	}

	b.Run("IsJsonFormat", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, tc := range testCases {
				if IsJsonFormat(tc.message) {
					// Simulate processing
					_ = len(tc.message)
				}
			}
		}
	})

	b.Run("StringContainsPayload", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, tc := range testCases {
				if strings.Contains(tc.message, "payload") {
					// Simulate processing
					_ = len(tc.message)
				}
			}
		}
	})
}

// Benchmark for edge cases and performance characteristics
func BenchmarkEdgeCases(b *testing.B) {
	// Very long strings to test performance with large inputs
	longJsonWithPayload := `{"payload": {"data": "` + strings.Repeat("very long string ", 1000) + `"}, "timestamp": 1234567890}`
	longJsonWithoutPayload := `{"data": "` + strings.Repeat("very long string ", 1000) + `", "timestamp": 1234567890}`
	longTextWithPayload := "payload: " + strings.Repeat("very long string ", 1000)
	longTextWithoutPayload := strings.Repeat("very long string ", 1000)

	testCases := []struct {
		name    string
		message string
	}{
		{"long_json_with_payload", longJsonWithPayload},
		{"long_json_without_payload", longJsonWithoutPayload},
		{"long_text_with_payload", longTextWithPayload},
		{"long_text_without_payload", longTextWithoutPayload},
	}

	b.Run("IsJsonFormat_LongStrings", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, tc := range testCases {
				IsJsonFormat(tc.message)
			}
		}
	})

	b.Run("StringContainsPayload_LongStrings", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, tc := range testCases {
				strings.Contains(tc.message, "payload")
			}
		}
	})
}
