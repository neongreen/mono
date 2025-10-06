# Technical Features for Software Engineers

diagram-dsl now includes powerful components specifically designed for technical presentations and documentation aimed at software engineers.

## Technical Components (5)

### 1. SequenceDiagram

Visualize message flow between actors over time. Perfect for API interactions, authentication flows, and microservices communication.

```tsx
<SequenceDiagram
  actors={[
    { id: 'user', name: 'User', type: 'user' },
    { id: 'api', name: 'API Server', type: 'service' },
    { id: 'db', name: 'Database', type: 'database' }
  ]}
  messages={[
    { from: 'user', to: 'api', message: 'GET /users/123', type: 'sync' },
    { from: 'api', to: 'db', message: 'SELECT * FROM users', type: 'sync' },
    { from: 'db', to: 'api', message: 'User data', type: 'return', style: 'dashed' },
    { from: 'api', to: 'user', message: '200 OK', type: 'return', style: 'dashed' }
  ]}
/>
```

**Features:**
- 4 actor types: user, service, database, system (with icons)
- 3 message types: sync, async, return
- Automatic lifeline rendering
- Message labels on arrows
- Dashed lines for return messages

**Use Cases:**
- OAuth/authentication flows
- API request/response cycles
- Microservices communication
- User journey mapping
- System interaction patterns

### 2. APIEndpoint

Document REST API endpoints with full request/response details.

```tsx
<APIEndpoint
  method="POST"
  path="/api/v1/users"
  description="Create a new user account"
  request={{
    params: { org_id: 'string (UUID)' },
    body: '{\n  "name": "John Doe",\n  "email": "john@example.com"\n}'
  }}
  response={{
    status: 201,
    body: '{\n  "id": "abc123",\n  "created_at": "2024-01-15T10:30:00Z"\n}'
  }}
/>
```

**Features:**
- Color-coded HTTP methods (GET=blue, POST=green, PUT=orange, DELETE=red, PATCH=purple)
- Method badges with colored backgrounds
- Monospace font for paths and code
- Separate request/response sections
- Status code badges (green for 2xx, red for errors)
- Dark code blocks for JSON payloads

**Use Cases:**
- API documentation
- OpenAPI/Swagger alternatives
- Developer onboarding docs
- API design presentations
- Technical specifications

### 3. Terminal

Display command-line interface with realistic terminal styling.

```tsx
<Terminal
  title="Production Deployment"
  commands={[
    'npm run build',
    'docker build -t myapp:latest .',
    'docker push registry.io/myapp:latest',
    'kubectl apply -f deployment.yaml',
    '# Deployment successful ✓'
  ]}
  theme="dark"
  prompt="$"
/>
```

**Features:**
- Realistic terminal window chrome (colored dots)
- Dark and light themes
- Customizable prompt symbol
- Monospace font (Monaco/Courier)
- Window title bar

**Use Cases:**
- Deployment instructions
- CLI tool documentation
- Setup/installation guides
- Debugging walkthroughs
- DevOps procedures

### 4. DataFlow

Visualize data processing pipelines and transformations.

```tsx
<DataFlow
  nodes={[
    { id: 'source', label: 'Data Source', type: 'input', description: 'CSV files' },
    { id: 'extract', label: 'Extract', type: 'process', description: 'Parse & validate' },
    { id: 'transform', label: 'Transform', type: 'process', description: 'Clean & enrich' },
    { id: 'load', label: 'Load', type: 'output', description: 'Insert to DB' },
    { id: 'storage', label: 'Database', type: 'storage', description: 'PostgreSQL' }
  ]}
  connections={[
    { from: 'source', to: 'extract', data: 'Raw data' },
    { from: 'extract', to: 'transform', data: 'Parsed records' },
    { from: 'transform', to: 'load', data: 'Clean data' },
    { from: 'load', to: 'storage', data: 'SQL inserts' }
  ]}
  orientation="horizontal"
/>
```

**Features:**
- 4 node types: input 📥, process ⚙️, output 📤, storage 💾
- Themed colors for each type
- Optional descriptions
- Data labels on connections
- Horizontal or vertical layout

**Use Cases:**
- ETL pipelines
- Data processing flows
- System integration diagrams
- Message queue patterns
- Stream processing

### 5. ComparisonTable

Display feature comparisons, benchmarks, and specifications in tabular format.

```tsx
<ComparisonTable
  columns={[
    { header: 'Feature', key: 'feature', width: 200, align: 'left' },
    { header: 'Basic', key: 'basic', width: 150, align: 'center' },
    { header: 'Pro', key: 'pro', width: 150, align: 'center' },
    { header: 'Enterprise', key: 'enterprise', width: 150, align: 'center' }
  ]}
  rows={[
    { feature: 'Users', basic: '10', pro: '100', enterprise: 'Unlimited' },
    { feature: 'API Access', basic: false, pro: true, enterprise: true },
    { feature: 'Support', basic: 'Community', pro: 'Email', enterprise: '24/7 Phone' },
    { feature: 'SLA', basic: false, pro: false, enterprise: true }
  ]}
  highlightColumn="pro"
  striped={true}
/>
```

**Features:**
- Boolean values shown as ✓/✗ (colored green/red)
- Column highlighting
- Optional striped rows
- Customizable column widths and alignment
- Professional table styling

**Use Cases:**
- Technology stack comparisons
- Feature tier matrices
- Performance benchmarks
- Tool evaluations
- Pricing tables

## Scrolling Page Mode

In addition to slide decks, diagram-dsl now supports generating single-page scrolling documentation.

### Why Scrolling Mode?

**Slides are great for:**
- Presentations
- Talks
- Step-by-step walkthroughs
- Discrete topics

**Scrolling pages are great for:**
- Technical documentation
- Long-form content
- Reference materials
- Comprehensive guides
- Self-paced learning

### Usage

```tsx
import { generateScrollingPage } from 'diagram-dsl';

const sections = [
  { name: 'introduction', component: <IntroSection /> },
  { name: 'architecture', component: <ArchitectureSection /> },
  { name: 'api-reference', component: <APISection /> },
  { name: 'deployment', component: <DeploymentSection /> }
];

await generateScrollingPage(sections, {
  outputDir: './docs',
  htmlTitle: 'Technical Documentation',
  width: 1200,
  sectionGap: 60
});
```

### Features

**Navigation:**
- Sidebar table of contents
- Smooth scrolling between sections
- Active section highlighting
- Back-to-top button
- Keyboard shortcuts (Home/End)

**Styling:**
- Professional sidebar design
- Boxed content sections with shadows
- Responsive layout
- Print-friendly CSS
- Clean, modern appearance

**Developer Experience:**
- Same components work for both modes
- Single source of truth
- Generate both formats from same content
- No duplication needed

## Complete Example

Here's a comprehensive technical presentation using all features:

```tsx
import { 
  Slide, Title, SequenceDiagram, APIEndpoint, Terminal, 
  DataFlow, ComparisonTable, generateSlideDeck, generateScrollingPage 
} from 'diagram-dsl';

// Define your content once
const sections = [
  { 
    name: 'architecture',
    component: (
      <Slide>
        <Title level={1}>System Architecture</Title>
        <DataFlow nodes={...} connections={...} />
      </Slide>
    )
  },
  {
    name: 'api',
    component: (
      <Slide>
        <Title level={1}>API Documentation</Title>
        <APIEndpoint method="GET" path="/api/users" {...} />
        <APIEndpoint method="POST" path="/api/users" {...} />
      </Slide>
    )
  },
  {
    name: 'auth-flow',
    component: (
      <Slide>
        <Title level={1}>Authentication</Title>
        <SequenceDiagram actors={...} messages={...} />
      </Slide>
    )
  },
  {
    name: 'deployment',
    component: (
      <Slide>
        <Title level={1}>Deployment</Title>
        <Terminal commands={['docker build...', 'kubectl apply...']} />
      </Slide>
    )
  },
  {
    name: 'comparison',
    component: (
      <Slide>
        <Title level={1}>Technology Comparison</Title>
        <ComparisonTable columns={...} rows={...} />
      </Slide>
    )
  }
];

// Generate both formats
ccccccccccccccccumberSlides(sections);
await generateSlideDeck(slides, { outputDir: './slides' });
await generateScrollingPage(sections,await generat 'awaicsawa);aw``

## Benefits for## Benefits foram## Benefits ngle Source of Truth
Write your technical content once, generate both presentation slides and documentation.

### 2. Version Control Friendly
All content is code (TSX), stored in Git with full history and diff capabilities.

### 3. Automated Generation
CI/CD pipelines can regenerate docs on every commit. No manual export/save steps.

### 4. Type Safety
Full TypeScript support catches errors before generation. No broken links or missing data.

### 5. Consistent Styling
All diagrams and components use consistent colors, fonts, and spacing automatically.

### 6. Developer-Focused
Components designed specifically for technical concepts engineers work with daily.

### 7. Flexible Output
Switch between presentation and documentation modes without rewriting content.

## Use Cases

### Architecture Documentation
- System diagrams with DataFlow
- Service interactions with SequenceDiagram
- Technology comparisons with ComparisonTable

### API Documentation
- Endpoint documentation with APIEndpoint
- Authentication flows with SequenceDiagram
- Example requests in Terminal

### Onboarding Materials
- Setup instructions in Terminal
- Architecture overview in DataFlow
- Feature comparisons in ComparisonTable

### Technical Presentations
- Conference talks (slide mode)
- Brown bag sessions (slide mode)
- Reference docs (scrolling mode)

### Design Documents
- System design proposals
- RFC-style technical specs
- Architecture decision records (ADRs)

## Comparison: Before vs. After

### Before (Traditional Tools)

**PowerPoint/Google Slides:**
- Manual diagram creation
- Inconsistent styling
- Hard to version control
- Difficult to maintain
- Export to PDF manually

**Mermaid/PlantUML:**
- Text-based (good!)
- Limited styling options
- No presentation mode
- Separate tools for different diagram types

**Confluence/Notion:**
- Good for docs
- Poor for presentations
- Vendor lock-in
- Limited diagram capabilities

### After (diagram-dsl)

**Single Tool:**
- All diagram types
- Presentations AND docs
- Consistent styling
- Full type safety
- Version controlled
- CI/CD friendly
- Professional output
- No manual work

## Performance

Generating presentations is fast:

```
8-slide presentation with complex diagrams: ~2 seconds
20-section scrolling page: ~5 seconds
50+ slides: ~15 seconds
```

All generation happens at build time, so runtime performance is instant (static SVG files).
