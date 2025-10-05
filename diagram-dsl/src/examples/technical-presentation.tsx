import React from 'react';
import { 
  Slide, Title, Subtitle, Text, Section, Divider, Badge, Well,
  SequenceDiagram, APIEndpoint, Terminal, DataFlow, ComparisonTable,
  CodeBlock, List, Panel, Cluster,
  generateSlideDeck, numberSlides, generateScrollingPage
} from '../index';

// Title Slide
const TitleSlide = () => (
  <Slide alignItems="center" justifyContent="center">
    <Badge text="Technical" variant="primary" size="large" marginBottom={20} />
    <Title level={1}>Microservices Architecture</Title>
    <Subtitle>Design Patterns & Best Practices</Subtitle>
    <Text fontSize={14} color="#666" marginTop={20}>
      A comprehensive guide for software engineers
    </Text>
  </Slide>
);

// Architecture Overview
const ArchitectureSlide = () => (
  <Slide>
    <Title level={1}>System Architecture</Title>
    
    <Cluster title="Microservices Ecosystem" variant="primary" width={1080} marginTop={30}>
      <DataFlow
        nodes={[
          { id: 'client', label: 'Client Apps', type: 'input', description: 'Web & Mobile' },
          { id: 'gateway', label: 'API Gateway', type: 'process', description: 'Kong/Nginx' },
          { id: 'services', label: 'Services', type: 'process', description: 'Business Logic' },
          { id: 'db', label: 'Databases', type: 'storage', description: 'PostgreSQL' }
        ]}
        connections={[
          { from: 'client', to: 'gateway', data: 'HTTPS' },
          { from: 'gateway', to: 'services', data: 'gRPC' },
          { from: 'services', to: 'db', data: 'SQL' }
        ]}
        orientation="horizontal"
      />
    </Cluster>
  </Slide>
);

// API Documentation
const APISlide = () => (
  <Slide>
    <Title level={1}>REST API Endpoints</Title>
    
    <APIEndpoint
      method="GET"
      path="/api/v1/users/:id"
      description="Retrieve user information by ID"
      request={{
        params: { id: 'string (UUID)' }
      }}
      response={{
        status: 200,
        body: '{\n  "id": "123",\n  "name": "John Doe",\n  "email": "john@example.com"\n}'
      }}
      marginTop={20}
    />
    
    <APIEndpoint
      method="POST"
      path="/api/v1/users"
      description="Create a new user"
      request={{
        body: '{\n  "name": "Jane Doe",\n  "email": "jane@example.com",\n  "role": "developer"\n}'
      }}
      response={{
        status: 201,
        body: '{\n  "id": "456",\n  "name": "Jane Doe",\n  "created_at": "2024-01-15T10:30:00Z"\n}'
      }}
      marginTop={20}
      width={800}
    />
  </Slide>
);

// Sequence Diagram
const SequenceSlide = () => (
  <Slide>
    <Title level={1}>Authentication Flow</Title>
    <Subtitle>OAuth 2.0 Authorization Code Flow</Subtitle>
    
    <SequenceDiagram
      actors={[
        { id: 'user', name: 'User', type: 'user' },
        { id: 'client', name: 'Client App', type: 'service' },
        { id: 'auth', name: 'Auth Server', type: 'service' },
        { id: 'api', name: 'API Server', type: 'system' }
      ]}
      messages={[
        { from: 'user', to: 'client', message: 'Login request', type: 'sync' },
        { from: 'client', to: 'auth', message: 'Authorization request', type: 'sync' },
        { from: 'auth', to: 'user', message: 'Login page', type: 'return', style: 'dashed' },
        { from: 'user', to: 'auth', message: 'Credentials', type: 'sync' },
        { from: 'auth', to: 'client', message: 'Auth code', type: 'return', style: 'dashed' },
        { from: 'client', to: 'auth', message: 'Token request', type: 'sync' },
        { from: 'auth', to: 'client', message: 'Access token', type: 'return', style: 'dashed' },
        { from: 'client', to: 'api', message: 'API call + token', type: 'sync' },
        { from: 'api', to: 'client', message: 'Protected resource', type: 'return', style: 'dashed' }
      ]}
      width={1080}
      marginTop={30}
    />
  </Slide>
);

// Command Line Examples
const CLISlide = () => (
  <Slide>
    <Title level={1}>Deployment Commands</Title>
    
    <Terminal
      title="Production Deployment"
      commands={[
        'docker build -t myapp:latest .',
        'docker tag myapp:latest registry.io/myapp:v1.2.3',
        'docker push registry.io/myapp:v1.2.3',
        '',
        'kubectl set image deployment/myapp myapp=registry.io/myapp:v1.2.3',
        'kubectl rollout status deployment/myapp',
        '',
        '# Deployment successful ✓'
      ]}
      theme="dark"
      width={900}
      marginTop={30}
    />
    
    <Well variant="info" width={900} marginTop={20}>
      <Text fontSize={13} fontWeight="bold" marginBottom={8}>💡 Pro Tip</Text>
      <Text fontSize={12}>
        Always test in staging environment before deploying to production.
        Use blue-green deployment strategy for zero-downtime updates.
      </Text>
    </Well>
  </Slide>
);

// Technology Comparison
const ComparisonSlide = () => (
  <Slide>
    <Title level={1}>Technology Stack Comparison</Title>
    
    <ComparisonTable
      columns={[
        { header: 'Feature', key: 'feature', width: 200, align: 'left' },
        { header: 'Node.js', key: 'nodejs', width: 150, align: 'center' },
        { header: 'Go', key: 'go', width: 150, align: 'center' },
        { header: 'Rust', key: 'rust', width: 150, align: 'center' }
      ]}
      rows={[
        { feature: 'Performance', nodejs: '⭐⭐⭐', go: '⭐⭐⭐⭐', rust: '⭐⭐⭐⭐⭐' },
        { feature: 'Learning Curve', nodejs: 'Easy', go: 'Medium', rust: 'Hard' },
        { feature: 'Async Support', nodejs: true, go: true, rust: true },
        { feature: 'Memory Safety', nodejs: false, go: false, rust: true },
        { feature: 'Concurrency', nodejs: 'Event Loop', go: 'Goroutines', rust: 'Tokio' },
        { feature: 'Package Manager', nodejs: 'npm', go: 'go mod', rust: 'cargo' },
        { feature: 'Enterprise Ready', nodejs: true, go: true, rust: true }
      ]}
      highlightColumn="go"
      width={900}
      marginTop={30}
    />
    
    <Badge text="Go Selected" variant="success" size="medium" marginTop={20} />
  </Slide>
);

// Code Example
const CodeSlide = () => (
  <Slide>
    <Title level={1}>Implementation Example</Title>
    
    <Panel
      header={<Text fontSize={14} fontWeight="bold">Go Microservice Handler</Text>}
      variant="secondary"
      width={900}
      marginTop={30}
    >
      <CodeBlock
        language="Go"
        code={[
          'func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {',
          '    id := chi.URLParam(r, "id")',
          '    ',
          '    user, err := h.userService.GetByID(r.Context(), id)',
          '    if err != nil {',
          '        http.Error(w, err.Error(), http.StatusNotFound)',
          '        return',
          '    }',
          '    ',
          '    json.NewEncoder(w).Encode(user)',
          '}'
        ]}
        lineNumbers={true}
        width={850}
      />
    </Panel>
  </Slide>
);

// Best Practices
const BestPracticesSlide = () => (
  <Slide>
    <Title level={1}>Best Practices</Title>
    
    <Section title="Design Principles" variant="primary" width={900} marginTop={30}>
      <List
        items={[
          'Single Responsibility: Each service does one thing well',
          'API First: Design APIs before implementation',
          'Loose Coupling: Services communicate through well-defined interfaces',
          'Fault Tolerance: Implement circuit breakers and retry logic',
          'Observability: Comprehensive logging, metrics, and tracing'
        ]}
        fontSize={13}
      />
    </Section>
    
    <Divider width={900} marginTop={24} marginBottom={24} />
    
    <Section title="Security Checklist" variant="danger" width={900}>
      <List
        items={[
          '✓ Use HTTPS for all communications',
          '✓ Implement authentication & authorization',
          '✓ Validate and sanitize all inputs',
          '✓ Keep dependencies updated',
          '✓ Use secrets management (Vault, AWS Secrets Manager)'
        ]}
        fontSize={13}
      />
    </Section>
  </Slide>
);

// Generate both formats
async function generate() {
  const sections = [
    { name: 'title', component: <TitleSlide /> },
    { name: 'architecture', component: <ArchitectureSlide /> },
    { name: 'api', component: <APISlide /> },
    { name: 'sequence', component: <SequenceSlide /> },
    { name: 'cli', component: <CLISlide /> },
    { name: 'comparison', component: <ComparisonSlide /> },
    { name: 'code', component: <CodeSlide /> },
    { name: 'best-practices', component: <BestPracticesSlide /> }
  ];

  console.log('=== Generating Technical Presentation ===\n');
  
  // Generate as slides
  console.log('1. Generating slide deck...');
  const slides = numberSlides(sections);
  await generateSlideDeck(slides, {
    outputDir: './technical-slides',
    htmlTitle: 'Microservices Architecture - Slides',
    width: 1200,
    height: 800
  });
  
  // Generate as scrolling page
  console.log('\n2. Generating scrolling page...');
  await generateScrollingPage(sections, {
    outputDir: './technical-scrolling',
    htmlTitle: 'Microservices Architecture - Documentation',
    width: 1200,
    sectionGap: 60
  });
  
  console.log('\n=== Generation Complete! ===\n');
  console.log('Slides:    ./technical-slides/index.html');
  console.log('Scrolling: ./technical-scrolling/index.html\n');
}

generate();
