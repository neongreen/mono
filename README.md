# Agent Demo Workspace

A monorepo for building presentations and diagrams with diagram-dsl.

## Projects

### diagram-dsl

A high-level DSL for creating diagrams and presentation slides using React and JSX.

**Features:**
- Declarative React-based API
- Automatic layout with Yoga
- SVG output
- Presentation helper components
- Semantic styling variants

**Location:** `diagram-dsl/`

**Documentation:**
- [Presentation Components Guide](diagram-dsl/PRESENTATION_COMPONENTS.md)
- Component examples in `diagram-dsl/src/examples/`

### Presentations

Collection of presentations built with diagram-dsl.

**Location:** `presentations/`

**Current presentations:**
- `llm-context-management/` - Context management strategies in LLM agents (8 slides)

## Getting Started

### Prerequisites

- Node.js 18+ (for ESM support)
- pnpm 8+

### Installation

```bash
# Install all dependencies
pnpm install

# Build diagram-dsl
pnpm run build
```

### Creating a Presentation

**Option 1: Use the template script**
```bash
./scripts/create-presentation.sh my-presentation
cd presentations/my-presentation
pnpm install
# Edit src/presentation.tsx
pnpm generate
```

**Option 2: Copy existing presentation**
```bash
cp -r presentations/llm-context-management presentations/my-presentation
cd presentations/my-presentation
# Edit src/presentation.tsx and package.json
pnpm generate
```

### Viewing a Presentation

```bash
cd presentations/llm-context-management
pnpm generate
open output/index.html
```

## Development

### Building diagram-dsl

```bash
cd diagram-dsl
pnpm build
```

### Running Examples

```bash
cd diagram-dsl
pnpm exec tsx src/examples/presentation-helpers.tsx
pnpm exec tsx src/examples/advanced-presentation.tsx
```

### Testing

```bash
cd diagram-dsl
pnpm test
```

## Workspace Structure

```
agentdemo/
├── diagram-dsl/              # Core diagram DSL library
│   ├── src/
│   │   ├── components/       # React components
│   │   ├── layout/          # Layout engine
│   │   ├── renderer/        # SVG renderer
│   │   ├── test/            # Test utilities
│   │   └── examples/        # Usage examples
│   ├── dist/                # Built output
│   └── package.json
├── presentations/           # Presentation projects
│   ├── llm-context-management/
│   │   ├── src/
│   │   │   ├── presentation.tsx      # Original
│   │   │   └── presentation-v2.tsx   # Refactored
│   │   ├── output/          # Generated SVGs
│   │   └── package.json
│   └── README.md
├── scripts/                 # Utility scripts
│   └── create-presentation.sh
├── package.json            # Workspace root
├── pnpm-workspace.yaml     # Workspace configuration
└── README.md
```

## Key Components

### Presentation Components

**Layout:**
- `Slide` - Standard slide container (1200x800)
- `Grid` - Multi-column layouts
- `Spacer` - Flexible spacing

**Content:**
- `Section` - Titled content sections with variants
- `List` - Bullet point lists
- `ProsCons` - Side-by-side comparison
- `Callout` - Highlighted important information
- `RichText` - Mixed text formatting

**Typography:**
- `Title` - Main headings (levels 1-3)
- `Subtitle` - Secondary headings
- `Label` - Small labels
- `Text` - Body text

**Containers:**
- `Card` - Styled boxes with variants
- `Box`, `Stack`, `Row` - Basic layout primitives

See [PRESENTATION_COMPONENTS.md](diagram-dsl/PRESENTATION_COMPONENTS.md) for complete documentation.

## Quick Example

```tsx
import { Slide, Title, Section, List, Callout } from 'diagram-dsl';

const MySlide = () => (
  <Slide>
    <Title level={1}>My Presentation</Title>
    
    <Section title="Key Points" variant="primary" width={900}>
      <List items={['Point 1', 'Point 2', 'Point 3']} />
    </Section>
    
    <Callout title="Important" variant="warning" width={900}>
      <Text>Don't forget this important detail!</Text>
    </Callout>
  </Slide>
);
```

## Best Practices

1. **Use presentation components** instead of manual layout for consistency
2. **Define each slide as a separate component** for better organization
3. **Keep slides focused** - one main idea per slide
4. **Use variants** for consistent theming throughout
5. **Test individual slides** before generating the full deck
6. **Leverage Spacer** for flexible spacing between sections

## Performance Tips

- Building diagram-dsl once is sufficient for multiple presentations
- Generate presentations in parallel for faster batch generation
- SVG files are typically small (1-10KB per slide)
- HTML viewer has no dependencies and works offline

## Contributing

When adding new components to diagram-dsl:

1. Create component in `diagram-dsl/src/components/`
2. Export from `diagram-dsl/src/index.ts`
3. Add examples to `diagram-dsl/src/examples/`
4. Update `diagram-dsl/PRESENTATION_COMPONENTS.md`
5. Build and test
6. Commit changes

## License

MIT

## Resources

- [diagram-dsl Documentation](diagram-dsl/)
- [Presentation Components Guide](diagram-dsl/PRESENTATION_COMPONENTS.md)
- [Presentations README](presentations/README.md)
- [Example Presentation](presentations/llm-context-management/)
