package main

import (
	"strings"
	"testing"
)

func TestFormatMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "simple paragraph",
			input: "This is a sentence. This is another sentence.",
			expected: `This is a sentence.
This is another sentence.
`,
		},
		{
			name: "heading",
			input: `# Heading One

This is text.`,
			expected: `# Heading One

This is text.
`,
		},
		{
			name: "unordered list",
			input: `- First item
- Second item
- Third item`,
			expected: `- First item
- Second item
- Third item
`,
		},
		{
			name: "ordered list",
			input: `1. First item
2. Second item`,
			expected: `1. First item
2. Second item
`,
		},
		{
			name:     "fenced code block",
			input:    "```go\nfunc main() {}\n```",
			expected: "```go\nfunc main() {}\n```\n\n",
		},
		{
			name:  "inline code",
			input: "This has `inline code` in it.",
			expected: `This has ` + "`inline code`" + ` in it.
`,
		},
		{
			name:  "emphasis",
			input: "This has *italic* and **bold** text.",
			expected: `This has *italic* and **bold** text.
`,
		},
		{
			name:  "link",
			input: "This has a [link](https://example.com).",
			expected: `This has a [link](https://example.com).
`,
		},
		{
			name:  "blockquote",
			input: `> This is a quote.`,
			expected: `> This is a quote.

`,
		},
		{
			name: "multiple sentences in list item",
			input: `- Item one. With multiple sentences. See?
- Item two`,
			expected: `- Item one.
  With multiple sentences.
  See?
- Item two
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := formatMarkdown([]byte(tt.input))
			if err != nil {
				t.Fatalf("formatMarkdown() error = %v", err)
			}

			got := string(output)
			if got != tt.expected {
				t.Errorf("formatMarkdown() mismatch\nGot:\n%q\nExpected:\n%q", got, tt.expected)
			}
		})
	}
}

func TestSplitIntoSentences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single sentence",
			input:    "This is a sentence.",
			expected: []string{"This is a sentence."},
		},
		{
			name:     "multiple sentences",
			input:    "First sentence. Second sentence. Third sentence.",
			expected: []string{"First sentence.", "Second sentence.", "Third sentence."},
		},
		{
			name:     "question and exclamation",
			input:    "Is this a question? Yes it is! Great!",
			expected: []string{"Is this a question?", "Yes it is!", "Great!"},
		},
		{
			name:     "no trailing punctuation",
			input:    "This has no end punctuation",
			expected: []string{"This has no end punctuation"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitIntoSentences(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("splitIntoSentences() length = %d, expected %d", len(got), len(tt.expected))
				return
			}
			for i := range got {
				if strings.TrimSpace(got[i]) != strings.TrimSpace(tt.expected[i]) {
					t.Errorf("splitIntoSentences()[%d] = %q, expected %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}
