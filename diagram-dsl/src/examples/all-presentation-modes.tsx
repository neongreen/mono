/**
 * Comprehensive example showcasing all three presentation modes:
 * 1. Traditional slides (left-to-right navigation)
 * 2. Scrolling slides (vertical scroll with slide gaps)
 * 3. Continuous page (pageless mode - no gaps, like Google Docs)
 * 
 * Also demonstrates theme switching capabilities
 */

import React from 'react';
import { writeFileSync, mkdirSync } from 'fs';
import { 
  Slide, Title, Subtitle, Text, Section, List, Card, Row, Stack,
  Badge, Divider, CodeBlock, Image, Spacer, Highlight, Well,
  Panel, Cluster, DataFlow,
  generateSlideDeck, generateScrollingPage, generateContinuousPage,
  setCurrentTheme,
  defaultTheme, darkTheme, nordTheme, githubTheme, draculaTheme,
  minimalTheme, professionalTheme, solarizedLightTheme
} from '../index';

// Sample slides content
const TitleSlideContent = () => (
  <Slide alignItems="center" justifyContent="center">
    <Badge text="Demo" variant="primary" size="large" marginBottom={20} />
    <Title level={1}>Presentation Modes Demo</Title>
    <Subtitle>Three ways to present your content</Subtitle>
    <Spacer size={40} />
    <Section title="Available Modes" variant="default" width={700}>
      <List
        items={[
          '📊 Slides Mode - Classic left-to-right navigation',
          '📜 Scrolling Slides - Vertical scroll with gaps',
          '📄 Continuous Page - Seamless flow (pageless)'
        ]}
        fontSize={14}
      />
    </Section>
  </Slide>
);

const ContentSlide1 = () => (
  <Slide>
    <Title level={1}>What is Context Management?</Title>
    <Subtitle>Understanding the fundamentals</Subtitle>
    
    <Row gap={24} marginTop={30}>
      <Card variant="primary" width={500}>
        <Title level={3}>The Challenge</Title>
        <List
          items={[
            'LLMs have limited context windows',
            'Long conversations exceed limits',
            'Costs scale with context size',
            'Need smart strategies'
          ]}
          fontSize={13}
        />
      </Card>
      
      <Card variant="success" width={500}>
        <Title level={3}>The Solution</Title>
        <List
          items={[
            'Selective context retention',
            'Intelligent summarization',
            'Vector-based retrieval',
            'Hybrid approaches'
          ]}
          fontSize={13}
        />
      </Card>
    </Row>
    
    <Well variant="info" width={1080} marginTop={30}>
      <Text fontSize={13} fontWeight="bold">💡 Key Insight</Text>
      <Text fontSize={12} marginTop={8}>
        The goal is to provide the most relevant information to the LLM
        while staying within token limits and managing costs effectively.
      </Text>
    </Well>
  </Slide>
);

const ContentSlide2 = () => (
  <Slide>
    <Title level={1}>Architecture Overview</Title>
    
    <Cluster title="Context Management System" variant="primary" width={1080} marginTop={30}>
      <DataFlow
        nodes={[
          { id: 'input', label: 'User Input', type: 'input', description: 'Conversation' },
          { id: 'manager', label: 'Context Manager', type: 'process', description: 'Decision Engine' },
          { id: 'storage', label: 'Vector DB', type: 'storage', description: 'Embeddings' },
          { id: 'llm', label: 'LLM', type: 'process', description: 'Generation' }
        ]}
        connections={[
          { from: 'input', to: 'manager', data: 'Message' },
          { from: 'manager', to: 'storage', data: 'Query' },
          { from: 'storage', to: 'manager', data: 'Context' },
          { from: 'manager', to: 'llm', data: 'Prompt' }
        ]}
        orientation="horizontal"
      />
    </Cluster>
  </Slide>
);

const CodeExampleSlide = () => (
  <Slide>
    <Title level={1}>Implementation Example</Title>
    
    <Panel header={<Text fontSize={14} fontWeight="bold">Context Manager in Python</Text>} 
           variant="secondary" width={900} marginTop={30}>
      <CodeBlock
        language="Python"
        code={[
          'class ContextManager:',
          '    def __init__(self, max_tokens=4000):',
          '        self.max_tokens = max_tokens',
          '        self.history = []',
          '    ',
          '    def add_message(self, message):',
          '        self.history.append(message)',
          '        self._trim_context()',
          '    ',
          '    def _trim_context(self):',
          '        # Keep only recent messages within limit',
          '        while self._count_tokens() > self.max_tokens:',
          '            self.history.pop(0)'
        ]}
        lineNumbers={true}
        width={850}
      />
    </Panel>
    
    <Badge text="Python 3.11+" variant="info" size="small" marginTop={20} />
  </Slide>
);

const ImageDemoSlide = () => (
  <Slide>
    <Title level={1}>Visual Content Support</Title>
    <Subtitle>Images and diagrams in your presentations</Subtitle>
    
    <Row gap={24} marginTop={30} justifyContent="center">
      <Image
        src="https://example.com/architecture-diagram.png"
        alt="System Architecture"
        width={400}
        height={300}
        borderRadius={8}
      />
      
      <Stack gap={16} width={500}>
        <Text fontSize={14} fontWeight="bold">Image Component Features:</Text>
        <List
          items={[
            'Support for external URLs',
            'Data URLs for embedded images',
            'Configurable dimensions',
            'Border radius control',
            'Alt text for accessibility'
          ]}
          fontSize={13}
        />
        <Text fontSize={12} color="#666" marginTop={10}>
          Note: Currently shows placeholders. Full image rendering coming soon!
        </Text>
      </Stack>
    </Row>
  </Slide>
);

const ThemeShowcaseSlide = () => (
  <Slide>
    <Title level={1}>Theme System</Title>
    <Subtitle>10 built-in themes to choose from</Subtitle>
    
    <Stack gap={16} marginTop={30}>
      <Highlight variant="info" width={1080}>
        <Text fontSize={14} fontWeight="bold">Available Themes:</Text>
        <Row gap={16} marginTop={12}>
          <Badge text="Default" variant="primary" size="medium" />
          <Badge text="Dark" variant="secondary" size="medium" />
          <Badge text="Professional" variant="default" size="medium" />
          <Badge text="Minimal" variant="default" size="medium" />
          <Badge text="Vibrant" variant="primary" size="medium" />
        </Row>
        <Row gap={16} marginTop={12}>
          <Badge text="Nord" variant="info" size="medium" />
          <Badge text="Dracula" variant="danger" size="medium" />
          <Badge text="GitHub" variant="default" size="medium" />
          <Badge text="Solarized" variant="warning" size="medium" />
          <Badge text="High Contrast" variant="default" size="medium" />
        </Row>
      </Highlight>
      
      <Divider width={1080} marginTop={16} marginBottom={16} />
      
      <Section title="Theme Usage" variant="primary" width={1080}>
        <CodeBlock
          language="TypeScript"
          code={[
            'import { setCurrentTheme, darkTheme } from "diagram-dsl";',
            '',
            '// Apply theme globally',
            'setCurrentTheme(darkTheme);',
            '',
            '// Or pass theme to generation functions',
            'await generateSlideDeck(slides, { theme: nordTheme });'
          ]}
          fontSize={12}
          width={1020}
        />
      </Section>
    </Stack>
  </Slide>
);

const SummarySlide = () => (
  <Slide alignItems="center" justifyContent="center">
    <Title level={1}>Choose Your Presentation Mode</Title>
    
    <Spacer size={40} />
    
    <Stack gap={24} width={900}>
      <Card variant="primary" width={900}>
        <Title level={3}>📊 Slides Mode</Title>
        <Text fontSize={13} marginTop={8}>
          Traditional slide deck with left-to-right navigation. Perfect for
          presentations with distinct sections and topics.
        </Text>
        <Badge text="generateSlideDeck()" variant="primary" size="small" marginTop={12} />
      </Card>
      
      <Card variant="secondary" width={900}>
        <Title level={3}>📜 Scrolling Slides</Title>
        <Text fontSize={13} marginTop={8}>
          Vertical scrolling with visible slide boundaries. Great for web-based
          presentations and documentation that needs clear section separation.
        </Text>
        <Badge text="generateScrollingPage()" variant="secondary" size="small" marginTop={12} />
      </Card>
      
      <Card variant="success" width={900}>
        <Title level={3}>📄 Continuous Page</Title>
        <Text fontSize={13} marginTop={8}>
          Seamless flow without slide boundaries, like Google Docs pageless mode.
          Ideal for long-form content, technical docs, and reports.
        </Text>
        <Badge text="generateContinuousPage()" variant="success" size="small" marginTop={12} />
      </Card>
    </Stack>
  </Slide>
);

// Generate examples in all three modes and multiple themes
async function generateExamples() {
  console.log('🎨 Generating presentation mode examples...\n');
  
  const outputDir = './presentation-modes-output';
  mkdirSync(outputDir, { recursive: true });
  
  const slideDefinitions = [
    { name: 'title', component: <TitleSlideContent /> },
    { name: 'content-1', component: <ContentSlide1 /> },
    { name: 'content-2', component: <ContentSlide2 /> },
    { name: 'code', component: <CodeExampleSlide /> },
    { name: 'images', component: <ImageDemoSlide /> },
    { name: 'themes', component: <ThemeShowcaseSlide /> },
    { name: 'summary', component: <SummarySlide /> },
  ];
  
  const contentBlocks = slideDefinitions.map(def => ({
    name: def.name,
    component: def.component,
    spacing: 0, // No extra spacing in continuous mode
  }));
  
  // Test with multiple themes
  const themes = [
    { name: 'default', theme: defaultTheme },
    { name: 'dark', theme: darkTheme },
    { name: 'nord', theme: nordTheme },
    { name: 'github', theme: githubTheme },
  ];
  
  for (const { name: themeName, theme } of themes) {
    console.log(`\n📚 Generating with ${theme.name} theme...\n`);
    setCurrentTheme(theme);
    
    // Mode 1: Traditional slides
    console.log(`  Generating traditional slides...`);
    const slideDeckSvgs = await generateSlideDeck(slideDefinitions, {
      theme,
      width: 1200,
      height: 800,
      backgroundColor: theme.background,
    });
    
    slideDeckSvgs.forEach((svg, index) => {
      const filename = `${outputDir}/mode1-slides-${themeName}-${index + 1}.svg`;
      writeFileSync(filename, svg);
    });
    console.log(`  ✓ Created ${slideDeckSvgs.length} slide files`);
    
    // Mode 2: Scrolling slides (vertical with gaps)
    console.log(`  Generating scrolling slides...`);
    const scrollingSvg = await generateScrollingPage(
      slideDefinitions.map(def => ({
        name: def.name,
        component: def.component,
      })),
      {
        theme,
        width: 1200,
        slideHeight: 800,
        gap: 60, // Gap between slides
        backgroundColor: theme.backgroundSecondary,
      }
    );
    
    const scrollingFilename = `${outputDir}/mode2-scrolling-${themeName}.svg`;
    writeFileSync(scrollingFilename, scrollingSvg);
    console.log(`  ✓ Created scrolling slides file`);
    
    // Mode 3: Continuous page (no gaps, pageless)
    console.log(`  Generating continuous page...`);
    const continuousSvg = await generateContinuousPage(contentBlocks, {
      theme,
      width: 1200,
      backgroundColor: theme.background,
      padding: 60,
      gap: 40, // Spacing between content blocks (much smaller than slides)
    });
    
    const continuousFilename = `${outputDir}/mode3-continuous-${themeName}.svg`;
    writeFileSync(continuousFilename, continuousSvg);
    console.log(`  ✓ Created continuous page file`);
  }
  
  console.log('\n✨ All examples generated successfully!');
  console.log(`\nOutput location: ${outputDir}/`);
  console.log(`\nGenerated files:`);
  console.log(`  • Traditional slides: mode1-slides-<theme>-<number>.svg`);
  console.log(`  • Scrolling slides: mode2-scrolling-<theme>.svg`);
  console.log(`  • Continuous pages: mode3-continuous-<theme>.svg`);
}

generateExamples().catch(console.error);
