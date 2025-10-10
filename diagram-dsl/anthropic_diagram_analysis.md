# Structural Analysis: Anthropic "Prompt engineering vs. context engineering" Diagram

## Executive Summary

This document provides a comprehensive structural analysis of the Anthropic diagram, including:
- Complete color palette with hex codes
- Hierarchical element tree with properties
- Layout constraints and sizing rules
- Arrow routing specifications
- Typography system
- Recommendations for improvements

## 1. Color Palette

### Background Colors
- **Canvas Background**: `#f7f7f5` (off-white/beige)
- **Panel Background**: `#fafafa` (light gray)
- **Message History Background**: `#ffffff` (white)

### Element Colors
- **Prompt Bubbles**: `#6fb4ff` (sky blue) with `#4a9ae5` border
- **Document Cards**: `#90EE90` (light green)
- **Tool Pills**: `#FFA500` (orange)
- **Memory Cards**: `#DDA0DD` (light violet/purple)
- **Instruction Cards**: `#6fb4ff` (cobalt blue)
- **Tool Call/Result**: `#f8e0ba` (tan/beige) with `#d4b896` border

### Border & Line Colors
- **Dashed Borders**: `#888888` (medium gray)
- **Panel Borders**: `#d0d0d0` (subtle gray)
- **Arrow Color**: `#4c4c4c` (charcoal)
- **Output Box Borders**: `#444444` (dark gray)

### Text Colors
- **Main Title**: `#000000` (black)
- **Section Headers**: `#666666` (medium gray)
- **Labels**: `#666666` (medium gray)
- **Content Text**: `#4c4c4c` (darker gray)

### Neural Network Colors
- **Circle Blue**: `#a3d5ff` (pastel blue)
- **Circle Green**: `#b8e6b8` (pastel green)
- **Circle Yellow**: `#ffe6a3` (pastel yellow)

### ⚠️ Color Contrast Issue
**Problem**: Contrast between canvas (#f7f7f5) and panels (#fafafa) is only ~1.06:1 (almost invisible)

**Recommendations**:
1. Keep background `#f7f7f5`, use `#ffffff` for panels (better)
2. Use `#eeeeee` for background, keep `#fafafa` for panels
3. Use `#f0f0f0` for background, use `#ffffff` for panels (best)

## 2. Hierarchical Element Structure

```
Canvas (#f7f7f5, padding: 24px)
├─ Title "Prompt engineering vs. context engineering" (32px bold, centered, margin-bottom: 24px)
└─ Row (gap: 40px, equal height children)
   ├─ Left Panel (width: ~450-500px, height: auto, bg: #fafafa, border-radius: 12px, padding: 20px)
   │  ├─ Header "Prompt engineering for single turn queries" (18px bold, margin-bottom: 16px)
   │  ├─ Label "Context window" (14px, margin-bottom: 6px) ← CLOSER to box below
   │  └─ Dashed Box (width: panel-40px, height: fill, border: 2px dashed #888, padding: 20px)
   │     └─ Stack (gap: 8px)
   │        ├─ Speech Bubble "System prompt" (140×40px, #6fb4ff)
   │        ├─ Speech Bubble "User message" (140×40px, #6fb4ff)
   │        ├─ Spacer (flexible height - represents unused token space)
   │        └─ Scissors Icon (30px, bottom-left)
   │
   ├─ Left Model Output (vertical stack, beside panel)
   │  ├─ Neural Grid (3×3, 10px circles, 6px gaps)
   │  └─ Output Box "Assistant message" (110×38px, white)
   │
   ├─ Right Panel (width: ~750-900px, height: auto, bg: #fafafa, border-radius: 12px, padding: 20px)
   │  ├─ Header "Context engineering for agents" (18px bold, margin-bottom: 16px)
   │  └─ Row (gap: 12px)
   │     ├─ Possible Context Container
   │     │  ├─ Label "Possible context to give model" (14px, margin-bottom: 6px)
   │     │  └─ Dashed Box (width: ~360-400px, height: fill, border: 2px dashed)
   │     │     └─ Stack (gap: 8px)
   │     │        ├─ Doc Card "Doc 1" (120×36px, green)
   │     │        ├─ Doc Card "Doc 2" (120×36px, green)
   │     │        ├─ Doc Card "Doc 3" (120×36px, green)
   │     │        ├─ Tool Pill "Tool 1" (120×36px, orange, rounded)
   │     │        ├─ Tool Pill "Tool 2" (120×36px, orange)
   │     │        ├─ Tool Pill "Tool 3" (120×36px, orange)
   │     │        ├─ Tool Pill "Tool 4" (120×36px, orange)
   │     │        ├─ Memory Card "Memory file" (120×36px, violet)
   │     │        ├─ Memory Card "Memory file" (120×36px, violet)
   │     │        ├─ Instruction "Comprehensive instructions" (180×36px, blue)
   │     │        ├─ Instruction "Domain knowledge" (180×36px, blue)
   │     │        └─ Message History (140×38px, white with border)
   │     │
   │     └─ Context Window Container
   │        ├─ Label "Context window" (14px, margin-bottom: 6px)
   │        └─ Dashed Box (width: ~360-400px, height: EQUAL to possible box)
   │           └─ Stack (gap: 8px)
   │              ├─ Speech Bubble "System prompt" (140×36px, blue)
   │              ├─ Row (gap: 6px)
   │              │  ├─ Doc "Doc 1" (67×36px)
   │              │  └─ Doc "Doc 2" (67×36px)
   │              ├─ Memory "Memory file" (140×36px, violet)
   │              ├─ Row (gap: 6px)
   │              │  ├─ Tool "Tool 1" (67×36px, orange pill)
   │              │  └─ Tool "Tool 2" (67×36px, orange pill)
   │              ├─ Speech Bubble "User message" (140×36px, blue)
   │              ├─ Message History (140×38px, white)
   │              ├─ Spacer (flexible)
   │              └─ Scissors Icon (30px, bottom-left)
   │
   └─ Right Model Output (vertical stack, beside panel)
      ├─ Neural Grid (3×3, 10px circles, 6px gaps)
      ├─ Output "Assistant message" (110×38px, white)
      ├─ Output "Tool call" (120×38px, tan)
      └─ Output "Tool result" (120×38px, tan)
```

## 3. Layout Constraints

### Global Constraints
1. **Canvas padding**: 24px all around
2. **Title centering**: Horizontally centered
3. **Title to panels spacing**: 24px margin-bottom
4. **Panel gap**: 40px horizontal gap between panels
5. **Equal heights**: Left and right panels MUST have equal heights
6. **Vertical stretch**: Panels stretch to accommodate content

### Panel Constraints
1. **Width ratio**: Left panel ~35-40% : Right panel ~52-60% (approximately 1:1.5 to 1:2)
2. **Suggested widths**:
   - Left panel: 450-500px
   - Right panel: 750-900px
3. **Internal padding**: 20px all sides
4. **Border radius**: 12px
5. **Background**: #fafafa

### Hierarchical Spacing Rules ⭐ CRITICAL
This is the KEY principle that makes the diagram clear:

**Rule**: Elements that logically belong together have SMALLER spacing than elements that are logically apart.

**Examples**:
- Section header → Label: 16px (larger gap, they're separate concepts)
- Label → Its box: 6px (smaller gap, they belong together)
- Elements within box: 8px gap (medium, they're related)

**Why this matters**: The visual hierarchy matches the logical hierarchy. When you see the spacing, you immediately understand which elements belong together.

### Dashed Box Constraints
1. **Inset**: ~20px from parent panel edges
2. **Height**: Fill remaining height after header/label
3. **Equal heights**: All dashed boxes in same panel have equal heights
4. **Border**: 2px dashed, #888888, dash 6px, gap 4px
5. **Border radius**: 8px
6. **Internal padding**: 12-20px

### Element Sizing Categories

#### Fixed Size Elements
- **Speech bubbles**: 140×40px (left panel), 140×36px (right panel)
- **Doc cards**: 120×36px (single), 67×36px (paired)
- **Tool pills**: 120×36px (single), 67×36px (paired)
- **Memory cards**: 120×36px (single), 140×36px (full width)
- **Instruction cards**: 180×36px
- **Message history**: 140×38px
- **Neural network grid**: 3×3, 10px circles, 6px gaps (total ~44×44px)
- **Output boxes**: 110-120×38px
- **Scissors icon**: 30px

#### Content-Based Size
- **Title**: Width based on text content
- **Headers**: Width based on text content
- **Labels**: Width based on text content

#### Fill-Available Size
- **Spacers inside dashed boxes**: Flexible height to represent unused token space
- **Dashed boxes**: Fill remaining height of parent panel

#### Parent-Based Size
- **Dashed box widths**: Parent width minus 40px (20px inset each side)

### Positioning Constraints
1. **Model grids**: OUTSIDE panels, positioned beside them
2. **Model vertical alignment**: Vertically centered relative to adjacent panel
3. **Scissors icons**: Bottom-left corner of dashed boxes
4. **Element alignment**: Left-aligned within containers
5. **Row elements**: Horizontally arranged with 6px gap

## 4. Arrow Routing Specifications

### Left Side Arrows
1. **Context to Model**:
   - From: left-context-window (center-right edge)
   - To: left-model (left side)
   - Style: 2px solid #4c4c4c, arrowhead
   - Routing: Straight line

2. **Model to Output**:
   - From: left-model (right side)
   - To: assistant-message-left (left side)
   - Style: 2px solid #4c4c4c, arrowhead
   - Routing: Straight line

### Right Side Arrows (Complex)

1. **Curation Arrow** ⭐ SPECIAL:
   - From: possible-context-box (center-right edge)
   - To: context-window-box (left edge, STOPS SHORT by ~10px)
   - Style: 2px solid #4c4c4c, arrowhead
   - Label: "Curation" (14px, monospace, gray, positioned ABOVE arrow)
   - Routing: Straight line
   - **Special property**: Arrow does NOT reach target (shortenEnd: 10)

2. **Bracket Connectors to Model**:
   - From: Multiple elements in context-window-box (right edges)
   - To: right-model (left side)
   - Style: 2px solid #4c4c4c, thin bracket-shaped lines
   - Routing: Bracket lines from each card → merge into single arrow → model
   - **Complex**: Multiple sources merge at a point before target

3. **Y-Shaped Fork** ⭐ SPECIAL:
   - From: right-model (right side)
   - To: TWO targets [assistant-message-right (up), tool-call (down)]
   - Style: 2px solid #4c4c4c, arrowheads on both branches
   - Routing: Single arrow splits into Y-shape
   - **Special property**: One source, two targets

4. **Tool Call to Result**:
   - From: tool-call (bottom)
   - To: tool-result (top)
   - Style: 2px solid #4c4c4c (may be slightly thicker)
   - Routing: Straight vertical line

5. **Feedback Loop** ⭐ SPECIAL:
   - From: tool-result (left side)
   - To: message-history in context-window-box (bottom edge)
   - Style: 2px solid #4c4c4c, arrowhead
   - Routing: Horizontal left → vertical up (orthogonal/step routing)
   - **Special property**: Forms feedback loop, enters panel from outside

## 5. Typography System

| Element | Font Size | Weight | Color | Family |
|---------|-----------|--------|-------|--------|
| Main Title | 32px | bold | #000000 | Inter/Helvetica |
| Section Header | 18px | bold | #666666 | Inter/Helvetica |
| Label | 14px | normal | #666666 | Inter/Helvetica |
| Content | 11-12px | normal | #4c4c4c | Inter/Helvetica |
| Content Bold | 11-12px | bold | #000000 | Inter/Helvetica |
| Arrow Label | 14px | normal | #666666 | Monospace |

## 6. Implementation Requirements

### What the Library Needs to Support

Based on this analysis, the diagram-dsl library needs:

1. **Layout Features**:
   - ✅ Flexbox-based layout (Stack, Row)
   - ✅ Equal height constraints
   - ✅ Content-based sizing
   - ✅ Fill-available sizing
   - ✅ Parent-based sizing with fixed offsets
   - ⚠️ Better support for hierarchical spacing rules

2. **Arrow Features**:
   - ✅ Basic arrows with fromSide/toSide
   - ⚠️ Arrow shortening (shortenEnd)
   - ⚠️ Y-shaped forks (one source → multiple targets)
   - ⚠️ Bracket merging (multiple sources → merge → one target)
   - ⚠️ Orthogonal routing (feedback loops)
   - ✅ Arrow labels

3. **Element Features**:
   - ✅ Dashed borders
   - ✅ Border radius
   - ✅ Speech bubbles (can use Box with border-radius)
   - ✅ Rounded pills (high border-radius)
   - ⚠️ Document cards with folded corner
   - ⚠️ Flexible spacers that fill available height

4. **Typography**:
   - ✅ Font sizes, weights, colors
   - ✅ Text alignment

### Priority Improvements

1. **HIGH**: Implement hierarchical spacing system (labels closer to their boxes)
2. **HIGH**: Arrow shortening (shortenEnd property)
3. **HIGH**: Y-shaped arrow forks (one source → multiple targets)
4. **MEDIUM**: Bracket merging for arrows (multiple sources → merge → target)
5. **MEDIUM**: Orthogonal/step routing for feedback loops
6. **MEDIUM**: Flexible spacers that fill available height
7. **LOW**: Document cards with folded corners (can use Box)
8. **LOW**: Better color contrast in default theme

## 7. Validation Checklist

Use this checklist to verify a diagram implementation:

### Layout
- [ ] Canvas has #f7f7f5 background
- [ ] Canvas has 24px padding
- [ ] Title is centered and 32px bold
- [ ] Title has 24px margin-bottom
- [ ] Panels have 40px horizontal gap
- [ ] Left and right panels have EQUAL heights
- [ ] Left panel width ~35-40% (450-500px suggested)
- [ ] Right panel width ~52-60% (750-900px suggested)
- [ ] Panel width ratio approximately 1:1.5 to 1:2
- [ ] Panels have #fafafa background (or better: #ffffff)
- [ ] Panels have 12px border-radius
- [ ] Panels have 20px internal padding

### Hierarchical Spacing
- [ ] Section header to label: 16px gap
- [ ] Label to box: 6px gap (SMALLER than header to label)
- [ ] Elements within boxes: 8px gap
- [ ] Spacing visually indicates logical grouping

### Dashed Boxes
- [ ] Dashed boxes have ~20px inset from panel edges
- [ ] Dashed boxes fill remaining height
- [ ] All dashed boxes in same panel have equal heights
- [ ] Border: 2px dashed #888888
- [ ] Dash pattern: 6px dash, 4px gap
- [ ] Border-radius: 8px
- [ ] Internal padding: 12-20px

### Elements
- [ ] All elements have correct fixed sizes
- [ ] Speech bubbles: 140×40px (left) or 140×36px (right)
- [ ] Doc cards: 120×36px or 67×36px
- [ ] Tool pills: Rounded (border-radius: 18px)
- [ ] Neural grid: 3×3, 10px circles, 6px gaps
- [ ] Output boxes: 110-120×38px
- [ ] Scissors: 30px, bottom-left of dashed boxes

### Arrows
- [ ] Left context → model: Straight line
- [ ] Left model → output: Straight line
- [ ] Curation arrow: Has label above, stops short of target
- [ ] Brackets merge into single arrow to model
- [ ] Y-fork: Model → two outputs
- [ ] Tool call → result: Vertical line
- [ ] Feedback loop: Orthogonal routing back to message history

### Colors
- [ ] All elements use correct colors from palette
- [ ] Background-to-panel contrast is visible
- [ ] Text colors match hierarchy (title black, headers gray, etc.)

### Typography
- [ ] Main title: 32px bold
- [ ] Section headers: 18px bold
- [ ] Labels: 14px normal
- [ ] Content: 11-12px
- [ ] Consistent font family (Inter/Helvetica)

---

## Conclusion

This analysis provides the complete structural specification needed to create a faithful replication of the Anthropic diagram. The key insights are:

1. **Hierarchical spacing** is critical for visual clarity
2. **Panel width ratio** should be approximately 1:1.5 to 1:2
3. **Equal panel heights** are a hard constraint
4. **Arrow features** need enhancements (shortening, Y-forks, merging, orthogonal routing)
5. **Color contrast** needs improvement (#fafafa vs #f7f7f5 is too subtle)

The diagram's complexity comes from:
- Nested boxes with specific size relationships
- Complex arrow routing (shortcuts, forks, merges, loops)
- Hierarchical spacing that conveys logical relationships
- Multiple element types with specific styling

By following this specification, a diagram library can create diagrams that match both the **style** (colors, borders, typography) and **substance** (layout constraints, spacing hierarchy, arrow complexity) of the original.
