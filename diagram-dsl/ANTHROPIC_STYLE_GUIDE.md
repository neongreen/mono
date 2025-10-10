# Creating Anthropic-Style Diagrams

This guide shows you how to create professional AI system architecture diagrams similar to those in Anthropic's documentation, using the diagram-dsl library.

## Overview

Anthropic-style diagrams typically feature:
- **Clear visual hierarchy** - Grouped components with distinct sections
- **Layered architecture** - Input → Processing → Output flow
- **Professional styling** - Clean cards with consistent colors
- **Informative labels** - Multiple text layers showing detail
- **Clear data flow** - Color-coded arrows showing connections
- **Metadata badges** - Key metrics and status information

## Key Components

### 1. Cluster Component

The `Cluster` component is essential for creating clear visual groupings:

```tsx
import { Cluster } from 'diagram-dsl';

<Cluster title="Input Processing" variant="primary" width={350} padding={20}>
  {/* Your components here */}
</Cluster>
```

**Variants:**
- `primary` - Blue theme, for user-facing or input components
- `accent` - Orange theme, for core processing/model components  
- `success` - Green theme, for output or successful operations
- `secondary` - Purple/gray theme, for data storage or auxiliary systems
- `warning` - Orange theme, for safety or validation steps
- `danger` - Red theme, for errors or critical paths

### 2. Card Component

Cards are the building blocks for individual components:

```tsx
import { Card, Label, Subtitle } from 'diagram-dsl';

<Card id="my-component" variant="primary" width={300} height={75}>
  <Stack gap={6} alignItems="center">
    <Label bold size="lg">Component Name</Label>
    <Subtitle>Brief description</Subtitle>
    <Label size="sm">Additional detail</Label>
  </Stack>
</Card>
```

**Multi-layer text structure:**
- `Label bold size="lg"` - Main component name (prominent)
- `Subtitle` - Secondary description (medium)
- `Label size="sm"` - Tertiary details (subtle)

### 3. Arrow Component

Arrows show data flow and connections:

```tsx
import { Arrow } from 'diagram-dsl';

{/* Basic arrow */}
<Arrow from="component1" to="component2" label="data" color="#1976d2" thickness="medium" />

{/* Bidirectional arrow */}
<Arrow from="cache" to="database" color="#2196f3" style="dashed" bidirectional={true} />

{/* Curved arrow */}
<Arrow from="input" to="output" label="response" color="#4caf50" thickness="thick" curve="arc" />
```

**Arrow properties:**
- `thickness`: `thin`, `medium`, `thick`, `very-thick`
- `style`: `solid`, `dashed`, `dotted`, `wave`
- `curve`: `straight`, `curved`, `step`, `arc`
- `bidirectional`: `true` to show two-way communication

### 4. Badge Component

Badges for metadata and key information:

```tsx
import { Badge } from 'diagram-dsl';

<Badge text="200K context" variant="success" />
<Badge text="Constitutional AI" variant="info" />
```

## Layout Patterns

### Pattern 1: Three-Column Flow

Perfect for showing input → processing → output:

```tsx
<Row gap={40} justifyContent="center" alignItems="flex-start">
  <Cluster title="Input" variant="primary" width={350}>
    {/* Input components */}
  </Cluster>
  
  <Cluster title="Processing" variant="accent" width={380}>
    {/* Processing components */}
  </Cluster>
  
  <Cluster title="Output" variant="success" width={350}>
    {/* Output components */}
  </Cluster>
</Row>
```

### Pattern 2: Layered Architecture

Stack layers vertically with dividers:

```tsx
<Stack gap={40}>
  {/* User Layer */}
  <Stack gap={20}>
    <Badge text="User Layer" variant="info" />
    <Row gap={30} justifyContent="center">
      <Card id="web" variant="primary" width={220} height={100}>
        {/* Web app card */}
      </Card>
      <Card id="mobile" variant="primary" width={220} height={100}>
        {/* Mobile app card */}
      </Card>
    </Row>
  </Stack>
  
  <Divider width={1500} />
  
  {/* API Layer */}
  <Stack gap={20}>
    <Badge text="API Layer" variant="success" />
    {/* API components */}
  </Stack>
  
  <Divider width={1500} />
  
  {/* Data Layer */}
  <Cluster title="Data Storage" variant="secondary">
    {/* Database components */}
  </Cluster>
</Stack>
```

### Pattern 3: Data Layer

Shared data components at the bottom:

```tsx
<Cluster title="Data & Storage" variant="secondary" width={1320}>
  <Row gap={30} justifyContent="center">
    <Card id="vector-db" variant="secondary" width={250}>
      <Stack gap={6} alignItems="center">
        <Label bold size="lg">Vector Database</Label>
        <Subtitle>Embeddings storage</Subtitle>
      </Stack>
    </Card>
    {/* More data stores */}
  </Row>
</Cluster>
```

## Color Coding

Use consistent colors for different types of connections:

- **Primary flow** - Blue (`#1976d2`, `#2196f3`) - Main data path
- **Success** - Green (`#4caf50`, `#66bb6a`) - Successful operations
- **Processing** - Orange (`#ff9800`, `#ffa726`) - Internal processing
- **Model operations** - Purple (`#ab47bc`, `#7b1fa2`) - AI/ML specific
- **Metadata** - Gray (`#9e9e9e`, `#607d8b`) - Logging, monitoring

## Complete Example

Here's a complete example showing all patterns:

```tsx
import React from 'react';
import {
  Stack, Row, Card, Title, Subtitle, Label, Arrow, Cluster,
  Badge, Divider, renderToSVG
} from 'diagram-dsl';

const AISystemDiagram = () => (
  <Stack width={1400} height={1100} padding={40} gap={35}>
    {/* Title */}
    <Stack gap={10} alignItems="center">
      <Title level={1}>AI System Architecture</Title>
      <Subtitle size="base">End-to-end conversation flow</Subtitle>
    </Stack>

    {/* Main Processing Flow */}
    <Row gap={40} justifyContent="center" alignItems="flex-start">
      {/* Input */}
      <Cluster title="Input Processing" variant="primary" width={350} padding={20}>
        <Stack gap={20} alignItems="center">
          <Card id="user-input" variant="primary" width={300} height={75}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">User Message</Label>
              <Subtitle>Text input</Subtitle>
            </Stack>
          </Card>
          <Card id="safety" variant="warning" width={300} height={75}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">Safety Check</Label>
              <Subtitle>Content moderation</Subtitle>
            </Stack>
          </Card>
        </Stack>
      </Cluster>

      {/* Processing */}
      <Cluster title="Model Processing" variant="accent" width={380} padding={20}>
        <Stack gap={20} alignItems="center">
          <Card id="model" variant="accent" width={330} height={110}>
            <Stack gap={10} alignItems="center">
              <Label bold size="lg">AI Model</Label>
              <Subtitle>Neural inference</Subtitle>
              <Badge text="200K tokens" variant="success" />
            </Stack>
          </Card>
        </Stack>
      </Cluster>

      {/* Output */}
      <Cluster title="Response" variant="success" width={350} padding={20}>
        <Stack gap={20} alignItems="center">
          <Card id="response" variant="success" width={300} height={75}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">Response</Label>
              <Subtitle>Formatted output</Subtitle>
            </Stack>
          </Card>
        </Stack>
      </Cluster>
    </Row>

    <Divider width={1320} />

    {/* Data Layer */}
    <Cluster title="Data Storage" variant="secondary" width={1320} padding={20}>
      <Row gap={30} justifyContent="center">
        <Card id="database" variant="secondary" width={250} height={80}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Database</Label>
            <Subtitle>Chat history</Subtitle>
          </Stack>
        </Card>
      </Row>
    </Cluster>

    {/* Arrows */}
    <Arrow from="user-input" to="safety" label="text" color="#1976d2" thickness="medium" />
    <Arrow from="safety" to="model" label="validated" color="#66bb6a" thickness="medium" />
    <Arrow from="model" to="response" label="output" color="#ab47bc" thickness="thick" />
    <Arrow from="response" to="database" label="save" color="#9e9e9e" style="dashed" />

    {/* Legend */}
    <Row gap={25} marginTop={20} justifyContent="center">
      <Badge text="→ Primary flow" variant="primary" />
      <Badge text="⋯ Async operations" variant="secondary" />
    </Row>
  </Stack>
);

// Render
(async () => {
  const svg = await renderToSVG(<AISystemDiagram />, {
    width: 1400,
    height: 1100,
    backgroundColor: 'white'
  });
  // Save or use the SVG
})();
```

## Best Practices

### 1. Consistent Sizing

- **Small cards**: 200-250px width, 70-80px height
- **Medium cards**: 280-330px width, 85-100px height  
- **Large cards**: 350-400px width, 110-140px height
- **Clusters**: 320-380px width for columns, 1200-1500px for full-width

### 2. Gap Spacing

- **Tight spacing** (8-12px): Within cards
- **Normal spacing** (20-30px): Between cards in same cluster
- **Relaxed spacing** (40-50px): Between clusters

### 3. Text Hierarchy

Always use 2-3 levels of text in cards:
1. **Primary**: `<Label bold size="lg">` - What it is
2. **Secondary**: `<Subtitle>` - What it does
3. **Tertiary**: `<Label size="sm">` - Technical details

### 4. Arrow Organization

- Draw arrows after all components are defined
- Group arrows by flow: forward flow, data access, monitoring
- Use comments to separate arrow groups

### 5. Color Consistency

- Use the same color for the same type of operation across diagrams
- Match cluster variants to the card variants inside them
- Reserve gray/dashed arrows for monitoring and logging

## Common Pitfalls

❌ **Don't**: Put too many cards in one cluster (max 4-5)
✅ **Do**: Split into multiple clusters or use sub-groups

❌ **Don't**: Use random colors for arrows
✅ **Do**: Use a consistent color scheme

❌ **Don't**: Overcrowd cards with too much text
✅ **Do**: Keep text concise, use 2-3 levels max

❌ **Don't**: Create overly complex arrow routing
✅ **Do**: Use curve types and bidirectional arrows to simplify

## Examples in This Repository

- `examples/anthropic-style-diagram.tsx` - Full layered architecture
- `examples/anthropic-improved.tsx` - Three-column flow with clusters
- `examples/showcase-agent-system.tsx` - Complex agent system

## Quick Start Template

```tsx
import React from 'react';
import { Stack, Row, Card, Title, Subtitle, Label, Arrow, Cluster, Badge, Divider, renderToSVG } from 'diagram-dsl';

const MyDiagram = () => (
  <Stack width={1400} height={1000} padding={40} gap={35}>
    <Stack gap={10} alignItems="center">
      <Title level={1}>Your Diagram Title</Title>
      <Subtitle>Subtitle here</Subtitle>
    </Stack>

    <Row gap={40} justifyContent="center" alignItems="flex-start">
      <Cluster title="Section 1" variant="primary" width={350} padding={20}>
        {/* Add your cards here */}
      </Cluster>
      
      <Cluster title="Section 2" variant="accent" width={380} padding={20}>
        {/* Add your cards here */}
      </Cluster>
    </Row>

    {/* Add your arrows here */}
    
    <Row gap={25} marginTop={20} justifyContent="center">
      {/* Add legend badges here */}
    </Row>
  </Stack>
);

(async () => {
  const svg = await renderToSVG(<MyDiagram />, { width: 1400, height: 1000 });
  // Use the SVG
})();
```

## Tips for Large Diagrams

For diagrams with many components:

1. **Break into stages**: Use multiple Cluster components
2. **Use sub-groups**: Group related cards with Row/Stack inside clusters
3. **Simplify arrows**: Show only critical paths, add "..." for implied connections
4. **Add legends**: Use badges to explain arrow types and colors
5. **Version iterations**: Start simple, add complexity gradually

## Conclusion

The diagram-dsl library makes it easy to create Anthropic-style diagrams by:
- Providing high-level components (Cluster, Card, Badge)
- Supporting professional styling out of the box
- Offering flexible arrow routing with multiple styles
- Enabling clear visual hierarchy through consistent design patterns

Happy diagramming! 🎨
