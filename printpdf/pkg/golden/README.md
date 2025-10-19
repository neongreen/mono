# Golden Test Suite for printpdf

This package provides a comprehensive golden test suite for printpdf that uses PDF-to-image conversion and visual comparison to detect regressions.

## Overview

The golden test suite:
1. **Generates PDFs** using all available converters (Typst, Prince, WeasyPrint)
2. **Converts PDFs to images** using `pdftoppm` 
3. **Compares images** against golden references using ImageMagick's `compare`
4. **Detects visual regressions** automatically
5. **Covers comprehensive test cases** including margins, zoom, columns, orientations, etc.

## Prerequisites

Before running the golden tests, install the required tools:

### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install poppler-utils imagemagick
```

### macOS
```bash
brew install poppler imagemagick
```

### Converters
For complete testing, install the PDF converters:

- **WeasyPrint**: `pip install weasyprint` (usually available)
- **Typst**: Install from https://github.com/typst/typst (auto-downloaded by printpdf)
- **Prince**: Install from https://www.princexml.com/ (commercial, optional)

## Usage

### Running Tests

```bash
# Run all golden tests
cd printpdf
go test ./pkg/golden -v

# Run with verbose output
go test ./pkg/golden -v -timeout=10m
```

### Creating/Updating Golden References

When running tests for the first time, or when intentionally changing output:

```bash
# Update golden reference images
UPDATE_GOLDENS=1 go test ./pkg/golden -v
```

This will:
1. Generate PDFs with current code
2. Convert PDFs to images  
3. Save images as new golden references
4. **Requires manual review** - always inspect the generated images before committing!

### Test Structure

```
printpdf/pkg/golden/
├── testdata/
│   ├── output/           # Generated test outputs (gitignored)
│   └── golden/           # Golden reference images (committed)
│       ├── basic-markdown-typst-images/
│       │   ├── page-1.png
│       │   └── page-2.png
│       ├── basic-markdown-weasyprint-images/
│       └── ...
├── golden_test.go        # Core test framework
├── testcases_test.go     # Test case definitions
└── README.md            # This file
```

## Test Cases

The suite includes comprehensive test cases:

### Basic Tests
- **basic-markdown**: Simple Markdown features (headers, lists, code, links)
- **complex-markdown**: Advanced features (nested lists, blockquotes, tables, special chars)
- **html-input**: Direct HTML input processing

### Layout Tests  
- **custom-margins**: Different margin configurations
- **three-columns**: Multi-column newspaper layout
- **landscape-orientation**: Landscape page orientation
- **first-page-guide**: Vertical guide line on first page

### Content Tests
- **zoom-150**: 150% font size scaling
- **code-blocks**: Various programming language code blocks
- **tables**: Table rendering and alignment
- **footnotes**: Footnote references and inline footnote bodies across converters

## Workflow

### 1. Development Workflow
```bash
# Make changes to printpdf code
vim pkg/converter/typst.go

# Run golden tests to check for regressions
go test ./pkg/golden -v

# If tests fail, investigate:
# - Are the changes intentional? Update goldens.
# - Are they bugs? Fix the code.
```

### 2. Adding New Test Cases
```go
// Add to testcases_test.go
func myNewTestCase() GoldenTestCase {
    return GoldenTestCase{
        Name:        "my-new-feature",
        ContentType: fetcher.ContentTypeMarkdown,
        Input:       `# Test new feature...`,
        Options:     converter.PageOptions{...},
        Converters:  []string{"typst", "weasyprint"},
    }
}

// Add to TestGoldenSuite
suite.AddTestCase(myNewTestCase())
```

### 3. Updating Golden References
```bash
# After intentional changes to output
UPDATE_GOLDENS=1 go test ./pkg/golden -v

# Review generated images
ls -la pkg/golden/testdata/golden/

# Commit new goldens after manual review
git add pkg/golden/testdata/golden/
git commit -m "Update golden references for new feature"
```

## CI/CD Integration

Add to `.github/workflows/printpdf.yml`:

```yaml
- name: Install PDF tools
  run: |
    sudo apt-get update
    sudo apt-get install poppler-utils imagemagick
    pip install weasyprint

- name: Run golden tests
  run: |
    cd printpdf
    go test ./pkg/golden -v -timeout=10m
```

## Troubleshooting

### Missing Tools
```
Error: PDF to image tool not found: pdftoppm
```
**Solution**: Install poppler-utils (`apt-get install poppler-utils`)

### Missing Converters
```
Converter typst not available
```
**Solution**: Converters are auto-skipped if not available. Install for complete testing.

### Image Differences
```
Image comparison failed for page-1.png: images differ (metric: 1234)
```
**Solution**: 
1. Check if difference is intentional (new feature)
2. If intentional: `UPDATE_GOLDENS=1 go test ./pkg/golden`
3. If bug: fix code and re-run tests

### Large Diffs
If you see many test failures after changes:
1. Run one test case: `go test ./pkg/golden -run TestGoldenSuite/basic-markdown`
2. Visually inspect outputs in `testdata/output/`
3. Compare with goldens in `testdata/golden/`
4. Update goldens only after confirming changes are correct

## Manual Review Process

**IMPORTANT**: Always manually review golden reference images before committing:

```bash
# Generate new goldens
UPDATE_GOLDENS=1 go test ./pkg/golden -v

# Review each generated image
find pkg/golden/testdata/golden -name "*.png" | xargs -I {} echo "Reviewing: {}"

# Use image viewer to check:
# - Text is clear and properly formatted
# - Margins are correct
# - Layouts look reasonable
# - No obvious rendering bugs

# Only commit after review
git add pkg/golden/testdata/golden/
git commit -m "Update golden references: describe what changed"
```

This ensures the golden references represent the intended output quality.
