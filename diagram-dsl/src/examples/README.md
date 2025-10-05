# diagram-dsl Examples

This directory contains example diagrams demonstrating the capabilities of the diagram-dsl library.

## Running Examples

Generate all basic examples:
```bash
pnpm dev
```

Generate advanced examples:
```bash
pnpm dev:advanced
```

## Example Files

### basic.tsx
- **basic-flowchart.svg** - Simple vertical flowchart showing a linear process
- **architecture-diagram.svg** - Three-tier architecture with frontend, API, and database

### advanced.tsx
- **multi-tier-architecture.svg** - Comprehensive multi-tier web application architecture with client, application, and data tiers
- **decision-flowchart.svg** - User authentication flow with conditional branches

## Features Demonstrated

### Basic Examples
- Vertical stacks with gaps
- Boxes with colors, borders, and rounded corners
- Text with different font sizes and weights
- Simple arrows between boxes
- Centered alignment

### Advanced Examples
- Multi-row layouts with Row component
- Complex nested structures (Stack > Row > Box)
- Multiple arrows showing data flow
- Labels on arrows
- Color-coded tiers/sections
- Different box shapes (rounded boxes for start/end states)
- Real-world architecture patterns

## Viewing the Output

The generated SVG files can be viewed in:
- Any modern web browser
- Image viewers that support SVG
- Integrated directly into documentation
- Presentation tools that support SVG import

## Customization

Feel free to modify these examples or create your own by:
1. Creating a new .tsx file in this directory
2. Importing the necessary components from '../index'
3. Composing your diagram using JSX
4. Calling `renderToSVG()` with your diagram
5. Saving the output to an SVG file
