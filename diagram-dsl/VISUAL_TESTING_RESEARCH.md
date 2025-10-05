# Automated Visual Testing for Diagram Generation - Research & Recommendations

## Problem Statement

AI agents struggle to detect visual layout issues in generated diagrams because:
- Text misalignment and overlap
- Incorrect content size estimation
- Off-center elements
- General measurement calculation errors

Vision models treating images as pictures aren't precise enough for layout verification.

## How Other Projects Approach This

### 1. **Snapshot/Visual Regression Testing**

Used by: React Native, Storybook, Chromatic, Percy

**Approach:**
- Generate SVG/image outputs
- Store "golden" reference files
- Compare new outputs pixel-by-pixel or via diff
- Flag differences for human review

**Tools:**
- `jest-image-snapshot` - Jest matcher for image comparison
- `pixelmatch` - Pixel-level image comparison
- `looks-same` - Image comparison library
- `svg-snapshot-testing` - SVG-specific snapshot testing

**Pros:**
- Catches any visual regression
- No need to enumerate all possible issues
- Works for complex layouts

**Cons:**
- Requires initial "good" baseline
- Brittle (tiny changes break tests)
- Doesn't explain *what* is wrong
- Storage overhead for golden files

### 2. **Geometric Property Testing**

Used by: Layout engines (Yoga, Morphdom), browser test suites

**Approach:**
- Assert on computed layout properties
- Test bounding boxes, positions, dimensions
- Verify relationships (alignment, spacing, containment)

**Example Tests:**
```javascript
// Test that text is centered in box
expect(text.x + text.width/2).toBeCloseTo(box.x + box.width/2, 1);
expect(text.y + text.height/2).toBeCloseTo(box.y + box.height/2, 1);

// Test that text doesn't overflow box
expect(text.x).toBeGreaterThanOrEqual(box.x + padding);
expect(text.x + text.width).toBeLessThanOrEqual(box.x + box.width - padding);

// Test gap between elements
expect(box2.y).toBeCloseTo(box1.y + box1.height + gap, 1);
```

**Pros:**
- Precise error messages
- Tests intent, not pixels
- Fast execution
- No golden files needed

**Cons:**
- Requires instrumenting the layout system
- Must enumerate all properties to test
- Doesn't catch visual rendering issues

### 3. **SVG/DOM Structure Testing**

Used by: D3.js, SVG libraries, React Testing Library

**Approach:**
- Parse SVG as XML/DOM
- Query elements by type, attributes, content
- Assert on element properties and relationships

**Example Tests:**
```javascript
const svg = parseSVG(output);
const texts = svg.querySelectorAll('text');
const rect = svg.querySelector('rect');

// Test text is inside rect bounds
texts.forEach(text => {
  const textX = parseFloat(text.getAttribute('x'));
  const rectX = parseFloat(rect.getAttribute('x'));
  const rectWidth = parseFloat(rect.getAttribute('width'));
  expect(textX).toBeGreaterThanOrEqual(rectX);
  expect(textX).toBeLessThanOrEqual(rectX + rectWidth);
});
```

**Pros:**
- Works with SVG output directly
- Can verify specific properties
- No rendering needed
- Fast and deterministic

**Cons:**
- Only tests SVG attributes, not visual result
- Doesn't account for text metrics
- Can't verify actual rendering

### 4. **Text Measurement Integration**

Used by: Canvas APIs, PDF generators, typography engines

**Approach:**
- Use proper text measurement API
- Calculate accurate bounding boxes
- Adjust layout based on actual metrics

**Libraries:**
- `canvas` npm package - Node.js canvas with measureText
- `opentype.js` - Font parsing and metrics
- `fontkit` - Font subsetting and metrics
- `text-metrics` - Text measurement without canvas

**Example:**
```javascript
import { createCanvas } from 'canvas';

function measureText(text, fontSize, fontFamily) {
  const canvas = createCanvas(1, 1);
  const ctx = canvas.getContext('2d');
  ctx.font = `${fontSize}px ${fontFamily}`;
  const metrics = ctx.measureText(text);
  return {
    width: metrics.width,
    height: metrics.actualBoundingBoxAscent + metrics.actualBoundingBoxDescent
  };
}
```

**Pros:**
- Accurate text dimensions
- Works with any font
- Platform-independent with canvas
- Can be integrated into layout calculation

**Cons:**
- Requires native dependencies (canvas)
- Adds complexity to build
- Different across platforms/fonts

### 5. **Property-Based Testing with Invariants**

Used by: QuickCheck, fast-check, layout algorithms

**Approach:**
- Generate random inputs
- Test universal invariants
- Verify mathematical properties

**Example Invariants:**
```javascript
// Text should always fit in its container
text.width <= container.width - 2 * padding

// Centered elements should be equidistant from edges
abs((text.x - box.x) - (box.x + box.width - text.x - text.width)) < 1

// Stacked elements shouldn't overlap
box2.y >= box1.y + box1.height

// Gap should be preserved
abs((box2.y - box1.y - box1.height) - gap) < 1
```

**Pros:**
- Finds edge cases
- Tests general properties
- Good coverage with few tests
- Self-documenting constraints

**Cons:**
- Requires defining invariants
- May miss specific bugs
- Can be slow with many iterations

## Recommended Approach for diagram-dsl

### **Phase 1: Computational Layout Testing (Immediate)**

Add a layout verification layer that extracts and tests computed properties:

```typescript
interface LayoutAssertions {
  // Test that computed layout matches expectations
  assertCentered(element: string, container: string): void;
  assertFitsInside(element: string, container: string, padding?: number): void;
  assertGap(element1: string, element2: string, expectedGap: number): void;
  assertAligned(elements: string[], alignment: 'left' | 'center' | 'right'): void;
  assertNoOverlap(element1: string, element2: string): void;
}
```

**Implementation:**
1. Expose computed layout tree from `renderToSVG`
2. Add helper functions to query elements by ID
3. Write assertions for common layout properties
4. Test each example against its invariants

**Example Test:**
```typescript
const { svg, layout } = await renderToSVGWithLayout(<Example1 />);

// Text should be centered in box
assertCentered(layout, 'text1', 'box1');

// Text should fit inside box with padding
assertFitsInside(layout, 'text1', 'box1', 20);

// Boxes should have correct gap
assertGap(layout, 'box1', 'box2', 20);
```

### **Phase 2: Accurate Text Measurement (Next)**

Replace estimation with real text measurement:

```bash
pnpm add canvas
```

```typescript
import { createCanvas } from 'canvas';

function measureText(text: string, fontSize: number, fontFamily: string) {
  const canvas = createCanvas(1, 1);
  const ctx = canvas.getContext('2d');
  ctx.font = `${fontSize}px ${fontFamily}`;
  const metrics = ctx.measureText(text);
  return {
    width: metrics.width,
    height: metrics.actualBoundingBoxAscent + metrics.actualBoundingBoxDescent
  };
}
```

Update `yoga-engine.ts` to use actual measurements instead of estimates.

### **Phase 3: Visual Regression Testing (Future)**

Once layouts are correct, add snapshot testing:

```bash
pnpm add -D jest-image-snapshot pixelmatch
```

```typescript
import { toMatchImageSnapshot } from 'jest-image-snapshot';

expect.extend({ toMatchImageSnapshot });

test('simple-box renders correctly', async () => {
  const svg = await renderToSVG(<SimpleBox />);
  const png = await convertSVGtoPNG(svg); // Using resvg-js or similar
  expect(png).toMatchImageSnapshot();
});
```

### **Phase 4: Property-Based Testing (Enhancement)**

Add randomized testing for edge cases:

```bash
pnpm add -D fast-check
```

```typescript
import fc from 'fast-check';

test('text always fits in box', () => {
  fc.assert(
    fc.property(
      fc.string({ minLength: 1, maxLength: 50 }),
      fc.integer({ min: 12, max: 48 }), // fontSize
      fc.integer({ min: 100, max: 500 }), // boxWidth
      (text, fontSize, boxWidth) => {
        const layout = computeLayout(
          <Box width={boxWidth} padding={20}>
            <Text fontSize={fontSize}>{text}</Text>
          </Box>
        );
        
        const textNode = layout.findById('text');
        const boxNode = layout.findById('box');
        
        // Text should fit inside box with padding
        expect(textNode.width).toBeLessThanOrEqual(boxWidth - 40);
      }
    )
  );
});
```

## Immediate Actions to Fix Current Issues

### 1. **Add Layout Extraction**

Modify `renderToSVG` to return layout data:

```typescript
export async function renderToSVG(element: ReactElement, options?: RenderOptions) {
  const layoutTree = elementToLayoutNode(element);
  const layoutEngine = await YogaLayoutEngine.create();
  const computedTree = layoutEngine.computeLayout(layoutTree, options.width, options.height);
  const svgRenderer = new SVGRenderer();
  const svg = svgRenderer.renderWithArrowMarkers(computedTree, options);
  
  return {
    svg,
    layout: computedTree  // Expose for testing
  };
}
```

### 2. **Create Layout Test Utilities**

```typescript
// src/test/layout-assertions.ts
export class LayoutAssertions {
  constructor(private layout: LayoutNode) {}
  
  findById(id: string): LayoutNode {
    // Traverse tree to find node with id
  }
  
  assertCentered(childId: string, containerId: string, axis: 'x' | 'y' | 'both' = 'both') {
    const child = this.findById(childId);
    const container = this.findById(containerId);
    
    if (axis === 'x' || axis === 'both') {
      const childCenterX = child.computed.x + child.computed.width / 2;
      const containerCenterX = container.computed.x + container.computed.width / 2;
      expect(Math.abs(childCenterX - containerCenterX)).toBeLessThan(1);
    }
    
    if (axis === 'y' || axis === 'both') {
      const childCenterY = child.computed.y + child.computed.height / 2;
      const containerCenterY = container.computed.y + container.computed.height / 2;
      expect(Math.abs(childCenterY - containerCenterY)).toBeLessThan(1);
    }
  }
  
  assertFitsInside(childId: string, containerId: string, padding = 0) {
    const child = this.findById(childId);
    const container = this.findById(containerId);
    
    expect(child.computed.x).toBeGreaterThanOrEqual(container.computed.x + padding);
    expect(child.computed.y).toBeGreaterThanOrEqual(container.computed.y + padding);
    expect(child.computed.x + child.computed.width).toBeLessThanOrEqual(
      container.computed.x + container.computed.width - padding
    );
    expect(child.computed.y + child.computed.height).toBeLessThanOrEqual(
      container.computed.y + container.computed.height - padding
    );
  }
}
```

### 3. **Write Specific Tests for Current Examples**

```typescript
test('simple box - text is centered', async () => {
  const { layout } = await renderToSVGWithLayout(<SimpleBox />);
  const assertions = new LayoutAssertions(layout);
  assertions.assertCentered('text', 'box');
});

test('basic flowchart - boxes are vertically stacked with gap', async () => {
  const { layout } = await renderToSVGWithLayout(<BasicFlowchart />);
  const assertions = new LayoutAssertions(layout);
  assertions.assertGap('start', 'process', 20);
  assertions.assertGap('process', 'end', 20);
});
```

## Summary

**Best approach without human in the loop:**

1. ✅ **Computational testing** - Test layout properties mathematically
2. ✅ **Text measurement** - Use canvas API for accurate dimensions
3. ✅ **Invariant testing** - Define and test universal rules
4. ⚠️ **Snapshot testing** - Good for regression, needs initial baseline
5. ❌ **Vision models** - Not precise enough for layout verification

**Priority order:**
1. Add layout extraction and assertion utilities (immediate, high impact)
2. Integrate real text measurement (fixes root cause)
3. Write property-based tests (finds edge cases)
4. Add visual snapshots (catches regressions)

The key insight: **Don't test the visual output, test the computational model that generates it.**
