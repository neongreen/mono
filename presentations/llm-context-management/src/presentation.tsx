import React from 'react';
import { Stack, Row, Box, Text, Arrow, renderToSVG, Card, Title, Subtitle, Label } from 'diagram-dsl';
import { writeFileSync, mkdirSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const outputDir = join(__dirname, '../output');

// Slide 1: Title Slide
const TitleSlide = () => (
  <Stack gap={40} padding={60} alignItems="center" justifyContent="center" width={1200} height={800}>
    <Title level={1}>Context Management Strategies</Title>
    <Title level={2}>in LLM Agent Implementations</Title>
    
    <Stack gap={16} marginTop={60}>
      <Text fontSize={20} textAlign="center" color="#666">
        Understanding how agents maintain coherent conversations
      </Text>
      <Text fontSize={20} textAlign="center" color="#666">
        while managing token limits and computational costs
      </Text>
    </Stack>

    <Stack gap={12} marginTop={80} alignItems="center">
      <Text fontSize={16} color="#999">Topics Covered</Text>
      <Text fontSize={14} color="#666">• Context Window Management</Text>
      <Text fontSize={14} color="#666">• Memory Architectures</Text>
      <Text fontSize={14} color="#666">• Compression Strategies</Text>
      <Text fontSize={14} color="#666">• Retrieval Augmented Generation</Text>
    </Stack>
  </Stack>
);

// Slide 2: The Context Challenge
const ContextChallenge = () => (
  <Stack gap={32} padding={60} width={1200} height={800}>
    <Title level={1}>The Context Challenge</Title>
    
    <Row gap={24} marginTop={20}>
      <Stack gap={16} flexGrow={1}>
        <Card id="user-input" variant="primary" width={350}>
          <Label>User Input</Label>
          <Subtitle>Long conversation history</Subtitle>
          <Text fontSize={12} color="#666">Previous messages</Text>
          <Text fontSize={12} color="#666">Tool call results</Text>
          <Text fontSize={12} color="#666">External data</Text>
        </Card>
      </Stack>

      <Stack gap={16} justifyContent="center">
        <Box
          width={80}
          height={80}
          backgroundColor="#fff3e0"
          borderColor="#ff9800"
          borderWidth={3}
          borderRadius={8}
          justifyContent="center"
          alignItems="center"
        >
          <Text fontSize={32}>⚠️</Text>
        </Box>
      </Stack>

      <Stack gap={16} flexGrow={1}>
        <Card id="llm-limits" variant="danger" width={350}>
          <Label>LLM Constraints</Label>
          <Subtitle>Token limits vary by model</Subtitle>
          <Text fontSize={12} color="#666">GPT-4: 8K-128K tokens</Text>
          <Text fontSize={12} color="#666">Claude: 100K-200K tokens</Text>
          <Text fontSize={12} color="#666">Cost per token increases</Text>
        </Card>
      </Stack>
    </Row>

    <Stack gap={16} marginTop={40} alignItems="center">
      <Box
        width={900}
        backgroundColor="#f5f5f5"
        borderColor="#999"
        borderWidth={2}
        borderRadius={8}
        padding={24}
      >
        <Text fontSize={18} fontWeight="bold" marginBottom={12}>Key Trade-offs</Text>
        <Text fontSize={14} marginBottom={8}>
          <Text fontWeight="bold">More Context:</Text> Better understanding, higher accuracy, higher costs
        </Text>
        <Text fontSize={14}>
          <Text fontWeight="bold">Less Context:</Text> Faster responses, lower costs, potential information loss
        </Text>
      </Box>
    </Stack>
  </Stack>
);

// Slide 3: Strategy 1 - Sliding Window
const SlidingWindow = () => (
  <Stack gap={32} padding={60} width={1200} height={800}>
    <Title level={1}>Strategy 1: Sliding Window</Title>
    <Subtitle>Keep only the most recent N messages</Subtitle>

    <Stack gap={24} marginTop={20}>
      <Row gap={12} justifyContent="center">
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

      <Row gap={40} justifyContent="center" marginTop={20}>
        <Box
          width={220}
          height={120}
          backgroundColor="#ffebee"
          borderColor="#c62828"
          borderWidth={2}
          borderRadius={8}
          padding={16}
        >
          <Text fontSize={14} fontWeight="bold" marginBottom={8} color="#c62828">Dropped</Text>
          <Text fontSize={11}>Old messages removed to save tokens</Text>
        </Box>

        <Box
          width={220}
          height={120}
          backgroundColor="#e8f5e9"
          borderColor="#2e7d32"
          borderWidth={2}
          borderRadius={8}
          padding={16}
        >
          <Text fontSize={14} fontWeight="bold" marginBottom={8} color="#2e7d32">Kept</Text>
          <Text fontSize={11}>Recent messages in active context</Text>
        </Box>
      </Row>

      <Card variant="default" width={900} marginTop={20}>
        <Label>Pros & Cons</Label>
        <Row gap={32}>
          <Stack gap={8} flexGrow={1}>
            <Text fontSize={13} fontWeight="bold" color="#2e7d32">✓ Pros</Text>
            <Text fontSize={11}>• Simple to implement</Text>
            <Text fontSize={11}>• Predictable token usage</Text>
            <Text fontSize={11}>• Low computational cost</Text>
          </Stack>
          <Stack gap={8} flexGrow={1}>
            <Text fontSize={13} fontWeight="bold" color="#c62828">✗ Cons</Text>
            <Text fontSize={11}>• Loses important history</Text>
            <Text fontSize={11}>• No semantic awareness</Text>
            <Text fontSize={11}>• Context can break mid-topic</Text>
          </Stack>
        </Row>
      </Card>
    </Stack>
  </Stack>
);

// Slide 4: Strategy 2 - Hierarchical Summarization
const HierarchicalSummarization = () => (
  <Stack gap={32} padding={60} width={1200} height={800}>
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

      <Stack gap={16} justifyContent="center" width={550}>
        <Box
          backgroundColor="#f5f5f5"
          borderColor="#999"
          borderWidth={2}
          borderRadius={8}
          padding={20}
        >
          <Text fontSize={14} fontWeight="bold" marginBottom={12}>How It Works</Text>
          
          <Stack gap={10}>
            <Text fontSize={12}>
              <Text fontWeight="bold">1. Recent messages:</Text> Kept in full detail
            </Text>
            <Text fontSize={12}>
              <Text fontWeight="bold">2. Medium-age messages:</Text> Summarized into key points
            </Text>
            <Text fontSize={12}>
              <Text fontWeight="bold">3. Old messages:</Text> Compressed into single paragraph
            </Text>
          </Stack>

          <Stack gap={8} marginTop={16} padding={12} backgroundColor="#e3f2fd" borderRadius={6}>
            <Text fontSize={11} fontWeight="bold">Example Compression</Text>
            <Text fontSize={10} color="#666">
              "User discussed implementing auth system. Decided on OAuth2.
              Then talked about database schema. Settled on PostgreSQL with
              normalized tables. Now working on API design."
            </Text>
          </Stack>
        </Box>

        <Card variant="default" width={550}>
          <Label>Trade-offs</Label>
          <Row gap={24}>
            <Stack gap={6} flexGrow={1}>
              <Text fontSize={12} fontWeight="bold" color="#2e7d32">✓ Pros</Text>
              <Text fontSize={10}>• Retains important info</Text>
              <Text fontSize={10}>• Scales to long convos</Text>
              <Text fontSize={10}>• Semantic awareness</Text>
            </Stack>
            <Stack gap={6} flexGrow={1}>
              <Text fontSize={12} fontWeight="bold" color="#c62828">✗ Cons</Text>
              <Text fontSize={10}>• Requires LLM calls</Text>
              <Text fontSize={10}>• Summarization costs</Text>
              <Text fontSize={10}>• May lose nuance</Text>
            </Stack>
          </Row>
        </Card>
      </Stack>
    </Row>
  </Stack>
);

// Slide 5: Strategy 3 - Vector Memory
const VectorMemory = () => (
  <Stack gap={32} padding={60} width={1200} height={800}>
    <Title level={1}>Strategy 3: Vector Memory (RAG)</Title>
    <Subtitle>Store embeddings and retrieve relevant context dynamically</Subtitle>

    <Stack gap={24} marginTop={20}>
      <Row gap={20} justifyContent="center">
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

      <Row gap={32} marginTop={20}>
        <Stack gap={12} flexGrow={1}>
          <Box
            backgroundColor="#f5f5f5"
            borderColor="#999"
            borderWidth={2}
            borderRadius={8}
            padding={20}
          >
            <Text fontSize={14} fontWeight="bold" marginBottom={12}>Process Flow</Text>
            <Stack gap={8}>
              <Text fontSize={12}>
                <Text fontWeight="bold">1.</Text> User query arrives
              </Text>
              <Text fontSize={12}>
                <Text fontWeight="bold">2.</Text> Query embedded as vector
              </Text>
              <Text fontSize={12}>
                <Text fontWeight="bold">3.</Text> Similar messages retrieved
              </Text>
              <Text fontSize={12}>
                <Text fontWeight="bold">4.</Text> Relevant context assembled
              </Text>
              <Text fontSize={12}>
                <Text fontWeight="bold">5.</Text> Sent to LLM with query
              </Text>
            </Stack>
          </Box>
        </Stack>

        <Stack gap={12} flexGrow={1}>
          <Box
            backgroundColor="#e8f5e9"
            borderColor="#2e7d32"
            borderWidth={2}
            borderRadius={8}
            padding={20}
          >
            <Text fontSize={14} fontWeight="bold" marginBottom={12} color="#2e7d32">Advantages</Text>
            <Stack gap={6}>
              <Text fontSize={11}>• Semantic search capability</Text>
              <Text fontSize={11}>• Handles unlimited history</Text>
              <Text fontSize={11}>• Retrieves relevant context</Text>
              <Text fontSize={11}>• Good for knowledge bases</Text>
            </Stack>
          </Box>

          <Box
            backgroundColor="#ffebee"
            borderColor="#c62828"
            borderWidth={2}
            borderRadius={8}
            padding={20}
          >
            <Text fontSize={14} fontWeight="bold" marginBottom={12} color="#c62828">Challenges</Text>
            <Stack gap={6}>
              <Text fontSize={11}>• Requires vector database</Text>
              <Text fontSize={11}>• Embedding costs</Text>
              <Text fontSize={11}>• May miss context order</Text>
              <Text fontSize={11}>• Complex infrastructure</Text>
            </Stack>
          </Box>
        </Stack>
      </Row>
    </Stack>
  </Stack>
);

// Slide 6: Strategy 4 - Hybrid Approach
const HybridApproach = () => (
  <Stack gap={32} padding={60} width={1200} height={800}>
    <Title level={1}>Strategy 4: Hybrid Approach</Title>
    <Subtitle>Combine multiple strategies for optimal results</Subtitle>

    <Stack gap={20} marginTop={20}>
      <Row gap={16} justifyContent="center">
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

      <Box
        width={1000}
        backgroundColor="#fff3e0"
        borderColor="#ff9800"
        borderWidth={2}
        borderRadius={8}
        padding={24}
        marginTop={16}
      >
        <Text fontSize={16} fontWeight="bold" marginBottom={16}>Example Implementation</Text>
        
        <Row gap={32}>
          <Stack gap={10} flexGrow={1}>
            <Text fontSize={13} fontWeight="bold">Context Assembly Order:</Text>
            <Text fontSize={11}>1. System prompt & instructions</Text>
            <Text fontSize={11}>2. Session summary (if exists)</Text>
            <Text fontSize={11}>3. Retrieved relevant messages (RAG)</Text>
            <Text fontSize={11}>4. Recent message summaries</Text>
            <Text fontSize={11}>5. Last 3-5 full messages</Text>
            <Text fontSize={11}>6. Current user query</Text>
          </Stack>

          <Stack gap={10} flexGrow={1}>
            <Text fontSize={13} fontWeight="bold">Token Budget Example:</Text>
            <Text fontSize={11}>• Model limit: 8,000 tokens</Text>
            <Text fontSize={11}>• System prompt: 500 tokens</Text>
            <Text fontSize={11}>• Session summary: 200 tokens</Text>
            <Text fontSize={11}>• RAG retrieved: 1,000 tokens</Text>
            <Text fontSize={11}>• Recent messages: 2,000 tokens</Text>
            <Text fontSize={11}>• Reserve for response: 2,000 tokens</Text>
            <Text fontSize={11} fontWeight="bold" color="#ff9800">
              Total used: 5,700 tokens
            </Text>
          </Stack>
        </Row>
      </Box>

      <Card variant="default" width={1000}>
        <Label>Why Hybrid Works Best</Label>
        <Text fontSize={12}>
          By combining strategies, agents maintain both immediate context and long-term memory,
          retrieve relevant information semantically, and stay within token budgets efficiently.
          This mirrors human memory: detailed recent memory, summarized medium-term memory,
          and searchable long-term memory.
        </Text>
      </Card>
    </Stack>
  </Stack>
);

// Slide 7: Practical Considerations
const PracticalConsiderations = () => (
  <Stack gap={32} padding={60} width={1200} height={800}>
    <Title level={1}>Practical Considerations</Title>
    <Subtitle>Implementation details and trade-offs</Subtitle>

    <Row gap={24} marginTop={20}>
      <Stack gap={16} flexGrow={1}>
        <Card variant="primary" width={550}>
          <Label>Token Counting & Budgeting</Label>
          <Stack gap={8}>
            <Text fontSize={12}>• Always count tokens accurately (tiktoken)</Text>
            <Text fontSize={12}>• Reserve buffer for response (20-30%)</Text>
            <Text fontSize={12}>• Monitor costs per conversation</Text>
            <Text fontSize={12}>• Set hard limits for safety</Text>
          </Stack>
        </Card>

        <Card variant="secondary" width={550}>
          <Label>Context Freshness</Label>
          <Stack gap={8}>
            <Text fontSize={12}>• Update summaries incrementally</Text>
            <Text fontSize={12}>• Invalidate stale embeddings</Text>
            <Text fontSize={12}>• Refresh vector DB periodically</Text>
            <Text fontSize={12}>• Balance freshness vs cost</Text>
          </Stack>
        </Card>
      </Stack>

      <Stack gap={16} flexGrow={1}>
        <Card variant="accent" width={550}>
          <Label>Performance Optimization</Label>
          <Stack gap={8}>
            <Text fontSize={12}>• Cache embeddings aggressively</Text>
            <Text fontSize={12}>• Batch vector operations</Text>
            <Text fontSize={12}>• Use streaming for summaries</Text>
            <Text fontSize={12}>• Parallel retrieval when possible</Text>
          </Stack>
        </Card>

        <Card variant="success" width={550}>
          <Label>Quality Assurance</Label>
          <Stack gap={8}>
            <Text fontSize={12}>• Test with various conversation lengths</Text>
            <Text fontSize={12}>• Validate summary quality</Text>
            <Text fontSize={12}>• Check retrieval relevance</Text>
            <Text fontSize={12}>• Monitor user satisfaction metrics</Text>
          </Stack>
        </Card>
      </Stack>
    </Row>

    <Box
      width={1100}
      backgroundColor="#f5f5f5"
      borderColor="#999"
      borderWidth={2}
      borderRadius={8}
      padding={24}
      marginTop={16}
    >
      <Text fontSize={15} fontWeight="bold" marginBottom={12}>Cost-Benefit Analysis</Text>
      <Row gap={40}>
        <Stack gap={8} flexGrow={1}>
          <Text fontSize={12} fontWeight="bold" color="#2e7d32">Low Cost Strategies</Text>
          <Text fontSize={11}>• Sliding window only</Text>
          <Text fontSize={11}>• Minimal summarization</Text>
          <Text fontSize={11}>• Small context windows</Text>
          <Text fontSize={10} color="#666">Best for: Simple tasks, budget-constrained apps</Text>
        </Stack>
        <Stack gap={8} flexGrow={1}>
          <Text fontSize={12} fontWeight="bold" color="#ff9800">Medium Cost Strategies</Text>
          <Text fontSize={11}>• Hierarchical summaries</Text>
          <Text fontSize={11}>• Moderate context windows</Text>
          <Text fontSize={11}>• Selective retrieval</Text>
          <Text fontSize={10} color="#666">Best for: Most production applications</Text>
        </Stack>
        <Stack gap={8} flexGrow={1}>
          <Text fontSize={12} fontWeight="bold" color="#c62828">High Cost Strategies</Text>
          <Text fontSize={11}>• Full vector RAG</Text>
          <Text fontSize={11}>• Large context windows</Text>
          <Text fontSize={11}>• Frequent re-embedding</Text>
          <Text fontSize={10} color="#666">Best for: Enterprise, research applications</Text>
        </Stack>
      </Row>
    </Box>
  </Stack>
);

// Slide 8: Summary & Recommendations
const SummarySlide = () => (
  <Stack gap={32} padding={60} width={1200} height={800}>
    <Title level={1}>Summary & Recommendations</Title>
    
    <Stack gap={20} marginTop={20}>
      <Card variant="primary" width={1100}>
        <Label>Choosing the Right Strategy</Label>
        <Stack gap={10} marginTop={8}>
          <Text fontSize={13}>
            <Text fontWeight="bold">Start Simple:</Text> Begin with sliding window, add complexity as needed
          </Text>
          <Text fontSize={13}>
            <Text fontWeight="bold">Measure First:</Text> Understand your token usage patterns before optimizing
          </Text>
          <Text fontSize={13}>
            <Text fontWeight="bold">User Experience:</Text> Maintain conversation continuity over perfect memory
          </Text>
          <Text fontSize={13}>
            <Text fontWeight="bold">Cost Awareness:</Text> Balance quality with operational expenses
          </Text>
        </Stack>
      </Card>

      <Row gap={24}>
        <Card variant="secondary" width={530}>
          <Label>For Most Applications</Label>
          <Text fontSize={12} marginTop={8}>
            Use a <Text fontWeight="bold">hybrid approach</Text> with:
          </Text>
          <Stack gap={6} marginTop={8}>
            <Text fontSize={11}>• 3-5 recent full messages</Text>
            <Text fontSize={11}>• Session summary</Text>
            <Text fontSize={11}>• Optional RAG for knowledge base</Text>
          </Stack>
          <Text fontSize={11} marginTop={12} color="#666">
            This provides the best balance of context quality,
            cost efficiency, and implementation complexity.
          </Text>
        </Card>

        <Card variant="accent" width={530}>
          <Label>Future Considerations</Label>
          <Stack gap={8} marginTop={8}>
            <Text fontSize={12}>• <Text fontWeight="bold">Longer context windows:</Text> Models improving</Text>
            <Text fontSize={12}>• <Text fontWeight="bold">Better compression:</Text> LLMs learning to summarize</Text>
            <Text fontSize={12}>• <Text fontWeight="bold">Cheaper embeddings:</Text> Vector operations cost dropping</Text>
            <Text fontSize={12}>• <Text fontWeight="bold">Smarter retrieval:</Text> More sophisticated RAG</Text>
          </Stack>
        </Card>
      </Row>

      <Box
        width={1100}
        backgroundColor="#e8f5e9"
        borderColor="#2e7d32"
        borderWidth={3}
        borderRadius={8}
        padding={24}
        marginTop={16}
      >
        <Text fontSize={16} fontWeight="bold" marginBottom={12} color="#2e7d32">Key Takeaway</Text>
        <Text fontSize={14}>
          Context management is not a solved problem but a series of trade-offs.
          The best strategy depends on your specific use case, budget constraints,
          and user expectations. Start simple, measure carefully, and iterate based
          on real usage patterns.
        </Text>
      </Box>
    </Stack>
  </Stack>
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

    console.log('Generating LLM Context Management presentation...\n');

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
  <title>Context Management Strategies in LLM Agents</title>
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
    <h1>Context Management Strategies in LLM Agent Implementations</h1>
    
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
    console.log('\nPresentation complete! Open output/index.html to view.');

  } catch (error) {
    console.error('Error generating presentation:', error);
    process.exit(1);
  }
}

generatePresentation();
