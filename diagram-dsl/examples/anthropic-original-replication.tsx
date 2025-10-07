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
  <Stack width={1400} height={850} padding={24} gap={24} backgroundColor="#f7f7f5">
    {/* Title */}
    <Text fontSize={32} fontWeight="bold" textAlign="center">
      Prompt engineering vs. context engineering
    </Text>

    {/* Two main panels side by side */}
    <Row gap={40} alignItems="flex-start" justifyContent="center">
      {/* Left Panel - Single Turn */}
      <Stack 
        width={580}
        backgroundColor="#fafafa"
        borderRadius={12}
        padding={20}
        gap={12}
      >
        {/* Header */}
        <Text fontSize={18} fontWeight="bold" color="#666">
          Prompt engineering for single turn queries
        </Text>

        {/* Context Window */}
        <Stack gap={6}>
          <Text fontSize={14} color="#666">
            Context window
          </Text>
          
          {/* Dashed box representing context window */}
          <Box
            id="left-context"
            width={540}
            height={400}
            backgroundColor="transparent"
            borderColor="#888888"
            borderWidth={2}
            borderRadius={8}
            padding={16}
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
              <Box height={240} width={100} />
              
              {/* Scissors icon */}
              <Text fontSize={20}>✂️</Text>
            </Stack>
          </Box>
        </Stack>

        {/* Model representation (3x3 grid) below */}
        <Row gap={12} justifyContent="center" marginTop={20}>
          <Box
            id="left-model"
            width={60}
            height={60}
            backgroundColor="transparent"
          >
            <Stack gap={4}>
              <Row gap={4}>
                <Box width={14} height={14} backgroundColor="#a3d5ff" borderRadius={7} />
                <Box width={14} height={14} backgroundColor="#b8e6b8" borderRadius={7} />
                <Box width={14} height={14} backgroundColor="#ffe6a3" borderRadius={7} />
              </Row>
              <Row gap={4}>
                <Box width={14} height={14} backgroundColor="#b8e6b8" borderRadius={7} />
                <Box width={14} height={14} backgroundColor="#a3d5ff" borderRadius={7} />
                <Box width={14} height={14} backgroundColor="#ffe6a3" borderRadius={7} />
              </Row>
              <Row gap={4}>
                <Box width={14} height={14} backgroundColor="#ffe6a3" borderRadius={7} />
                <Box width={14} height={14} backgroundColor="#b8e6b8" borderRadius={7} />
                <Box width={14} height={14} backgroundColor="#a3d5ff" borderRadius={7} />
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
        </Row>
      </Stack>

      {/* Right Panel - Multi Turn with Agent */}
      <Stack 
        width={600}
        backgroundColor="#fafafa"
        borderRadius={12}
        padding={20}
        gap={12}
      >
        {/* Header */}
        <Text fontSize={18} fontWeight="bold" color="#666">
          Context engineering for agents
        </Text>

        {/* Two dashed boxes side by side */}
        <Row gap={12} alignItems="flex-start">
          {/* Left: Possible context */}
          <Stack gap={6}>
            <Text fontSize={14} color="#666">
              Possible context to give model
            </Text>
            
            <Box
              id="right-possible"
              width={260}
              height={400}
              backgroundColor="transparent"
              borderColor="#888888"
              borderWidth={2}
              borderRadius={8}
              padding={10}
            >
              <Stack gap={6} alignItems="flex-start">
                {/* Green doc cards */}
                <Box width={110} height={32} backgroundColor="#90EE90" borderRadius={4} padding={4}
                     justifyContent="center">
                  <Text fontSize={10}>Doc 1</Text>
                </Box>
                <Box width={110} height={32} backgroundColor="#90EE90" borderRadius={4} padding={4}
                     justifyContent="center">
                  <Text fontSize={10}>Doc 2</Text>
                </Box>

                {/* Orange tool pills */}
                <Box width={110} height={32} backgroundColor="#FFA500" borderRadius={16} padding={4}
                     justifyContent="center">
                  <Text fontSize={10}>Tool 1</Text>
                </Box>
                <Box width={110} height={32} backgroundColor="#FFA500" borderRadius={16} padding={4}
                     justifyContent="center">
                  <Text fontSize={10}>Tool 2</Text>
                </Box>

                {/* Violet memory files */}
                <Box width={110} height={32} backgroundColor="#DDA0DD" borderRadius={4} padding={4}
                     justifyContent="center">
                  <Text fontSize={10}>Memory file</Text>
                </Box>

                {/* Wide blue cards */}
                <Box width={160} height={32} backgroundColor="#6fb4ff" borderRadius={4} padding={4}
                     justifyContent="center">
                  <Text fontSize={10} fontWeight="bold">Instructions</Text>
                </Box>

                {/* White message history */}
                <Box id="right-history" width={130} height={34} backgroundColor="white" 
                     borderColor="#888" borderWidth={1} borderRadius={4} padding={4}
                     justifyContent="center">
                  <Text fontSize={10}>Message history</Text>
                </Box>
              </Stack>
            </Box>
          </Stack>

          {/* Right: Curated context window */}
          <Stack gap={6}>
            <Text fontSize={14} color="#666">
              Context window
            </Text>
            
            <Box
              id="right-context"
              width={260}
              height={400}
              backgroundColor="transparent"
              borderColor="#888888"
              borderWidth={2}
              borderRadius={8}
              padding={10}
            >
              <Stack gap={6} alignItems="flex-start">
                {/* System prompt */}
                <Box width={130} height={32} backgroundColor="#6fb4ff" borderRadius={4} padding={4}
                     justifyContent="center">
                  <Text fontSize={10} fontWeight="bold">System prompt</Text>
                </Box>

                {/* Doc cards side by side */}
                <Row gap={4}>
                  <Box width={63} height={32} backgroundColor="#90EE90" borderRadius={4} padding={3}
                       justifyContent="center">
                    <Text fontSize={9}>Doc 1</Text>
                  </Box>
                  <Box width={63} height={32} backgroundColor="#90EE90" borderRadius={4} padding={3}
                       justifyContent="center">
                    <Text fontSize={9}>Doc 2</Text>
                  </Box>
                </Row>

                {/* Memory file */}
                <Box width={130} height={32} backgroundColor="#DDA0DD" borderRadius={4} padding={4}
                     justifyContent="center">
                  <Text fontSize={10}>Memory file</Text>
                </Box>

                {/* Tool pills side by side */}
                <Row gap={4}>
                  <Box width={63} height={32} backgroundColor="#FFA500" borderRadius={16} padding={3}
                       justifyContent="center">
                    <Text fontSize={9}>Tool 1</Text>
                  </Box>
                  <Box width={63} height={32} backgroundColor="#FFA500" borderRadius={16} padding={3}
                       justifyContent="center">
                    <Text fontSize={9}>Tool 2</Text>
                  </Box>
                </Row>

                {/* User message */}
                <Box width={130} height={32} backgroundColor="#6fb4ff" borderRadius={4} padding={4}
                     justifyContent="center">
                  <Text fontSize={10} fontWeight="bold">User message</Text>
                </Box>

                {/* Message history */}
                <Box width={130} height={34} backgroundColor="white" borderColor="#888" 
                     borderWidth={1} borderRadius={4} padding={4}
                     justifyContent="center">
                  <Text fontSize={10}>Message history</Text>
                </Box>

                {/* Empty space and scissors */}
                <Box height={60} />
                <Text fontSize={18}>✂️</Text>
              </Stack>
            </Box>
          </Stack>
        </Row>

        {/* Model and outputs below */}
        <Row gap={12} justifyContent="center" marginTop={12}>
          <Box
            id="right-model"
            width={60}
            height={60}
            backgroundColor="transparent"
          >
            <Stack gap={4}>
              <Row gap={4}>
                <Box width={14} height={14} backgroundColor="#a3d5ff" borderRadius={7} />
                <Box width={14} height={14} backgroundColor="#b8e6b8" borderRadius={7} />
                <Box width={14} height={14} backgroundColor="#ffe6a3" borderRadius={7} />
              </Row>
              <Row gap={4}>
                <Box width={14} height={14} backgroundColor="#b8e6b8" borderRadius={7} />
                <Box width={14} height={14} backgroundColor="#a3d5ff" borderRadius={7} />
                <Box width={14} height={14} backgroundColor="#ffe6a3" borderRadius={7} />
              </Row>
              <Row gap={4}>
                <Box width={14} height={14} backgroundColor="#ffe6a3" borderRadius={7} />
                <Box width={14} height={14} backgroundColor="#b8e6b8" borderRadius={7} />
                <Box width={14} height={14} backgroundColor="#a3d5ff" borderRadius={7} />
              </Row>
            </Stack>
          </Box>

          <Stack gap={10}>
            {/* Assistant message output */}
            <Box
              id="right-output"
              width={110}
              height={36}
              backgroundColor="white"
              borderColor="#444444"
              borderWidth={1}
              borderRadius={4}
              padding={5}
              justifyContent="center"
              alignItems="center"
            >
              <Text fontSize={11}>Assistant message</Text>
            </Box>

            {/* Tool call */}
            <Box
              id="right-tool"
              width={110}
              height={36}
              backgroundColor="#f8e0ba"
              borderColor="#d4b896"
              borderWidth={1}
              borderRadius={4}
              padding={5}
              justifyContent="center"
              alignItems="center"
            >
              <Text fontSize={11}>Tool call</Text>
            </Box>
          </Stack>
        </Row>

        {/* Tool result at bottom (for feedback loop) */}
        <Box
          id="right-result"
          width={110}
          height={36}
          backgroundColor="#f8e0ba"
          borderColor="#d4b896"
          borderWidth={1}
          borderRadius={4}
          padding={5}
          marginLeft={240}
          justifyContent="center"
          alignItems="center"
        >
          <Text fontSize={11}>Tool result</Text>
        </Box>
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
