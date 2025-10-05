/**
 * LLM Context Management Presentation - Version 4
 * 
 * This version uses the new Phase 1 components to showcase:
 * - Agent loops and state machines
 * - Data flow in context retrieval
 * - Memory and storage visualization
 * - Process flows and decisions
 */

import React from 'react';
import { writeFileSync, mkdirSync } from 'fs';
import { 
  Slide, Stack, Row, Box, Text, Arrow, Card, Title, Subtitle,
  List, Section, Well, Badge, Spacer, Divider,
  StateNode, DecisionNode, ProcessNode, MemoryBlock, ContextWindow,
  Timeline, TimelineEvent,
  generateSlideDeck, setCurrentTheme, nordTheme
} from 'diagram-dsl';

// Set theme
setCurrentTheme(nordTheme);

// Slide 1: Title
const TitleSlide = () => (
  <Slide alignItems="center" justifyContent="center">
    <Badge text="LLM Agent Architectures" variant="primary" size="large" marginBottom={20} />
    <Title level={1}>Context Engineering</Title>
    <Subtitle>Agent Loops, Data Flow, and Memory Management</Subtitle>
    
    <Spacer size={60} />
    
    <Section title="Topics Covered" variant="default" width={700}>
      <List
        items={[
          'Agent state machines and loops',
          'Context retrieval data flows',
          'Memory and storage patterns',
          'Decision points in context selection'
        ]}
        fontSize={14}
      />
    </Section>
  </Slide>
);

// Slide 2: Agent State Machine
const AgentStateMachine = () => (
  <Slide>
    <Title level={1}>Agent Lifecycle: State Machine</Title>
    <Subtitle>Understanding the agent's execution states</Subtitle>
    
    <Stack gap={40} marginTop={40} alignItems="center">
      <Row gap={40}>
        <StateNode 
          id="idle_s2" 
          label="Idle" 
          stateType="initial" 
          icon="⏸️"
          width={140}
          height={70}
        />
        
        <StateNode 
          id="retrieving_s2" 
          label="Retrieving Context" 
          stateType="active" 
          icon="🔍"
          width={180}
          height={70}
        />
      </Row>
      
      <Row gap={40}>
        <StateNode 
          id="responding_s2" 
          label="Responding" 
          stateType="default" 
          icon="💬"
          width={140}
          height={70}
        />
        
        <StateNode 
          id="processing_s2" 
          label="Processing" 
          stateType="active" 
          icon="⚙️"
          width={140}
          height={70}
        />
      </Row>
    </Stack>
    
    <Arrow from="idle_s2" to="retrieving_s2" label="user query" color="#88c0d0" thickness="medium" />
    <Arrow from="retrieving_s2" to="processing_s2" label="context ready" color="#a3be8c" thickness="medium" />
    <Arrow from="processing_s2" to="responding_s2" label="LLM complete" color="#ebcb8b" thickness="medium" />
    <Arrow from="responding_s2" to="idle_s2" label="response sent" color="#d08770" style="dashed" curve="curved" />
    
    <Well variant="info" width={1000} marginTop={40}>
      <Text fontSize={14}>
        The agent cycles through states, managing context at each transition.
        The loop back to Idle allows continuous query processing.
      </Text>
    </Well>
  </Slide>
);

// Slide 3: Decision Flow
const ContextDecisionFlow = () => (
  <Slide>
    <Title level={1}>Context Retrieval Decision Tree</Title>
    <Subtitle>How agents decide what context to retrieve</Subtitle>
    
    <Stack gap={30} marginTop={30} alignItems="center">
      <ProcessNode 
        id="query_s3" 
        label="Incoming Query" 
        nodeType="start"
        width={180}
        height={60}
      />
      
      <Arrow from="query_s3" to="classify_s3" color="#88c0d0" />
      
      <DecisionNode 
        id="classify_s3" 
        label="Need context?" 
        width={180}
        height={90}
      />
      
      <Row gap={100} marginTop={20}>
        <Stack gap={20} alignItems="center">
          <ProcessNode 
            id="retrieve_s3" 
            label="Vector Retrieval" 
            nodeType="subprocess"
            status="active"
            width={180}
          />
          
          <MemoryBlock
            id="vectordb_s3"
            label="Vector Store"
            capacity={10000}
            used={7500}
            unit="docs"
            showBar={true}
            width={200}
            height={100}
          />
        </Stack>
        
        <ProcessNode 
          id="direct_s3" 
          label="Direct Prompt" 
          nodeType="process"
          width={180}
        />
      </Row>
      
      <ProcessNode 
        id="llm_s3" 
        label="Send to LLM" 
        nodeType="end"
        width={180}
        marginTop={30}
      />
    </Stack>
    
    <Arrow from="classify_s3" to="retrieve_s3" label="complex" color="#a3be8c" curve="curved" />
    <Arrow from="classify_s3" to="direct_s3" label="simple" color="#bf616a" curve="curved" />
    <Arrow from="retrieve_s3" to="llm_s3" color="#88c0d0" />
    <Arrow from="direct_s3" to="llm_s3" color="#88c0d0" />
    <Arrow from="vectordb_s3" to="retrieve_s3" color="#d08770" style="dashed" bidirectional={true} label="query" />
  </Slide>
);

// Slide 4: Context Window Management
const ContextWindowSlide = () => (
  <Slide>
    <Title level={1}>Context Window Budget</Title>
    <Subtitle>How tokens are allocated across different components</Subtitle>
    
    <Stack gap={30} marginTop={40} alignItems="center">
      <ContextWindow
        id="context_s4"
        capacity={4096}
        sections={[
          { label: 'System Prompt', tokens: 150, color: '#5e81ac' },
          { label: 'Conversation History', tokens: 2400, color: '#88c0d0' },
          { label: 'Retrieved Context', tokens: 900, color: '#a3be8c' },
          { label: 'User Query', tokens: 300, color: '#ebcb8b' },
          { label: 'Reserved for Output', tokens: 346, color: '#d8dee9' }
        ]}
        showLabels={true}
        showPercentages={true}
        width={900}
        height={120}
      />
      
      <Spacer size={20} />
      
      <Row gap={40}>
        <Card variant="primary" width={400}>
          <Title level={4}>💡 Key Insight</Title>
          <Text fontSize={13} marginTop={10}>
            Each component competes for limited token budget. 
            Smart allocation is crucial for performance.
          </Text>
        </Card>
        
        <Card variant="warning" width={400}>
          <Title level={4}>⚠️ Trade-offs</Title>
          <Text fontSize={13} marginTop={10}>
            More context = better accuracy but higher cost and latency.
            Balance is key.
          </Text>
        </Card>
      </Row>
      
      <Divider width={900} marginTop={20} marginBottom={20} />
      
      <Row gap={30}>
        <MemoryBlock
          id="input_mem"
          label="Input Tokens"
          capacity={4096}
          used={3750}
          unit="tokens"
          showBar={true}
          width={260}
          height={100}
        />
        
        <MemoryBlock
          id="output_mem"
          label="Output Budget"
          capacity={2048}
          used={346}
          unit="tokens"
          showBar={true}
          width={260}
          height={100}
        />
        
        <MemoryBlock
          id="cost_mem"
          label="Est. Cost"
          capacity={100}
          used={42}
          unit="cents"
          showBar={true}
          width={260}
          height={100}
        />
      </Row>
    </Stack>
  </Slide>
);

// Slide 5: RAG Pipeline
const RAGPipeline = () => (
  <Slide>
    <Title level={1}>RAG Pipeline: Complete Data Flow</Title>
    <Subtitle>From query to response with context retrieval</Subtitle>
    
    <Stack gap={35} marginTop={30} alignItems="center">
      <ProcessNode 
        id="query_s5" 
        label="User Query" 
        nodeType="start"
        width={160}
        height={60}
      />
      
      <Arrow from="query_s5" to="embed_s5" color="#88c0d0" label="text" />
      
      <ProcessNode 
        id="embed_s5" 
        label="Embed Query" 
        nodeType="process"
        status="complete"
        width={160}
        height={60}
      />
      
      <Arrow from="embed_s5" to="search_s5" color="#a3be8c" label="vector" />
      
      <Row gap={40}>
        <ProcessNode 
          id="search_s5" 
          label="Vector Search" 
          nodeType="subprocess"
          status="active"
          width={160}
          height={60}
        />
        
        <MemoryBlock
          id="vs_s5"
          label="Vector Store"
          capacity={10000}
          used={7500}
          unit="docs"
          showBar={true}
          width={200}
          height={100}
        />
      </Row>
      
      <Arrow from="search_s5" to="assemble_s5" color="#ebcb8b" label="top-k" />
      
      <ProcessNode 
        id="assemble_s5" 
        label="Assemble Context" 
        nodeType="process"
        width={180}
        height={60}
      />
      
      <Arrow from="assemble_s5" to="llm_s5" color="#b48ead" label="prompt" thickness="thick" />
      
      <ProcessNode 
        id="llm_s5" 
        label="LLM Generate" 
        nodeType="subprocess"
        width={160}
        height={60}
      />
      
      <Arrow from="llm_s5" to="output_s5" color="#88c0d0" label="response" />
      
      <ProcessNode 
        id="output_s5" 
        label="Return to User" 
        nodeType="end"
        width={160}
        height={60}
      />
    </Stack>
    
    <Arrow from="vs_s5" to="search_s5" color="#d08770" style="dashed" bidirectional={true} />
  </Slide>
);

// Slide 6: Timeline
const ExecutionTimeline = () => (
  <Slide>
    <Title level={1}>Execution Timeline</Title>
    <Subtitle>Time breakdown of a typical RAG query</Subtitle>
    
    <Timeline orientation="horizontal" showAxis={true} width={900} height={280} marginTop={40}>
      <Row gap={70} marginTop={30}>
        <TimelineEvent
          id="t1"
          time="0ms"
          label="Query Start"
          description="0ms"
          color="#88c0d0"
          icon="📥"
        />
        
        <TimelineEvent
          id="t2"
          time="10ms"
          label="Embedding"
          description="10ms"
          color="#a3be8c"
          icon="🔢"
        />
        
        <TimelineEvent
          id="t3"
          time="55ms"
          label="Retrieval"
          description="45ms"
          color="#ebcb8b"
          icon="🔍"
        />
        
        <TimelineEvent
          id="t4"
          time="2100ms"
          label="LLM Call"
          description="2045ms"
          color="#b48ead"
          icon="🤖"
        />
        
        <TimelineEvent
          id="t5"
          time="2110ms"
          label="Complete"
          description="10ms"
          color="#88c0d0"
          icon="✅"
        />
      </Row>
    </Timeline>
    
    <Stack gap={20} marginTop={40} alignItems="center">
      <Badge text="Total Time: 2.11 seconds" variant="primary" size="large" />
      
      <Row gap={30} marginTop={20}>
        <Card variant="info" width={350}>
          <Title level={4}>⚡ Fastest</Title>
          <Text fontSize={13} marginTop={8}>
            Query & Embedding: ~10ms
          </Text>
        </Card>
        
        <Card variant="warning" width={350}>
          <Title level={4}>🐌 Slowest</Title>
          <Text fontSize={13} marginTop={8}>
            LLM Generation: ~2 seconds (97% of time)
          </Text>
        </Card>
      </Row>
    </Stack>
  </Slide>
);

// Slide 7: Memory Hierarchy
const MemoryHierarchy = () => (
  <Slide>
    <Title level={1}>Memory Hierarchy</Title>
    <Subtitle>Multi-tier storage for efficient context retrieval</Subtitle>
    
    <Stack gap={30} marginTop={40} alignItems="center">
      <MemoryBlock
        id="hot_cache"
        label="Hot Cache (RAM)"
        capacity={1000}
        used={950}
        unit="entries"
        showBar={true}
        showPercentage={true}
        width={350}
        height={120}
        backgroundColor="#bf616a20"
      />
      
      <Arrow from="hot_cache" to="warm_cache" color="#d08770" label="miss" style="dashed" />
      
      <MemoryBlock
        id="warm_cache"
        label="Warm Cache (SSD)"
        capacity={10000}
        used={7500}
        unit="entries"
        showBar={true}
        showPercentage={true}
        width={350}
        height={120}
        backgroundColor="#ebcb8b20"
      />
      
      <Arrow from="warm_cache" to="cold_storage" color="#d08770" label="miss" style="dashed" />
      
      <MemoryBlock
        id="cold_storage"
        label="Cold Storage (Disk)"
        capacity={1000000}
        used={850000}
        unit="entries"
        showBar={true}
        showPercentage={true}
        width={350}
        height={120}
        backgroundColor="#a3be8c20"
      />
    </Stack>
    
    <Well variant="info" width={900} marginTop={40}>
      <Title level={4}>💡 Caching Strategy</Title>
      <Text fontSize={13} marginTop={10}>
        Hot cache: Frequently accessed context (sub-millisecond access)
      </Text>
      <Text fontSize={13} marginTop={5}>
        Warm cache: Recently used context (5-10ms access)
      </Text>
      <Text fontSize={13} marginTop={5}>
        Cold storage: Full vector database (20-50ms access)
      </Text>
    </Well>
  </Slide>
);

// Slide 8: Summary
const Summary = () => (
  <Slide alignItems="center" justifyContent="center">
    <Title level={1}>Key Takeaways</Title>
    
    <Stack gap={25} marginTop={40} width={900}>
      <Card variant="primary" width={900}>
        <Row gap={20}>
          <Badge text="1" variant="primary" size="large" />
          <Stack gap={8}>
            <Title level={3}>Agent State Machines</Title>
            <Text fontSize={14}>
              Agents cycle through well-defined states. Understanding these transitions
              helps optimize context retrieval timing.
            </Text>
          </Stack>
        </Row>
      </Card>
      
      <Card variant="secondary" width={900}>
        <Row gap={20}>
          <Badge text="2" variant="secondary" size="large" />
          <Stack gap={8}>
            <Title level={3}>Data Flow Optimization</Title>
            <Text fontSize={14}>
              The RAG pipeline has clear bottlenecks. LLM calls dominate latency,
              but smart caching can reduce retrieval overhead.
            </Text>
          </Stack>
        </Row>
      </Card>
      
      <Card variant="accent" width={900}>
        <Row gap={20}>
          <Badge text="3" variant="accent" size="large" />
          <Stack gap={8}>
            <Title level={3}>Memory Hierarchy</Title>
            <Text fontSize={14}>
              Multi-tier caching dramatically improves performance. Design your
              system with hot/warm/cold storage in mind.
            </Text>
          </Stack>
        </Row>
      </Card>
    </Stack>
  </Slide>
);

// Generate slide deck
const slides = [
  { name: 'TitleSlide', component: <TitleSlide /> },
  { name: 'AgentStateMachine', component: <AgentStateMachine /> },
  { name: 'ContextDecisionFlow', component: <ContextDecisionFlow /> },
  { name: 'ContextWindowSlide', component: <ContextWindowSlide /> },
  { name: 'RAGPipeline', component: <RAGPipeline /> },
  { name: 'ExecutionTimeline', component: <ExecutionTimeline /> },
  { name: 'MemoryHierarchy', component: <MemoryHierarchy /> },
  { name: 'Summary', component: <Summary /> },
];

console.log('Generating Context Engineering presentation with Agent Focus...');

try {
  mkdirSync('output-v4-agent', { recursive: true });
} catch (e) {
  // Directory already exists
}

(async () => {
  await generateSlideDeck(slides, {
    outputDir: 'output-v4-agent',
    width: 1200,
    height: 800,
    backgroundColor: '#ffffff'
  });

  console.log('✓ Presentation generated in output-v4-agent/');
  console.log('✓ 8 slides created showcasing agent loops and data flows');
})();
