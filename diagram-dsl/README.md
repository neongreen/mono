# diagram-dsl

A high-level TypeScript DSL for creating diagrams and flowchart-heavy slides using React and JSX. Generate beautiful SVG diagrams without dragging boxes around or fighting with layout engines.

## Philosophy

Most graphical diagram tools require manual positioning and constant tweaking. Text-based tools like Mermaid, D2, or Graphviz try to auto-layout everything, which often requires hacks when you want specific layouts. This library takes a different approach:

- **Declarative JSX syntax** - Describe your diagrams like you describe UIs
- **Flexible layout primitives** - Stack, Row, Column components with padding, gaps, and alignment
- **Yoga layout engine** - Battle-tested flexbox layout (same engine used by React Native)
- **SVG output** - No browser needed, pure SVG generation
- **TypeScript first** - Full type safety and IntelliSense support

## Installation

```bash
npm install diagram-dsl
```

## Quick Start

```tsx
import React from 'react';
import { Stack, Box, Text, Arrow, renderToSVG } from 'diagram-dsl';
import { writeFileSync } from 'fs';

const MyDiagram = () => (
  <Stack gap={20} padding={40} alignItems="center">
    <Text fontSize={32} fontWeight="bold">My Flowchart</Text>
    
    <Box
      id="step1"
      width={200}
      height={80}
      backgroundColor="#e3f2fd"
      borderColor="#1976d2"
      borderWidth={2}
      borderRadius={8}
      padding={20}
      justifyContent="center"
      alignItems="center"
    >
      <Text fontSize={18}>First Step</Text>
    </Box>

    <Box
      id="step2"
      width={200}
      height={80}
      backgroundColor="#f3e5f5"
      borderColor="#7b1fa2"
      borderWidth={2}
      borderRadius={8}
      padding={20}
      justifyContent="center"
      alignItems="center"
    >
      <Text fontSize={18}>Second Step</Text>
    </Box>

    <Arrow from="step1" to="step2" color="#1976d2" strokeWidth={2} />
  </Stack>
);

async function generate() {
  const svg = await renderToSVG(<MyDiagram />, { 
    width: 800, 
    height: 600,
    backgroundColor: 'white'
  });
  writeFileSync('diagram.svg', svg);
}

generate();
```

## Components

### Box

The basic building block. Can contain other components and has full layout control.

```tsx
<Box
  width={200}
  height={100}
  backgroundColor="#f0f0f0"
  borderColor="black"
  borderWidth={2}
  borderRadius={8}
  padding={20}
  margin={10}
  alignItems="center"
  justifyContent="center"
  position="relative"
  id="mybox"
>
  {/* children */}
</Box>
```

### Stack

Arranges children vertically (default) or horizontally with automatic spacing.

```tsx
<Stack
  direction="vertical" // or "horizontal"
  gap={20}
  alignItems="center"
  justifyContent="space-between"
  padding={40}
>
  {/* children */}
</Stack>
```

### Row

Arranges children horizontally. Shorthand for `<Stack direction="horizontal">`.

```tsx
<Row gap={10} alignItems="center">
  <Box width={100} height={100} backgroundColor="red" />
  <Box width={100} height={100} backgroundColor="blue" />
  <Box width={100} height={100} backgroundColor="green" />
</Row>
```

### Column

Arranges children vertically. Shorthand for `<Stack direction="vertical">`.

```tsx
<Column gap={10} alignItems="center">
  <Box width={200} height={50} backgroundColor="red" />
  <Box width={200} height={50} backgroundColor="blue" />
</Column>
```

### Text

Renders text with full typography control.

```tsx
<Text
  fontSize={24}
  fontWeight="bold"
  color="#333"
  fontFamily="Arial, sans-serif"
  textAlign="center"
>
  Hello World
</Text>
```

### Arrow

Draws an arrow between two boxes (identified by their `id` props).

```tsx
<Arrow
  from="box1"
  to="box2"
  color="#1976d2"
  strokeWidth={2}
  label="connects to"
/>
```

## Layout Props

All components support these layout properties:

- `width`, `height` - Fixed dimensions or `"auto"`
- `minWidth`, `minHeight`, `maxWidth`, `maxHeight` - Size constraints
- `padding`, `paddingTop`, `paddingBottom`, `paddingLeft`, `paddingRight`
- `margin`, `marginTop`, `marginBottom`, `marginLeft`, `marginRight`
- `gap` - Spacing between children (Stack/Row/Column)

## Alignment Props

- `alignItems` - `"flex-start"`, `"center"`, `"flex-end"`, `"stretch"`
- `justifyContent` - `"flex-start"`, `"center"`, `"flex-end"`, `"space-between"`, `"space-around"`, `"space-evenly"`

## Position Props

- `position` - `"relative"` (default) or `"absolute"`
- `top`, `bottom`, `left`, `right` - Absolute positioning offsets

## Examples

Check the `examples/` directory for:
- `basic-flowchart.svg` - Simple vertical flowchart
- `architecture-diagram.svg` - Multi-layer architecture diagram

Run the examples:

```bash
npm run dev
```

## API

### renderToSVG(element, options)

Renders a React element to SVG string.

```typescript
async function renderToSVG(
  element: ReactElement,
  options?: {
    width?: number;      // Default: 800
    height?: number;     // Default: 600
    backgroundColor?: string;  // Default: 'white'
  }
): Promise<string>
```

## Why Yoga?

[Yoga](https://yogalayout.dev/) is Meta's cross-platform layout engine implementing CSS Flexbox. It's battle-tested, used by React Native, and provides predictable layouts without manual calculations.

## Use Cases

- Technical documentation with diagrams
- RFC proposals with architecture diagrams
- Process flowcharts
- System architecture diagrams
- Data flow diagrams
- Slide presentations focused on diagrams

## License

MIT
