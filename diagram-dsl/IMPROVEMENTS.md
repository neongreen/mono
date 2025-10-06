# Text Measurement and Layout Testing

## Accurate Text Measurement

Text dimensions are measured using the `canvas` npm package for accurate layouts.

### Files
- `src/layout/text-measurement.ts` - Text measurement using canvas API
- `src/layout/yoga-engine.ts` - Uses `measureText()` for Text nodes

### Key Features
- `measureText()` uses canvas API for accurate dimensions
- Supports font size, font family, and font weight
- Returns bounding box metrics for vertical alignment

## Layout Testing

Framework for testing computed layout properties.

### Files
- `src/test/layout-assertions.ts` - Layout assertion utilities
- `src/test/layout-tests.tsx` - Layout tests

### LayoutAssertions Class
```typescript
assertCentered(childId, containerId, axis?)
assertFitsInside(childId, containerId, padding?)
assertGap(elem1Id, elem2Id, expectedGap)
assertAligned(elementIds[], alignment)
assertNoOverlap(elem1Id, elem2Id)
findById(id)
findByType(type)
getLayoutInfo(id)
```

### API
```typescript
renderToSVGWithLayout(element, options) -> { svg, layout }
```

Returns SVG string and computed layout tree.

## Tests

14 tests total:
- 7 SVG rendering tests
- 7 layout assertion tests

## Dependencies

- `canvas@3.2.0` - For text measurement

## Usage

```typescript
import { renderToSVGWithLayout, LayoutAssertions, measureText } from 'diagram-dsl';

// Render with layout data
const { svg, layout } = await renderToSVGWithLayout(<MyDiagram />);

// Test layout
const assertions = new LayoutAssertions(layout);
assertions.assertCentered('text', 'box');

// Measure text directly
const metrics = measureText('Hello', 24, 'Arial', 'bold');
```
