# D2 Comparison Feature

This document describes the D2 comparison feature added to diagram-dsl.

## Overview

We've added D2 language equivalents for all diagram-dsl examples to allow direct comparison between the two diagramming approaches. This helps users understand the trade-offs between using diagram-dsl (React/JSX with explicit flexbox layout) versus D2 (specialized DSL with automatic layout).

## What was added

### D2 Example Files

Located in `examples/d2/`:
- `simple-box.d2` - A simple box with text
- `basic-flowchart.d2` - Three-step vertical flowchart
- `architecture-diagram.d2` - Three-tier architecture (Frontend → API → Database)
- `styled-flowchart.d2` - Modern flowchart with colored cards
- `styled-architecture.d2` - Three-tier architecture with presentation/business/data layers
- `multi-tier-architecture.d2` - Comprehensive multi-tier web application
- `decision-flowchart.d2` - User authentication flow with conditional branches

### Generation Scripts

Located in `scripts/`:

1. **`generate-d2-examples.sh`** - Generates SVG files from D2 source files
   - Input: `examples/d2/*.d2`
   - Output: `examples/d2-output/*.svg`
   - Uses D2 CLI tool with theme 200

2. **`generate-all-examples.sh`** - Master script that generates both versions
   - Runs `npm run examples` (diagram-dsl examples)
   - Runs `generate-d2-examples.sh` (D2 examples)
   - Provides unified output

### NPM Scripts

Added to `package.json`:
- `npm run examples:d2` - Generate D2 examples only
- `npm run examples:all` - Generate both diagram-dsl and D2 examples

### Documentation

1. **README.md updates**:
   - Added "Comparison with D2" section
   - Includes 7 side-by-side comparison tables
   - Shows rendered output for both tools
   - Explains key differences between approaches

2. **examples/d2/README.md**:
   - Explains what D2 is
   - Lists all D2 files
   - Provides generation instructions
   - Compares advantages of each tool

### Configuration

- Updated `.gitignore` to exclude `examples/d2-output/` (generated files)

## Usage

### Generating Examples

```bash
# Generate only D2 examples
npm run examples:d2

# Generate all examples (both diagram-dsl and D2)
npm run examples:all
```

### Viewing Comparisons

The README.md now includes inline comparison tables showing both outputs. Open the README in GitHub or any Markdown viewer to see the visual comparisons.

Alternatively, SVG files can be viewed directly:
- diagram-dsl outputs: `examples/*.svg`
- D2 outputs: `examples/d2-output/*.svg`

## Requirements

To generate D2 examples, you need D2 installed:

```bash
# macOS
brew install d2

# Linux
curl -fsSL https://d2lang.com/install.sh | sh -s --

# Or download from GitHub releases
# https://github.com/terrastruct/d2/releases
```

## Key Differences

**diagram-dsl**:
- React/JSX syntax
- Explicit flexbox layout (Yoga)
- Precise control over spacing, sizing, alignment
- TypeScript type safety
- Predictable, reproducible layouts

**D2**:
- Custom DSL syntax
- Automatic layout (Dagre/ELK)
- More concise for simple diagrams
- Built-in themes
- Less control over exact positioning

## Examples Not Included

- `title-hierarchy.svg` - This is a typography showcase specific to diagram-dsl's semantic components, not a diagram type that makes sense for D2 comparison.

## File Structure

```
diagram-dsl/
├── examples/
│   ├── *.svg                    # diagram-dsl outputs (committed)
│   ├── d2/
│   │   ├── README.md            # D2 documentation
│   │   └── *.d2                 # D2 source files (committed)
│   └── d2-output/
│       └── *.svg                # D2 generated SVGs (gitignored)
├── scripts/
│   ├── generate-d2-examples.sh      # D2 generation script
│   └── generate-all-examples.sh     # Master generation script
└── README.md                    # Updated with comparison section
```

## Future Enhancements

Possible improvements:
- Add more complex examples (e.g., sequence diagrams, entity-relationship diagrams)
- Create automated visual regression testing
- Add interactive comparison tool
- Include performance benchmarks
- Add more layout algorithm comparisons (force-directed, hierarchical, etc.)
