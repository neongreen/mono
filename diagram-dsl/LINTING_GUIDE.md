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

### 3. Arrowhead Crowding

**What it checks:** Detects when two arrowheads land too close together (closer than ~16px)

**Why it matters:** Crowded arrowheads are hard to parse visually and can look like a single arrow. Spacing them out improves clarity and makes directionality obvious.

**Example warning:**
```
⚠ Arrowheads for "source-a" → "target" and "source-b" → "target" are very close (8.0px).
  Increase spacing or adjust positioning so arrowheads have at least 16px separation.
```

**How to fix:**
- Increase spacing between the target and source elements so attachment points land farther apart
- Slightly offset the target node so arrows connect to different sides
- Consider routing one arrow through an intermediate connector

### 4. Arrow Crossings

**What it checks:** Detects when two straight connectors intersect each other.

**Why it matters:** Crossing arrows create visual ambiguity about which node connects to which. Clean diagrams avoid intersecting lines or use explicit junction markers.

**Example warning:**
```
⚠ Arrows "top-left" → "bottom-right" and "top-right" → "bottom-left" cross each other. Adjust layout or reroute arrows to avoid intersecting connectors.
```

**How to fix:**
- Increase spacing between rows/columns so arrows can connect without crossing
- Route one arrow to a different side of the target node
- Introduce intermediate connector nodes to fan out connections cleanly


### 5. Arrow Label Overlap

**What it checks:** Warns when an arrow label overlaps a node it does not belong to (or another connector), making the diagram difficult to read.

**Why it matters:** Labels should enhance clarity, not obscure other content. Overlapping labels often signal that nodes are packed too tightly or need rerouting.

**Example warning:**
```
⚠ Arrow label for "left-node" → "right-node" overlaps node "blocker". Nudge layout or reroute the arrow so labels remain readable.
```

**How to fix:**
- Increase spacing between source/target and surrounding nodes
- Route the arrow so the label lands in white space (different side or connector path)
- Introduce intermediary nodes/connectors to reposition the label safely


### 6. Text Overflow

**What it checks:** Detects when measured text width exceeds the usable width inside its container (accounting for padding).

**Why it matters:** Overflowing copy either gets clipped or forces awkward scaling when exported to other formats. Keeping text within bounds preserves legibility.

**Example warning:**
```
⚠ Text "narrow-label" exceeds available width inside "narrow-card". Reduce copy length, adjust padding, or widen the container.
```

**How to fix:**
- Shorten or rephrase the copy so it fits comfortably
- Increase the container width or padding as appropriate
- Reduce font size (as a last resort) to avoid overflow


### 7. Tight Text Stack Spacing

**What it checks:** Spots vertically adjacent text nodes (labels, subtitles, etc.) that render with less than ~6px separation.

**Why it matters:** When successive text lines are packed too tightly they blur together, especially once exported to slides or PDFs. Setting an explicit gap keeps typographic rhythm consistent.

**Example warning:**
```
⚠ Text nodes "tight-heading" and "tight-subtitle" inside "tight-card" are only 2.0px apart. Use a Stack with gap or adjust layout so they have at least 6px separation.
```

**How to fix:**
- Wrap the text nodes in a `<Stack gap={6}>` (or similar) so Yoga enforces consistent spacing
- Increase padding or container height to create breathing room
- Consider combining the copy into a single text block if it belongs together


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
pnpm lint
```

Test the linting system:

```bash
pnpm test:lint
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

## Future Enhancements

Potential future lints could include:
- Detecting text overflow in boxes
- Checking for inconsistent spacing patterns
- Validating arrow label positioning
- Detecting overlapping elements
- Checking minimum font sizes for readability

## Feedback

The linting system is designed to be helpful, not restrictive. If you find false positives or have suggestions for new lints, please open an issue!
