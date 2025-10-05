import React from 'react';
import { 
  Slide, Title, Subtitle, Section, List, FlowDiagram, TwoColumn, Callout, Badge,
  generateSlideDeck, numberSlides
} from '../index';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Define your slides as components
const IntroSlide = () => (
  <Slide alignItems="center" justifyContent="center">
    <Badge text="Example" variant="success" size="large" marginBottom={20} />
    <Title level={1}>Using generateSlideDeck</Title>
    <Subtitle>Simplified presentation generation</Subtitle>
  </Slide>
);

const BenefitsSlide = () => (
  <Slide>
    <Title level={1}>Benefits</Title>
    <Section title="Key Advantages" variant="primary" width={900} marginTop={20}>
      <List
        items={[
          'Automatic slide numbering',
          'Consistent HTML viewer generation',
          'Easy slide skipping',
          'Less boilerplate code'
        ]}
      />
    </Section>
  </Slide>
);

const FlowSlide = () => (
  <Slide>
    <Title level={1}>Development Flow</Title>
    <FlowDiagram
      steps={[
        { id: 'define', label: 'Define Slides', variant: 'primary' },
        { id: 'number', label: 'Number Slides', variant: 'secondary' },
        { id: 'generate', label: 'Generate Deck', variant: 'success' }
      ]}
      marginTop={40}
    />
  </Slide>
);

const CodeExample = () => (
  <Slide>
    <Title level={1}>Code Example</Title>
    <TwoColumn
      gap={24}
      marginTop={20}
      left={
        <Section title="Before" variant="danger" width={500}>
          <List
            items={[
              'Manual slide numbering',
              'Custom HTML generation',
              'Repetitive renderToSVG calls'
            ]}
            fontSize={12}
          />
        </Section>
      }
      right={
        <Section title="After" variant="success" width={500}>
          <List
            items={[
              'Automatic numbering',
              'Built-in HTML viewer',
              'Single function call'
            ]}
            fontSize={12}
          />
        </Section>
      }
    />
  </Slide>
);

// Main generation function
async function generate() {
  const slides = numberSlides([
    { name: 'intro', component: <IntroSlide /> },
    { name: 'benefits', component: <BenefitsSlide /> },
    { name: 'flow', component: <FlowSlide /> },
    { name: 'code-example', component: <CodeExample /> },
  ]);

  await generateSlideDeck(slides, {
    outputDir: join(__dirname, '../deck-helper-output'),
    htmlTitle: 'Slide Deck Helper Example',
    width: 1200,
    height: 800,
  });
}

generate();
