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

The linter now includes 6 different checks to help you create better diagrams:

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

### 2. Internal vs External Spacing (Visual Hierarchy)

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

### 3. Overlapping Elements

**What it checks:** Detects when two boxes or cards physically overlap each other

**Why it matters:** Overlapping elements obscure content and create visual confusion. Each element should have its own clear space in the diagram.

**Example warning:**
```
⚠ Elements "box1" and "box2" are overlapping. 
  This may obscure content and create visual confusion.
```

**How to fix:**
- Increase spacing between elements
- Adjust element sizes to fit better
- Reorganize layout to prevent overlap
- Check for absolute positioning conflicts

**Example:**
```tsx
// Before (overlapping due to tight layout)
<Stack gap={10}>
  <Box id="box1" width={200} height={100} position="absolute" top={0} left={0} />
  <Box id="box2" width={200} height={100} position="absolute" top={50} left={50} />
</Stack>

// After (no overlap)
<Stack gap={20}>
  <Box id="box1" width={200} height={100} />
  <Box id="box2" width={200} height={100} />
</Stack>
```

### 4. Minimum Font Size

**What it checks:** Detects text with font sizes smaller than 10px

**Why it matters:** Text smaller than 10px can be difficult to read, especially on lower-resolution displays or when printed. This affects accessibility and overall diagram usability.

**Example warning:**
```
ℹ Text has very small font size (8px). 
  Minimum recommended: 10px for readability.
```

**How to fix:**
- Increase `fontSize` prop to at least 10px
- Consider 12-14px for body text
- Use 16px or larger for headings

**Example:**
```tsx
// Before (too small)
<Text fontSize={8}>Important info</Text>

// After (readable)
<Text fontSize={12}>Important info</Text>
```

### 5. Inconsistent Spacing

**What it checks:** Detects when a container (Stack or Row) has highly variable gaps between its children

**Why it matters:** Inconsistent spacing creates visual noise and makes the diagram look unpolished. Uniform spacing creates better visual rhythm and professionalism.

**Example warning:**
```
ℹ Container "main-stack" has inconsistent spacing between children 
  (10.0px to 40.0px). Consider using uniform gaps for visual consistency.
```

**How to fix:**
- Use the same `gap` value for all children in a container
- If different spacing is needed, consider using nested containers
- Be intentional about spacing variations

**Example:**
```tsx
// Before (inconsistent - gaps vary)
<Stack gap={10}>
  <Card height={50} />
  <Card height={50} marginTop={30} />  {/* Extra margin creates inconsistency */}
  <Card height={50} />
</Stack>

// After (consistent)
<Stack gap={20}>
  <Card height={50} />
  <Card height={50} />
  <Card height={50} />
</Stack>
```

### 6. Arrow Crossings

**What it checks:** Detects when arrows cross each other in the diagram

**Why it matters:** Crossed arrows make it harder to follow connections and understand the flow of information. Clean, non-crossing paths are easier to trace visually.

**Example warning:**
```
ℹ Arrows "box1→box2" and "box3→box4" are crossing. 
  Consider rearranging elements to avoid crossed connections.
```

**How to fix:**
- Rearrange elements to minimize crossings
- Change the order of boxes in the layout
- Consider a different layout structure (e.g., use layers/tiers)
- For complex diagrams, accept some crossings as unavoidable

**Example:**
```tsx
// Before (arrows cross)
<Stack gap={20}>
  <Row gap={20}>
    <Box id="A" />
    <Box id="B" />
  </Row>
  <Row gap={20}>
    <Box id="C" />
    <Box id="D" />
  </Row>
  <Arrow from="A" to="D" />  {/* These arrows */}
  <Arrow from="B" to="C" />  {/* will cross */}
</Stack>

// After (no crossing - reordered)
<Stack gap={20}>
  <Row gap={20}>
    <Box id="A" />
    <Box id="B" />
  </Row>
  <Row gap={20}>
    <Box id="D" />
    <Box id="C" />
  </Row>
  <Arrow from="A" to="D" />
  <Arrow from="B" to="C" />
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

## Lint Severity Levels

The linter uses two severity levels:

- **⚠ Warning** - Issues that likely affect usability or visual clarity (short arrows, overlapping elements, poor visual hierarchy)
- **ℹ Info** - Suggestions for improvement that may be acceptable in context (font sizes, inconsistent spacing, arrow crossings)

## Best Practices

1. **Review lints regularly** - Run the linter after making spacing changes
2. **Address warnings first** - Warnings (⚠) indicate more serious issues
3. **Don't ignore short arrow warnings** - They usually indicate cramped layouts
4. **Follow the internal < external rule** - It creates better visual hierarchy
5. **Consider info lints contextually** - Info (ℹ) suggestions may be acceptable depending on your needs
6. **Use consistent spacing** - Uniform gaps create visual rhythm
7. **Avoid overlapping elements** - Each element needs its own space
8. **Keep text readable** - Use font sizes of 10px or larger
9. **Minimize arrow crossings** - Clean paths are easier to follow
10. **Lints are suggestions** - Sometimes you may have good reasons to ignore them

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

## Summary of All Checks

| Check | Type | What it detects |
|-------|------|----------------|
| Short Arrow Detection | Warning | Arrows shorter than 20px |
| Internal vs External Spacing | Warning | Padding larger than gaps |
| Overlapping Elements | Warning | Elements that physically overlap |
| Minimum Font Size | Info | Text smaller than 10px |
| Inconsistent Spacing | Info | Variable gaps in containers |
| Arrow Crossings | Info | Arrows that intersect each other |

## Future Enhancements

Potential future lints could include:
- Detecting text overflow in boxes (when text is wider/taller than container)
- Validating arrow label positioning and collisions
- Checking for proper alignment of related elements
- Detecting color contrast issues for accessibility
- Suggesting optimal container sizes based on content

## Feedback

The linting system is designed to be helpful, not restrictive. If you find false positives or have suggestions for new lints, please open an issue!
