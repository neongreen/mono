import React from 'react';
import { 
  Slide, Stack, Row, Box, Text, Arrow, Card, Title, Subtitle, Label,
  List, ProsCons, Section, Callout, RichText, Spacer, Grid,
  renderToSVG 
} from 'diagram-dsl';
import { writeFileSync, mkdirSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const outputDir = join(__dirname, '../output-v2');

// Slide 1: Title Slide
const TitleSlide = () => (
  <Slide alignItems="center" justifyContent="center">
    <Title level={1}>Context Management Strategies</Title>
    <Title level={2}>in LLM Agent Implementations</Title>
    
    <Spacer size={40} />
    
    <Stack gap={16} alignItems="center">
      <Text fontSize={20} textAlign="center" color="#666">
        Understanding how agents maintain coherent conversations
      </Text>
      <Text fontSize={20} textAlign="center" color="#666">
        while managing token limits and computational costs
      </Text>
    </Stack>

    <Spacer size={60} />
    
    <Section title="Topics Covered" variant="default" width={600}>
      <List
        items={[
          'Context Window Management',
          'Memory Architectures',
          'Compression Strategies',
          'Retrieval Augmented Generation'
        ]}
        fontSize={14}
      />
    </Section>
  </Slide>
);

// Slide 2: The Context Challenge
const ContextChallenge = () => (
  <Slide>
    <Title level={1}>The Context Challenge</Title>
    
    <Row gap={24} marginTop={20}>
      <Card id="user-input" variant="primary" width={350}>
        <Label>User Input</Label>
        <Subtitle>Long conversation history</Subtitle>
        <List items={['Previous messages', 'Tool call results', 'External data']} fontSize={12} />
      </Card>

      <Stack gap={16} justifyContent="center">
        <Text fontSize={48}>⚠️</Text>
      </Stack>

      <Card id="llm-limits" variant="danger" width={350}>
        <Label>LLM Constraints</Label>
        <Subtitle>Token limits vary by model</Subtitle>
        <List items={['GPT-4: 8K-128K tokens', 'Claude: 100K-200K tokens', 'Cost per token increases']} fontSize={12} />
      </Card>
    </Row>

    <Spacer size={20} />

    <Callout title="Key Trade-offs" variant="warning" width={900}>
      <RichText
        segments={[
          { text: 'More Context:', bold: true },
          ' Better understanding, higher accuracy, higher costs'
        ]}
        fontSize={14}
      />
      <Spacer size={8} />
      <RichText
        segments={[
          { text: 'Less Context:', bold: true },
          ' Faster responses, lower costs, potential information loss'
        ]}
        fontSize={14}
      />
    </Callout>
  </Slide>
);

// Slide 3: Strategy 1 - Sliding Window
const SlidingWindow = () => (
  <Slide>
    <Title level={1}>Strategy 1: Sliding Window</Title>
    <Subtitle>Keep only the most recent N messages</Subtitle>

    <Row gap={12} justifyContent="center" marginTop={20}>
      {['Msg 1', 'Msg 2', 'Msg 3', 'Msg 4', 'Msg 5', 'Msg 6', 'Msg 7', 'Msg 8'].map((msg, i) => (
        <Box
          key={i}
          id={`msg-${i}`}
          width={80}
          height={60}
          backgroundColor={i < 4 ? '#ffebee' : '#e8f5e9'}
          borderColor={i < 4 ? '#c62828' : '#2e7d32'}
          borderWidth={2}
          borderRadius={6}
          justifyContent="center"
          alignItems="center"
          opacity={i < 4 ? 0.4 : 1}
        >
          <Text fontSize={12} fontWeight="bold">{msg}</Text>
        </Box>
      ))}
    </Row>

    <Grid columns={2} gap={40} marginTop={20}>
      <Section title="Dropped" variant="danger" width={220}>
        <Text fontSize={11}>Old messages removed to save tokens</Text>
      </Section>

      <Section title="Kept" variant="success" width={220}>
        <Text fontSize={11}>Recent messages in active context</Text>
      </Section>
    </Grid>

    <Spacer size={20} />

    <Card variant="default" width={900}>
      <ProsCons
        pros={['Simple to implement', 'Predictable token usage', 'Low computational cost']}
        cons={['Loses important history', 'No semantic awareness', 'Context can break mid-topic']}
        fontSize={11}
      />
    </Card>
  </Slide>
);

// Slide 4: Strategy 2 - Hierarchical Summarization
const HierarchicalSummarization = () => (
  <Slide>
    <Title level={1}>Strategy 2: Hierarchical Summarization</Title>
    <Subtitle>Compress old context into summaries at multiple levels</Subtitle>

    <Row gap={40} marginTop={20} justifyContent="center">
      <Stack gap={16} alignItems="center">
        <Card id="detailed" variant="primary" width={200}>
          <Label>Detailed Messages</Label>
          <Text fontSize={11}>Last 5-10 messages</Text>
          <Text fontSize={11}>Full content preserved</Text>
        </Card>

        <Card id="recent" variant="secondary" width={200}>
          <Label>Recent Summary</Label>
          <Text fontSize={11}>Last 50 messages</Text>
          <Text fontSize={11}>Key points extracted</Text>
        </Card>

        <Card id="session" variant="accent" width={200}>
          <Label>Session Summary</Label>
          <Text fontSize={11}>Entire conversation</Text>
          <Text fontSize={11}>High-level overview</Text>
        </Card>
      </Stack>

      <Section title="How It Works" variant="default" width={550}>
        <List
          items={[
            { text: '1. Recent messages: ', bold: true },
            'Kept in full detail',
            { text: '2. Medium-age messages: ', bold: true },
            'Summarized into key points',
            { text: '3. Old messages: ', bold: true },
            'Compressed into single paragraph'
          ]}
          fontSize={12}
          bullet=""
        />
        
        <Spacer size={12} />
        
        <Callout title="Example Compression" variant="info">
          <Text fontSize={10} color="#666">
            "User discussed implementing auth system. Decided on OAuth2.
            Then talked about database schema. Settled on PostgreSQL with
            normalized tables. Now working on API design."
          </Text>
        </Callout>

        <Spacer size={12} />

        <ProsCons
          pros={['Retains important info', 'Scales to long convos', 'Semantic awareness']}
          cons={['Requires LLM calls', 'Summarization costs', 'May lose nuance']}
          fontSize={10}
        />
      </Section>
    </Row>
  </Slide>
);

// Slide 5: Strategy 3 - Vector Memory
const VectorMemory = () => (
  <Slide>
    <Title level={1}>Strategy 3: Vector Memory (RAG)</Title>
    <Subtitle>Store embeddings and retrieve relevant context dynamically</Subtitle>

    <Row gap={20} justifyContent="center" marginTop={20}>
      <Card id="messages" variant="primary" width={180}>
        <Label>All Messages</Label>
        <Text fontSize={10}>Full conversation</Text>
        <Text fontSize={10}>history stored</Text>
      </Card>

      <Stack justifyContent="center">
        <Arrow from="messages" to="embeddings" color="#1976d2" strokeWidth={2} />
      </Stack>

      <Card id="embeddings" variant="secondary" width={180}>
        <Label>Vector DB</Label>
        <Text fontSize={10}>Semantic embeddings</Text>
        <Text fontSize={10}>of all messages</Text>
      </Card>

      <Stack justifyContent="center">
        <Arrow from="embeddings" to="retrieval" color="#7b1fa2" strokeWidth={2} />
      </Stack>

      <Card id="retrieval" variant="accent" width={180}>
        <Label>Smart Retrieval</Label>
        <Text fontSize={10}>Query-relevant</Text>
        <Text fontSize={10}>context only</Text>
      </Card>
    </Row>

    <Spacer size={20} />

    <Grid columns={2} gap={32}>
      <Section title="Process Flow" variant="default">
        <List
          items={[
            '1. User query arrives',
            '2. Query embedded as vector',
            '3. Similar messages retrieved',
            '4. Relevant context assembled',
            '5. Sent to LLM with query'
          ]}
          fontSize={12}
          bullet=""
        />
      </Section>

      <Stack gap={12}>
        <Section title="Advantages" variant="success">
          <List
            items={[
              'Semantic search capability',
              'Handles unlimited history',
              'Retrieves relevant context',
              'Good for knowledge bases'
            ]}
            fontSize={11}
          />
        </Section>

        <Section title="Challenges" variant="danger">
          <List
            items={[
              'Requires vector database',
              'Embedding costs',
              'May miss context order',
              'Complex infrastructure'
            ]}
            fontSize={11}
          />
        </Section>
      </Stack>
    </Grid>
  </Slide>
);

// Slide 6: Strategy 4 - Hybrid Approach
const HybridApproach = () => (
  <Slide>
    <Title level={1}>Strategy 4: Hybrid Approach</Title>
    <Subtitle>Combine multiple strategies for optimal results</Subtitle>

    <Row gap={16} justifyContent="center" marginTop={20}>
      <Card id="recent-full" variant="primary" width={220}>
        <Label>Recent (Full)</Label>
        <Text fontSize={11}>Last 3-5 messages</Text>
        <Text fontSize={11}>Complete details</Text>
        <Text fontSize={10} fontWeight="bold" color="#1976d2" marginTop={8}>
          Sliding Window
        </Text>
      </Card>

      <Card id="summary-layer" variant="secondary" width={220}>
        <Label>Summary Layer</Label>
        <Text fontSize={11}>Medium history</Text>
        <Text fontSize={11}>Compressed points</Text>
        <Text fontSize={10} fontWeight="bold" color="#7b1fa2" marginTop={8}>
          Summarization
        </Text>
      </Card>

      <Card id="vector-search" variant="accent" width={220}>
        <Label>Vector Search</Label>
        <Text fontSize={11}>Related context</Text>
        <Text fontSize={11}>From full history</Text>
        <Text fontSize={10} fontWeight="bold" color="#f57c00" marginTop={8}>
          RAG Retrieval
        </Text>
      </Card>

      <Card id="system-context" variant="success" width={220}>
        <Label>System Context</Label>
        <Text fontSize={11}>Instructions</Text>
        <Text fontSize={11}>Agent capabilities</Text>
        <Text fontSize={10} fontWeight="bold" color="#2e7d32" marginTop={8}>
          Always Present
        </Text>
      </Card>
    </Row>

    <Spacer size={16} />

    <Callout title="Example Implementation" variant="warning" width={1000}>
      <Grid columns={2} gap={32}>
        <Stack gap={8}>
          <Text fontSize={13} fontWeight="bold">Context Assembly Order:</Text>
          <List
            items={[
              '1. System prompt & instructions',
              '2. Session summary (if exists)',
              '3. Retrieved relevant messages (RAG)',
              '4. Recent message summaries',
              '5. Last 3-5 full messages',
              '6. Current user query'
            ]}
            fontSize={11}
            bullet=""
          />
        </Stack>

        <Stack gap={8}>
          <Text fontSize={13} fontWeight="bold">Token Budget Example:</Text>
          <List
            items={[
              'Model limit: 8,000 tokens',
              'System prompt: 500 tokens',
              'Session summary: 200 tokens',
              'RAG retrieved: 1,000 tokens',
              'Recent messages: 2,000 tokens',
              'Reserve for response: 2,000 tokens'
            ]}
            fontSize={11}
          />
          <Text fontSize={11} fontWeight="bold" color="#ff9800">
            Total used: 5,700 tokens
          </Text>
        </Stack>
      </Grid>
    </Callout>

    <Spacer size={12} />

    <Callout title="Why Hybrid Works Best" variant="info" width={1000}>
      <Text fontSize={12}>
        By combining strategies, agents maintain both immediate context and long-term memory,
        retrieve relevant information semantically, and stay within token budgets efficiently.
        This mirrors human memory: detailed recent memory, summarized medium-term memory,
        and searchable long-term memory.
      </Text>
    </Callout>
  </Slide>
);

// Slide 7: Practical Considerations
const PracticalConsiderations = () => (
  <Slide>
    <Title level={1}>Practical Considerations</Title>
    <Subtitle>Implementation details and trade-offs</Subtitle>

    <Grid columns={2} gap={24} marginTop={20}>
      <Section title="Token Counting & Budgeting" variant="primary">
        <List
          items={[
            'Always count tokens accurately (tiktoken)',
            'Reserve buffer for response (20-30%)',
            'Monitor costs per conversation',
            'Set hard limits for safety'
          ]}
          fontSize={12}
        />
      </Section>

      <Section title="Context Freshness" variant="secondary">
        <List
          items={[
            'Update summaries incrementally',
            'Invalidate stale embeddings',
            'Refresh vector DB periodically',
            'Balance freshness vs cost'
          ]}
          fontSize={12}
        />
      </Section>

      <Section title="Performance Optimization" variant="accent">
        <List
          items={[
            'Cache embeddings aggressively',
            'Batch vector operations',
            'Use streaming for summaries',
            'Parallel retrieval when possible'
          ]}
          fontSize={12}
        />
      </Section>

      <Section title="Quality Assurance" variant="success">
        <List
          items={[
            'Test with various conversation lengths',
            'Validate summary quality',
            'Check retrieval relevance',
            'Monitor user satisfaction metrics'
          ]}
          fontSize={12}
        />
      </Section>
    </Grid>

    <Spacer size={16} />

    <Callout title="Cost-Benefit Analysis" variant="default" width={1100}>
      <Grid columns={3} gap={40}>
        <Stack gap={8}>
          <Text fontSize={12} fontWeight="bold" color="#2e7d32">Low Cost Strategies</Text>
          <List items={['Sliding window only', 'Minimal summarization', 'Small context windows']} fontSize={11} />
          <Text fontSize={10} color="#666">Best for: Simple tasks, budget-constrained apps</Text>
        </Stack>
        
        <Stack gap={8}>
          <Text fontSize={12} fontWeight="bold" color="#ff9800">Medium Cost Strategies</Text>
          <List items={['Hierarchical summaries', 'Moderate context windows', 'Selective retrieval']} fontSize={11} />
          <Text fontSize={10} color="#666">Best for: Most production applications</Text>
        </Stack>
        
        <Stack gap={8}>
          <Text fontSize={12} fontWeight="bold" color="#c62828">High Cost Strategies</Text>
          <List items={['Full vector RAG', 'Large context windows', 'Frequent re-embedding']} fontSize={11} />
          <Text fontSize={10} color="#666">Best for: Enterprise, research applications</Text>
        </Stack>
      </Grid>
    </Callout>
  </Slide>
);

// Slide 8: Summary & Recommendations
const SummarySlide = () => (
  <Slide>
    <Title level={1}>Summary & Recommendations</Title>
    
    <Section title="Choosing the Right Strategy" variant="primary" width={1100} marginTop={20}>
      <Stack gap={10}>
        <RichText segments={[{ text: 'Start Simple:', bold: true }, ' Begin with sliding window, add complexity as needed']} fontSize={13} />
        <RichText segments={[{ text: 'Measure First:', bold: true }, ' Understand your token usage patterns before optimizing']} fontSize={13} />
        <RichText segments={[{ text: 'User Experience:', bold: true }, ' Maintain conversation continuity over perfect memory']} fontSize={13} />
        <RichText segments={[{ text: 'Cost Awareness:', bold: true }, ' Balance quality with operational expenses']} fontSize={13} />
      </Stack>
    </Section>

    <Spacer size={20} />

    <Grid columns={2} gap={24}>
      <Section title="For Most Applications" variant="secondary">
        <RichText segments={['Use a ', { text: 'hybrid approach', bold: true }, ' with:']} fontSize={12} />
        <Spacer size={8} />
        <List items={['3-5 recent full messages', 'Session summary', 'Optional RAG for knowledge base']} fontSize={11} />
        <Spacer size={12} />
        <Text fontSize={11} color="#666">
          This provides the best balance of context quality,
          cost efficiency, and implementation complexity.
        </Text>
      </Section>

      <Section title="Future Considerations" variant="accent">
        <List
          items={[
            { text: 'Longer context windows: ', bold: true },
            'Models improving',
            { text: 'Better compression: ', bold: true },
            'LLMs learning to summarize',
            { text: 'Cheaper embeddings: ', bold: true },
            'Vector operations cost dropping',
            { text: 'Smarter retrieval: ', bold: true },
            'More sophisticated RAG'
          ]}
          fontSize={12}
          bullet=""
        />
      </Section>
    </Grid>

    <Spacer size={16} />

    <Callout title="Key Takeaway" variant="success" width={1100}>
      <Text fontSize={14}>
        Context management is not a solved problem but a series of trade-offs.
        The best strategy depends on your specific use case, budget constraints,
        and user expectations. Start simple, measure carefully, and iterate based
        on real usage patterns.
      </Text>
    </Callout>
  </Slide>
);

// Generate all slides
async function generatePresentation() {
  try {
    // Create output directory
    mkdirSync(outputDir, { recursive: true });

    const slides = [
      { name: '01-title', component: <TitleSlide /> },
      { name: '02-context-challenge', component: <ContextChallenge /> },
      { name: '03-sliding-window', component: <SlidingWindow /> },
      { name: '04-hierarchical-summarization', component: <HierarchicalSummarization /> },
      { name: '05-vector-memory', component: <VectorMemory /> },
      { name: '06-hybrid-approach', component: <HybridApproach /> },
      { name: '07-practical-considerations', component: <PracticalConsiderations /> },
      { name: '08-summary', component: <SummarySlide /> },
    ];

    console.log('Generating LLM Context Management presentation (v2 with new components)...\n');

    for (const slide of slides) {
      const svg = await renderToSVG(slide.component, {
        width: 1200,
        height: 800,
        backgroundColor: 'white',
      });
      
      const filename = `${slide.name}.svg`;
      writeFileSync(join(outputDir, filename), svg);
      console.log(`✓ Generated ${filename}`);
    }

    // Generate an HTML viewer
    const htmlContent = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Context Management Strategies in LLM Agents (v2)</title>
  <style>
    body {
      margin: 0;
      padding: 20px;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
      background: #f5f5f5;
    }
    .container {
      max-width: 1240px;
      margin: 0 auto;
      background: white;
      padding: 20px;
      border-radius: 8px;
      box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    }
    h1 {
      text-align: center;
      color: #333;
      margin-bottom: 30px;
    }
    .slide {
      margin-bottom: 40px;
      border: 1px solid #ddd;
      border-radius: 4px;
      overflow: hidden;
    }
    .slide img {
      width: 100%;
      height: auto;
      display: block;
    }
    .controls {
      text-align: center;
      margin-top: 20px;
      position: sticky;
      top: 20px;
      background: white;
      padding: 10px;
      border-radius: 4px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }
    button {
      margin: 0 10px;
      padding: 10px 20px;
      font-size: 16px;
      cursor: pointer;
      background: #1976d2;
      color: white;
      border: none;
      border-radius: 4px;
    }
    button:hover {
      background: #1565c0;
    }
    .slide-number {
      display: inline-block;
      margin: 0 20px;
      font-size: 18px;
      font-weight: bold;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>Context Management Strategies in LLM Agent Implementations (v2)</h1>
    
    <div class="controls">
      <button onclick="previousSlide()">← Previous</button>
      <span class="slide-number" id="slideNumber">Slide 1 of 8</span>
      <button onclick="nextSlide()">Next →</button>
    </div>

    <div id="slideContainer">
      ${slides.map((slide, i) => `
      <div class="slide" id="slide-${i}" style="display: ${i === 0 ? 'block' : 'none'}">
        <img src="${slide.name}.svg" alt="Slide ${i + 1}">
      </div>
      `).join('')}
    </div>
  </div>

  <script>
    let currentSlide = 0;
    const totalSlides = 8;

    function showSlide(n) {
      document.querySelectorAll('.slide').forEach(slide => {
        slide.style.display = 'none';
      });
      
      currentSlide = Math.max(0, Math.min(n, totalSlides - 1));
      document.getElementById('slide-' + currentSlide).style.display = 'block';
      document.getElementById('slideNumber').textContent = 'Slide ' + (currentSlide + 1) + ' of ' + totalSlides;
    }

    function nextSlide() {
      showSlide(currentSlide + 1);
    }

    function previousSlide() {
      showSlide(currentSlide - 1);
    }

    document.addEventListener('keydown', (e) => {
      if (e.key === 'ArrowRight' || e.key === ' ') {
        nextSlide();
      } else if (e.key === 'ArrowLeft') {
        previousSlide();
      }
    });
  </script>
</body>
</html>`;

    writeFileSync(join(outputDir, 'index.html'), htmlContent);
    console.log('\n✓ Generated index.html viewer');
    console.log('\nPresentation complete! Open output-v2/index.html to view.');

  } catch (error) {
    console.error('Error generating presentation:', error);
    process.exit(1);
  }
}

generatePresentation();
