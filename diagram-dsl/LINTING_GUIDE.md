# Layout Linting Guide

The diagram-dsl now includes a layout linting system that helps identify potential visual hierarchy and spacing issues in your diagrams. These are **suggestions, not errors** - they help you create more professional-looking diagrams by following design best practices.

## Philosophy

The linter is based on the principle of **proximity and visual grouping**:

> **Internal distances should be smaller than external distances**

This means:
- The space between a box's border and its content (padding) should be smaller than the space between that box and other boxes
- Arrows should be long enough to be clearly visible and not appear cramped
- Visual elements that belong together should be closer to each other than to elements they don't belong with

## Available Lints

### 1. Short Arrow Detection

**What it checks:** Detects arrows that are too short (shorter than 20px)

**Why it matters:** Very short arrows can be hard to see because they're almost as short as the arrowhead marker itself (~10px). This often happens when boxes are placed too close together.

**Example warning:**
```
⚠ Arrow from "box1" to "box2" is very short (8.0px). 
  Consider increasing spacing between elements (minimum recommended: 20px).
```

**How to fix:**
- Increase the `gap` prop in your Stack or Row component
- Add more space between vertically or horizontally adjacent elements

**Example:**
```tsx
// Before (too tight)
<Stack gap={8}>  
  <Card id="box1">...</Card>
  <Card id="box2">...</Card>
  <Arrow from="box1" to="box2" />
</Stack>

// After (better spacing)
<Stack gap={40}>  
  <Card id="box1">...</Card>
  <Card id="box2">...</Card>
  <Arrow from="box1" to="box2" />
</Stack>
```

### 2. Internal vs External Spacing

**What it checks:** Detects when a box's internal spacing (padding) is larger than the external gap to adjacent boxes

**Why it matters:** When internal spacing is larger than external spacing, visual grouping becomes confusing. The content inside a box appears more distant from its own border than from neighboring elements, breaking the visual hierarchy.

**Example warning:**
```
⚠ Box "card1" has internal spacing (16px) > external gap (10.0px). 
  Internal distances should be smaller than external distances for better visual grouping.
```

**How to fix:**
- **Option 1:** Increase the gap between boxes (recommended)
- **Option 2:** Reduce the internal padding of the boxes
- **Option 3:** Adjust both to create better visual hierarchy

**Example:**
```tsx
// Before (poor visual hierarchy)
<Stack gap={10}>  {/* External gap: 10px */}
  <Card id="card1" padding={16}>  {/* Internal spacing: 16px */}
    <Label>Content</Label>
  </Card>
  <Card id="card2" padding={16}>
    <Label>Content</Label>
  </Card>
</Stack>

// After Option 1: Increase external gap (recommended)
<Stack gap={20}>  {/* External gap: 20px > internal 16px */}
  <Card id="card1" padding={16}>
    <Label>Content</Label>
  </Card>
  <Card id="card2" padding={16}>
    <Label>Content</Label>
  </Card>
</Stack>

// After Option 2: Reduce internal padding
<Stack gap={10}>
  <Card id="card1" padding={8}>  {/* Internal spacing: 8px < external 10px */}
    <Label>Content</Label>
  </Card>
  <Card id="card2" padding={8}>
    <Label>Content</Label>
  </Card>
</Stack>
```

## Using the Linter

### In Your Code

```tsx
import { renderToSVGWithLayout, LayoutLinter } from 'diagram-dsl';

const MyDiagram = () => (
  <Stack gap={20}>
    {/* ... your diagram ... */}
  </Stack>
);

// Render and lint
const { svg, layout } = await renderToSVGWithLayout(<MyDiagram />);

// Run linter
const linter = new LayoutLinter(layout);
const lints = linter.runAllLints();

// Display results
if (lints.length > 0) {
  console.log(LayoutLinter.formatLints(lints));
} else {
  console.log('No layout issues found!');
}
```

### Command Line

Lint your examples with:

```bash
npm run lint
```

Test the linting system:

```bash
npm run test:lint
```

## Lint Object Structure

Each lint warning has this structure:

```typescript
interface LayoutLint {
  type: 'warning' | 'info';
  message: string;        // Human-readable description
  elementId?: string;     // ID of the affected element
  details?: any;          // Additional diagnostic information
}
```

## Best Practices

1. **Review lints regularly** - Run the linter after making spacing changes
2. **Don't ignore short arrow warnings** - They usually indicate cramped layouts
3. **Follow the internal < external rule** - It creates better visual hierarchy
4. **Lints are suggestions** - Sometimes you may have good reasons to ignore them
5. **Consider context** - What works for flowcharts may differ from architecture diagrams

## Advanced: Custom Spacing Guidelines

For different diagram types, consider these spacing guidelines:

### Flowcharts
- Gap between boxes: 30-40px minimum
- Internal card padding: 12-16px
- Arrow lengths: >20px for readability

### Architecture Diagrams
- Gap between tiers: 24-32px
- Gap between cards in same tier: 16-20px (horizontal), 16-24px (vertical)
- Internal card padding: 16px
- Card internal gaps (label to subtitle): 8-10px

### Decision Trees
- Gap between decision points: 40-50px
- Branch spacing: 60-80px horizontal
- Internal padding: 12-16px


