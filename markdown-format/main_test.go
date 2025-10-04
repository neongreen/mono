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

func TestNestedElements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "list with code block inside",
			input: `- First item with text. Another sentence.
- Second item with code:
  ` + "```" + `go
  func main() {}
  ` + "```" + `
- Third item`,
			expected: `- First item with text.
  Another sentence.
- Second item with code:` + "```go\nfunc main() {}\n```\n" + `- Third item
`,
		},
		{
			name: "nested unordered lists",
			input: `- Outer item one. With a sentence.
  - Nested item one
  - Nested item two. With text.
- Outer item two`,
			expected: `- Outer item one.
  With a sentence.- Nested item one
- Nested item two.
  With text.
- Outer item two
`,
		},
		{
			name: "blockquote with multiple paragraphs",
			input: `> First paragraph in quote. Second sentence here.
>
> Second paragraph. Another sentence!`,
			expected: `> First paragraph in quote.
> Second sentence here.
> 
> Second paragraph.
> Another sentence!

`,
		},
		{
			name: "nested blockquotes",
			input: `> Outer quote. Another sentence.
> > Nested quote here. More text!`,
			expected: `> Outer quote.
> Another sentence.
> 
> > Nested quote here.
> > More text!

`,
		},
		{
			name: "list inside blockquote",
			input: `> Here's a list inside a quote:
> - Item one. Sentence two.
> - Item two`,
			expected: `> Here's a list inside a quote:
> 
> - Item one.
>   Sentence two.
> - Item two

`,
		},
		{
			name: "code block in blockquote",
			input: "> Here's code:\n>\n> ```python\n> def test():\n>     pass\n> ```",
			expected: `> Here's code:
> 
> ` + "```python\n> def test():\n>     pass\n> ```" + `

`,
		},
		{
			name: "mixed inline elements in list",
			input: `- Item with **bold** and *italic*. Another sentence with [link](url).
- Item with ` + "`code`" + ` and more. Final sentence!`,
			expected: `- Item with **bold** and *italic*.
  Another sentence with [link](url).
- Item with ` + "`code`" + ` and more.
  Final sentence!
`,
		},
		{
			name: "heading followed by list",
			input: `## Section Title

- First item. Second sentence.
- Second item`,
			expected: `## Section Title

- First item.
  Second sentence.
- Second item
`,
		},
		{
			name: "multiple blocks in sequence",
			input: `# Title

Paragraph one. Sentence two.

## Subtitle

> Quote here. Another sentence!

- List item one
- List item two. With text.

` + "```" + `
code block
` + "```" + `

Final paragraph. Last sentence.`,
			expected: `# Title

Paragraph one.
Sentence two.

## Subtitle

> Quote here.
> Another sentence!

- List item one
- List item two.
  With text.

` + "```\ncode block\n```" + `

Final paragraph.
Last sentence.
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

func TestLargeOrderedList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "ordered list with more than 10 items",
			input: `1. First item. With text.
2. Second item
3. Third item. Multiple sentences here. See?
4. Fourth item
5. Fifth item. Another sentence.
6. Sixth item
7. Seventh item
8. Eighth item. More text here!
9. Ninth item
10. Tenth item. Double digits now.
11. Eleventh item. Still going!
12. Twelfth item
13. Thirteenth item. Lucky number.
14. Fourteenth item
15. Fifteenth item. Final one here!`,
			expected: `1. First item.
  With text.
2. Second item
3. Third item.
  Multiple sentences here.
  See?
4. Fourth item
5. Fifth item.
  Another sentence.
6. Sixth item
7. Seventh item
8. Eighth item.
  More text here!
9. Ninth item
10. Tenth item.
  Double digits now.
11. Eleventh item.
  Still going!
12. Twelfth item
13. Thirteenth item.
  Lucky number.
14. Fourteenth item
15. Fifteenth item.
  Final one here!
`,
		},
		{
			name: "ordered list starting from different number",
			input: `5. Item five. Starting here.
6. Item six
7. Item seven. More text.
8. Item eight
9. Item nine
10. Item ten. Double digits!
11. Item eleven
12. Item twelve. Last one.`,
			expected: `5. Item five.
  Starting here.
6. Item six
7. Item seven.
  More text.
8. Item eight
9. Item nine
10. Item ten.
  Double digits!
11. Item eleven
12. Item twelve.
  Last one.
`,
		},
		{
			name: "unordered list with many items",
			input: `- Item 1. Sentence here.
- Item 2
- Item 3. More text!
- Item 4
- Item 5. Another one.
- Item 6
- Item 7
- Item 8. Almost done.
- Item 9
- Item 10
- Item 11. More than ten!
- Item 12
- Item 13. Unlucky?
- Item 14
- Item 15. Final item here!`,
			expected: `- Item 1.
  Sentence here.
- Item 2
- Item 3.
  More text!
- Item 4
- Item 5.
  Another one.
- Item 6
- Item 7
- Item 8.
  Almost done.
- Item 9
- Item 10
- Item 11.
  More than ten!
- Item 12
- Item 13.
  Unlucky?
- Item 14
- Item 15.
  Final item here!
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

func TestComplexNestedStructures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "deeply nested blockquotes",
			input: `> Level 1 quote. First sentence.
> > Level 2 quote. Nested here.
> > > Level 3 quote. Deep nesting! Amazing stuff.`,
			expected: `> Level 1 quote.
> First sentence.
> 
> > Level 2 quote.
> > Nested here.
> > 
> > > Level 3 quote.
> > > Deep nesting!
> > > Amazing stuff.

`,
		},
		{
			name: "ordered list with nested unordered list",
			input: `1. First ordered item. With text.
   - Nested unordered one
   - Nested unordered two. More here.
2. Second ordered item. Another sentence!
3. Third ordered item`,
			expected: `1. First ordered item.
  With text.- Nested unordered one
- Nested unordered two.
  More here.
2. Second ordered item.
  Another sentence!
3. Third ordered item
`,
		},
		{
			name: "complex document structure",
			input: `# Main Title

Introduction paragraph. Second sentence here. Third one too!

## Section 1

First section paragraph. Another sentence.

> Important quote here. Multiple sentences work! See this?

### Subsection 1.1

- Point one. Details here.
- Point two. More details. Even more!
  - Nested point. With text.
  - Another nested point
- Point three

## Section 2

1. Step one. Do this first.
2. Step two. Then this. Important!
3. Step three
4. Step four. Almost done.
5. Step five. Final step!

` + "```python" + `
def example():
    """Example code block."""
    return True
` + "```" + `

Conclusion paragraph. Final thoughts here. That's all!`,
			expected: `# Main Title

Introduction paragraph.
Second sentence here.
Third one too!

## Section 1

First section paragraph.
Another sentence.

> Important quote here.
> Multiple sentences work!
> See this?

### Subsection 1.1

- Point one.
  Details here.
- Point two.
  More details.
  Even more!- Nested point.
  With text.
  - Another nested point
- Point three

## Section 2

1. Step one.
  Do this first.
2. Step two.
  Then this.
  Important!
3. Step three
4. Step four.
  Almost done.
5. Step five.
  Final step!

` + "```python\ndef example():\n    \"\"\"Example code block.\"\"\"\n    return True\n```" + `

Conclusion paragraph.
Final thoughts here.
That's all!
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
