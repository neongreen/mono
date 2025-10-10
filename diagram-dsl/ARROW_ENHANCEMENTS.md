# Arrow Enhancements for Complex Diagrams

## Overview

This document describes the new arrow capabilities added to support complex diagram structures like the Anthropic "Prompt engineering vs. context engineering" diagram.

## New Arrow Properties

### 1. Custom Attachment Points

**`fromSide` and `toSide`**
Specify exactly which side of a box the arrow should connect to:
- `'top'` - Connect to top edge
- `'bottom'` - Connect to bottom edge
- `'left'` - Connect to left edge
- `'right'` - Connect to right edge
- `'auto'` (default) - Auto-detect closest sides

```tsx
<Arrow from="box1" to="box2" fromSide="right" toSide="left" />
```

### 2. Offset from Center

**`fromOffset` and `toOffset`**
Offset the connection point along the edge (useful for corners or specific points):

```tsx
// Connect from top-right corner of box1 to bottom-left corner of box2
<Arrow 
  from="box1" 
  to="box2" 
  fromSide="top" 
  fromOffset={50}  // 50px to the right of center
  toSide="bottom" 
  toOffset={-50}   // 50px to the left of center
/>
```

### 3. Arrow Shortening

**`shortenStart` and `shortenEnd`**
Shorten arrows from either end (useful for arrows that don't quite reach their target):

```tsx
// Arrow that stops 20px short of the target
<Arrow from="box1" to="box2" shortenEnd={20} />

// Arrow that starts 10px away from the source
<Arrow from="box1" to="box2" shortenStart={10} />
```

### 4. Multi-Target Arrows (Y-shaped Forks)

**`toMultiple`**
Create forked arrows that split to multiple destinations:

```tsx
// Single arrow that forks to two targets
<Arrow from="model" toMultiple={["assistant", "tool"]} color="#4c4c4c" />
```

## Use Cases

### Corner-to-Corner Connections

```tsx
<Box id="source" width={200} height={100} />
<Box id="target" width={200} height={100} />

{/* Connect bottom-right corner to top-left corner */}
<Arrow 
  from="source" 
  to="target" 
  fromSide="bottom" 
  fromOffset={100}  // Half of width
  toSide="top" 
  toOffset={-100}   // Negative half of width
/>
```

### Arrows That Don't Reach

```tsx
{/* "Curation" arrow that stops before entering the box */}
<Arrow 
  from="possible-context" 
  to="context-window" 
  label="Curation"
  shortenEnd={30}
/>
```

### Y-Shaped Decision Points

```tsx
<Box id="decision" width={100} height={60} />
<Box id="path-a" width={100} height={60} />
<Box id="path-b" width={100} height={60} />

{/* Fork from decision to both paths */}
<Arrow from="decision" toMultiple={["path-a", "path-b"]} />
```

### Feedback Loops

```tsx
{/* Vertical arrow down */}
<Arrow from="tool-call" to="tool-result" fromSide="bottom" toSide="top" />

{/* Horizontal arrow back */}
<Arrow from="tool-result" to="history" fromSide="left" toSide="bottom" curve="step" />
```

## Complete Example

```tsx
import { Stack, Row, Box, Text, Arrow } from 'diagram-dsl';

const ComplexDiagram = () => (
  <Stack width={1000} height={600} padding={40} gap={30}>
    <Row gap={50}>
      <Box id="source" width={200} height={100} backgroundColor="#e3f2fd">
        <Text>Source Box</Text>
      </Box>
      
      <Box id="middle" width={150} height={80} backgroundColor="#fff3e0">
        <Text>Decision</Text>
      </Box>
      
      <Stack gap={20}>
        <Box id="target-a" width={180} height={70} backgroundColor="#e8f5e9">
          <Text>Target A</Text>
        </Box>
        <Box id="target-b" width={180} height={70} backgroundColor="#e8f5e9">
          <Text>Target B</Text>
        </Box>
      </Stack>
    </Row>

    {/* Arrow with custom attachment */}
    <Arrow 
      from="source" 
      to="middle" 
      fromSide="right" 
      toSide="left"
      label="process"
    />

    {/* Y-shaped fork to multiple targets */}
    <Arrow 
      from="middle" 
      toMultiple={["target-a", "target-b"]}
      fromSide="right"
      color="#666"
    />

    {/* Arrow that stops short */}
    <Arrow 
      from="target-b" 
      to="source" 
      fromSide="left" 
      toSide="bottom"
      shortenEnd={20}
      style="dashed"
      label="feedback"
    />
  </Stack>
);
```

## Implementation Notes

### Attachment Point Calculation

The renderer now includes a helper function that calculates attachment points with offsets:

```typescript
const getAttachmentPoint = (pos: any, side: string, offset: number = 0) => {
  const centerX = pos.x + pos.width / 2;
  const centerY = pos.y + pos.height / 2;
  
  switch (side) {
    case 'top':
      return { x: centerX + offset, y: pos.y };
    case 'bottom':
      return { x: centerX + offset, y: pos.y + pos.height };
    case 'left':
      return { x: pos.x, y: centerY + offset };
    case 'right':
      return { x: pos.x + pos.width, y: centerY + offset };
    default:
      return { x: centerX, y: centerY };
  }
};
```

### Arrow Shortening

Arrows are shortened by calculating the direction vector and moving the endpoints:

```typescript
if (node.props.shortenStart || node.props.shortenEnd) {
  const dx = x2 - x1;
  const dy = y2 - y1;
  const length = Math.sqrt(dx * dx + dy * dy);
  const dirX = dx / length;
  const dirY = dy / length;

  if (node.props.shortenStart) {
    x1 += dirX * node.props.shortenStart;
    y1 += dirY * node.props.shortenStart;
  }

  if (node.props.shortenEnd) {
    x2 -= dirX * node.props.shortenEnd;
    y2 -= dirY * node.props.shortenEnd;
  }
}
```

## Benefits

1. **Precise Control**: Specify exactly where arrows connect
2. **Complex Routing**: Support corner-to-corner and diagonal connections
3. **Visual Clarity**: Arrows that stop short improve readability
4. **Decision Trees**: Y-shaped forks for branching logic
5. **Feedback Loops**: Better support for circular data flow

## Backwards Compatibility

All new properties are optional. Existing diagrams will continue to work with automatic edge detection. The `'auto'` value for `fromSide` and `toSide` provides the same behavior as before.

## See Also

- `examples/anthropic-original-replication.tsx` - Full replication using these features
- `ANTHROPIC_STYLE_GUIDE.md` - General Anthropic-style diagram patterns
- `QUICK_REFERENCE.md` - Quick reference for all arrow properties
