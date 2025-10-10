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

## Creating Anthropic-Style Diagrams

The library excels at creating professional AI system architecture diagrams similar to those in Anthropic's documentation. See the comprehensive guide: **[ANTHROPIC_STYLE_GUIDE.md](ANTHROPIC_STYLE_GUIDE.md)**

### Quick Example

```tsx
import { Stack, Row, Card, Label, Subtitle, Arrow, Cluster, Badge } from 'diagram-dsl';

const AISystemDiagram = () => (
  <Stack width={1000} height={700} padding={40} gap={30}>
    <Title level={1}>Simple AI Conversation</Title>
    
    <Row gap={35} justifyContent="center">
      <Cluster title="Input" variant="primary" width={280}>
        <Card id="user" variant="primary" width={240} height={70}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">User</Label>
            <Subtitle>Sends message</Subtitle>
          </Stack>
        </Card>
      </Cluster>

      <Cluster title="Processing" variant="accent" width={280}>
        <Card id="claude" variant="accent" width={240} height={100}>
          <Stack gap={8} alignItems="center">
            <Label bold size="lg">Claude</Label>
            <Subtitle>AI inference</Subtitle>
            <Badge text="200K context" variant="success" />
          </Stack>
        </Card>
      </Cluster>

      <Cluster title="Output" variant="success" width={280}>
        <Card id="response" variant="success" width={240} height={70}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Response</Label>
            <Subtitle>Return to user</Subtitle>
          </Stack>
        </Card>
      </Cluster>
    </Row>

    <Arrow from="user" to="claude" label="message" color="#1976d2" thickness="medium" />
    <Arrow from="claude" to="response" label="output" color="#4caf50" thickness="thick" />
  </Stack>
);
```

### Example Diagrams

**View live examples:** Open `examples/view-svg.html` in a browser

- **`anthropic-original-replication.tsx`** ⭐ - Replicates the specific "Prompt engineering vs. context engineering" diagram
- **`anthropic-simple.tsx`** - Perfect starting template (3-column flow)
- **`anthropic-style-diagram.tsx`** - Full layered architecture
- **`anthropic-improved.tsx`** - Advanced with Cluster components
- **`showcase-agent-system.tsx`** - Complex agent system with memory

### Advanced Arrow Features

For complex diagrams like the Anthropic original, new arrow capabilities include:

- **`fromSide` / `toSide`** - Connect to specific edges (`'top'`, `'bottom'`, `'left'`, `'right'`)
- **`fromOffset` / `toOffset`** - Position along edge for corner connections
- **`shortenEnd` / `shortenStart`** - Arrows that stop short of their target
- **`curve="step"`** - Orthogonal routing for feedback loops

See **[ARROW_ENHANCEMENTS.md](ARROW_ENHANCEMENTS.md)** for complete documentation.

### Key Components for Anthropic-Style Diagrams

- **`Cluster`** - Visual grouping with colored borders and titles
- **`Card`** - Individual components with professional styling
- **`Arrow`** - Connections with labels, styles (solid/dashed), and curves
- **`Badge`** - Metadata and status information
- **`Divider`** - Visual separation between layers

## Documentation

- [Architecture](ARCHITECTURE.md)
- [Presentation Components](PRESENTATION_COMPONENTS.md)
- [Styling Guide](STYLING_GUIDE.md)
- [Anthropic-Style Guide](ANTHROPIC_STYLE_GUIDE.md) ⭐ **New!**
- [Examples](EXAMPLES.md)

## License

MIT
