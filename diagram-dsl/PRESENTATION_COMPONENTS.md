# Presentation Components Guide

This guide covers the presentation helper components in diagram-dsl, designed to make creating slides easier and more consistent.

## Core Concepts

All presentation components build on top of the base layout components (Box, Stack, Row, etc.) and provide:
- Consistent styling and spacing
- Variant-based theming
- Less boilerplate code
- Semantic meaning

## Components

### Slide

A container for presentation slides with standard dimensions.

```tsx
<Slide>
  <Title level={1}>My Slide Title</Title>
  {/* slide content */}
</Slide>
```

**Props:**
- All BoxProps
- `width` (default: 1200)
- `height` (default: 800)
- `padding` (default: 60)
- `gap` (default: 32)
- `backgroundColor` (default: 'white')

### List

Renders bullet point lists with customizable bullets.

```tsx
<List
  items={[
    'First item',
    'Second item',
    'Third item'
  ]}
  bullet="•"
  fontSize={14}
/>
```

**Props:**
- `items`: Array of strings or ReactNodes
- `bullet` (default: '•')
- `fontSize` (default: 14)
- `color`: Text color
- `gap` (default: 8)
- All BoxProps

### ProsCons

Displays pros and cons in a side-by-side layout.

```tsx
<ProsCons
  pros={['Fast', 'Simple', 'Reliable']}
  cons={['Limited features', 'Higher cost']}
/>
```

**Props:**
- `pros`: Array of strings
- `cons`: Array of strings
- `prosTitle` (default: '✓ Pros')
- `consTitle` (default: '✗ Cons')
- `fontSize` (default: 11)
- `gap` (default: 32)

### Section

A titled container with variant-based styling.

```tsx
<Section title="Important Information" variant="success">
  <Text>Content goes here</Text>
</Section>
```

**Props:**
- `title`: Section title
- `children`: Content
- `titleSize` (default: 14)
- `titleColor`: Override title color
- `variant`: 'default' | 'primary' | 'secondary' | 'accent' | 'success' | 'warning' | 'danger'
- `padding` (default: 20)
- `borderRadius` (default: 8)
- `borderWidth` (default: 2)
- All BoxProps

**Variants:**
- `default`: Gray background
- `primary`: Blue background
- `secondary`: Purple background
- `accent`: Orange background
- `success`: Green background
- `warning`: Orange background
- `danger`: Red background

### Callout

Highlights important information with an icon and title.

```tsx
<Callout title="Key Takeaway" variant="success">
  <Text>This is important information.</Text>
</Callout>
```

**Props:**
- `title`: Optional title
- `children`: Content
- `variant`: 'default' | 'info' | 'success' | 'warning' | 'danger'
- `icon`: Custom icon (overrides default)
- `padding` (default: 24)
- `borderRadius` (default: 8)
- `borderWidth` (default: 3)
- All BoxProps

**Default Icons:**
- `default`: 📌
- `info`: ℹ️
- `success`: ✓
- `warning`: ⚠️
- `danger`: ✗

### Highlight

A simple highlighted content box.

```tsx
<Highlight variant="warning">
  <Text>Highlighted text</Text>
</Highlight>
```

**Props:**
- `children`: Content
- `color`: Custom color
- `variant`: 'info' | 'success' | 'warning' | 'danger'
- `padding` (default: 12)
- `borderRadius` (default: 6)
- All BoxProps

### RichText

Renders text with mixed formatting in a single line.

```tsx
<RichText
  segments={[
    'This is ',
    { text: 'bold text', bold: true },
    ' and ',
    { text: 'colored', color: '#ff0000' }
  ]}
/>
```

**Props:**
- `segments`: Array of strings or TextSegments
- `fontSize` (default: 14)
- `color` (default: 'black')
- `gap` (default: 4)

**TextSegment:**
- `text`: The text content
- `bold`: Make text bold
- `color`: Text color
- `fontSize`: Override fontSize

### Spacer

Adds space between elements.

```tsx
<Spacer size={20} />
{/* or flexible spacer */}
<Spacer flexible />
```

**Props:**
- `size` (default: 20): Fixed size in pixels
- `flexible` (default: false): If true, uses flexGrow instead

### Grid

Arranges children in a grid layout.

```tsx
<Grid columns={3} gap={16}>
  <Card>Item 1</Card>
  <Card>Item 2</Card>
  <Card>Item 3</Card>
</Grid>
```

**Props:**
- `children`: Grid items
- `columns` (default: 2): Number of columns
- `gap` (default: 16): Space between items
- All BoxProps

## Usage Examples

### Simple Slide with List

```tsx
<Slide>
  <Title level={1}>My Topic</Title>
  <Subtitle>A brief introduction</Subtitle>
  
  <Section title="Key Points" variant="primary" width={800} marginTop={20}>
    <List
      items={[
        'First key point',
        'Second key point',
        'Third key point'
      ]}
    />
  </Section>
</Slide>
```

### Slide with Pros/Cons

```tsx
<Slide>
  <Title level={1}>Decision Analysis</Title>
  
  <Card variant="default" width={900} marginTop={20}>
    <ProsCons
      pros={['Benefit 1', 'Benefit 2', 'Benefit 3']}
      cons={['Drawback 1', 'Drawback 2']}
    />
  </Card>
</Slide>
```

### Slide with Grid Layout

```tsx
<Slide>
  <Title level={1}>Feature Comparison</Title>
  
  <Grid columns={3} gap={20} marginTop={20}>
    <Section title="Feature 1" variant="success">
      <Text>Description</Text>
    </Section>
    <Section title="Feature 2" variant="primary">
      <Text>Description</Text>
    </Section>
    <Section title="Feature 3" variant="accent">
      <Text>Description</Text>
    </Section>
  </Grid>
</Slide>
```

### Slide with Callouts

```tsx
<Slide>
  <Title level={1}>Important Information</Title>
  
  <Callout title="Success!" variant="success" width={900} marginTop={20}>
    <Text>Everything is working correctly.</Text>
  </Callout>
  
  <Spacer size={20} />
  
  <Callout title="Watch Out" variant="warning" width={900}>
    <List items={['Point 1', 'Point 2']} />
  </Callout>
</Slide>
```

### Complex Layout Example

```tsx
<Slide>
  <Title level={1}>Comprehensive Example</Title>
  <Subtitle>Using multiple components together</Subtitle>
  
  <Section title="Overview" variant="primary" width={1080} marginTop={16}>
    <RichText
      segments={[
        'This demonstrates ',
        { text: 'multiple components', bold: true },
        ' working together seamlessly.'
      ]}
    />
  </Section>
  
  <Spacer size={20} />
  
  <Grid columns={2} gap={16}>
    <Callout title="Benefits" variant="success" width={520}>
      <List items={['Easy to use', 'Consistent', 'Flexible']} fontSize={11} />
    </Callout>
    
    <Callout title="Usage" variant="info" width={520}>
      <List items={['Import components', 'Compose layout', 'Generate SVG']} fontSize={11} />
    </Callout>
  </Grid>
</Slide>
```

## Best Practices

1. **Use Slide as your root container** for consistent dimensions
2. **Leverage variants** for consistent theming across slides
3. **Use Spacer** instead of manual margins for flexible spacing
4. **Combine Grid with Section/Card** for clean multi-column layouts
5. **Use RichText sparingly** - prefer separate Text elements when possible
6. **Keep font sizes consistent** within similar content types
7. **Use Callout for emphasis** on important information
8. **Use ProsCons** for any comparison or trade-off discussion

## Migration from Manual Layout

### Before:
```tsx
<Stack gap={32} padding={60} width={1200} height={800}>
  <Text fontSize={32} fontWeight="bold">Title</Text>
  <Text fontSize={20} color="#666">Subtitle</Text>
  
  <Box
    backgroundColor="#e3f2fd"
    borderColor="#1976d2"
    borderWidth={2}
    borderRadius={8}
    padding={20}
    width={900}
  >
    <Text fontSize={14} fontWeight="bold" marginBottom={12}>Section Title</Text>
    <Text fontSize={12}>• Item 1</Text>
    <Text fontSize={12}>• Item 2</Text>
    <Text fontSize={12}>• Item 3</Text>
  </Box>
</Stack>
```

### After:
```tsx
<Slide>
  <Title level={1}>Title</Title>
  <Subtitle>Subtitle</Subtitle>
  
  <Section title="Section Title" variant="primary" width={900}>
    <List items={['Item 1', 'Item 2', 'Item 3']} fontSize={12} />
  </Section>
</Slide>
```

Much cleaner, more readable, and easier to maintain!

## Tips

- **Width management**: Sections, Cards, and Callouts often need explicit widths. Common values: 520, 900, 1000, 1080
- **Spacing**: Use Spacer between major sections, gap props within containers
- **Typography**: Keep font sizes consistent - 12-14 for body text, 11 for secondary text
- **Colors**: Let variants handle colors unless you need custom styling
- **Flexibility**: All components accept additional BoxProps for customization

## Component Relationships

```
Slide (root container)
├── Title, Subtitle (typography)
├── Section (themed container)
│   ├── List (bullet points)
│   └── ProsCons (comparison)
├── Callout (emphasis)
│   └── Any content
├── Grid (layout)
│   ├── Section
│   ├── Card
│   └── Callout
└── RichText (inline formatting)
```

## Advanced Components (v2)

### FlowDiagram

Creates automatic flow diagrams with connected steps.

```tsx
<FlowDiagram
  steps={[
    { id: 'step1', label: 'Input', subtitle: 'User query', variant: 'primary' },
    { id: 'step2', label: 'Process', subtitle: 'Analyze', variant: 'secondary' },
    { id: 'step3', label: 'Output', subtitle: 'Response', variant: 'success' }
  ]}
  orientation="horizontal"
  showArrows={true}
/>
```

**Props:**
- `steps`: Array of `FlowStep` objects
- `orientation`: 'horizontal' | 'vertical' (default: 'horizontal')
- `stepWidth`: Width of each step (default: 180)
- `stepHeight`: Height of each step (optional)
- `showArrows`: Show connecting arrows (default: true)
- `arrowColor`: Color of arrows (default: '#1976d2')
- `gap`: Space between steps (default: 20)

**FlowStep:**
- `id`: Unique identifier for arrow connections
- `label`: Step name
- `subtitle`: Optional description
- `variant`: Color variant

### TwoColumn & ThreeColumn

Flexible column layouts without manual width calculations.

```tsx
<TwoColumn
  left={<Section title="Left">Content</Section>}
  right={<Section title="Right">Content</Section>}
  gap={24}
/>

<ThreeColumn
  left={<Text>Column 1</Text>}
  center={<Text>Column 2</Text>}
  right={<Text>Column 3</Text>}
/>
```

### CodeBlock

Display code with syntax highlighting and line numbers.

```tsx
<CodeBlock
  code={[
    'function hello() {',
    '  console.log("Hello!");',
    '}'
  ]}
  language="JavaScript"
  lineNumbers={true}
  fontSize={11}
/>
```

**Props:**
- `code`: String or array of strings
- `language`: Language label (optional)
- `fontSize`: Font size (default: 11)
- `lineNumbers`: Show line numbers (default: false)
- `backgroundColor`: Background color (default: '#1e1e1e')

### Quote

Blockquote with optional author attribution.

```tsx
<Quote
  text="This is an inspiring quote"
  author="Famous Person"
  variant="primary"
/>
```

**Props:**
- `text`: Quote text
- `author`: Optional attribution
- `fontSize`: Font size (default: 16)
- `variant`: 'default' | 'primary' | 'accent'

### Badge

Small labels or tags with variants.

```tsx
<Badge text="NEW" variant="success" size="medium" />
```

**Props:**
- `text`: Badge text
- `variant`: 'primary' | 'secondary' | 'success' | 'warning' | 'danger' | 'info' | 'default'
- `size`: 'small' | 'medium' | 'large' (default: 'medium')

### Divider

Visual separators between sections.

```tsx
<Divider width={900} variant="solid" />
<Divider width={900} variant="dashed" />
```

**Props:**
- `variant`: 'solid' | 'dashed' (default: 'solid')
- `thickness`: Line thickness (default: 1)
- `color`: Line color (default: '#ddd')
- `orientation`: 'horizontal' | 'vertical' (default: 'horizontal')
- `margin`: Spacing around divider (default: 16)

## Presentation Themes

Customize the look and feel of your presentations with themes.

### Built-in Themes

```tsx
import { 
  defaultTheme, 
  professionalTheme, 
  darkTheme, 
  vibrantTheme, 
  minimalTheme 
} from 'diagram-dsl';

// Use a theme (conceptual - for reference)
const myColors = {
  primary: professionalTheme.primary,
  success: professionalTheme.success
};
```

**Available Themes:**
- `defaultTheme`: Blue and purple color scheme
- `professionalTheme`: Muted, corporate colors
- `darkTheme`: Dark background with bright accents
- `vibrantTheme`: Bold, energetic colors
- `minimalTheme`: Black and white minimalism

### Creating Custom Themes

```tsx
import { createCustomTheme } from 'diagram-dsl';

const myTheme = createCustomTheme({
  primary: '#00bcd4',
  secondary: '#ff4081',
  accent: '#ffc107',
  slideWidth: 1920,
  slideHeight: 1080,
});
```

## Slide Deck Generation

Simplify presentation generation with the `generateSlideDeck` helper.

### Basic Usage

```tsx
import { generateSlideDeck, numberSlides } from 'diagram-dsl';

const slides = numberSlides([
  { name: 'intro', component: <IntroSlide /> },
  { name: 'content', component: <ContentSlide /> },
  { name: 'conclusion', component: <ConclusionSlide /> }
]);

await generateSlideDeck(slides, {
  outputDir: './output',
  htmlTitle: 'My Presentation',
  width: 1200,
  height: 800
});
```

### Benefits

- **Automatic numbering**: Slides numbered as 01-intro, 02-content, etc.
- **Built-in HTML viewer**: Professional navigation with keyboard shortcuts
- **Progress reporting**: See generation progress in real-time
- **Skip slides**: Mark slides with `skip: true` to exclude from output

### Advanced Options

```tsx
await generateSlideDeck(slides, {
  outputDir: './output',
  width: 1920,          // Custom dimensions
  height: 1080,
  backgroundColor: '#f0f0f0',
  createHTML: true,      // Generate viewer (default: true)
  htmlTitle: 'My Deck'
});
```

## Component Comparison

| Component | Use Case | Code Reduction |
|-----------|----------|----------------|
| Slide | Every slide | 3-5 lines → 1 line |
| List | Bullet points | 3-5 lines per item → 1 array |
| ProsCons | Comparisons | ~40 lines → ~8 lines |
| Section | Titled content | ~10 lines → ~3 lines |
| FlowDiagram | Process flows | ~50 lines → ~10 lines |
| TwoColumn | Side-by-side | ~15 lines → ~5 lines |
| generateSlideDeck | Full presentation | ~50 lines → ~10 lines |

## Complete Example

Here's a complete presentation using all the new features:

```tsx
import React from 'react';
import { 
  Slide, Title, Subtitle, Section, List, FlowDiagram, 
  TwoColumn, CodeBlock, Quote, Badge, Divider,
  generateSlideDeck, numberSlides 
} from 'diagram-dsl';

const TitleSlide = () => (
  <Slide alignItems="center" justifyContent="center">
    <Badge text="v2.0" variant="success" size="large" marginBottom={20} />
    <Title level={1}>Modern Presentations</Title>
    <Subtitle>With diagram-dsl</Subtitle>
  </Slide>
);

const FeaturesSlide = () => (
  <Slide>
    <Title level={1}>New Features</Title>
    <TwoColumn
      left={
        <Section title="Components" variant="primary">
          <List items={['FlowDiagram', 'CodeBlock', 'Quote', 'Badge']} />
        </Section>
      }
      right={
        <Section title="Helpers" variant="secondary">
          <List items={['Themes', 'generateSlideDeck', 'numberSlides']} />
        </Section>
      }
    />
  </Slide>
);

const FlowSlide = () => (
  <Slide>
    <Title level={1}>Development Process</Title>
    <FlowDiagram
      steps={[
        { id: 'design', label: 'Design', variant: 'primary' },
        { id: 'develop', label: 'Develop', variant: 'secondary' },
        { id: 'deploy', label: 'Deploy', variant: 'success' }
      ]}
      marginTop={40}
    />
  </Slide>
);

const CodeSlide = () => (
  <Slide>
    <Title level={1}>Code Example</Title>
    <CodeBlock
      language="TypeScript"
      code={[
        'const slides = numberSlides([',
        '  { name: "intro", component: <Intro /> },',
        '  { name: "features", component: <Features /> }',
        ']);',
        '',
        'await generateSlideDeck(slides, {',
        '  outputDir: "./output"',
        '});'
      ]}
      lineNumbers={true}
      width={800}
      marginTop={20}
    />
  </Slide>
);

const QuoteSlide = () => (
  <Slide>
    <Title level={1}>Testimonial</Title>
    <Quote
      text="diagram-dsl makes creating presentations a breeze!"
      author="Happy Developer"
      variant="primary"
      width={900}
      marginTop={40}
    />
    <Divider width={900} marginTop={30} marginBottom={30} />
    <Section title="Benefits" variant="success" width={900}>
      <List items={['75% less code', 'Faster development', 'Better consistency']} />
    </Section>
  </Slide>
);

async function generate() {
  const slides = numberSlides([
    { name: 'title', component: <TitleSlide /> },
    { name: 'features', component: <FeaturesSlide /> },
    { name: 'flow', component: <FlowSlide /> },
    { name: 'code', component: <CodeSlide /> },
    { name: 'quote', component: <QuoteSlide /> }
  ]);

  await generateSlideDeck(slides, {
    outputDir: './output',
    htmlTitle: 'Modern Presentations with diagram-dsl'
  });
}

generate();
```

This example demonstrates:
- All major components working together
- Proper use of variants for consistent theming
- Flow diagrams for process visualization
- Code examples with syntax highlighting
- Professional quotes and badges
- Automated slide deck generation

The result is a professional, maintainable presentation with minimal code!
