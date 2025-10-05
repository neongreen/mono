# diagram-dsl Architecture

This document explains the internal architecture of the diagram-dsl library.

## Overview

The library follows a three-stage pipeline:

1. **JSX Parsing** - React elements → Layout tree
2. **Layout Computation** - Yoga layout engine calculates positions
3. **SVG Rendering** - Layout tree → SVG string

```
React JSX → Layout Tree → Yoga Engine → Computed Tree → SVG Renderer → SVG String
```

## Directory Structure

```
diagram-dsl/
├── src/
│   ├── components/       # React component definitions
│   │   ├── Box.tsx
│   │   ├── Stack.tsx
│   │   ├── Row.tsx
│   │   ├── Column.tsx
│   │   ├── Text.tsx
│   │   └── Arrow.tsx
│   ├── layout/           # Layout engine wrapper
│   │   └── yoga-engine.ts
│   ├── renderer/         # SVG rendering logic
│   │   ├── index.ts      # Main rendering pipeline
│   │   └── svg-renderer.ts
│   ├── types/            # TypeScript type definitions
│   │   └── index.ts
│   ├── examples/         # Example diagrams
│   │   ├── basic.tsx
│   │   └── advanced.tsx
│   ├── test/             # Test suite
│   │   └── index.test.tsx
│   └── index.ts          # Public API exports
├── examples/             # Generated SVG files
└── dist/                 # Compiled JavaScript output
```

## Component Layer

### React Components

All components are simple functional React components that use `React.createElement` to create element trees. They don't render to DOM - instead, they create a tree structure that will be processed by the layout engine.

**Key Components:**
- `Box` - Base container with layout properties
- `Stack` - Vertical or horizontal layout container
- `Row` - Horizontal layout (shorthand for Stack)
- `Column` - Vertical layout (shorthand for Stack)
- `Text` - Text rendering with typography options
- `Arrow` - Connection between boxes (doesn't participate in layout)

### Why React?

Using React gives us:
- Familiar JSX syntax for developers
- Component composition and reusability
- Type-safe props with TypeScript
- No need to reinvent a component model

## Layout Layer

### Yoga Layout Engine

[Yoga](https://yogalayout.dev/) is Facebook's implementation of CSS Flexbox in C++, with bindings for many languages including JavaScript. It's the same layout engine used by React Native.

**Why Yoga?**
- Battle-tested (used by millions of React Native apps)
- Implements CSS Flexbox specification
- Fast and predictable
- Cross-platform consistency

### YogaLayoutEngine Class

The `YogaLayoutEngine` class wraps the Yoga library and handles:

1. **Node Creation** - Converts layout tree to Yoga nodes
2. **Property Mapping** - Maps our props to Yoga properties
3. **Layout Computation** - Calls Yoga to calculate positions
4. **Result Extraction** - Extracts computed positions back to our tree
5. **Memory Management** - Frees Yoga nodes after use

**Special Handling:**
- Arrow nodes don't participate in layout (return null)
- Text nodes get estimated dimensions based on font size
- Flex direction is inferred from component type

## Rendering Layer

### Element to Layout Tree

The `elementToLayoutNode` function recursively converts React elements to a simpler tree structure:

```typescript
interface LayoutNode {
  type: string;
  props: any;
  children: LayoutNode[];
  computed?: {
    x: number;
    y: number;
    width: number;
    height: number;
  };
}
```

**Key transformations:**
- Function components are evaluated to get their elements
- Text strings are wrapped in Text nodes
- Props are cleaned up (children removed from props)
- Arrow nodes are preserved but marked as non-layout

### SVG Renderer

The `SVGRenderer` class generates SVG markup from the computed layout tree:

**Two-pass rendering:**
1. **First pass** - Render all layout elements (boxes, text)
2. **Second pass** - Render arrows using collected positions

**Features:**
- Box rendering with borders, backgrounds, and border radius
- Text rendering with font properties and alignment
- Arrow rendering with proper arrowhead markers
- XML escaping for text content
- Marker definitions for arrow colors

### Arrow Positioning

Arrows are special because they:
1. Don't participate in layout (no Yoga node)
2. Need to know positions of other elements (via `id` props)
3. Calculate connection points between box centers
4. Adjust endpoints to box edges (not overlapping)

## Type System

All layout, alignment, and position properties are strongly typed:

```typescript
interface LayoutProps {
  width?: number | 'auto';
  height?: number | 'auto';
  padding?: number;
  margin?: number;
  gap?: number;
  // ... more properties
}

interface AlignmentProps {
  alignItems?: 'flex-start' | 'center' | 'flex-end' | 'stretch';
  justifyContent?: 'flex-start' | 'center' | 'flex-end' | ...;
}
```

This provides:
- IntelliSense in IDEs
- Compile-time error checking
- Self-documenting API

## Async Rendering

The `renderToSVG` function is async because Yoga is loaded asynchronously (using top-level await in ES modules). This is a limitation of the yoga-layout package but doesn't impact the API significantly.

```typescript
const svg = await renderToSVG(<MyDiagram />, options);
```

## Testing

Tests verify:
1. Individual component rendering
2. Layout composition (Stack, Row, Column)
3. Text rendering and alignment
4. Arrow rendering between boxes
5. Styling properties (borders, colors)

Tests use the same public API as end users, ensuring the external interface works as expected.

## Performance Considerations

- Yoga layout computation is very fast (C++ native code)
- SVG string generation is simple string concatenation
- No DOM rendering or browser APIs needed
- Can be used in Node.js servers or build tools


