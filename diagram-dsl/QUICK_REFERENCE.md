# Quick Reference: Anthropic-Style Diagrams

## Template (Copy & Paste)

```tsx
import React from 'react';
import {
  Stack, Row, Card, Title, Subtitle, Label, Arrow, Cluster,
  Badge, Divider, renderToSVG
} from 'diagram-dsl';

const MyDiagram = () => (
  <Stack width={1200} height={900} padding={40} gap={30}>
    <Stack gap={8} alignItems="center">
      <Title level={1}>Your Title Here</Title>
      <Subtitle>Your subtitle</Subtitle>
    </Stack>

    <Row gap={35} justifyContent="center">
      <Cluster title="Section 1" variant="primary" width={320} padding={20}>
        <Card id="comp1" variant="primary" width={280} height={80}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Component</Label>
            <Subtitle>Description</Subtitle>
          </Stack>
        </Card>
      </Cluster>

      <Cluster title="Section 2" variant="accent" width={320} padding={20}>
        <Card id="comp2" variant="accent" width={280} height={80}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Component</Label>
            <Subtitle>Description</Subtitle>
          </Stack>
        </Card>
      </Cluster>
    </Row>

    <Arrow from="comp1" to="comp2" label="data" color="#1976d2" thickness="medium" />
  </Stack>
);

(async () => {
  const svg = await renderToSVG(<MyDiagram />, { width: 1200, height: 900 });
  // Use svg...
})();
```

## Component Quick Reference

### Cluster (Visual Grouping)
```tsx
<Cluster title="Title" variant="primary" width={350} padding={20}>
  {/* Cards go here */}
</Cluster>
```

**Variants:** `primary` (blue) | `accent` (orange) | `success` (green) | `secondary` (gray) | `warning` (yellow) | `danger` (red)

**Sizes:** Width 280-400px, padding 20px

### Card (Individual Component)
```tsx
<Card id="unique-id" variant="primary" width={280} height={80}>
  <Stack gap={6} alignItems="center">
    <Label bold size="lg">Name</Label>
    <Subtitle>What it does</Subtitle>
    <Label size="sm">Details</Label>
  </Stack>
</Card>
```

**Variants:** Same as Cluster
**Sizes:** 
- Small: 200×70px
- Medium: 280×80px
- Large: 330×110px

### Typography
```tsx
<Title level={1}>Main Title</Title>          // 36px, bold
<Title level={2}>Section Title</Title>       // 24px, bold
<Title level={3}>Subsection</Title>          // 20px, bold
<Subtitle>Secondary text</Subtitle>          // 14px, gray
<Label bold size="lg">Large label</Label>    // 16px, bold
<Label>Normal label</Label>                  // 14px
<Label size="sm">Small label</Label>         // 12px
```

### Arrow (Connections)
```tsx
// Basic
<Arrow from="id1" to="id2" label="text" color="#1976d2" thickness="medium" />

// Advanced
<Arrow from="id1" to="id2" 
       label="text" 
       color="#2196f3" 
       thickness="thick"
       style="dashed"
       curve="arc"
       bidirectional={true} />
```

**Thickness:** `thin` | `medium` | `thick` | `very-thick`
**Style:** `solid` | `dashed` | `dotted`
**Curve:** `straight` | `curved` | `step` | `arc`

### Badge (Metadata)
```tsx
<Badge text="Info" variant="primary" />
<Badge text="Success" variant="success" />
<Badge text="Warning" variant="warning" />
```

### Layout
```tsx
<Stack gap={20}>           // Vertical stack
<Row gap={20}>             // Horizontal row
<Divider width={1200} />   // Horizontal line
```

## Color Palette

| Purpose | Color | Hex |
|---------|-------|-----|
| Primary flow | Blue | `#1976d2`, `#2196f3` |
| Success | Green | `#4caf50`, `#66bb6a` |
| Processing | Orange | `#ff9800`, `#ffa726` |
| Model | Purple | `#ab47bc`, `#7b1fa2` |
| Metadata | Gray | `#9e9e9e`, `#607d8b` |

## Common Patterns

### Pattern 1: Three-Column Flow
```tsx
<Row gap={35} justifyContent="center">
  <Cluster title="Input" variant="primary" width={320}>
    {/* Cards */}
  </Cluster>
  <Cluster title="Processing" variant="accent" width={320}>
    {/* Cards */}
  </Cluster>
  <Cluster title="Output" variant="success" width={320}>
    {/* Cards */}
  </Cluster>
</Row>
```

### Pattern 2: Layered Architecture
```tsx
<Stack gap={35}>
  <Badge text="Layer 1" variant="info" />
  <Row gap={30}>{/* Layer 1 cards */}</Row>
  <Divider width={1200} />
  <Badge text="Layer 2" variant="success" />
  <Row gap={30}>{/* Layer 2 cards */}</Row>
</Stack>
```

### Pattern 3: Data Layer
```tsx
<Cluster title="Data Storage" variant="secondary" width={1200} padding={20}>
  <Row gap={30} justifyContent="center">
    <Card id="db1" variant="secondary" width={250} height={80}>
      {/* Database card */}
    </Card>
    <Card id="db2" variant="secondary" width={250} height={80}>
      {/* Cache card */}
    </Card>
  </Row>
</Cluster>
```

## Sizing Guidelines

### Canvas
- Simple: 1000×700px
- Medium: 1200×900px
- Complex: 1400×1100px
- Full: 1600×1200px

### Gaps
- Tight: 8-12px (within cards)
- Normal: 20-30px (between cards)
- Relaxed: 35-50px (between clusters)

### Clusters
- Column: 280-380px wide
- Full-width: 1200-1500px wide
- Padding: 20px

### Cards
- Width: 200-330px
- Height: 70-140px
- Small: 200×70
- Medium: 280×85
- Large: 330×110

## Arrow Organization

Group arrows by type:
```tsx
{/* Forward flow */}
<Arrow from="a" to="b" label="main" color="#1976d2" thickness="medium" />
<Arrow from="b" to="c" label="process" color="#4caf50" thickness="thick" />

{/* Data access */}
<Arrow from="b" to="db" color="#2196f3" style="dashed" bidirectional={true} />

{/* Monitoring */}
<Arrow from="c" to="logs" color="#9e9e9e" style="dashed" />
```

## Common Pitfalls

❌ Too many cards in one cluster (max 4-5)
❌ Random arrow colors
❌ Inconsistent sizing
❌ Too much text in cards
❌ Missing IDs on cards (needed for arrows)

✅ Split into multiple clusters
✅ Use consistent color scheme
✅ Follow sizing guidelines
✅ Keep text concise (2-3 levels max)
✅ Always set ID when arrows point to it

## Examples

See full examples in:
- `examples/anthropic-simple.tsx` - Start here!
- `examples/anthropic-improved.tsx` - Advanced
- `examples/anthropic-style-diagram.tsx` - Full featured

## Full Guide

See `ANTHROPIC_STYLE_GUIDE.md` for complete documentation.
