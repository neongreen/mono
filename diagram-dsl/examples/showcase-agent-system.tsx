/**
 * Showcase: Complete Agent System with Context Management
 * 
 * This demonstrates the full power of Phase 1 components
 * in a single comprehensive diagram showing:
 * - Agent state machine
 * - Data flow through the system
 * - Memory and storage
 * - Context window management
 * - Decision points
 */

import React from 'react';
import { writeFileSync } from 'fs';
import {
  Stack, Row, Box, Text, Title,
  StateNode, DecisionNode, ProcessNode, MemoryBlock, ContextWindow,
  Arrow, Badge, Divider,
  renderToSVG
} from '../src';

const CompleteAgentSystem = () => (
  <Stack width={1400} height={900} padding={40} gap={15}>
    <Title level={1}>LLM Agent System: Complete Architecture</Title>
    <Text fontSize={14} color="#666">Real-time context-aware agent with RAG and memory management</Text>
    
    {/* Top Row: User Input and Agent States */}
    <Row gap={50} marginTop={20} justifyContent="space-between">
      {/* User Input */}
      <ProcessNode 
        id="user_input" 
        label="User Query" 
        nodeType="start"
        width={140}
        height={60}
      />
      
      {/* Agent State Machine */}
      <Stack gap={15}>
        <Badge text="Agent States" variant="primary" />
        <Row gap={20}>
          <StateNode 
            id="idle" 
            label="Idle" 
            stateType="initial"
            width={100}
            height={50}
          />
          <StateNode 
            id="active" 
            label="Processing" 
            stateType="active"
            width={110}
            height={50}
          />
          <StateNode 
            id="done" 
            label="Done" 
            stateType="final"
            width={100}
            height={50}
          />
        </Row>
      </Stack>
      
      {/* System Metrics */}
      <Stack gap={10}>
        <Badge text="System Status" variant="success" />
        <MemoryBlock
          id="sys_mem"
          label="Available Memory"
          capacity={1000}
          used={650}
          unit="MB"
          showBar={true}
          width={180}
          height={90}
        />
      </Stack>
    </Row>
    
    <Divider width={1320} marginTop={10} marginBottom={10} />
    
    {/* Main Processing Flow */}
    <Row gap={40}>
      {/* Left Column: Query Processing */}
      <Stack gap={25} width={400}>
        <Badge text="Query Processing" variant="secondary" />
        
        <ProcessNode 
          id="embed_query" 
          label="Embed Query" 
          nodeType="process"
          status="complete"
          width={180}
          height={60}
        />
        
        <DecisionNode 
          id="need_context" 
          label="Need context?" 
          width={160}
          height={80}
        />
        
        <Row gap={30}>
          <ProcessNode 
            id="direct" 
            label="Direct" 
            nodeType="process"
            width={100}
            height={50}
          />
          
          <ProcessNode 
            id="retrieve" 
            label="Retrieve" 
            nodeType="subprocess"
            status="active"
            width={100}
            height={50}
          />
        </Row>
      </Stack>
      
      {/* Middle Column: Memory & Storage */}
      <Stack gap={20} width={350}>
        <Badge text="Memory Hierarchy" variant="info" />
        
        <MemoryBlock
          id="hot_cache"
          label="Hot Cache"
          capacity={100}
          used={95}
          unit="queries"
          showBar={true}
          showPercentage={true}
          width={300}
          height={100}
          backgroundColor="#ffebee"
        />
        
        <MemoryBlock
          id="vector_db"
          label="Vector Database"
          capacity={50000}
          used={38500}
          unit="docs"
          showBar={true}
          showPercentage={true}
          width={300}
          height={100}
          backgroundColor="#e3f2fd"
        />
        
        <MemoryBlock
          id="conversation"
          label="Conv. History"
          capacity={50}
          used={23}
          unit="messages"
          showBar={true}
          showPercentage={true}
          width={300}
          height={100}
          backgroundColor="#f3e5f5"
        />
      </Stack>
      
      {/* Right Column: Context Assembly & LLM */}
      <Stack gap={25} width={450}>
        <Badge text="Context Assembly & Generation" variant="accent" />
        
        <ContextWindow
          id="context_window"
          capacity={4096}
          sections={[
            { label: 'System', tokens: 150, color: '#5e81ac' },
            { label: 'History', tokens: 1500, color: '#88c0d0' },
            { label: 'Retrieved', tokens: 1200, color: '#a3be8c' },
            { label: 'Query', tokens: 300, color: '#ebcb8b' },
            { label: 'Reserved', tokens: 946, color: '#d8dee9' }
          ]}
          showLabels={true}
          showPercentages={false}
          width={420}
          height={90}
        />
        
        <ProcessNode 
          id="llm_call" 
          label="LLM Generate" 
          nodeType="subprocess"
          status="active"
          width={200}
          height={70}
        />
        
        <ProcessNode 
          id="output" 
          label="Return Response" 
          nodeType="end"
          width={200}
          height={60}
        />
      </Stack>
    </Row>
    
    {/* Data Flow Arrows */}
    {/* User to Processing */}
    <Arrow from="user_input" to="embed_query" label="text" color="#2196f3" thickness="medium" />
    <Arrow from="embed_query" to="need_context" label="vector" color="#4caf50" />
    
    {/* Decision branches */}
    <Arrow from="need_context" to="direct" label="simple" color="#9e9e9e" curve="curved" />
    <Arrow from="need_context" to="retrieve" label="complex" color="#ff9800" curve="curved" thickness="thick" />
    
    {/* Memory interactions */}
    <Arrow from="retrieve" to="hot_cache" color="#f44336" style="dashed" bidirectional={true} label="check" />
    <Arrow from="retrieve" to="vector_db" color="#2196f3" style="dashed" bidirectional={true} label="search" />
    <Arrow from="retrieve" to="conversation" color="#9c27b0" style="dashed" label="recent" />
    
    {/* Assembly */}
    <Arrow from="hot_cache" to="context_window" label="cached" color="#ff5722" curve="curved" />
    <Arrow from="vector_db" to="context_window" label="docs" color="#03a9f4" curve="curved" />
    <Arrow from="conversation" to="context_window" label="history" color="#ab47bc" curve="curved" />
    <Arrow from="direct" to="context_window" label="query" color="#607d8b" curve="arc" />
    <Arrow from="retrieve" to="context_window" label="enriched" color="#ffa726" curve="arc" />
    
    {/* Generation */}
    <Arrow from="context_window" to="llm_call" label="prompt" color="#1976d2" thickness="very-thick" />
    <Arrow from="llm_call" to="output" label="completion" color="#388e3c" thickness="thick" />
    
    {/* State transitions */}
    <Arrow from="idle" to="active" color="#9e9e9e" style="dotted" />
    <Arrow from="active" to="done" color="#9e9e9e" style="dotted" />
    
    {/* Bottom Info */}
    <Row gap={20} marginTop={20} justifyContent="center">
      <Badge text="💡 Hot Cache: <50ms" variant="success" />
      <Badge text="🔍 Vector Search: 20-50ms" variant="info" />
      <Badge text="🤖 LLM Call: 500-2000ms" variant="warning" />
      <Badge text="📊 Total Budget: 4096 tokens" variant="secondary" />
    </Row>
  </Stack>
);

console.log('🎨 Generating Complete Agent System Showcase...\n');

(async () => {
  const svg = await renderToSVG(<CompleteAgentSystem />);
  writeFileSync('showcase-complete-agent-system.svg', svg);
  console.log('✅ Generated showcase-complete-agent-system.svg');
  console.log('\nThis diagram shows:');
  console.log('  • Agent state machine (idle → processing → done)');
  console.log('  • Query processing with decision logic');
  console.log('  • 3-tier memory hierarchy (hot cache, vector DB, conversation history)');
  console.log('  • Context window token allocation');
  console.log('  • Complete data flow from user input to LLM output');
  console.log('  • Bidirectional memory access');
  console.log('  • Enhanced arrows (thick, dashed, curved, bidirectional)');
  console.log('\n✨ This is the power of Phase 1 components!');
})();
