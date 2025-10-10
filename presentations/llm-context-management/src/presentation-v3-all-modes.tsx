/**
 * LLM Context Management Presentation - Version 3
 * 
 * This version showcases all three presentation modes and multiple themes
 * Demonstrates the full power of diagram-dsl's presentation system
 */

import React from 'react';
import { writeFileSync, mkdirSync } from 'fs';
import { 
  Slide, Stack, Row, Box, Text, Arrow, Card, Title, Subtitle,
  List, ProsCons, Section, Callout, RichText, Spacer, Badge,
  Highlight, Well, Panel, CodeBlock, Image, Divider, Steps,
  DataFlow, ComparisonTable, SequenceDiagram,
  generateSlideDeck, generateScrollingPage, generateContinuousPage,
  setCurrentTheme,
  defaultTheme, darkTheme, nordTheme, draculaTheme, professionalTheme
} from 'diagram-dsl';

// Slide 1: Title
const TitleSlide = () => (
  <Slide alignItems="center" justifyContent="center">
    <Badge text="LLM Systems" variant="primary" size="large" marginBottom={20} />
    <Title level={1}>Context Management Strategies</Title>
    <Subtitle>in Large Language Model Agent Implementations</Subtitle>
    
    <Spacer size={60} />
    
    <Section title="What You'll Learn" variant="default" width={700}>
      <List
        items={[
          'Core challenges of context window limitations',
          'Strategic approaches to context management',
          'Memory architectures for long conversations',
          'Practical implementation patterns'
        ]}
        fontSize={14}
      />
    </Section>
  </Slide>
);

// Slide 2: The Problem
const ProblemSlide = () => (
  <Slide>
    <Title level={1}>The Context Challenge</Title>
    <Subtitle>Why context management matters</Subtitle>
    
    <Row gap={24} marginTop={30}>
      <Card id="constraints" variant="danger" width={500}>
        <Title level={3}>⚠️ The Constraints</Title>
        <List items={[
          'Token limits: 4K-200K tokens',
          'Cost scales with context size',
          'Latency increases with length',
          'Information can be lost'
        ]} fontSize={13} />
      </Card>

      <Card id="needs" variant="primary" width={500}>
        <Title level={3}>✅ What We Need</Title>
        <List items={[
          'Remember past interactions',
          'Access relevant information',
          'Stay within token limits',
          'Maintain conversation quality'
        ]} fontSize={13} />
      </Card>
    </Row>

    <Well variant="warning" width={1080} marginTop={30}>
      <Text fontSize={14} fontWeight="bold">💡 Key Insight</Text>
      <Text fontSize={13} marginTop={8}>
        The art of context management is selecting the RIGHT information, 
        not just cramming in as much as possible.
      </Text>
    </Well>
  </Slide>
);

// Slide 3: Strategy Overview
const StrategyOverview = () => (
  <Slide>
    <Title level={1}>Four Core Strategies</Title>
    
    <Stack gap={20} marginTop={30}>
      <Card variant="primary" width={1080}>
        <Row gap={20}>
          <Badge text="1" variant="primary" size="large" />
          <Stack gap={8}>
            <Title level={3}>Sliding Window</Title>
            <Text fontSize={13}>Keep only the N most recent messages</Text>
          </Stack>
        </Row>
      </Card>
      
      <Card variant="secondary" width={1080}>
        <Row gap={20}>
          <Badge text="2" variant="secondary" size="large" />
          <Stack gap={8}>
            <Title level={3}>Summarization</Title>
            <Text fontSize={13}>Compress old context into summaries</Text>
          </Stack>
        </Row>
      </Card>
      
      <Card variant="info" width={1080}>
        <Row gap={20}>
          <Badge text="3" variant="info" size="large" />
          <Stack gap={8}>
            <Title level={3}>Vector Retrieval (RAG)</Title>
            <Text fontSize={13}>Retrieve relevant context from embeddings</Text>
          </Stack>
        </Row>
      </Card>
      
      <Card variant="success" width={1080}>
        <Row gap={20}>
          <Badge text="4" variant="success" size="large" />
          <Stack gap={8}>
            <Title level={3}>Hybrid Approach</Title>
            <Text fontSize={13}>Combine multiple strategies intelligently</Text>
          </Stack>
        </Row>
      </Card>
    </Stack>
  </Slide>
);

// Slide 4: Sliding Window
const SlidingWindowSlide = () => (
  <Slide>
    <Title level={1}>Strategy 1: Sliding Window</Title>
    <Subtitle>Simple and effective for short-term memory</Subtitle>
    
    <Panel 
      header={<Text fontSize={14} fontWeight="bold">Concept</Text>}
      variant="primary"
      width={1080}
      marginTop={30}
    >
      <Text fontSize={13} marginBottom={16}>
        Keep only the most recent N messages in the context. As new messages arrive,
        old ones are dropped. Like a moving window over the conversation history.
      </Text>
      
      <CodeBlock
        language="Python"
        code={[
          'class SlidingWindowContext:',
          '    def __init__(self, max_messages=10):',
          '        self.messages = []',
          '        self.max_messages = max_messages',
          '    ',
          '    def add_message(self, message):',
          '        self.messages.append(message)',
          '        if len(self.messages) > self.max_messages:',
          '            self.messages.pop(0)  # Remove oldest',
          '    ',
          '    def get_context(self):',
          '        return self.messages'
        ]}
        lineNumbers={true}
        fontSize={11}
        width={1020}
      />
    </Panel>
    
    <Row gap={24} marginTop={24}>
      <Highlight variant="success" width={520}>
        <Text fontSize={12} fontWeight="bold">✅ Pros</Text>
        <List items={[
          'Simple to implement',
          'Predictable token usage',
          'Low computational cost',
          'Works well for recent context'
        ]} fontSize={11} />
      </Highlight>
      
      <Highlight variant="danger" width={520}>
        <Text fontSize={12} fontWeight="bold">❌ Cons</Text>
        <List items={[
          'Loses all old information',
          'No long-term memory',
          'May drop important context',
          'Not suitable for complex tasks'
        ]} fontSize={11} />
      </Highlight>
    </Row>
  </Slide>
);

// Slide 5: Summarization
const SummarizationSlide = () => (
  <Slide>
    <Title level={1}>Strategy 2: Summarization</Title>
    <Subtitle>Compress history while retaining key information</Subtitle>
    
    <DataFlow
      nodes={[
        { id: 'old', label: 'Old Messages', type: 'storage', description: '100 messages' },
        { id: 'llm', label: 'LLM Summarizer', type: 'process', description: 'GPT-4' },
        { id: 'summary', label: 'Summary', type: 'storage', description: '5 messages' },
        { id: 'recent', label: 'Recent', type: 'input', description: '10 messages' },
        { id: 'context', label: 'Final Context', type: 'output', description: '15 messages' }
      ]}
      connections={[
        { from: 'old', to: 'llm', data: 'Compress' },
        { from: 'llm', to: 'summary', data: 'Summary' },
        { from: 'summary', to: 'context', data: 'Include' },
        { from: 'recent', to: 'context', data: 'Include' }
      ]}
      orientation="horizontal"
      width={1080}
      marginTop={30}
    />
    
    <ComparisonTable
      columns={[
        { header: 'Aspect', key: 'aspect', width: 200, align: 'left' },
        { header: 'Before', key: 'before', width: 300, align: 'left' },
        { header: 'After', key: 'after', width: 300, align: 'left' }
      ]}
      rows={[
        { aspect: 'Messages', before: '100 messages', after: '15 messages' },
        { aspect: 'Tokens', before: '~50K tokens', after: '~5K tokens' },
        { aspect: 'Cost', before: '$0.50', after: '$0.05' },
        { aspect: 'Information', before: '100%', after: '~80% key points' }
      ]}
      width={900}
      marginTop={30}
    />
  </Slide>
);

// Slide 6: RAG
const RAGSlide = () => (
  <Slide>
    <Title level={1}>Strategy 3: RAG (Retrieval Augmented Generation)</Title>
    <Subtitle>Retrieve only relevant context on-demand</Subtitle>
    
    <SequenceDiagram
      actors={[
        { id: 'user', name: 'User', type: 'user' },
        { id: 'agent', name: 'Agent', type: 'service' },
        { id: 'vector', name: 'Vector DB', type: 'system' },
        { id: 'llm', name: 'LLM', type: 'service' }
      ]}
      messages={[
        { from: 'user', to: 'agent', message: 'New question', type: 'sync' },
        { from: 'agent', to: 'vector', message: 'Query embeddings', type: 'sync' },
        { from: 'vector', to: 'agent', message: 'Top K relevant', type: 'return', style: 'dashed' },
        { from: 'agent', to: 'llm', message: 'Prompt + context', type: 'sync' },
        { from: 'llm', to: 'agent', message: 'Response', type: 'return', style: 'dashed' },
        { from: 'agent', to: 'user', message: 'Answer', type: 'return', style: 'dashed' }
      ]}
      width={1080}
      marginTop={30}
    />
    
    <Well variant="info" width={1080} marginTop={30}>
      <Text fontSize={14} fontWeight="bold">🔍 How It Works</Text>
      <List items={[
        '1. Store all past conversations as embeddings in a vector database',
        '2. When user asks a question, convert it to an embedding',
        '3. Retrieve the K most similar past conversations',
        '4. Include only these relevant chunks in the LLM context',
        '5. LLM generates response with perfect, targeted context'
      ]} fontSize={12} />
    </Well>
  </Slide>
);

// Slide 7: Hybrid Approach
const HybridSlide = () => (
  <Slide>
    <Title level={1}>Strategy 4: Hybrid Approach</Title>
    <Subtitle>Combine strategies for optimal results</Subtitle>
    
    <Stack gap={24} marginTop={30}>
      <Card variant="success" width={1080}>
        <Title level={3}>The Best of All Worlds</Title>
        <Text fontSize={13} marginTop={12}>
          Use multiple strategies together to get the benefits of each while
          mitigating their individual weaknesses.
        </Text>
      </Card>
      
      <Panel header={<Text fontSize={14} fontWeight="bold">Example Architecture</Text>} 
             variant="primary" width={1080}>
        <Steps
          steps={[
            {
              number: 1,
              title: 'Sliding Window (Last 5 messages)',
              description: 'Always keep immediate context'
            },
            {
              number: 2,
              title: 'Summarization (Last 50 messages)',
              description: 'Compress recent history'
            },
            {
              number: 3,
              title: 'RAG (All history)',
              description: 'Retrieve relevant past context'
            },
            {
              number: 4,
              title: 'Combine',
              description: 'Merge all three into final context'
            }
          ]}
        />
      </Panel>
      
      <Row gap={24}>
        <Badge text="Recent: 5 msgs" variant="primary" size="medium" />
        <Badge text="Summary: 2 msgs" variant="secondary" size="medium" />
        <Badge text="Retrieved: 3 msgs" variant="info" size="medium" />
        <Badge text="Total: 10 msgs" variant="success" size="medium" />
      </Row>
    </Stack>
  </Slide>
);

// Slide 8: Implementation
const ImplementationSlide = () => (
  <Slide>
    <Title level={1}>Implementation Considerations</Title>
    
    <Row gap={24} marginTop={30}>
      <Section title="Performance" variant="warning" width={500}>
        <List items={[
          'Embedding generation latency',
          'Vector search performance',
          'Summarization overhead',
          'Caching strategies',
          'Batch processing'
        ]} fontSize={12} />
      </Section>
      
      <Section title="Cost Management" variant="danger" width={500}>
        <List items={[
          'Token usage tracking',
          'API call optimization',
          'Embedding storage costs',
          'LLM call batching',
          'Cache hit rates'
        ]} fontSize={12} />
      </Section>
    </Row>
    
    <Divider width={1080} marginTop={24} marginBottom={24} />
    
    <Section title="Quality Metrics" variant="info" width={1080}>
      <List items={[
        'Context relevance score - Are we retrieving the right information?',
        'Response coherence - Does the LLM maintain conversation flow?',
        'Information retention - Can it recall important past details?',
        'User satisfaction - Does it feel natural to users?',
        'Error rate - How often does it fail to understand context?'
      ]} fontSize={13} />
    </Section>
  </Slide>
);

// Slide 9: Summary
const SummarySlide = () => (
  <Slide alignItems="center" justifyContent="center">
    <Title level={1}>Key Takeaways</Title>
    
    <Spacer size={40} />
    
    <Stack gap={24} width={900}>
      <Callout title="Choose Wisely" variant="success" width={900}>
        <Text fontSize={13}>
          Select the strategy that matches your use case. Simple sliding windows
          work great for chat. Complex systems need hybrid approaches.
        </Text>
      </Callout>
      
      <Callout title="Measure Everything" variant="info" width={900}>
        <Text fontSize={13}>
          Track token usage, costs, latency, and quality. You can't optimize
          what you don't measure.
        </Text>
      </Callout>
      
      <Callout title="Iterate & Improve" variant="warning" width={900}>
        <Text fontSize={13}>
          Context management is not one-size-fits-all. Experiment, measure,
          and refine based on your specific requirements.
        </Text>
      </Callout>
    </Stack>
    
    <Spacer size={40} />
    
    <Badge text="Thank You!" variant="primary" size="large" />
  </Slide>
);

// Generate all modes
async function generate() {
  console.log('🎨 Generating LLM Context Management Presentation v3\n');
  
  const outputDir = './output-v3';
  mkdirSync(outputDir, { recursive: true });
  
  const slides = [
    { name: '01-title', component: <TitleSlide /> },
    { name: '02-problem', component: <ProblemSlide /> },
    { name: '03-strategies', component: <StrategyOverview /> },
    { name: '04-sliding-window', component: <SlidingWindowSlide /> },
    { name: '05-summarization', component: <SummarizationSlide /> },
    { name: '06-rag', component: <RAGSlide /> },
    { name: '07-hybrid', component: <HybridSlide /> },
    { name: '08-implementation', component: <ImplementationSlide /> },
    { name: '09-summary', component: <SummarySlide /> },
  ];
  
  const themes = [
    { name: 'default', theme: defaultTheme },
    { name: 'dark', theme: darkTheme },
    { name: 'nord', theme: nordTheme },
    { name: 'professional', theme: professionalTheme },
  ];
  
  for (const { name: themeName, theme } of themes) {
    console.log(`\n📚 Generating with ${theme.name} theme...\n`);
    setCurrentTheme(theme);
    
    // Mode 1: Traditional slides
    console.log(`  Mode 1: Traditional slides...`);
    const slideSvgs = await generateSlideDeck(slides, {
      theme,
      width: 1200,
      height: 800,
    });
    
    slideSvgs.forEach((svg, i) => {
      const filename = `${outputDir}/${themeName}-slide-${String(i + 1).padStart(2, '0')}.svg`;
      writeFileSync(filename, svg);
    });
    console.log(`    ✓ Created ${slideSvgs.length} slides`);
    
    // Mode 2: Scrolling slides
    console.log(`  Mode 2: Scrolling slides...`);
    const scrollingSvg = await generateScrollingPage(
      slides.map(s => ({ name: s.name, component: s.component })),
      { theme, gap: 60 }
    );
    writeFileSync(`${outputDir}/${themeName}-scrolling.svg`, scrollingSvg);
    console.log(`    ✓ Created scrolling page`);
    
    // Mode 3: Continuous page
    console.log(`  Mode 3: Continuous page...`);
    const continuousSvg = await generateContinuousPage(
      slides.map(s => ({ name: s.name, component: s.component })),
      { theme, gap: 40 }
    );
    writeFileSync(`${outputDir}/${themeName}-continuous.svg`, continuousSvg);
    console.log(`    ✓ Created continuous page`);
  }
  
  console.log('\n✨ All presentations generated successfully!');
  console.log(`\nOutput: ${outputDir}/`);
  console.log(`\nFiles per theme:`);
  console.log(`  • 9 individual slides`);
  console.log(`  • 1 scrolling page`);
  console.log(`  • 1 continuous page`);
  console.log(`\nTotal: ${4 * (9 + 1 + 1)} = 44 files\n`);
}

generate().catch(console.error);
