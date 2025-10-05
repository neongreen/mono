# D2 Comparison Examples

This directory contains D2 language equivalents of the diagram-dsl examples for comparison purposes.

## What is D2?

D2 is a modern diagram scripting language that compiles to SVG. It uses automatic layout algorithms to position elements, making it more concise but less predictable than diagram-dsl's explicit flexbox-based layout.

## Files

- `simple-box.d2` - A simple box with text
- `basic-flowchart.d2` - Three-step vertical flowchart
- `architecture-diagram.d2` - Three-tier architecture (Frontend → API → Database)
- `styled-flowchart.d2` - Modern flowchart with colored cards
- `styled-architecture.d2` - Three-tier architecture with presentation/business/data layers
- `multi-tier-architecture.d2` - Comprehensive multi-tier web application
- `decision-flowchart.d2` - User authentication flow with conditional branches

## Generating SVGs

To generate SVG files from these D2 diagrams:

```bash
# From the diagram-dsl root directory
pnpm run examples:d2

# Or generate both at once
pnpm run examples:all

# Or generate both diagram-dsl and D2 examples
pnpm run examples:all
```

The generated SVG files will be placed alongside the DSL outputs in `examples/*-d2.svg`.

## D2 Installation

The generation script (`pnpm run examples:d2`) automatically downloads D2 from GitHub releases if it's not already installed. No manual installation is required!

## Comparison Notes

**diagram-dsl advantages:**
- Precise control over layout and positioning
- Familiar React/JSX syntax for developers
- Explicit width, height, padding, gaps
- Flexbox layout engine (Yoga)
- TypeScript type safety

**D2 advantages:**
- More concise syntax
- Automatic layout (less manual positioning)
- Simpler for basic diagrams
- Built-in themes

**When to use diagram-dsl:**
- When you need precise control over spacing and alignment
- When you want to integrate diagrams with React-based tooling
- When you need predictable, reproducible layouts
- When you're already familiar with React/JSX/flexbox

**When to use D2:**
- For quick, simple diagrams
- When automatic layout is sufficient
- When you prefer a specialized DSL over JSX
- When you want minimal code

Both tools produce high-quality SVG output suitable for documentation, presentations, and technical materials.
