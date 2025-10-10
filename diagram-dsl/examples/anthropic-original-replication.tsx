/**
 * Replication of Original Anthropic Diagram
 * "Prompt engineering vs. context engineering"
 * 
 * This example replicates the specific diagram from Anthropic's documentation
 * showing the comparison between single-turn prompt engineering and
 * multi-turn context engineering for agents.
 * 
 * Demonstrates new arrow features:
 * - Custom attachment points (fromSide, toSide)
 * - Arrow offsets (fromOffset, toOffset)
 * - Arrow shortening (shortenEnd)
 * - Multi-target forks (toMultiple)
 */

import React from 'react';
import { writeFileSync } from 'fs';
import {
  Stack, Row, Box, Text, Arrow,
  renderToSVG
} from '../src';

const AnthropicOriginalDiagram = () => (
  <Stack width={1600} height={950} padding={24} gap={0} backgroundColor="#f7f7f5" alignItems="center">
    {/* Title - centered */}
    <Text fontSize={32} fontWeight="bold" textAlign="center" marginBottom={24}>
      Prompt engineering vs. context engineering
    </Text>

    {/* Two main panels side by side with 40px gap */}
    {/* Left panel narrower, right panel wider (approximately 1:1.5 ratio) */}
    <Row gap={40} alignItems="flex-start" justifyContent="center">
      {/* Left Panel - Single Turn - Narrower panel (about 35% of total width) */}
      <Stack 
        width={550}
        height={780}
        backgroundColor="#fafafa"
        borderRadius={12}
        padding={20}
        gap={0}
      >
        {/* Header - top aligned with good separation from label below */}
        <Text fontSize={18} fontWeight="bold" color="#666" marginBottom={16}>
          Prompt engineering for single turn queries
        </Text>

        {/* Context Window label - closer to its box (6px) than to header (16px) */}
        <Text fontSize={14} color="#666" marginBottom={6}>
          Context window
        </Text>
        
        {/* Dashed box representing context window - fills remaining height with proper inset */}
        <Box
          id="left-context"
          width={510}
          height={690}
          backgroundColor="transparent"
          borderColor="#888888"
          borderWidth={2}
          borderRadius={8}
          padding={20}
        >
          <Stack gap={8} alignItems="flex-start">
            {/* System Prompt bubble */}
            <Box
              id="left-sys-prompt"
              width={140}
              height={40}
              backgroundColor="#6fb4ff"
              borderColor="#4a9ae5"
              borderWidth={2}
              borderRadius={6}
              padding={6}
              justifyContent="center"
              alignItems="center"
            >
              <Text fontSize={12} fontWeight="bold">System prompt</Text>
            </Box>

            {/* User Message bubble */}
            <Box
              id="left-user-msg"
              width={140}
              height={40}
              backgroundColor="#6fb4ff"
              borderColor="#4a9ae5"
              borderWidth={2}
              borderRadius={6}
              padding={6}
              justifyContent="center"
              alignItems="center"
            >
              <Text fontSize={12} fontWeight="bold">User message</Text>
            </Box>

            {/* Large empty space to show unused context */}
            <Box height={480} width={100} />
            
            {/* Scissors icon at bottom */}
            <Text fontSize={30}>✂️</Text>
          </Stack>
        </Box>
      </Stack>

      {/* Model and output for left side - positioned outside/beside panel */}
      <Stack gap={20} marginTop={350}>
        <Box
          id="left-model"
          width={60}
          height={60}
          backgroundColor="transparent"
        >
          <Stack gap={6}>
            <Row gap={6}>
              <Box width={10} height={10} backgroundColor="#a3d5ff" borderRadius={5} />
              <Box width={10} height={10} backgroundColor="#b8e6b8" borderRadius={5} />
              <Box width={10} height={10} backgroundColor="#ffe6a3" borderRadius={5} />
            </Row>
            <Row gap={6}>
              <Box width={10} height={10} backgroundColor="#b8e6b8" borderRadius={5} />
              <Box width={10} height={10} backgroundColor="#a3d5ff" borderRadius={5} />
              <Box width={10} height={10} backgroundColor="#ffe6a3" borderRadius={5} />
            </Row>
            <Row gap={6}>
              <Box width={10} height={10} backgroundColor="#ffe6a3" borderRadius={5} />
              <Box width={10} height={10} backgroundColor="#b8e6b8" borderRadius={5} />
              <Box width={10} height={10} backgroundColor="#a3d5ff" borderRadius={5} />
            </Row>
          </Stack>
        </Box>

        {/* Assistant message output */}
        <Box
          id="left-output"
          width={110}
          height={38}
          backgroundColor="white"
          borderColor="#444444"
          borderWidth={1}
          borderRadius={4}
          padding={6}
          justifyContent="center"
          alignItems="center"
        >
          <Text fontSize={11}>Assistant message</Text>
        </Box>
      </Stack>

      {/* Right Panel - Multi Turn with Agent - Wider panel (about 52% of total width) */}
      <Stack 
        width={820}
        height={780}
        backgroundColor="#fafafa"
        borderRadius={12}
        padding={20}
        gap={0}
      >
        {/* Header - top aligned with good separation from labels below */}
        <Text fontSize={18} fontWeight="bold" color="#666" marginBottom={16}>
          Context engineering for agents
        </Text>

        {/* Two dashed boxes side by side */}
        <Row gap={12} alignItems="flex-start">
          {/* Left: Possible context */}
          <Stack gap={0}>
            {/* Label - closer to its box (6px) than to header (16px) */}
            <Text fontSize={14} color="#666" marginBottom={6}>
              Possible context to give model
            </Text>
            
            <Box
              id="right-possible"
              width={360}
              height={690}
              backgroundColor="transparent"
              borderColor="#888888"
              borderWidth={2}
              borderRadius={8}
              padding={12}
            >
              <Stack gap={8} alignItems="flex-start">
                {/* Green doc cards (folded corner style) */}
                <Box width={120} height={36} backgroundColor="#90EE90" borderRadius={4} padding={6}
                     justifyContent="center">
                  <Text fontSize={10}>Doc 1</Text>
                </Box>
                <Box width={120} height={36} backgroundColor="#90EE90" borderRadius={4} padding={6}
                     justifyContent="center">
                  <Text fontSize={10}>Doc 2</Text>
                </Box>
                <Box width={120} height={36} backgroundColor="#90EE90" borderRadius={4} padding={6}
                     justifyContent="center">
                  <Text fontSize={10}>Doc 3</Text>
                </Box>

                {/* Orange tool pills */}
                <Box width={120} height={36} backgroundColor="#FFA500" borderRadius={18} padding={6}
                     justifyContent="center">
                  <Text fontSize={10}>Tool 1</Text>
                </Box>
                <Box width={120} height={36} backgroundColor="#FFA500" borderRadius={18} padding={6}
                     justifyContent="center">
                  <Text fontSize={10}>Tool 2</Text>
                </Box>
                <Box width={120} height={36} backgroundColor="#FFA500" borderRadius={18} padding={6}
                     justifyContent="center">
                  <Text fontSize={10}>Tool 3</Text>
                </Box>
                <Box width={120} height={36} backgroundColor="#FFA500" borderRadius={18} padding={6}
                     justifyContent="center">
                  <Text fontSize={10}>Tool 4</Text>
                </Box>

                {/* Violet memory files */}
                <Box width={120} height={36} backgroundColor="#DDA0DD" borderRadius={4} padding={6}
                     justifyContent="center">
                  <Text fontSize={10}>Memory file</Text>
                </Box>
                <Box width={120} height={36} backgroundColor="#DDA0DD" borderRadius={4} padding={6}
                     justifyContent="center">
                  <Text fontSize={10}>Memory file</Text>
                </Box>

                {/* Wide blue cards */}
                <Box width={180} height={36} backgroundColor="#6fb4ff" borderRadius={4} padding={6}
                     justifyContent="center">
                  <Text fontSize={10} fontWeight="bold">Comprehensive instructions</Text>
                </Box>
                <Box width={180} height={36} backgroundColor="#6fb4ff" borderRadius={4} padding={6}
                     justifyContent="center">
                  <Text fontSize={10} fontWeight="bold">Domain knowledge</Text>
                </Box>

                {/* White message history */}
                <Box id="right-history" width={140} height={38} backgroundColor="white" 
                     borderColor="#888" borderWidth={1} borderRadius={4} padding={6}
                     justifyContent="center">
                  <Text fontSize={10}>Message history</Text>
                </Box>
              </Stack>
            </Box>
          </Stack>

          {/* Right: Curated context window */}
          <Stack gap={0}>
            {/* Label - closer to its box (6px) than to header (16px) */}
            <Text fontSize={14} color="#666" marginBottom={6}>
              Context window
            </Text>
            
            <Box
              id="right-context"
              width={360}
              height={690}
              backgroundColor="transparent"
              borderColor="#888888"
              borderWidth={2}
              borderRadius={8}
              padding={12}
            >
              <Stack gap={8} alignItems="flex-start">
                {/* System prompt */}
                <Box width={140} height={36} backgroundColor="#6fb4ff" borderRadius={4} padding={6}
                     justifyContent="center">
                  <Text fontSize={10} fontWeight="bold">System prompt</Text>
                </Box>

                {/* Doc cards side by side */}
                <Row gap={6}>
                  <Box width={67} height={36} backgroundColor="#90EE90" borderRadius={4} padding={4}
                       justifyContent="center">
                    <Text fontSize={9}>Doc 1</Text>
                  </Box>
                  <Box width={67} height={36} backgroundColor="#90EE90" borderRadius={4} padding={4}
                       justifyContent="center">
                    <Text fontSize={9}>Doc 2</Text>
                  </Box>
                </Row>

                {/* Memory file */}
                <Box width={140} height={36} backgroundColor="#DDA0DD" borderRadius={4} padding={6}
                     justifyContent="center">
                  <Text fontSize={10}>Memory file</Text>
                </Box>

                {/* Tool pills side by side */}
                <Row gap={6}>
                  <Box width={67} height={36} backgroundColor="#FFA500" borderRadius={18} padding={4}
                       justifyContent="center">
                    <Text fontSize={9}>Tool 1</Text>
                  </Box>
                  <Box width={67} height={36} backgroundColor="#FFA500" borderRadius={18} padding={4}
                       justifyContent="center">
                    <Text fontSize={9}>Tool 2</Text>
                  </Box>
                </Row>

                {/* User message */}
                <Box width={140} height={36} backgroundColor="#6fb4ff" borderRadius={4} padding={6}
                     justifyContent="center">
                  <Text fontSize={10} fontWeight="bold">User message</Text>
                </Box>

                {/* Message history */}
                <Box width={140} height={38} backgroundColor="white" borderColor="#888" 
                     borderWidth={1} borderRadius={4} padding={6}
                     justifyContent="center">
                  <Text fontSize={10}>Message history</Text>
                </Box>

                {/* Empty space and scissors */}
                <Box height={280} />
                <Text fontSize={30}>✂️</Text>
              </Stack>
            </Box>
          </Stack>
        </Row>
      </Stack>

      {/* Model and outputs for right side - positioned outside/beside panel */}
      <Stack gap={20} marginTop={350}>
        <Box
          id="right-model"
          width={60}
          height={60}
          backgroundColor="transparent"
        >
          <Stack gap={6}>
            <Row gap={6}>
              <Box width={10} height={10} backgroundColor="#a3d5ff" borderRadius={5} />
              <Box width={10} height={10} backgroundColor="#b8e6b8" borderRadius={5} />
              <Box width={10} height={10} backgroundColor="#ffe6a3" borderRadius={5} />
            </Row>
            <Row gap={6}>
              <Box width={10} height={10} backgroundColor="#b8e6b8" borderRadius={5} />
              <Box width={10} height={10} backgroundColor="#a3d5ff" borderRadius={5} />
              <Box width={10} height={10} backgroundColor="#ffe6a3" borderRadius={5} />
            </Row>
            <Row gap={6}>
              <Box width={10} height={10} backgroundColor="#ffe6a3" borderRadius={5} />
              <Box width={10} height={10} backgroundColor="#b8e6b8" borderRadius={5} />
              <Box width={10} height={10} backgroundColor="#a3d5ff" borderRadius={5} />
            </Row>
          </Stack>
        </Box>

        {/* Outputs */}
        <Stack gap={12}>
          {/* Assistant message output */}
          <Box
            id="right-output"
            width={110}
            height={38}
            backgroundColor="white"
            borderColor="#444444"
            borderWidth={1}
            borderRadius={4}
            padding={6}
            justifyContent="center"
            alignItems="center"
          >
            <Text fontSize={11}>Assistant message</Text>
          </Box>

          {/* Tool call */}
          <Box
            id="right-tool"
            width={120}
            height={38}
            backgroundColor="#f8e0ba"
            borderColor="#d4b896"
            borderWidth={1}
            borderRadius={4}
            padding={6}
            justifyContent="center"
            alignItems="center"
          >
            <Text fontSize={11}>Tool call</Text>
          </Box>

          {/* Tool result */}
          <Box
            id="right-result"
            width={120}
            height={38}
            backgroundColor="#f8e0ba"
            borderColor="#d4b896"
            borderWidth={1}
            borderRadius={4}
            padding={6}
            justifyContent="center"
            alignItems="center"
          >
            <Text fontSize={11}>Tool result</Text>
          </Box>
        </Stack>
      </Stack>
    </Row>

    {/* Left side arrows */}
    <Arrow from="left-context" to="left-model" 
           fromSide="right" toSide="left" 
           color="#4c4c4c" strokeWidth={2} />
    <Arrow from="left-model" to="left-output" 
           fromSide="right" toSide="left"
           color="#4c4c4c" strokeWidth={2} />
    
    {/* Right side arrows with custom attachment */}
    <Arrow from="right-possible" to="right-context" 
           label="Curation" 
           fromSide="right" toSide="left"
           shortenEnd={10}
           color="#4c4c4c" strokeWidth={2} />
    <Arrow from="right-context" to="right-model" 
           fromSide="right" toSide="left"
           color="#4c4c4c" strokeWidth={2} />
    
    {/* Y-shaped fork from model to outputs */}
    <Arrow from="right-model" to="right-output"
           fromSide="right" toSide="left"
           color="#4c4c4c" strokeWidth={2} />
    <Arrow from="right-model" to="right-tool"
           fromSide="right" toSide="left"
           color="#4c4c4c" strokeWidth={2} />
    
    {/* Tool to result */}
    <Arrow from="right-tool" to="right-result"
           fromSide="bottom" toSide="top"
           color="#4c4c4c" strokeWidth={2} />
    
    {/* Feedback loop from result back to history */}
    <Arrow from="right-result" to="right-history"
           fromSide="left" toSide="bottom"
           curve="step"
           color="#4c4c4c" strokeWidth={2} />
  </Stack>
);

console.log('🎨 Generating original Anthropic diagram replication...\n');

(async () => {
  try {
    const svg = await renderToSVG(<AnthropicOriginalDiagram />, {
      width: 1400,
      height: 900,
      backgroundColor: '#f7f7f5'
    });
    
    writeFileSync('examples/anthropic-original-replication.svg', svg);
    console.log('✅ Generated anthropic-original-replication.svg');
    console.log('\nThis diagram replicates the specific Anthropic diagram:');
    console.log('  "Prompt engineering vs. context engineering"');
    console.log('  • Two-panel comparison layout');
    console.log('  • Dashed context window boxes');
    console.log('  • Speech bubbles, document cards, tool pills');
    console.log('  • Neural network visualization (3x3 grid)');
    console.log('  • Complex arrow routing with labels');
    console.log('  • Feedback loop visualization');
  } catch (error) {
    console.error('Error generating diagram:', error);
    console.error('Stack:', error instanceof Error ? error.stack : String(error));
    process.exit(1);
  }
})();
