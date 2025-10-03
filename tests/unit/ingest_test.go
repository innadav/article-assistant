package unit

import (
	"testing"

	"article-assistant/internal/ingest"
)

func TestStripHTMLBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple HTML",
			input:    "<p>Hello <strong>world</strong>!</p>",
			expected: "Hello world!",
		},
		{
			name:     "with script tags",
			input:    "<p>Hello</p><script>alert('test')</script><p>World</p>",
			expected: "HelloWorld",
		},
		{
			name:     "with style tags",
			input:    "<p>Hello</p><style>body{color:red}</style><p>World</p>",
			expected: "HelloWorld",
		},
		{
			name:     "nested tags",
			input:    "<div><p>Hello <span>world</span></p></div>",
			expected: "Hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ingest.StripHTMLBasic(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestExtractBetween(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    string
		end      string
		expected string
	}{
		{
			name:     "extract title",
			input:    "<html><head><title>Test Title</title></head></html>",
			start:    "<title>",
			end:      "</title>",
			expected: "Test Title",
		},
		{
			name:     "case insensitive",
			input:    "<html><head><TITLE>Test Title</TITLE></head></html>",
			start:    "<title>",
			end:      "</title>",
			expected: "Test Title",
		},
		{
			name:     "not found",
			input:    "<html><head></head></html>",
			start:    "<title>",
			end:      "</title>",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ingest.ExtractBetween(tt.input, tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
