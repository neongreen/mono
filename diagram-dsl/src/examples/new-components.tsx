import React from 'react';
import { 
  Slide, Title, Subtitle, Text, 
  FlowDiagram, TwoColumn, ThreeColumn, CodeBlock, Quote, Badge, Divider,
  Section, List, Callout,
  renderToSVG 
} from '../index';
import { writeFileSync } from 'fs';

// Example 1: FlowDiagram
const FlowDiagramExample = () => (
  <Slide>
    <Title level={1}>FlowDiagram Component</Title>
    <Subtitle>Automatic flow charts with arrows</Subtitle>
    
    <FlowDiagram
      steps={[
        { id: 'input', label: 'User Input', subtitle: 'Query arrives', variant: 'primary' },
        { id: 'process', label: 'Process', subtitle: 'Analyze request', variant: 'secondary' },
        { id: 'retrieve', label: 'Retrieve', subtitle: 'Get context', variant: 'accent' },
        { id: 'generate', label: 'Generate', subtitle: 'Create response', variant: 'success' }
      ]}
      marginTop={40}
    />
    
    <Callout title="Key Benefits" variant="info" width={900} marginTop={40}>
      <List
        items={[
          'Automatic arrow connections between steps',
          'No manual positioning needed',
          'Consistent styling with variants'
        ]}
        fontSize={12}
      />
    </Callout>
  </Slide>
);

// Example 2: TwoColumn and ThreeColumn
const ColumnLayoutsExample = () => (
  <Slide>
    <Title level={1}>Column Layouts</Title>
    <Subtitle>Flexible multi-column layouts</Subtitle>
    
    <Section title="Two Column Layout" variant="primary" width={1080} marginTop={20}>
      <TwoColumn
        left={
          <Text fontSize={13}>
            Left column content automatically takes up 50% of the available space.
            Perfect for side-by-side comparisons.
          </Text>
        }
        right={
          <Text fontSize={13}>
            Right column content. Both columns grow equally to fill the container.
            No manual width calculations needed!
          </Text>
        }
      />
    </Section>
    
    <Section title="Three Column Layout" variant="secondary" width={1080} marginTop={20}>
      <ThreeColumn
        left={<Text fontSize={12}>Column 1: Automatically sized</Text>}
        center={<Text fontSize={12}>Column 2: All equal width</Text>}
        right={<Text fontSize={12}>Column 3: No calculations</Text>}
      />
    </Section>
  </Slide>
);

// Example 3: CodeBlock
const CodeBlockExample = () => (
  <Slide>
    <Title level={1}>CodeBlock Component</Title>
    <Subtitle>Display code with syntax highlighting</Subtitle>
    
    <CodeBlock
      language="TypeScript"
      code={[
        'function createPresentation() {',
        '  return (',
        '    <Slide>',
        '      <Title>My Slide</Title>',
        '      <Text>Content here</Text>',
        '    </Slide>',
        '  );',
        '}'
      ]}
      lineNumbers={true}
      width={800}
      marginTop={20}
    />
    
    <Callout title="Features" variant="success" width={800} marginTop={20}>
      <List
        items={[
          'Monospace font for code',
          'Optional line numbers',
          'Language label support',
          'Dark theme by default'
        ]}
        fontSize={12}
      />
    </Callout>
  </Slide>
);

// Example 4: Quote and Badge
const QuoteAndBadgeExample = () => (
  <Slide>
    <Title level={1}>Quote & Badge Components</Title>
    
    <Quote
      text="These new components make presentations 10x easier to create."
      author="Happy Developer"
      variant="primary"
      width={900}
      marginTop={20}
    />
    
    <Divider width={900} marginTop={30} marginBottom={30} />
    
    <Section title="Badges for Labels" variant="default" width={900}>
      <Text fontSize={14} marginBottom={16}>Use badges to highlight important information:</Text>
      <TwoColumn
        left={
          <>
            <Badge text="NEW" variant="success" size="small" />
            <Text fontSize={12} marginLeft={8}>FlowDiagram component</Text>
          </>
        }
        right={
          <>
            <Badge text="v2.0" variant="info" size="small" />
            <Text fontSize={12} marginLeft={8}>Latest release</Text>
          </>
        }
      />
    </Section>
  </Slide>
);

// Example 5: Divider
const DividerExample = () => (
  <Slide>
    <Title level={1}>Divider Component</Title>
    <Subtitle>Visual separators for content</Subtitle>
    
    <Section title="Section 1" variant="primary" width={900} marginTop={20}>
      <Text fontSize={13}>First section content</Text>
    </Section>
    
    <Divider width={900} />
    
    <Section title="Section 2" variant="secondary" width={900}>
      <Text fontSize={13}>Second section content</Text>
    </Section>
    
    <Divider width={900} variant="dashed" />
    
    <Section title="Section 3" variant="accent" width={900}>
      <Text fontSize={13}>Third section content with dashed divider above</Text>
    </Section>
  </Slide>
);

// Example 6: Comprehensive example
const ComprehensiveExample = () => (
  <Slide>
    <Title level={1}>All Components Together</Title>
    
    <Badge text="NEW" variant="success" size="medium" marginBottom={20} />
    
    <ThreeColumn gap={16} marginTop={20}>
      <Section title="Flow" variant="primary">
        <Text fontSize={11}>FlowDiagram for processes</Text>
      </Section>
      <Section title="Layout" variant="secondary">
        <Text fontSize={11}>Column layouts for structure</Text>
      </Section>
      <Section title="Content" variant="accent">
        <Text fontSize={11}>Code, quotes, badges</Text>
      </Section>
    </ThreeColumn>
    
    <Divider width={1080} marginTop={20} marginBottom={20} />
    
    <TwoColumn gap={24}>
      <CodeBlock
        code={['const x = 1;', 'const y = 2;', 'return x + y;']}
        language="JavaScript"
        width={500}
      />
      <Quote
        text="Clean, composable, powerful"
        author="diagram-dsl"
        variant="accent"
        width={500}
      />
    </TwoColumn>
  </Slide>
);

async function generate() {
  const examples = [
    { name: 'new-1-flowdiagram', component: <FlowDiagramExample /> },
    { name: 'new-2-columns', component: <ColumnLayoutsExample /> },
    { name: 'new-3-codeblock', component: <CodeBlockExample /> },
    { name: 'new-4-quote-badge', component: <QuoteAndBadgeExample /> },
    { name: 'new-5-divider', component: <DividerExample /> },
    { name: 'new-6-comprehensive', component: <ComprehensiveExample /> },
  ];

  console.log('Generating new components examples...\n');

  for (const example of examples) {
    const svg = await renderToSVG(example.component, {
      width: 1200,
      height: 800,
      backgroundColor: 'white',
    });
    
    const filename = `${example.name}.svg`;
    writeFileSync(filename, svg);
    console.log(`✓ Generated ${filename}`);
  }
  
  console.log('\nDone!');
}

generate();
