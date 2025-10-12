# ReactFlow Layout System

A constraint-based layout system for building beautiful, responsive documents with ReactFlow.

## Overview

This project explores different implementation approaches for a layout system that makes it impossible to build something ugly. The system provides high-level building blocks (Stack, Card, Space) with intelligent constraints that prevent common layout issues like text overflow, poor spacing, and awkward proportions.

## Features

- **Constraint-Based Sizing**: Specify min/max/preferred dimensions and let the system handle the rest
- **Automatic Text Handling**: Text never overflows containers
- **Smart Defaults**: Sensible spacing and proportions without manual configuration
- **Multiple Approaches**: Compare three different implementation strategies

## Approaches

### Approach 1: Flexbox-based Constraint System
Uses CSS Flexbox with smart defaults and constraints. Best for simpler layouts with automatic flow.

**Pros:**
- Simple and predictable behavior
- Great browser support
- Minimal code complexity

**Cons:**
- Less control over complex grid layouts
- One-dimensional layout paradigm

### Approach 2: Grid-based Proportional System
Uses CSS Grid with proportional units (fr). Better for complex multi-dimensional layouts.

**Pros:**
- Powerful grid-based layouts
- Excellent for two-dimensional designs
- Clean fr-based proportions

**Cons:**
- Slightly more complex mental model
- May be overkill for simple layouts

### Approach 3: Constrained Absolute Positioning
Uses absolute positioning with intelligent constraint resolution.

**Pros:**
- Explicit control when needed
- Calculated optimal sizes
- Handles edge cases well

**Cons:**
- More computational overhead
- Requires more logic

## Components

### Stack
Arranges children vertically or horizontally with automatic spacing.

```tsx
<Stack direction="horizontal" gap={16} padding={20}>
  {/* children */}
</Stack>
```

**Props:**
- `direction`: 'horizontal' | 'vertical'
- `gap`: spacing between children (px)
- `align`: 'start' | 'center' | 'end' | 'stretch'
- `distribute`: 'start' | 'center' | 'end' | 'space-between' | 'space-around' | 'space-evenly'
- `padding`: internal padding (px)
- `constraints`: LayoutConstraints

### Card
A styled container with optional title and content.

```tsx
<Card
  title="My Card"
  content="Card content goes here"
  variant="primary"
  constraints={{ minWidth: 200, maxWidth: 400 }}
/>
```

**Props:**
- `title`: optional card title
- `content`: card content
- `variant`: 'default' | 'primary' | 'secondary' | 'success' | 'warning'
- `constraints`: LayoutConstraints

### Space
A flexible space that can grow/shrink within a stack.

```tsx
<Space grow={2} shrink={1}>
  {/* children */}
</Space>
```

**Props:**
- `grow`: flex grow factor
- `shrink`: flex shrink factor
- `basis`: flex basis (number or 'auto')
- `constraints`: LayoutConstraints

## Development

```bash
# Install dependencies
pnpm install

# Run development server
pnpm dev

# Build for production
pnpm build

# Preview production build
pnpm preview
```

## Examples

The showcase includes three examples that demonstrate different use cases:

1. **Dashboard Layout**: Header, sidebar, main content area with metrics
2. **Multi-Section Form**: Form with various sections and input groups
3. **Complex Grid**: Nested stacks with varied content and proportions

Each example is rendered with all three implementation approaches so you can compare them side-by-side.

## Deployment

This project is configured for deployment on Vercel. Simply connect your repository to Vercel and it will automatically deploy on push.

The `vercel.json` configuration is already set up to:
- Build with PNPM
- Output to the `dist` directory
- Handle client-side routing

## License

MIT
