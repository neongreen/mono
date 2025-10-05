/**
 * Agent Loops & Data Flow Visualization Demo
 * 
 * Demonstrates the new Phase 1 components for visualizing:
 * - Agent state machines and loops
 * - Data transformations
 * - Process flows
 * - Timeline sequences
 * - Memory/context management
 */

import React from 'react';
import { writeFileSync } from 'fs';
import {
  Stack, Row, Box, Text, Title, Subtitle,
  StateNode, DecisionNode, ProcessNode, Arrow,
  MemoryBlock, ContextWindow, Timeline, TimelineEvent,
  Card, Badge, Divider,
  renderToSVG
} from '../src';

// Demo 1: Simple Agent Loop
const AgentLoopDemo = () => (
  <Stack width={900} height={500} padding={40} gap={20}>
    <Title level={2}>Agent Loop: Query Processing</Title>
    
    <Row gap={40} marginTop={20}>
      {/* State nodes */}
      <StateNode 
        id="idle" 
        label="Idle" 
        stateType="initial" 
        icon="⏸️"
        width={140}
        height={70}
      />
      
      <StateNode 
        id="retrieving" 
        label="Retrieving Context" 
        stateType="active" 
        icon="🔍"
        width={160}
        height={70}
      />
      
      <StateNode 
        id="processing" 
        label="Processing" 
        stateType="active" 
        icon="⚙️"
        width={140}
        height={70}
      />
      
      <StateNode 
        id="responding" 
        label="Responding" 
        stateType="default" 
        icon="💬"
        width={140}
        height={70}
      />
    </Row>
    
    {/* State transitions */}
    <Arrow from="idle" to="retrieving" label="user query" color="#2196f3" />
    <Arrow from="retrieving" to="processing" label="context ready" color="#4caf50" />
    <Arrow from="processing" to="responding" label="LLM complete" color="#ff9800" />
    <Arrow from="responding" to="idle" label="done" color="#9e9e9e" style="dashed" />
  </Stack>
);

// Demo 2: Decision Flow with Process Nodes
const DecisionFlowDemo = () => (
  <Stack width={900} height={600} padding={40} gap={20}>
    <Title level={2}>Context Retrieval Decision Flow</Title>
    
    <Stack gap={30} marginTop={20} alignItems="center">
      <ProcessNode 
        id="start" 
        label="Receive Query" 
        nodeType="start"
        status="complete"
        width={180}
      />
      
      <Arrow from="start" to="check" color="#2196f3" />
      
      <DecisionNode 
        id="check" 
        label="Context needed?" 
        width={160}
        height={90}
      />
      
      <Row gap={80} marginTop={20}>
        <ProcessNode 
          id="retrieve" 
          label="Retrieve from Vector DB" 
          nodeType="process"
          status="active"
          width={180}
        />
        
        <ProcessNode 
          id="direct" 
          label="Use Direct Prompt" 
          nodeType="process"
          status="default"
          width={180}
        />
      </Row>
      
      <ProcessNode 
        id="llm" 
        label="Send to LLM" 
        nodeType="subprocess"
        status="default"
        width={180}
        marginTop={30}
      />
      
      <Arrow from="check" to="retrieve" label="yes" color="#4caf50" curve="curved" />
      <Arrow from="check" to="direct" label="no" color="#f44336" curve="curved" />
      <Arrow from="retrieve" to="llm" color="#2196f3" />
      <Arrow from="direct" to="llm" color="#2196f3" />
    </Stack>
  </Stack>
);

// Demo 3: Memory & Context Window
const MemoryDemo = () => (
  <Stack width={900} height={500} padding={40} gap={30}>
    <Title level={2}>Memory & Context Management</Title>
    
    <Row gap={30} marginTop={20}>
      <MemoryBlock
        id="vectordb"
        label="Vector Database"
        capacity={10000}
        used={7500}
        unit="embeddings"
        showBar={true}
        showPercentage={true}
        width={280}
        height={120}
      />
      
      <MemoryBlock
        id="cache"
        label="Hot Cache"
        capacity={1000}
        used={950}
        unit="tokens"
        showBar={true}
        showPercentage={true}
        width={280}
        height={120}
      />
    </Row>
    
    <Divider marginTop={20} marginBottom={20} />
    
    <Title level={3}>Context Window Allocation</Title>
    <ContextWindow
      id="context"
      capacity={4096}
      sections={[
        { label: 'System Prompt', tokens: 150, color: '#2196f3' },
        { label: 'Conversation History', tokens: 2800, color: '#4caf50' },
        { label: 'Retrieved Context', tokens: 800, color: '#ff9800' },
        { label: 'User Query', tokens: 200, color: '#f44336' },
        { label: 'Available', tokens: 146, color: '#e0e0e0' }
      ]}
      showLabels={true}
      showPercentages={true}
      width={800}
      height={100}
      marginTop={20}
    />
  </Stack>
);

// Demo 4: Enhanced Arrows
const ArrowStylesDemo = () => (
  <Stack width={900} height={600} padding={40} gap={20}>
    <Title level={2}>Enhanced Arrow Capabilities</Title>
    
    <Row gap={30} marginTop={30}>
      <Box id="a1" width={120} height={60} backgroundColor="#e3f2fd" borderColor="#2196f3" borderWidth={2} borderRadius={6}>
        <Text>Source A</Text>
      </Box>
      <Box id="a2" width={120} height={60} backgroundColor="#f3e5f5" borderColor="#9c27b0" borderWidth={2} borderRadius={6}>
        <Text>Target A</Text>
      </Box>
    </Row>
    
    <Row gap={30} marginTop={30}>
      <Box id="b1" width={120} height={60} backgroundColor="#e3f2fd" borderColor="#2196f3" borderWidth={2} borderRadius={6}>
        <Text>Source B</Text>
      </Box>
      <Box id="b2" width={120} height={60} backgroundColor="#f3e5f5" borderColor="#9c27b0" borderWidth={2} borderRadius={6}>
        <Text>Target B</Text>
      </Box>
    </Row>
    
    <Row gap={30} marginTop={30}>
      <Box id="c1" width={120} height={60} backgroundColor="#e3f2fd" borderColor="#2196f3" borderWidth={2} borderRadius={6}>
        <Text>Source C</Text>
      </Box>
      <Box id="c2" width={120} height={60} backgroundColor="#f3e5f5" borderColor="#9c27b0" borderWidth={2} borderRadius={6}>
        <Text>Target C</Text>
      </Box>
    </Row>
    
    <Row gap={30} marginTop={30}>
      <Box id="d1" width={120} height={60} backgroundColor="#e3f2fd" borderColor="#2196f3" borderWidth={2} borderRadius={6}>
        <Text>Node D1</Text>
      </Box>
      <Box id="d2" width={120} height={60} backgroundColor="#f3e5f5" borderColor="#9c27b0" borderWidth={2} borderRadius={6}>
        <Text>Node D2</Text>
      </Box>
    </Row>
    
    {/* Various arrow styles */}
    <Arrow from="a1" to="a2" label="solid, medium" color="#4caf50" thickness="medium" />
    <Arrow from="b1" to="b2" label="dashed, thick" style="dashed" color="#ff9800" thickness="thick" />
    <Arrow from="c1" to="c2" label="dotted, curved" style="dotted" curve="curved" color="#f44336" />
    <Arrow from="d1" to="d2" bidirectional={true} label="bidirectional" color="#9c27b0" thickness="thick" />
  </Stack>
);

// Demo 5: Timeline Visualization
const TimelineDemo = () => (
  <Stack width={900} height={400} padding={40} gap={20}>
    <Title level={2}>Agent Execution Timeline</Title>
    
    <Timeline orientation="horizontal" showAxis={true} width={800} height={250} marginTop={30}>
      <Row gap={60} marginTop={20}>
        <TimelineEvent
          id="t1"
          time="t0"
          label="Query Received"
          description="0ms"
          color="#2196f3"
          icon="📥"
        />
        
        <TimelineEvent
          id="t2"
          time="t1"
          label="Context Retrieved"
          description="45ms"
          color="#4caf50"
          icon="🔍"
        />
        
        <TimelineEvent
          id="t3"
          time="t2"
          label="LLM Called"
          description="50ms"
          color="#ff9800"
          icon="🤖"
        />
        
        <TimelineEvent
          id="t4"
          time="t3"
          label="Response Ready"
          description="2100ms"
          color="#9c27b0"
          icon="✅"
        />
      </Row>
    </Timeline>
    
    <Badge text="Total: 2.1 seconds" variant="primary" marginTop={20} />
  </Stack>
);

// Demo 6: Complete Agent Loop with Data Flow
const CompleteAgentLoop = () => (
  <Stack width={1100} height={700} padding={40} gap={20}>
    <Title level={2}>Complete Agent Loop: RAG System</Title>
    
    <Stack gap={40} marginTop={30}>
      {/* Row 1: Input */}
      <Row justifyContent="center">
        <ProcessNode 
          id="input" 
          label="User Query" 
          nodeType="start"
          width={160}
          height={60}
        />
      </Row>
      
      {/* Row 2: Embedding */}
      <Row justifyContent="center">
        <ProcessNode 
          id="embed" 
          label="Embed Query" 
          nodeType="process"
          status="complete"
          width={160}
          height={60}
        />
      </Row>
      
      {/* Row 3: Retrieval and Memory */}
      <Row gap={40} justifyContent="center">
        <ProcessNode 
          id="search" 
          label="Vector Search" 
          nodeType="subprocess"
          status="active"
          width={160}
          height={60}
        />
        
        <MemoryBlock
          id="vectorstore"
          label="Vector Store"
          capacity={10000}
          used={7500}
          unit="docs"
          showBar={true}
          width={200}
          height={100}
        />
      </Row>
      
      {/* Row 4: Context Assembly */}
      <Row justifyContent="center">
        <ContextWindow
          id="contextwin"
          capacity={4096}
          sections={[
            { label: 'System', tokens: 150, color: '#2196f3' },
            { label: 'Retrieved', tokens: 800, color: '#4caf50' },
            { label: 'Query', tokens: 200, color: '#ff9800' },
            { label: 'Available', tokens: 2946, color: '#e0e0e0' }
          ]}
          showLabels={true}
          showPercentages={false}
          width={500}
          height={80}
        />
      </Row>
      
      {/* Row 5: LLM */}
      <Row justifyContent="center">
        <ProcessNode 
          id="llm" 
          label="LLM Generate" 
          nodeType="subprocess"
          width={160}
          height={60}
        />
      </Row>
      
      {/* Row 6: Output */}
      <Row justifyContent="center">
        <ProcessNode 
          id="output" 
          label="Return Response" 
          nodeType="end"
          width={160}
          height={60}
        />
      </Row>
    </Stack>
    
    {/* Arrows showing flow */}
    <Arrow from="input" to="embed" color="#2196f3" label="text" />
    <Arrow from="embed" to="search" color="#4caf50" label="embedding" />
    <Arrow from="search" to="contextwin" color="#ff9800" label="top-k docs" />
    <Arrow from="contextwin" to="llm" color="#9c27b0" label="prompt" thickness="thick" />
    <Arrow from="llm" to="output" color="#2196f3" label="completion" />
    
    {/* Memory connection */}
    <Arrow from="vectorstore" to="search" color="#757575" style="dashed" bidirectional={true} label="query" />
  </Stack>
);

// Render all demos
console.log('Rendering Agent Loops & Data Flow demos...');

const demos = [
  { name: 'agent-loop', component: <AgentLoopDemo /> },
  { name: 'decision-flow', component: <DecisionFlowDemo /> },
  { name: 'memory-demo', component: <MemoryDemo /> },
  { name: 'arrow-styles', component: <ArrowStylesDemo /> },
  { name: 'timeline', component: <TimelineDemo /> },
  { name: 'complete-agent', component: <CompleteAgentLoop /> }
];

(async () => {
  for (const { name, component } of demos) {
    const svg = await renderToSVG(component);
    writeFileSync(`phase1-${name}.svg`, svg);
    console.log(`✓ Generated phase1-${name}.svg`);
  }
  
  console.log('\n✓ All Phase 1 demos generated successfully!');
})();
