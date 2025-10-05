import React from 'react';
import { 
  Slide, Title, Subtitle, Section, List, ProsCons, Callout, 
  RichText, Spacer, Grid, Card, Text, renderToSVG 
} from '../index';
import { writeFileSync } from 'fs';

// Example with RichText
const RichTextExample = () => (
  <Slide>
    <Title level={1}>RichText Component</Title>
    <Subtitle>Mixed formatting in a single line</Subtitle>
    
    <Section title="Examples" variant="default" width={1000} marginTop={20}>
      <RichText
        segments={[
          'This is ',
          { text: 'bold text', bold: true },
          ' and this is ',
          { text: 'colored text', color: '#1976d2', bold: true },
          ' in one line.'
        ]}
        fontSize={14}
      />
      
      <Spacer size={12} />
      
      <RichText
        segments={[
          { text: 'Important:', bold: true, color: '#c62828' },
          ' Always test your changes before committing.'
        ]}
        fontSize={13}
      />
    </Section>
  </Slide>
);

// Example with Callout
const CalloutExample = () => (
  <Slide>
    <Title level={1}>Callout Component</Title>
    
    <Callout title="Key Takeaway" variant="success" width={1000} marginTop={20}>
      <Text fontSize={14}>
        The new presentation components make it much easier to create
        consistent and professional-looking slides.
      </Text>
    </Callout>
    
    <Spacer size={20} />
    
    <Callout title="Watch Out!" variant="warning" width={1000}>
      <List
        items={[
          'Always rebuild diagram-dsl after making changes',
          'Test your slides before generating the full deck',
          'Use consistent sizing across slides'
        ]}
        fontSize={12}
      />
    </Callout>
  </Slide>
);

// Example with Grid
const GridExample = () => (
  <Slide>
    <Title level={1}>Grid Layout</Title>
    
    <Grid columns={3} gap={20} marginTop={20}>
      <Card variant="primary" width={350}>
        <Title level={3}>Card 1</Title>
        <Text fontSize={12}>Content in grid layout</Text>
      </Card>
      
      <Card variant="secondary" width={350}>
        <Title level={3}>Card 2</Title>
        <Text fontSize={12}>Auto-arranged</Text>
      </Card>
      
      <Card variant="accent" width={350}>
        <Title level={3}>Card 3</Title>
        <Text fontSize={12}>Easy to use</Text>
      </Card>
      
      <Card variant="success" width={350}>
        <Title level={3}>Card 4</Title>
        <Text fontSize={12}>Flexible columns</Text>
      </Card>
      
      <Card variant="danger" width={350}>
        <Title level={3}>Card 5</Title>
        <Text fontSize={12}>Responsive design</Text>
      </Card>
    </Grid>
  </Slide>
);

// Comprehensive example
const ComprehensiveExample = () => (
  <Slide>
    <Title level={1}>All Components Together</Title>
    <Subtitle>A complete example</Subtitle>
    
    <Section title="Overview" variant="primary" width={1080} marginTop={16}>
      <RichText
        segments={[
          'These components work ',
          { text: 'seamlessly together', bold: true },
          ' to create beautiful presentations.'
        ]}
        fontSize={13}
      />
    </Section>
    
    <Spacer size={16} />
    
    <Grid columns={2} gap={16}>
      <Callout title="Benefits" variant="success" width={520}>
        <List
          items={['Less code', 'Consistency', 'Flexibility']}
          fontSize={11}
          gap={6}
        />
      </Callout>
      
      <Callout title="Usage" variant="info" width={520}>
        <List
          items={['Import from diagram-dsl', 'Compose components', 'Generate SVG']}
          fontSize={11}
          gap={6}
        />
      </Callout>
    </Grid>
  </Slide>
);

async function generate() {
  const examples = [
    { name: 'advanced-1-richtext', component: <RichTextExample /> },
    { name: 'advanced-2-callout', component: <CalloutExample /> },
    { name: 'advanced-3-grid', component: <GridExample /> },
    { name: 'advanced-4-comprehensive', component: <ComprehensiveExample /> },
  ];

  console.log('Generating advanced presentation examples...\n');

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
