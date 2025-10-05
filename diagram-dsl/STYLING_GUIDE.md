# Styling Guide for diagram-dsl

This guide shows how to use the new semantic components and theme system to create beautiful diagrams with minimal code.

## Philosophy

The new semantic components provide **sensible defaults** so you can focus on content rather than styling. Instead of manually specifying colors, fonts, sizes, and spacing for every element, you can use components that already look good.

## Semantic Components

### Title

Use `Title` for diagram titles and section headings. Supports 3 levels of hierarchy.

```tsx
import { Title } from 'diagram-dsl';

// Level 1: Main diagram title (36px, bold)
<Title level={1}>My Diagram</Title>

// Level 2: Section heading (24px, bold)
<Title level={2}>Section Name</Title>

// Level 3: Subsection (20px, bold)
<Title level={3}>Subsection</Title>
```

**Default styling:**
- Centered text alignment
- Bold font weight
- Professional dark gray color
- Appropriate font size for hierarchy

### Subtitle

Use `Subtitle` for secondary information, descriptions, or smaller labels.

```tsx
import { Subtitle } from 'diagram-dsl';

// Small subtitle (12px, gray)
<Subtitle>Additional context</Subtitle>

// Base size subtitle (14px, gray)
<Subtitle size="base">Larger description</Subtitle>
```

**Default styling:**
- Centered text alignment
- Regular font weight
- Secondary gray color (#757575)
- Smaller font size

### Label

Use `Label` for body text and standard labels in your diagrams.

```tsx
import { Label } from 'diagram-dsl';

// Regular label (14px)
<Label>Normal text</Label>

// Small label (12px)
<Label size="sm">Compact text</Label>

// Large label (16px)
<Label size="lg">Prominent text</Label>

// Bold label
<Label bold>Important text</Label>
```

**Default styling:**
- Centered text alignment
- Appropriate font size
- Primary text color
- Optional bold weight

### Card

Use `Card` for boxes/containers with professional styling. Supports color-coded variants.

```tsx
import { Card } from 'diagram-dsl';

// Primary blue card
<Card variant="primary" width={200} height={100}>
  <Label>Content</Label>
</Card>

// Success green card
<Card variant="success" width={200} height={100}>
  <Label>Success</Label>
</Card>
```

**Available variants:**
- `primary` - Blue (default for main elements)
- `secondary` - Purple (for secondary elements)
- `success` - Green (for positive states)
- `warning` - Orange (for attention items)
- `error` - Red (for error states)
- `info` - Cyan (for informational items)
- `default` - Light gray (neutral)

**Default styling:**
- 16px padding (comfortable spacing)
- 8px border radius (modern rounded corners)
- 2px border width
- Centered content
- Color-coordinated background and border

## Rich Text in Boxes

One of the key improvements is the ability to easily combine different text styles within a box:

```tsx
import { Card, Stack, Label, Subtitle } from 'diagram-dsl';

<Card variant="primary" width={200} height={100}>
  <Stack gap={6} alignItems="center">
    <Label bold size="lg">Main Title</Label>
    <Subtitle>Supporting text</Subtitle>
  </Stack>
</Card>
```

This creates a card with:
- A prominent label on top
- A smaller, gray subtitle below
- Proper spacing between them (6px)
- All content centered

## Theme System

All semantic components use values from a centralized theme:

```tsx
import { theme } from 'diagram-dsl';

// Access theme values directly if needed
const myColor = theme.colors.primary.main;  // '#2196f3'
const mySpacing = theme.spacing.lg;         // 16
const myFontSize = theme.typography.fontSize.xl;  // 20
```

### Color Palette

```typescript
colors: {
  primary: { main: '#2196f3', light: '#e3f2fd', dark: '#1976d2' },
  secondary: { main: '#9c27b0', light: '#f3e5f5', dark: '#7b1fa2' },
  success: { main: '#4caf50', light: '#e8f5e9', dark: '#388e3c' },
  warning: { main: '#ff9800', light: '#fff3e0', dark: '#f57c00' },
  error: { main: '#f44336', light: '#ffebee', dark: '#d32f2f' },
  info: { main: '#00bcd4', light: '#e0f7fa', dark: '#0097a7' },
  gray: { 50-900 },
  text: { primary: '#212121', secondary: '#757575', disabled: '#bdbdbd' }
}
```

### Typography Scale

```typescript
typography: {
  fontSize: {
    xs: 10, sm: 12, base: 14, lg: 16, xl: 20,
    '2xl': 24, '3xl': 30, '4xl': 36, '5xl': 48
  }
}
```

### Spacing Scale

Based on a 4px grid for consistency:

```typescript
spacing: {
  xs: 4, sm: 8, md: 12, lg: 16, xl: 20,
  '2xl': 24, '3xl': 32, '4xl': 40, '5xl': 48
}
```

## Before and After Comparison

### Before: Manual Styling (verbose)

```tsx
<Box
  id="start"
  width={200}
  height={80}
  backgroundColor="#e0f7fa"
  borderColor="#00acc1"
  borderWidth={2}
  borderRadius={8}
  padding={20}
  justifyContent="center"
  alignItems="center"
>
  <Text fontSize={18}>Start Process</Text>
</Box>
```

### After: Semantic Components (concise)

```tsx
<Card
  id="start"
  variant="info"
  width={200}
  height={80}
>
  <Label size="lg">Start Process</Label>
</Card>
```

### With Rich Text (before)

```tsx
<Box
  width={200}
  height={100}
  backgroundColor="#e3f2fd"
  borderColor="#1976d2"
  borderWidth={2}
  borderRadius={8}
  padding={15}
  justifyContent="center"
  alignItems="center"
>
  <Text fontSize={16} fontWeight="bold">API Gateway</Text>
  <Text fontSize={12}>REST/GraphQL</Text>
</Box>
```

### With Rich Text (after)

```tsx
<Card variant="primary" width={200} height={100}>
  <Stack gap={6} alignItems="center">
    <Label bold size="lg">API Gateway</Label>
    <Subtitle>REST/GraphQL</Subtitle>
  </Stack>
</Card>
```

## Best Practices

### 1. Use Title Hierarchy

```tsx
<Stack gap={32}>
  <Title level={1}>Main Diagram Title</Title>
  
  <Stack gap={12}>
    <Title level={3}>Section 1</Title>
    {/* content */}
  </Stack>
  
  <Stack gap={12}>
    <Title level={3}>Section 2</Title>
    {/* content */}
  </Stack>
</Stack>
```

### 2. Inner and Outer Spacing

Use consistent spacing that follows the rule: **elements within a group should be closer to each other than to elements outside the group**.

```tsx
<Stack gap={32}>  {/* Large gap between major sections */}
  <Title level={2}>Section Title</Title>
  
  <Stack gap={12}>  {/* Medium gap within section */}
    <Card>
      <Stack gap={6}>  {/* Small gap within card */}
        <Label bold>Item 1</Label>
        <Subtitle>Description</Subtitle>
      </Stack>
    </Card>
    
    <Card>
      <Stack gap={6}>
        <Label bold>Item 2</Label>
        <Subtitle>Description</Subtitle>
      </Stack>
    </Card>
  </Stack>
</Stack>
```

### 3. Prevent Text Overlap

The semantic components use proper padding and spacing by default:
- Cards have 16px padding by default
- Stack gaps provide breathing room
- Text uses appropriate line height

### 4. Consistent Alignment

```tsx
// Center-align content in cards (default)
<Card variant="primary">
  <Label>Centered</Label>
</Card>

// Left-align for lists or detailed content
<Card variant="default" alignItems="flex-start">
  <Stack gap={4} alignItems="flex-start">
    <Label>Item 1</Label>
    <Label>Item 2</Label>
    <Label>Item 3</Label>
  </Stack>
</Card>
```

### 5. Color Coding by Function

Use semantic color variants to indicate purpose:

```tsx
// User interface elements
<Card variant="primary">Frontend</Card>

// Backend services  
<Card variant="success">API Server</Card>

// Data storage
<Card variant="secondary">Database</Card>

// Warnings or actions needed
<Card variant="warning">Review Required</Card>

// Errors or problems
<Card variant="error">Failed</Card>

// Information
<Card variant="info">Help Text</Card>
```

## Complete Example

Here's a complete example showing best practices:

```tsx
import React from 'react';
import { Stack, Row, Card, Title, Subtitle, Label, Arrow } from 'diagram-dsl';

const MyDiagram = () => (
  <Stack gap={32} padding={40}>
    {/* Title section with good spacing */}
    <Stack gap={8} alignItems="center">
      <Title level={1}>System Architecture</Title>
      <Subtitle size="base">Microservices deployment overview</Subtitle>
    </Stack>
    
    {/* Frontend tier */}
    <Stack gap={12}>
      <Title level={3}>Frontend</Title>
      <Row gap={16} justifyContent="center">
        <Card id="web" variant="primary" width={180} height={90}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Web App</Label>
            <Subtitle>React</Subtitle>
          </Stack>
        </Card>
        
        <Card id="mobile" variant="primary" width={180} height={90}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Mobile</Label>
            <Subtitle>React Native</Subtitle>
          </Stack>
        </Card>
      </Row>
    </Stack>
    
    {/* Backend tier */}
    <Stack gap={12}>
      <Title level={3}>Backend</Title>
      <Card id="api" variant="success" width={180} height={90}>
        <Stack gap={6} alignItems="center">
          <Label bold size="lg">API Gateway</Label>
          <Subtitle>GraphQL</Subtitle>
        </Stack>
      </Card>
    </Stack>
    
    <Arrow from="web" to="api" color="#1976d2" strokeWidth={2} />
    <Arrow from="mobile" to="api" color="#1976d2" strokeWidth={2} />
  </Stack>
);
```

## Examples

See these generated examples:
- `examples/styled-flowchart.svg` - Simple flowchart with semantic components
- `examples/styled-architecture.svg` - Three-tier architecture
- `examples/title-hierarchy.svg` - Typography showcase

Run the examples:
```bash
pnpm dev:styled
```
