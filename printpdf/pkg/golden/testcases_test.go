package golden

import (
	"testing"

	"github.com/neongreen/mono/printpdf/pkg/converter"
	"github.com/neongreen/mono/printpdf/pkg/fetcher"
)

// TestGoldenSuite runs the complete golden test suite with comprehensive test cases
func TestGoldenSuite(t *testing.T) {
	config := DefaultTestConfig()
	suite := NewGoldenTestSuite(config)

	// Add comprehensive test cases covering all major features
	suite.AddTestCase(basicMarkdownTestCase())
	suite.AddTestCase(complexMarkdownTestCase())
	suite.AddTestCase(htmlTestCase())
	suite.AddTestCase(marginTestCase())
	suite.AddTestCase(zoomTestCase())
	suite.AddTestCase(twoColumnsTestCase())
	suite.AddTestCase(threeColumnsTestCase())
	suite.AddTestCase(landscapeTestCase())
	suite.AddTestCase(firstPageGuideTestCase())
	suite.AddTestCase(codeBlockTestCase())
	suite.AddTestCase(ligatureTestCase())
	suite.AddTestCase(tableTestCase())
	suite.AddTestCase(footnoteTestCase())
	suite.AddTestCase(footnoteMultiPageTestCase())
	suite.AddTestCase(zoom80TestCase())
	suite.AddTestCase(zoom120TestCase())
	suite.AddTestCase(zoom200TestCase())

	suite.Run(t)
}

// basicMarkdownTestCase tests basic Markdown features
func basicMarkdownTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "basic-markdown",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Basic Markdown Test

This is a simple test document with basic Markdown features.

## Section 1

Regular paragraph with **bold text** and *italic text*.

## Section 2

- List item 1
- List item 2
- List item 3

1. Numbered item 1
2. Numbered item 2
3. Numbered item 3

## Code

Inline code: ` + "`fmt.Println(\"Hello\")`" + `

Block code:
` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

## Links

[Link to example](https://example.com)

## Conclusion

This covers basic Markdown features.
`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        100,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// complexMarkdownTestCase tests advanced Markdown features
func complexMarkdownTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "complex-markdown",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Complex Markdown Document

This document tests more advanced Markdown features and edge cases.

## Nested Lists

- Top level item 1
  - Nested item 1.1
  - Nested item 1.2
    - Double nested 1.2.1
    - Double nested 1.2.2
- Top level item 2
  1. Mixed nested numbered 2.1
  2. Mixed nested numbered 2.2

## Blockquotes

> This is a simple blockquote.

> This is a blockquote with multiple paragraphs.
>
> This is the second paragraph in the blockquote.

> ## Blockquote with header
>
> - List in blockquote
> - Another item
>
> ` + "`code in blockquote`" + `

## Tables

| Feature | Typst | Prince | WeasyPrint |
|---------|--------|--------|------------|
| Auto-download | ✓ | ✗ | ✗ |
| Open source | ✓ | ✗ | ✓ |
| Typography | Excellent | Excellent | Good |
| Speed | Fast | Medium | Slow |

## Code Blocks with Languages

Python:
` + "```python" + `
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

print(fibonacci(10))
` + "```" + `

JavaScript:
` + "```javascript" + `
const fibonacci = (n) => {
    if (n <= 1) return n;
    return fibonacci(n - 1) + fibonacci(n - 2);
};

console.log(fibonacci(10));
` + "```" + `

## Images and Links

![Alt text](https://via.placeholder.com/300x200?text=Sample+Image)

[External link](https://github.com/typst/typst) and [another link](https://www.princexml.com/).

## Special Characters

Testing special characters: α β γ δ ε → ← ↑ ↓ © ® ™ § ¶

Math-like: x² + y² = z², f(x) = ∫ x dx, Σ(i=1 to n)

## Horizontal Rules

---

Above and below this line should be horizontal rules.

***

## Mixed Content

Here's a paragraph with **bold**, *italic*, ` + "`inline code`" + `, [a link](https://example.com), and a footnote[^1].

[^1]: This is a footnote (may not be supported by all converters).

### Final Section

This concludes the complex Markdown test.
`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        100,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// htmlTestCase tests HTML input processing
func htmlTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "html-input",
		ContentType: fetcher.ContentTypeHTML,
		Input: `<!DOCTYPE html>
<html>
<head>
    <title>HTML Test Document</title>
    <style>
        body { font-family: Arial, sans-serif; }
        .highlight { background-color: yellow; }
        .code { font-family: monospace; background-color: #f5f5f5; padding: 2px 4px; }
    </style>
</head>
<body>
    <header>
        <h1>HTML Test Document</h1>
        <nav>
            <ul>
                <li><a href="#section1">Section 1</a></li>
                <li><a href="#section2">Section 2</a></li>
            </ul>
        </nav>
    </header>

    <main>
        <article>
            <h2 id="section1">Section 1: Basic HTML</h2>

            <p>This is a paragraph with <strong>bold text</strong>, <em>italic text</em>,
            and <span class="highlight">highlighted text</span>.</p>

            <p>Here's some <span class="code">inline code</span> and a
            <a href="https://example.com">link to example.com</a>.</p>

            <h3>Lists</h3>
            <ul>
                <li>Unordered item 1</li>
                <li>Unordered item 2</li>
                <li>Unordered item 3</li>
            </ul>

            <ol>
                <li>Ordered item 1</li>
                <li>Ordered item 2</li>
                <li>Ordered item 3</li>
            </ol>

            <h2 id="section2">Section 2: Advanced HTML</h2>

            <blockquote>
                <p>This is a blockquote. It should be styled differently from regular paragraphs.</p>
                <footer>— Citation source</footer>
            </blockquote>

            <h3>Table</h3>
            <table border="1">
                <thead>
                    <tr>
                        <th>Header 1</th>
                        <th>Header 2</th>
                        <th>Header 3</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>Cell 1</td>
                        <td>Cell 2</td>
                        <td>Cell 3</td>
                    </tr>
                    <tr>
                        <td>Cell 4</td>
                        <td>Cell 5</td>
                        <td>Cell 6</td>
                    </tr>
                </tbody>
            </table>

            <h3>Code Block</h3>
            <pre><code>function hello() {
    console.log("Hello, World!");
}

hello();</code></pre>

            <h3>Images</h3>
            <p>An image should appear below:</p>
            <img src="https://via.placeholder.com/400x200?text=HTML+Test+Image" alt="Test image" />

        </article>
    </main>

    <footer>
        <p>This is the footer content.</p>
    </footer>
</body>
</html>`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        100,
		},
		Converters: []string{"prince", "weasyprint"}, // Typst doesn't support HTML input yet
	}
}

// marginTestCase tests different margin configurations
func marginTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "custom-margins",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Margin Test Document

This document tests custom margin settings.

## Page Layout

This page should have:
- Top margin: 1cm
- Right margin: 3cm
- Bottom margin: 2cm
- Left margin: 4cm

## Content

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

### More Content

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

The different margins should be clearly visible in the PDF output.
`,
		Options: converter.PageOptions{
			Columns:      1,
			Orientation:  "portrait",
			Margin:       "2cm", // Default margin (will be overridden)
			MarginTop:    "1cm",
			MarginRight:  "3cm",
			MarginBottom: "2cm",
			MarginLeft:   "4cm",
			Zoom:         100,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// zoomTestCase tests zoom functionality
func zoomTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "zoom-150",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Zoom Test Document (150%)

This document tests the zoom functionality at 150%.

## Text Sizes

All text in this document should be rendered at 150% of the normal size.

### Regular Text

This is regular paragraph text that should appear larger than normal.

### Code

` + "`inline code`" + ` should also be larger.

Block code:
` + "```" + `
This code block should also be zoomed to 150%.
Each line should be clearly larger than normal.
` + "```" + `

### Lists

- List item 1 (150% size)
- List item 2 (150% size)
- List item 3 (150% size)

The zoom should affect all text consistently.
`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        150,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// twoColumnsTestCase tests two-column newspaper layout
func twoColumnsTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "two-columns",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Two Column Layout Test

This document tests the two-column newspaper layout feature.

## Column Layout

This content should be laid out in two columns, similar to a traditional newspaper or magazine layout. Two columns provide a good balance between readability and space efficiency.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

## Section 1: Reading Flow

Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.

Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.

## Section 2: Content Distribution

Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem.

At vero eos et accusamus et iusto odio dignissimos ducimus qui blanditiis praesentium voluptatum deleniti atque corrupti quos dolores et quas molestias excepturi sint occaecati cupiditate non provident.

The text should flow naturally from the left column to the right column, creating an easy reading experience typical of newspapers and magazines.
`,
		Options: converter.PageOptions{
			Columns:     2,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        100,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// threeColumnsTestCase tests three-column layout
func threeColumnsTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "three-columns",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Three Column Layout Test

This document tests the three-column layout feature.

## Column Layout

This content should be laid out in three columns, similar to a newspaper layout.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

### Section 1

Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.

### Section 2

Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.

### Section 3

Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem.

## More Content

At vero eos et accusamus et iusto odio dignissimos ducimus qui blanditiis praesentium voluptatum deleniti atque corrupti quos dolores et quas molestias excepturi sint occaecati cupiditate non provident.

The text should flow naturally across the three columns.
`,
		Options: converter.PageOptions{
			Columns:     3,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        100,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// landscapeTestCase tests landscape orientation
func landscapeTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "landscape-orientation",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Landscape Orientation Test

This document should be rendered in landscape orientation.

## Wide Content

The page should be wider than it is tall, suitable for content that benefits from horizontal space.

| Column 1 | Column 2 | Column 3 | Column 4 | Column 5 | Column 6 |
|----------|----------|----------|----------|----------|----------|
| Data 1   | Data 2   | Data 3   | Data 4   | Data 5   | Data 6   |
| Data 7   | Data 8   | Data 9   | Data 10  | Data 11  | Data 12  |
| Data 13  | Data 14  | Data 15  | Data 16  | Data 17  | Data 18  |

## Code Blocks

Wide code blocks should fit better in landscape:

` + "```javascript" + `
const veryLongFunctionName = (parameterOne, parameterTwo, parameterThree, parameterFour) => {
    return parameterOne + parameterTwo + parameterThree + parameterFour;
};
` + "```" + `

The landscape orientation should be clearly visible in the output.
`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "landscape",
			Margin:      "2cm",
			Zoom:        100,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// firstPageGuideTestCase tests the first page guide feature
func firstPageGuideTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "first-page-guide",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# First Page Guide Test

This document tests the first page guide feature.

## Guide Line

There should be a thin vertical line 3cm from the left edge of the first page only.

### Page 1 Content

This is the first page. You should see a vertical guide line 3cm from the left margin.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.

### More Content

Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

### Even More Content

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.

## Page Break

This content should continue to a second page if the document is long enough.

The second page should NOT have the guide line - it should only appear on the first page.

Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium.
`,
		Options: converter.PageOptions{
			Columns:        1,
			Orientation:    "portrait",
			Margin:         "2cm",
			Zoom:           100,
			FirstPageGuide: "3cm",
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// codeBlockTestCase tests code block rendering
func codeBlockTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "code-blocks",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Code Block Test

This document focuses on testing code block rendering.

## Inline Code

Here's some ` + "`inline code`" + ` that should be formatted differently.

## Code Blocks

### Go Code

` + "```go" + `
package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("Hello, World!")

    if len(os.Args) > 1 {
        fmt.Printf("Arguments: %v\n", os.Args[1:])
    }
}
` + "```" + `

### Python Code

` + "```python" + `
def fibonacci(n):
    """Calculate the nth Fibonacci number."""
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

# Test the function
for i in range(10):
    print(f"F({i}) = {fibonacci(i)}")
` + "```" + `

### JavaScript Code

` + "```javascript" + `
class Calculator {
    constructor() {
        this.history = [];
    }

    add(a, b) {
        const result = a + b;
        this.history.push({operation: 'add', a, b, result});
        return result;
    }

    getHistory() {
        return this.history;
    }
}

const calc = new Calculator();
console.log(calc.add(5, 3));
` + "```" + `

### Plain Text Code

` + "```" + `
This is a plain code block without syntax highlighting.
It should still be formatted as code.

Line 1
Line 2
Line 3
` + "```" + `

Code blocks should have consistent formatting and be easily distinguishable from regular text.
`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        100,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// ligatureTestCase tests ligature handling in code blocks
func ligatureTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "ligatures-in-code",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Ligature Test in Code Blocks

This document tests that ligatures (like fi, fl, ffi) are properly disabled in code blocks.

## Regular Text (Ligatures OK)

In regular text, ligatures should work fine: official,ffle, office, difficult.

## Code Blocks (Ligatures Should Be Disabled)

### Test Cases with Common Ligatures

` + "```javascript" + `
// These should NOT have ligatures - each letter should be separate
function official() {
    // fi ligature should be disabled
    const config = {
        filter: true,
        offline: false
    };

    // fl ligature should be disabled
    const flags = ["--flag", "--profile"];

    // ffi ligature should be disabled
    const difficult = "coefficient";

    return config;
}
` + "```" + `

### More Ligature Test Cases

` + "```go" + `
// fi combinations
if filter != nil {
    profile := "default"
    config := GetConfig()
}

// fl combinations
flag := "--profile"
float64Value := 3.14

// ffi combinations
coefficient := 1.23
office := "main"
` + "```" + `

### Python Ligature Test

` + "```python" + `
# fi combinations
def filter_files(profile):
    config = load_config()

# fl combinations
flags = ["--profile", "--filter"]
floating_point = 3.14159

# ffi combinations
coefficient = 2.5
office_hours = "9-5"
` + "```" + `

## Expected Behavior

In code blocks above, you should see:
- **fi** rendered as separate f + i (not as ligature)
- **fl** rendered as separate f + l (not as ligature)
- **ffi** rendered as separate f + f + i (not as ligature)

This ensures code readability and prevents confusion between similar-looking characters.
`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        100,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// tableTestCase tests table rendering
func tableTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "tables",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Table Test

This document tests table rendering capabilities.

## Simple Table

| Feature | Status | Priority |
|---------|--------|----------|
| Basic conversion | ✓ | High |
| Multi-column layout | ✓ | Medium |
| Custom margins | ✓ | Medium |
| Zoom support | ✓ | Low |

## Table with Alignment

| Left Aligned | Center Aligned | Right Aligned |
|:-------------|:--------------:|--------------:|
| Left 1       | Center 1       | Right 1       |
| Left 2       | Center 2       | Right 2       |
| Left 3       | Center 3       | Right 3       |

## Complex Table

| Converter | Auto-Download | License | Pros | Cons |
|-----------|---------------|---------|------|------|
| Typst | ✓ | Apache 2.0 | Modern typography, fast | Newer, less features |
| Prince | ✗ | Commercial | Professional output | Expensive license |
| WeasyPrint | ✗ | BSD | Open source, reliable | Slower rendering |

## Table with Long Content

| Description | Details |
|-------------|---------|
| Short | Brief |
| Medium length content | This is a medium length cell with more text that should wrap properly within the cell boundaries |
| Very long content that spans multiple lines | This is a very long cell content that should demonstrate how tables handle text wrapping and cell expansion when the content is much longer than the typical cell width would accommodate |

Tables should render consistently across all converters with proper borders, alignment, and text wrapping.
`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        100,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// footnoteTestCase verifies footnote rendering across converters
func footnoteTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "footnotes",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Footnote Rendering

This paragraph references a standard footnote marker[^intro] to ensure placement.

Another sentence introduces a second footnote[^second] that contains formatted text.

Sometimes a footnote needs multiple sentences[^long] to explain a concept thoroughly.

[^intro]: First footnote with **bold emphasis** and inline code ` + "`print(\"footnote\")`" + `.
[^second]: Second footnote with a [link](https://example.com) and *italic text* for variety.
[^long]: This longer footnote spans two sentences. It ensures our layout handles extended explanations without breaking the page flow.`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        100,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// footnoteMultiPageTestCase tests footnote rendering across multiple pages
func footnoteMultiPageTestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "footnotes-multipage",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Multi-Page Footnote Test

This document tests footnote rendering when content and footnotes span multiple pages.

## Section 1: Introduction

The first section introduces several concepts with extensive footnotes[^1] that may cause page breaks.

Additional content with another reference[^2] to test layout continuity.

More text to ensure we have enough content[^3] to span multiple pages in the PDF output.

## Section 2: Detailed Analysis

This section continues with more footnotes[^4] and additional explanatory text[^5] to verify proper rendering.

We need to ensure that footnote markers[^6] remain correctly positioned throughout the document.

## Section 3: Additional Content

Lorem ipsum dolor sit amet[^7], consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore[^8] et dolore magna aliqua.

Ut enim ad minim veniam[^9], quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat[^10].

## Section 4: More References

Duis aute irure dolor[^11] in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur[^12].

Excepteur sint occaecat cupidatat non proident[^13], sunt in culpa qui officia deserunt mollit anim id est laborum[^14].

## Section 5: Extended Discussion

This section provides even more content[^15] with detailed footnotes to ensure multi-page layout works correctly[^16].

Final paragraph with concluding references[^17] to verify all footnotes render properly[^18].

[^1]: First footnote with substantial content. This footnote contains multiple sentences to test how longer footnotes are rendered. It includes **bold text** and *italic text* as well as inline code ` + "`example()`" + ` to verify formatting is preserved.
[^2]: Second footnote with a [link to example](https://example.com) and additional explanatory text.
[^3]: Third footnote demonstrating that footnotes can contain lists: 1) First item, 2) Second item, 3) Third item.
[^4]: Fourth footnote with technical details about implementation and rendering behavior.
[^5]: Fifth footnote explaining concepts in depth with extended prose that may wrap across lines.
[^6]: Sixth footnote testing marker positioning and alignment with the main text flow.
[^7]: Seventh footnote with Latin text: Lorem ipsum dolor sit amet, consectetur adipiscing elit.
[^8]: Eighth footnote containing code snippets: ` + "`function test() { return true; }`" + `.
[^9]: Ninth footnote with a longer explanation that spans multiple sentences. This tests whether footnotes that extend beyond a single line maintain proper formatting and readability.
[^10]: Tenth footnote providing context and background information for the referenced statement.
[^11]: Eleventh footnote with comparative analysis and detailed examination of the topic.
[^12]: Twelfth footnote containing a blockquote-style note about related concepts and considerations.
[^13]: Thirteenth footnote with methodological notes about the approach and rationale.
[^14]: Fourteenth footnote discussing implications and consequences of the referenced statement.
[^15]: Fifteenth footnote with extended commentary and supplementary information for interested readers.
[^16]: Sixteenth footnote verifying that page breaks don't disrupt the footnote rendering flow.
[^17]: Seventeenth footnote with concluding remarks and final observations about the topic.
[^18]: Eighteenth and final footnote ensuring complete coverage of multi-page footnote scenarios.`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        100,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// zoom80TestCase tests 80% zoom (smaller text for dense content)
func zoom80TestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "zoom-80",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Zoom Test Document (80%)

This document tests the zoom functionality at 80% - useful for dense content or fitting more text per page.

## Text Sizes

All text in this document should be rendered at 80% of the normal size, making it smaller and more compact.

### Use Cases for 80% Zoom

- Dense technical documentation
- Reference materials
- Fitting more content per page
- Reducing paper usage

### Sample Content

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

### Code at 80%

` + "```javascript" + `
function smallerCode() {
    // This code should be 80% of normal size
    const config = { compact: true };
    return config;
}
` + "```" + `

The 80% zoom should make all text consistently smaller while maintaining readability.
`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        80,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// zoom120TestCase tests 120% zoom (larger text for accessibility)
func zoom120TestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "zoom-120",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Zoom Test Document (120%)

This document tests the zoom functionality at 120% - useful for improved readability and accessibility.

## Text Sizes

All text in this document should be rendered at 120% of the normal size, making it larger and easier to read.

### Use Cases for 120% Zoom

- Improved accessibility for visually impaired users
- Better readability for older adults
- Presentations or display materials
- Reducing eye strain

### Sample Content

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

### Code at 120%

` + "```python" + `
def larger_code():
    # This code should be 120% of normal size
    return {"readable": True}
` + "```" + `

The 120% zoom should make all text consistently larger while maintaining good layout.
`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        120,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}

// zoom200TestCase tests 200% zoom (very large text)
func zoom200TestCase() GoldenTestCase {
	return GoldenTestCase{
		Name:        "zoom-200",
		ContentType: fetcher.ContentTypeMarkdown,
		Input: `# Zoom Test Document (200%)

This document tests the zoom functionality at 200% - very large text for special accessibility needs.

## Text Sizes

All text in this document should be rendered at 200% of the normal size, making it significantly larger.

### Use Cases for 200% Zoom

- High accessibility requirements
- Large format displays
- Vision assistance
- Presentation materials

### Sample Content

This text should be very large and clearly visible.

### Code at 200%

` + "```go" + `
func veryLargeCode() {
    fmt.Println("Big text!")
}
` + "```" + `

The 200% zoom should make all text very large while maintaining layout integrity.
`,
		Options: converter.PageOptions{
			Columns:     1,
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        200,
		},
		Converters: []string{"typst", "prince", "weasyprint"},
	}
}
