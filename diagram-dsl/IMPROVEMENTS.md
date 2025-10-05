# Improvements: Accurate Text Measurement and Layout Testing

This document describes the improvements made to address layout accuracy and testing issues.

## Changes Made

### 1. Accurate Text Measurement (Addresses Root Cause)

**Problem:** Text dimensions were estimated using a simple formula (`width = text.length * fontSize * 0.6`), leading to inaccurate layouts and text misalignment.

**Solution:** Integrated the `canvas` npm package to use actual browser-like text measurement.

#### New Files
- `src/layout/text-measurement.ts` - Text measurement utilities using canvas API

#### Key Features
- `measureText()` function uses canvas `measureText()` API for accurate dimensions
- Supports font size, font family, and font weight
- Calculates exact width and height based on actual font metrics
- Returns actual bounding box ascent/descent for precise vertical alignment

#### Changes to Existing Files
- `src/layout/yoga-engine.ts` - Now uses `measureText()` instead of estimates for Text nodes

**Impact:**
- Text is now positioned more accurately within containers
- Better vertical alignment of text elements
- More reliable layout calculations for text-heavy diagrams

### 2. Computational Layout Testing (Enables Automated Verification)

**Problem:** No automated way to verify layout correctness without human visual inspection.

**Solution:** Added layout extraction and assertion framework for testing computed layout properties.

#### New Files
- `src/test/layout-assertions.ts` - Layout assertion utilities
- `src/test/layout-tests.tsx` - Comprehensive layout tests

#### Key Features

**LayoutAssertions Class:**
```typescript
- assertCentered(childId, containerId, axis?) - Verify centering
- assertFitsInside(childId, containerId, padding?) - Verify containment
- assertGap(elem1Id, elem2Id, expectedGap) - Verify spacing
- assertAligned(elementIds[], alignment) - Verify alignment
- assertNoOverlap(elem1Id, elem2Id) - Verify no overlap
- findById(id) - Find element by ID
- findByType(type) - Find all elements of type
- getLayoutInfo(id) - Debug layout information
```

**New API Functions:**
```typescript
renderToSVGWithLayout(element, options) -> { svg, layout }
```

Returns both SVG string and computed layout tree for testing.

#### Changes to Existing Files
- `src/renderer/index.ts` - Added `renderToSVGWithLayout()` function
- `src/types/index.ts` - Added `id` property to `TextProps`
- `src/index.ts` - Exported new testing utilities
- `src/test/index.test.tsx` - Integrated layout tests

**Impact:**
- Can verify layout properties mathematically
- Tests are precise and deterministic
- Clear error messages pinpoint exact issues
- No human inspection needed for basic layout verification

## Test Results

### Before Improvements
- 7 SVG rendering tests (all passing)
- No layout verification tests

### After Improvements
- 7 SVG rendering tests (all passing)
- 7 layout assertion tests (all passing)

**New Layout Tests:**
1. ✓ Text is centered in box
2. ✓ Text fits inside box with padding
3. ✓ Gap between stacked boxes is correct
4. ✓ Boxes are vertically centered in row
5. ✓ Boxes don't overlap
6. ✓ Text has accurate measurements
7. ✓ Multiple text sizes measured correctly

## Dependencies Added

- `canvas@3.2.0` - For accurate text measurement
  - Zero vulnerabilities (verified)
  - Native module with pre-built binaries
  - Cross-platform support

## Example Improvements

### Simple Box Example
- **Before:** Text width estimated at ~130px
- **After:** Text width accurately measured at 102px
- **Result:** Better centering within the box

### Text Positioning
- **Before:** Vertical position calculated with rough estimate (fontSize * 1.2)
- **After:** Vertical position based on actual ascent/descent metrics
- **Result:** More precise vertical centering

## API Changes

### New Exports

```typescript
// Render with layout data for testing
import { renderToSVGWithLayout, RenderResult } from 'diagram-dsl';

const { svg, layout } = await renderToSVGWithLayout(<MyDiagram />);

// Layout assertions for testing
import { LayoutAssertions } from 'diagram-dsl';

const assertions = new LayoutAssertions(layout);
assertions.assertCentered('text', 'box');

// Direct text measurement
import { measureText, TextMetrics } from 'diagram-dsl';

const metrics = measureText('Hello', 24, 'Arial', 'bold');
console.log(metrics.width, metrics.height);
```

### Backward Compatibility

All existing APIs remain unchanged:
- `renderToSVG()` still returns a string (uses `renderToSVGWithLayout()` internally)
- All component props remain the same
- All examples work without modifications

## Usage in Tests

```typescript
import { renderToSVGWithLayout, LayoutAssertions } from 'diagram-dsl';

// Test that text is centered
const { layout } = await renderToSVGWithLayout(
  <Box id="container" justifyContent="center" alignItems="center">
    <Text id="text">Centered</Text>
  </Box>
);

const assertions = new LayoutAssertions(layout);
assertions.assertCentered('text', 'container');
assertions.assertFitsInside('text', 'container', 10);
```

## Performance Impact

- **Text Measurement:** Minimal overhead (~1ms per text element)
- **Layout Testing:** Only used in tests, no production impact
- **Build Time:** Unchanged
- **Example Generation:** ~10-20ms slower due to canvas initialization (still < 100ms total)

## Future Enhancements

These improvements lay the groundwork for:

1. **Property-Based Testing** - Generate random diagrams and test invariants
2. **Visual Regression Testing** - Compare SVG snapshots once layouts are correct
3. **Font Loading** - Support custom fonts for accurate measurement
4. **Multi-line Text** - Extend measurement for text wrapping
5. **Advanced Metrics** - Letter spacing, line height, etc.

## Migration Guide

No migration needed - all changes are backward compatible.

To use new features:
1. `pnpm install` to get canvas package
2. Use `renderToSVGWithLayout()` for testing
3. Import `LayoutAssertions` for layout tests
4. Run `pnpm test` to see new tests in action

## Verification

Run the following to verify improvements:

```bash
# Build with canvas support
pnpm build

# Run all tests (14 total: 7 SVG + 7 layout)
pnpm test

# Regenerate examples with improved text measurement
pnpm run examples
```

All tests should pass and examples should be generated with more accurate text positioning.
