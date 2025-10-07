/**
 * Faithful Replication of Anthropic "Prompt engineering vs. context engineering" Diagram
 * 
 * Based on comprehensive structural analysis in anthropic_diagram_analysis.md
 * 
 * Key Features Demonstrated:
 * - Two-panel comparison with proper width ratio (1:1.5)
 * - Nested dashed boxes representing context windows
 * - Hierarchical spacing (labels closer to their boxes)
 * - Multiple element types with proper sizing
 * - Flexible spacers filling unused token space
 * - Complex arrow routing (Y-forks, shortening, feedback loops)
 * - Accurate color palette from original
 */

import React from 'react';
import { Stack, Row, Box, Text, Arrow, Spacer, renderToSVG } from '../src';
import { writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Color palette from analysis
const colors = {
  canvas: '#f7f7f5', // Off-white/beige background
  panel: '#ffffff', // Using white instead of #fafafa for better contrast
  dashedBorder: '#888888',
  prompt: '#6fb4ff', // Sky blue
  promptBorder: '#4a9ae5',
  doc: '#90EE90', // Light green
  tool: '#FFA500', // Orange
  memory: '#DDA0DD', // Light violet
  instruction: '#6fb4ff', // Cobalt blue
  toolCall: '#f8e0ba', // Tan
  toolCallBorder: '#d4b896',
  white: '#ffffff',
  whiteBorder: '#cccccc',
  
  // Text colors
  titleBlack: '#000000',
  headerGray: '#666666',
  labelGray: '#666666',
  contentGray: '#4c4c4c',
  arrowGray: '#4c4c4c',
  
  // Neural network colors
  neuralBlue: '#a3d5ff',
  neuralGreen: '#b8e6b8',
  neuralYellow: '#ffe6a3',
};

// Helper component for speech bubbles
const SpeechBubble = ({ id, text, width = 140, height = 40 }: any) => (
  <Box
    id={id}
    width={width}
    height={height}
    backgroundColor={colors.prompt}
    borderColor={colors.promptBorder}
    borderWidth={2}
    borderRadius={8}
    padding={8}
    justifyContent="center"
    alignItems="center"
  >
    <Text fontSize={12} fontWeight="bold" color={colors.white}>
      {text}
    </Text>
  </Box>
);

// Helper component for doc cards
const DocCard = ({ id, text, width = 120, height = 36 }: any) => (
  <Box
    id={id}
    width={width}
    height={height}
    backgroundColor={colors.doc}
    borderRadius={4}
    padding={6}
    justifyContent="center"
    alignItems="center"
  >
    <Text fontSize={11} fontWeight="bold">
      {text}
    </Text>
  </Box>
);

// Helper component for tool pills
const ToolPill = ({ id, text, width = 120, height = 36 }: any) => (
  <Box
    id={id}
    width={width}
    height={height}
    backgroundColor={colors.tool}
    borderRadius={18} // Rounded pill
    padding={6}
    justifyContent="center"
    alignItems="center"
  >
    <Text fontSize={11} fontWeight="bold" color={colors.white}>
      {text}
    </Text>
  </Box>
);

// Helper component for memory cards
const MemoryCard = ({ id, text, width = 120, height = 36 }: any) => (
  <Box
    id={id}
    width={width}
    height={height}
    backgroundColor={colors.memory}
    borderRadius={4}
    padding={6}
    justifyContent="center"
    alignItems="center"
  >
    <Text fontSize={11} fontWeight="bold">
      {text}
    </Text>
  </Box>
);

// Helper component for instruction cards
const InstructionCard = ({ id, text, width = 180, height = 36 }: any) => (
  <Box
    id={id}
    width={width}
    height={height}
    backgroundColor={colors.instruction}
    borderRadius={4}
    padding={6}
    justifyContent="center"
    alignItems="center"
  >
    <Text fontSize={11} fontWeight="bold" color={colors.white}>
      {text}
    </Text>
  </Box>
);

// Helper component for message history
const MessageHistory = ({ id, width = 140, height = 38 }: any) => (
  <Box
    id={id}
    width={width}
    height={height}
    backgroundColor={colors.white}
    borderColor={colors.whiteBorder}
    borderWidth={1}
    borderRadius={6}
    padding={6}
    justifyContent="center"
    alignItems="center"
  >
    <Text fontSize={11}>Message history</Text>
  </Box>
);

// Helper component for neural network grid
const NeuralGrid = ({ id }: { id: string }) => (
  <Box id={id} width={44} height={44} padding={2}>
    <Stack gap={6}>
      <Row gap={6} justifyContent="center">
        <Box width={10} height={10} backgroundColor={colors.neuralBlue} borderRadius={5} />
        <Box width={10} height={10} backgroundColor={colors.neuralGreen} borderRadius={5} />
        <Box width={10} height={10} backgroundColor={colors.neuralYellow} borderRadius={5} />
      </Row>
      <Row gap={6} justifyContent="center">
        <Box width={10} height={10} backgroundColor={colors.neuralGreen} borderRadius={5} />
        <Box width={10} height={10} backgroundColor={colors.neuralYellow} borderRadius={5} />
        <Box width={10} height={10} backgroundColor={colors.neuralBlue} borderRadius={5} />
      </Row>
      <Row gap={6} justifyContent="center">
        <Box width={10} height={10} backgroundColor={colors.neuralYellow} borderRadius={5} />
        <Box width={10} height={10} backgroundColor={colors.neuralBlue} borderRadius={5} />
        <Box width={10} height={10} backgroundColor={colors.neuralGreen} borderRadius={5} />
      </Row>
    </Stack>
  </Box>
);

// Helper component for output boxes
const OutputBox = ({ id, text, width = 110, bg = colors.white, borderColor = '#444444' }: any) => (
  <Box
    id={id}
    width={width}
    height={38}
    backgroundColor={bg}
    borderColor={borderColor}
    borderWidth={1}
    borderRadius={4}
    padding={6}
    justifyContent="center"
    alignItems="center"
  >
    <Text fontSize={11}>{text}</Text>
  </Box>
);

// Helper component for scissors icon (simplified as a small box with "✂" symbol)
const ScissorsIcon = () => (
  <Box width={30} height={30} justifyContent="center" alignItems="center">
    <Text fontSize={24}>✂️</Text>
  </Box>
);

const AnthropicFaithfulReplication = () => (
  <Stack
    width={1400}
    height={800}
    backgroundColor={colors.canvas}
    padding={24}
    gap={24}
    alignItems="center"
  >
    {/* Title */}
    <Text fontSize={32} fontWeight="bold" color={colors.titleBlack} textAlign="center">
      Prompt engineering vs. context engineering
    </Text>

    {/* Main content row with proper spacing for model grids */}
    <Row gap={40} alignItems="stretch" height={700}>
      {/* LEFT PANEL */}
      <Stack
        width={470}
        backgroundColor={colors.panel}
        borderRadius={12}
        padding={20}
        gap={0} // No gap, we'll use individual margins for hierarchical spacing
      >
        {/* Header */}
        <Text fontSize={18} fontWeight="bold" color={colors.headerGray} marginBottom={16}>
          Prompt engineering for single turn queries
        </Text>
        
        {/* Context window label - CLOSER to its box (6px) than to header (16px above) */}
        <Text fontSize={14} color={colors.labelGray} marginBottom={6}>
          Context window
        </Text>
        
        {/* Dashed context window box */}
        <Box
          id="left-context"
          flexGrow={1} // Fill remaining height
          borderColor={colors.dashedBorder}
          borderWidth={2}
          borderStyle="dashed"
          borderDashArray="6 4"
          borderRadius={8}
          padding={20}
        >
          <Stack gap={8}>
            <SpeechBubble id="left-system-prompt" text="System prompt" width={140} height={40} />
            <SpeechBubble id="left-user-message" text="User message" width={140} height={40} />
            
            {/* Flexible spacer representing unused token space */}
            <Spacer flexible minHeight={100} />
            
            {/* Scissors icon at bottom-left */}
            <Box width={140} alignItems="flex-start">
              <ScissorsIcon />
            </Box>
          </Stack>
        </Box>
      </Stack>

      {/* LEFT MODEL & OUTPUT (Outside panel) */}
      <Stack gap={12} justifyContent="center" width={140}>
        <NeuralGrid id="left-model" />
        <OutputBox id="left-assistant" text="Assistant message" width={120} />
      </Stack>

      {/* RIGHT PANEL */}
      <Stack
        width={780}
        backgroundColor={colors.panel}
        borderRadius={12}
        padding={20}
        gap={0}
      >
        {/* Header */}
        <Text fontSize={18} fontWeight="bold" color={colors.headerGray} marginBottom={16}>
          Context engineering for agents
        </Text>
        
        {/* Two dashed boxes side by side */}
        <Row gap={12} flexGrow={1}>
          {/* Possible context box */}
          <Stack gap={0} flexGrow={1}>
            <Text fontSize={14} color={colors.labelGray} marginBottom={6}>
              Possible context to give model
            </Text>
            
            <Box
              id="possible-context"
              flexGrow={1}
              borderColor={colors.dashedBorder}
              borderWidth={2}
              borderStyle="dashed"
              borderDashArray="6 4"
              borderRadius={8}
              padding={12}
            >
              <Stack gap={8}>
                <DocCard id="doc-1" text="Doc 1" />
                <DocCard id="doc-2" text="Doc 2" />
                <DocCard id="doc-3" text="Doc 3" />
                <ToolPill id="tool-1" text="Tool 1" />
                <ToolPill id="tool-2" text="Tool 2" />
                <ToolPill id="tool-3" text="Tool 3" />
                <ToolPill id="tool-4" text="Tool 4" />
                <MemoryCard id="memory-1" text="Memory file" />
                <MemoryCard id="memory-2" text="Memory file" />
                <InstructionCard id="instructions" text="Comprehensive instructions" />
                <InstructionCard id="knowledge" text="Domain knowledge" />
                <MessageHistory id="possible-history" />
              </Stack>
            </Box>
          </Stack>

          {/* Context window box */}
          <Stack gap={0} flexGrow={1}>
            <Text fontSize={14} color={colors.labelGray} marginBottom={6}>
              Context window
            </Text>
            
            <Box
              id="right-context"
              flexGrow={1}
              borderColor={colors.dashedBorder}
              borderWidth={2}
              borderStyle="dashed"
              borderDashArray="6 4"
              borderRadius={8}
              padding={12}
            >
              <Stack gap={8}>
                <SpeechBubble id="right-system-prompt" text="System prompt" width={140} height={36} />
                
                {/* Two docs side by side */}
                <Row gap={6}>
                  <DocCard id="curated-doc-1" text="Doc 1" width={67} />
                  <DocCard id="curated-doc-2" text="Doc 2" width={67} />
                </Row>
                
                <MemoryCard id="curated-memory" text="Memory file" width={140} />
                
                {/* Two tools side by side */}
                <Row gap={6}>
                  <ToolPill id="curated-tool-1" text="Tool 1" width={67} />
                  <ToolPill id="curated-tool-2" text="Tool 2" width={67} />
                </Row>
                
                <SpeechBubble id="right-user-message" text="User message" width={140} height={36} />
                <MessageHistory id="right-history" width={140} />
                
                {/* Flexible spacer */}
                <Spacer flexible minHeight={50} />
                
                {/* Scissors icon */}
                <Box width={140} alignItems="flex-start">
                  <ScissorsIcon />
                </Box>
              </Stack>
            </Box>
          </Stack>
        </Row>
      </Stack>

      {/* RIGHT MODEL & OUTPUTS (Outside panel) */}
      <Stack gap={12} justifyContent="center" width={140}>
        <NeuralGrid id="right-model" />
        <OutputBox id="right-assistant" text="Assistant message" width={120} />
        <OutputBox 
          id="tool-call" 
          text="Tool call" 
          width={120} 
          bg={colors.toolCall} 
          borderColor={colors.toolCallBorder} 
        />
        <OutputBox 
          id="tool-result" 
          text="Tool result" 
          width={120} 
          bg={colors.toolCall} 
          borderColor={colors.toolCallBorder} 
        />
      </Stack>
    </Row>

    {/* Arrows */}
    {/* Left side: context -> model -> output */}
    <Arrow from="left-context" to="left-model" fromSide="right" toSide="left" color={colors.arrowGray} strokeWidth={2} />
    <Arrow from="left-model" to="left-assistant" fromSide="right" toSide="left" color={colors.arrowGray} strokeWidth={2} />
    
    {/* Right side: Curation arrow (stops short) */}
    <Arrow 
      from="possible-context" 
      to="right-context" 
      fromSide="right" 
      toSide="left"
      color={colors.arrowGray} 
      strokeWidth={2}
      label="Curation"
      shortenEnd={10}
    />
    
    {/* Right context window -> model (would ideally have bracket merging) */}
    <Arrow from="right-context" to="right-model" fromSide="right" toSide="left" color={colors.arrowGray} strokeWidth={2} />
    
    {/* Y-shaped fork: model -> assistant + tool call (currently showing as two separate arrows) */}
    <Arrow from="right-model" to="right-assistant" fromSide="right" toSide="left" color={colors.arrowGray} strokeWidth={2} />
    <Arrow from="right-model" to="tool-call" fromSide="right" toSide="left" color={colors.arrowGray} strokeWidth={2} />
    
    {/* Tool call -> tool result */}
    <Arrow from="tool-call" to="tool-result" fromSide="bottom" toSide="top" color={colors.arrowGray} strokeWidth={2} />
    
    {/* Feedback loop: tool result -> message history (orthogonal routing) */}
    <Arrow 
      from="tool-result" 
      to="right-history" 
      fromSide="left" 
      toSide="bottom"
      color={colors.arrowGray} 
      strokeWidth={2}
      curve="step"
    />
  </Stack>
);

async function generate() {
  const svg = await renderToSVG(<AnthropicFaithfulReplication />, {
    width: 1400,
    height: 800,
    backgroundColor: colors.canvas,
  });
  
  const outputPath = join(__dirname, '../examples/anthropic-faithful-replication.svg');
  writeFileSync(outputPath, svg);
  console.log(`Generated: ${outputPath}`);
}

generate().catch(console.error);
