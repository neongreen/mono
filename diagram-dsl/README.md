# diagram-dsl

A DSL for creating diagrams and presentations using React and JSX.

## Features

- React-based declarative API
- Automatic layout with Yoga
- SVG output
- Presentation components
- Semantic styling

## Installation

```bash
npm install
npm run build
```

## Usage

```bash
npm test
npm run examples
```

## Components

**Layout:**
- `Box`, `Stack`, `Row`, `Column`
- `Slide`, `Grid`, `Spacer`

**Typography:**
- `Text`, `Title`, `Subtitle`, `Label`

**Content:**
- `Card`, `Section`, `List`, `Callout`
- `CodeBlock`, `Quote`, `Badge`

**Diagrams:**
- `Arrow`, `StateNode`, `ProcessNode`, `DecisionNode`
- `MemoryBlock`, `ContextWindow`, `Timeline`

See [PRESENTATION_COMPONENTS.md](PRESENTATION_COMPONENTS.md) for details.

## Example

```tsx
import { Stack, Title, Card, Label } from 'diagram-dsl';

const Diagram = () => (
  <Stack gap={20} padding={40}>
    <Title level={1}>My Diagram</Title>
    <Card variant="primary" width={200} height={100}>
      <Label>Content</Label>
    </Card>
  </Stack>
);
```

## Documentation

- [Architecture](ARCHITECTURE.md)
- [Presentation Components](PRESENTATION_COMPONENTS.md)
- [Styling Guide](STYLING_GUIDE.md)
- [Examples](EXAMPLES.md)

## License

MIT
