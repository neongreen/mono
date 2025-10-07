/**
 * Simple Anthropic-Style Diagram - Perfect Starting Point
 * 
 * This is a minimal but complete example showing the essential
 * components and patterns for creating Anthropic-style diagrams.
 * Use this as a template for your own diagrams.
 */

import React from 'react';
import { writeFileSync } from 'fs';
import {
  Stack, Row, Card, Title, Subtitle, Label, Arrow, Cluster,
  Badge,
  renderToSVG
} from '../src';

const SimpleAnthropicDiagram = () => (
  <Stack width={1000} height={700} padding={40} gap={30}>
    {/* Title Section */}
    <Stack gap={8} alignItems="center">
      <Title level={1}>Simple AI Conversation</Title>
      <Subtitle>User → Claude → Response</Subtitle>
    </Stack>

    {/* Main Flow: Three Columns */}
    <Row gap={35} justifyContent="center">
      {/* Column 1: Input */}
      <Cluster title="Input" variant="primary" width={280} padding={20}>
        <Stack gap={16} alignItems="center">
          <Card id="user" variant="primary" width={240} height={70}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">User</Label>
              <Subtitle>Sends message</Subtitle>
            </Stack>
          </Card>
          
          <Card id="safety" variant="warning" width={240} height={70}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">Safety Filter</Label>
              <Subtitle>Check content</Subtitle>
            </Stack>
          </Card>
        </Stack>
      </Cluster>

      {/* Column 2: Processing */}
      <Cluster title="Processing" variant="accent" width={280} padding={20}>
        <Stack gap={16} alignItems="center">
          <Card id="claude" variant="accent" width={240} height={100}>
            <Stack gap={8} alignItems="center">
              <Label bold size="lg">Claude</Label>
              <Subtitle>AI model inference</Subtitle>
              <Badge text="200K context" variant="success" />
            </Stack>
          </Card>
        </Stack>
      </Cluster>

      {/* Column 3: Output */}
      <Cluster title="Output" variant="success" width={280} padding={20}>
        <Stack gap={16} alignItems="center">
          <Card id="format" variant="success" width={240} height={70}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">Format</Label>
              <Subtitle>Structure response</Subtitle>
            </Stack>
          </Card>
          
          <Card id="response" variant="success" width={240} height={70}>
            <Stack gap={6} alignItems="center">
              <Label bold size="lg">Response</Label>
              <Subtitle>Return to user</Subtitle>
            </Stack>
          </Card>
        </Stack>
      </Cluster>
    </Row>

    {/* Data Storage */}
    <Cluster title="Storage" variant="secondary" width={920} padding={20}>
      <Row gap={25} justifyContent="center">
        <Card id="history" variant="secondary" width={200} height={70}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Chat History</Label>
            <Subtitle>Previous messages</Subtitle>
          </Stack>
        </Card>
        
        <Card id="logs" variant="secondary" width={200} height={70}>
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Logs</Label>
            <Subtitle>Analytics data</Subtitle>
          </Stack>
        </Card>
      </Row>
    </Cluster>

    {/* Arrows: Forward Flow */}
    <Arrow from="user" to="safety" label="message" color="#1976d2" thickness="medium" />
    <Arrow from="safety" to="claude" label="validated" color="#66bb6a" thickness="medium" />
    <Arrow from="claude" to="format" label="raw output" color="#ab47bc" thickness="thick" />
    <Arrow from="format" to="response" label="formatted" color="#4caf50" thickness="medium" />

    {/* Arrows: Data Access */}
    <Arrow from="claude" to="history" color="#2196f3" style="dashed" bidirectional={true} />
    <Arrow from="response" to="logs" label="log" color="#9e9e9e" style="dashed" />

    {/* Legend */}
    <Row gap={20} marginTop={15} justifyContent="center">
      <Badge text="→ Main flow" variant="primary" />
      <Badge text="⋯ Data access" variant="secondary" />
    </Row>
  </Stack>
);

console.log('🎨 Generating simple Anthropic-style diagram...\n');

(async () => {
  try {
    const svg = await renderToSVG(<SimpleAnthropicDiagram />, {
      width: 1000,
      height: 700,
      backgroundColor: 'white'
    });
    
    writeFileSync('examples/anthropic-simple.svg', svg);
    console.log('✅ Generated anthropic-simple.svg');
    console.log('\nPerfect starting point demonstrating:');
    console.log('  • Three-column layout (Input → Processing → Output)');
    console.log('  • Color-coded Cluster components for visual grouping');
    console.log('  • Professional Card components with clear hierarchy');
    console.log('  • Solid arrows for main flow, dashed for data access');
    console.log('  • Shared data layer at the bottom');
    console.log('  • Legend for arrow types');
    console.log('\n📋 Copy this as a template for your own diagrams!');
  } catch (error) {
    console.error('Error generating diagram:', error);
    console.error('Stack:', error instanceof Error ? error.stack : String(error));
    process.exit(1);
  }
})();
