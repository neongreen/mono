import React from 'react';
import { Slide, Title, Subtitle, Section, List, ProsCons, Highlight, Card, renderToSVG } from '../index';
import { writeFileSync } from 'fs';

// Example 1: Simple slide with title and list
const Example1 = () => (
  <Slide>
    <Title level={1}>New Presentation Components</Title>
    <Subtitle>Making slides easier to build</Subtitle>
    
    <Section title="New Components" variant="primary" width={1000} marginTop={20}>
      <List
        items={[
          'Slide - Consistent slide container',
          'List - Bullet point lists',
          'ProsCons - Side-by-side pros and cons',
          'Section - Titled content sections',
          'Highlight - Emphasized content blocks'
        ]}
        fontSize={13}
      />
    </Section>
  </Slide>
);

// Example 2: ProsCons demonstration
const Example2 = () => (
  <Slide>
    <Title level={1}>ProsCons Component</Title>
    
    <Card variant="default" width={1000} marginTop={20}>
      <ProsCons
        pros={[
          'Simple to implement',
          'Consistent styling',
          'Less code to write'
        ]}
        cons={[
          'Learning curve',
          'Less flexibility in some cases'
        ]}
      />
    </Card>
  </Slide>
);

// Example 3: Highlights
const Example3 = () => (
  <Slide>
    <Title level={1}>Highlight Component</Title>
    
    <Highlight variant="success" width={1000} marginTop={20}>
      <Title level={3}>Success!</Title>
      <List 
        items={['This is highlighted content', 'With automatic styling', 'Based on variant']}
        fontSize={12}
      />
    </Highlight>
    
    <Highlight variant="warning" width={1000} marginTop={20}>
      <Title level={3}>Warning</Title>
      <List 
        items={['Important information', 'Needs attention']}
        fontSize={12}
      />
    </Highlight>
  </Slide>
);

async function generate() {
  const examples = [
    { name: 'presentation-helpers-1', component: <Example1 /> },
    { name: 'presentation-helpers-2', component: <Example2 /> },
    { name: 'presentation-helpers-3', component: <Example3 /> },
  ];

  console.log('Generating presentation helper examples...\n');

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
