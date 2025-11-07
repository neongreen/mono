# Testing Guide for printpdf

This document describes the testing strategy and how to run tests for printpdf.

## Test Types

### 1. Unit Tests
Traditional Go unit tests for individual components:

```bash
# Run all unit tests
mise run printpdf:test

# Or manually
go test ./pkg/converter ./pkg/fetcher
```

**Location**: `pkg/*/test.go` files
**Coverage**: 
- Option parsing and validation
- Content type detection  
- URL parsing
- Zoom calculations

### 2. Golden Tests (Visual Regression)
Comprehensive end-to-end tests that generate PDFs and compare visual output:

```bash
# Run golden tests (requires tools)
mise run printpdf:test-golden

# Update golden references
mise run printpdf:update-goldens
```

**Location**: `pkg/golden/`
**Coverage**:
- All 3 converters (Typst, Prince, WeasyPrint)
- Markdown and HTML input types
- All page options (margins, zoom, columns, orientation)
- Complex content (tables, code blocks, nested lists)

## Prerequisites for Golden Tests

### Required Tools
```bash
# Ubuntu/Debian
sudo apt-get install poppler-utils imagemagick

# macOS  
brew install poppler imagemagick
```

### PDF Converters
- **WeasyPrint**: `pip install weasyprint` (most important)
- **Typst**: Auto-downloaded by printpdf
- **Prince**: Manual install from princexml.com (optional)

## Running Tests

### Quick Unit Tests
```bash
mise run printpdf:test
```

### Full Golden Test Suite  
```bash
# Check prerequisites first
which pdftoppm imagemagick weasyprint

# Run golden tests
mise run printpdf:test-golden
```

### Single Test Case
```bash
cd printpdf
go test ./pkg/golden -run TestGoldenSuite/basic-markdown -v
```

## Creating Golden References

**IMPORTANT**: Always manually review golden images before committing!

```bash
# Generate new golden references
mise run printpdf:update-goldens

# Review generated images
ls -la pkg/golden/testdata/golden/

# Example review process
find pkg/golden/testdata/golden -name "*.png" | head -5 | xargs open

# Only commit after manual verification
git add pkg/golden/testdata/golden/
git commit -m "printpdf: Update golden test references"
```

## Test Case Structure

Golden test cases are defined in `pkg/golden/testcases_test.go`:

```go
func myTestCase() GoldenTestCase {
    return GoldenTestCase{
        Name:        "my-test-case",
        ContentType: fetcher.ContentTypeMarkdown,
        Input:       `# Test Content...`,
        Options: converter.PageOptions{
            Columns:     1,
            Orientation: "portrait", 
            Margin:      "2cm",
            Zoom:        100,
        },
        Converters: []string{"typst", "weasyprint"},
    }
}
```

## Adding New Test Cases

1. **Define test case** in `testcases_test.go`
2. **Add to suite** in `TestGoldenSuite()`
3. **Generate goldens**: `UPDATE_GOLDENS=1 go test ./pkg/golden`
4. **Review images** manually
5. **Commit goldens** after verification

## CI/CD Integration

The golden tests should be added to `.github/workflows/printpdf.yml`:

```yaml
- name: Install PDF tools
  run: |
    sudo apt-get update
    sudo apt-get install poppler-utils imagemagick
    pip install weasyprint

- name: Run tests
  run: |
    cd printpdf
    go test ./...
    go test ./pkg/golden -v -timeout=10m
```

## Troubleshooting

### Missing Tools
```
Error: PDF to image tool not found: pdftoppm
```
**Fix**: `sudo apt-get install poppler-utils`

### Missing Converters
Tests auto-skip unavailable converters. For complete testing:
- Install WeasyPrint: `pip install weasyprint`
- Install Typst: Auto-downloaded
- Install Prince: Manual (optional)

### Image Differences
```
Image comparison failed for page-1.png: images differ (metric: 1234)
```

**Investigation steps**:
1. Check `pkg/golden/testdata/output/` for current output
2. Compare with `pkg/golden/testdata/golden/` references
3. If change is intentional: `UPDATE_GOLDENS=1 go test ./pkg/golden`
4. If change is a bug: fix code and re-run tests

### Large Test Output
```bash
# Run specific test
go test ./pkg/golden -run TestGoldenSuite/basic-markdown/weasyprint -v

# Check specific converter
go test ./pkg/golden -run TestGoldenSuite/.*weasyprint -v
```

## Test Coverage

Current golden test cases:

| Test Case | Description | Converters |
|-----------|-------------|------------|
| basic-markdown | Simple Markdown features | All |
| complex-markdown | Advanced Markdown | All |
| html-input | HTML input processing | Prince, WeasyPrint |
| custom-margins | Individual margin settings | All |
| zoom-150 | Font scaling | All |
| three-columns | Multi-column layout | All |
| landscape-orientation | Landscape pages | All |
| first-page-guide | Vertical guide line | All |
| code-blocks | Code syntax highlighting | All |
| tables | Table rendering | All |

## Best Practices

### For Development
1. **Run unit tests first** - they're faster
2. **Run golden tests before commits** - catch regressions
3. **Update goldens carefully** - always review visually

### For Adding Features  
1. **Add unit tests** for new logic
2. **Add golden test case** for visual features
3. **Test with all converters** when possible

### For Bug Fixes
1. **Add test case** that reproduces the bug
2. **Fix the bug**
3. **Verify test passes**
4. **Update goldens if output intentionally changed**

This comprehensive testing strategy ensures printpdf maintains high quality output across all supported converters and input types.