/**
 * Improved Anthropic-Style Diagram with Better Grouping
 * 
 * This example shows how to use Cluster components for clear visual hierarchy
 * and demonstrates best practices for creating professional AI system diagrams.
 */

import React from 'react';
import { writeFileSync } from 'fs';
import {
  Stack, Row, Card, Title, Subtitle, Label, Arrow, Cluster,
  Badge, Divider,
  renderToSVG
} from '../src';

const ImprovedAnthropicDiagram = () => (
  <Stack width={1400} height={1100} padding={40} gap={35}>
    {/* Title */}
    <Stack gap={10} alignItems="center">
      <Title level={1}>AI Agent Conversation Flow</Title>
      <Subtitle size="base">Request processing with safety layers and context management</Subtitle>
    </Stack>

    {/* Main Flow */}
    <Row gap={40} justifyContent="center" alignItems="flex-start">
      {/* Left Column: Input Processing */}
      <Cluster title="Input Processing" variant="primary" width={350} padding={20}>
        <Stack gap={20} alignItems="center">
          <Card id="user-input" variant="primary" width={300} height={75}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">User Message</Label>
              <Subtitle>Text input</Subtitle>
            </Stack>
          </Card>

          <Card id="safety-check" variant="warning" width={300} height={75}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">Safety Check</Label>
              <Subtitle>Content moderation</Subtitle>
            </Stack>
          </Card>

          <Card id="context-retrieval" variant="success" width={300} height={75}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">Context Retrieval</Label>
              <Subtitle>RAG / Memory</Subtitle>
            </Stack>
          </Card>
        </Stack>
      </Cluster>

      {/* Middle Column: Model Processing */}
      <Cluster title="Model Processing" variant="accent" width={380} padding={20}>
        <Stack gap={20} alignItems="center">
          <Card id="prompt-assembly" variant="accent" width={330} height={85}>
            <Stack gap={8} alignItems="center">
              <Label bold size="lg">Prompt Assembly</Label>
              <Subtitle>System + Context + User</Subtitle>
              <Label size="sm">Token budget: 200K</Label>
            </Stack>
          </Card>

          <Card id="claude-inference" variant="accent" width={330} height={110}>
            <Stack gap={10} alignItems="center">
              <Label bold size="lg">Claude Model</Label>
              <Subtitle>Neural inference</Subtitle>
              <Row gap={8} marginTop={4}>
                <Badge text="Constitutional AI" variant="success" />
                <Badge text="RLHF" variant="info" />
              </Row>
            </Stack>
          </Card>

          <Card id="output-processing" variant="accent" width={330} height={75}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">Output Processing</Label>
              <Subtitle>Format & validate</Subtitle>
            </Stack>
          </Card>
        </Stack>
      </Cluster>

      {/* Right Column: Response & Feedback */}
      <Cluster title="Response & Feedback" variant="success" width={350} padding={20}>
        <Stack gap={20} alignItems="center">
          <Card id="response-delivery" variant="success" width={300} height={75}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">Response Delivery</Label>
              <Subtitle>Streaming / Complete</Subtitle>
            </Stack>
          </Card>

          <Card id="feedback-collection" variant="secondary" width={300} height={75}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">Feedback Collection</Label>
              <Subtitle>Thumbs up/down</Subtitle>
            </Stack>
          </Card>

          <Card id="model-improvement" variant="secondary" width={300} height={75}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">Model Improvement</Label>
              <Subtitle>Training data</Subtitle>
            </Stack>
          </Card>
        </Stack>
      </Cluster>
    </Row>

    {/* Data Store Layer */}
    <Divider width={1320} />

    <Cluster title="Data & Storage" variant="secondary" width={1320} padding={20}>
      <Row gap={30} justifyContent="center">
        <Card id="vector-db" variant="secondary" width={250} height={80}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Vector Database</Label>
            <Subtitle>Embeddings storage</Subtitle>
            <Label size="sm">Semantic search</Label>
          </Stack>
        </Card>
        <Card id="conversation-store" variant="secondary" width={250} height={80}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Conversation Store</Label>
            <Subtitle>Chat history</Subtitle>
            <Label size="sm">Session management</Label>
          </Stack>
        </Card>
        <Card id="analytics" variant="secondary" width={250} height={80}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Analytics DB</Label>
            <Subtitle>Usage metrics</Subtitle>
            <Label size="sm">Performance tracking</Label>
          </Stack>
        </Card>
        <Card id="training-data" variant="secondary" width={250} height={80}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Training Data</Label>
            <Subtitle>RLHF samples</Subtitle>
            <Label size="sm">Model fine-tuning</Label>
          </Stack>
        </Card>
      </Row>
    </Cluster>

    {/* Arrows: Forward flow */}
    <Arrow from="user-input" to="safety-check" label="text" color="#1976d2" thickness="medium" />
    <Arrow from="safety-check" to="context-retrieval" label="validated" color="#66bb6a" thickness="medium" />
    <Arrow from="context-retrieval" to="prompt-assembly" label="with context" color="#1976d2" thickness="medium" />
    <Arrow from="prompt-assembly" to="claude-inference" label="prompt" color="#ff9800" thickness="thick" />
    <Arrow from="claude-inference" to="output-processing" label="response" color="#ab47bc" thickness="thick" />
    <Arrow from="output-processing" to="response-delivery" label="formatted" color="#66bb6a" thickness="medium" />
    <Arrow from="response-delivery" to="feedback-collection" label="delivered" color="#4caf50" thickness="medium" />
    <Arrow from="feedback-collection" to="model-improvement" label="feedback" color="#9e9e9e" thickness="medium" />

    {/* Arrows: Data interactions */}
    <Arrow from="context-retrieval" to="vector-db" color="#2196f3" style="dashed" bidirectional={true} />
    <Arrow from="context-retrieval" to="conversation-store" color="#2196f3" style="dashed" bidirectional={true} />
    <Arrow from="response-delivery" to="conversation-store" label="save" color="#9e9e9e" style="dashed" />
    <Arrow from="response-delivery" to="analytics" label="log" color="#9e9e9e" style="dashed" />
    <Arrow from="model-improvement" to="training-data" label="store" color="#7b1fa2" style="dashed" />

    {/* Legend */}
    <Row gap={25} marginTop={20} justifyContent="center">
      <Badge text="→ Primary data flow" variant="primary" />
      <Badge text="⟷ Bidirectional access" variant="info" />
      <Badge text="⋯ Async operations" variant="secondary" />
      <Badge text="200K token context" variant="success" />
    </Row>
  </Stack>
);

console.log('🎨 Generating improved Anthropic-style diagram...\n');

(async () => {
  try {
    const svg = await renderToSVG(<ImprovedAnthropicDiagram />, {
      width: 1400,
      height: 1100,
      backgroundColor: 'white'
    });
    
    writeFileSync('examples/anthropic-improved.svg', svg);
    console.log('✅ Generated anthropic-improved.svg');
    console.log('\nKey improvements:');
    console.log('  • Cluster components for clear visual grouping');
    console.log('  • Three-column layout for logical flow separation');
    console.log('  • Color-coded clusters by function (Input, Processing, Response)');
    console.log('  • Separate data layer with shared storage components');
    console.log('  • Bidirectional arrows for data access patterns');
    console.log('  • Cleaner, more scannable layout');
    console.log('\n✨ Much easier to create and maintain!');
  } catch (error) {
    console.error('Error generating diagram:', error);
    console.error('Stack:', error instanceof Error ? error.stack : String(error));
    process.exit(1);
  }
})();
