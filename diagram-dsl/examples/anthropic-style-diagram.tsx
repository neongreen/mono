/**
 * Anthropic-Style AI System Diagram
 * 
 * This example replicates typical Anthropic documentation diagrams
 * showing an AI system architecture with clear data flow,
 * component grouping, and professional styling.
 */

import React from 'react';
import { writeFileSync } from 'fs';
import {
  Stack, Row, Card, Title, Subtitle, Label, Arrow, Group,
  Panel, Badge, Divider,
  renderToSVG
} from '../src';

const AnthropicStyleDiagram = () => (
  <Stack width={1600} height={1200} padding={50} gap={40} backgroundColor="#fafafa">
    {/* Title */}
    <Stack gap={12} alignItems="center">
      <Title level={1}>Claude API System Architecture</Title>
      <Subtitle size="base">End-to-end request flow with safety layers</Subtitle>
    </Stack>

    {/* User Layer */}
    <Stack gap={20}>
      <Row gap={12} alignItems="center">
        <Badge text="User Layer" variant="info" />
        <Subtitle size="sm">Application integration points</Subtitle>
      </Row>
      <Row gap={30} justifyContent="center">
        <Card id="web-app" variant="primary" width={220} height={100}>
          <Stack gap={10} alignItems="center" padding={10}>
            <Label bold size="lg">Web Application</Label>
            <Subtitle>React / Vue / Angular</Subtitle>
            <Label size="sm">Direct API calls</Label>
          </Stack>
        </Card>
        <Card id="mobile-app" variant="primary" width={220} height={100}>
          <Stack gap={10} alignItems="center" padding={10}>
            <Label bold size="lg">Mobile App</Label>
            <Subtitle>iOS / Android</Subtitle>
            <Label size="sm">SDK integration</Label>
          </Stack>
        </Card>
        <Card id="backend" variant="primary" width={220} height={100}>
          <Stack gap={10} alignItems="center" padding={10}>
            <Label bold size="lg">Backend Service</Label>
            <Subtitle>Python / Node.js</Subtitle>
            <Label size="sm">Server-to-server</Label>
          </Stack>
        </Card>
      </Row>
    </Stack>

    <Divider width={1500} />

    {/* API Gateway Layer */}
    <Stack gap={20}>
      <Row gap={12} alignItems="center">
        <Badge text="API Gateway" variant="success" />
        <Subtitle size="sm">Request routing and authentication</Subtitle>
      </Row>
      <Row gap={40} justifyContent="center">
        <Card id="auth" variant="success" width={200} height={90}>
          <Stack gap={8} alignItems="center" padding={8}>
            <Label bold size="lg">Authentication</Label>
            <Subtitle>API Key validation</Subtitle>
            <Label size="sm">Rate limiting</Label>
          </Stack>
        </Card>
        <Card id="load-balancer" variant="success" width={200} height={90}>
          <Stack gap={8} alignItems="center" padding={8}>
            <Label bold size="lg">Load Balancer</Label>
            <Subtitle>Request distribution</Subtitle>
            <Label size="sm">Health checks</Label>
          </Stack>
        </Card>
        <Card id="queue" variant="success" width={200} height={90}>
          <Stack gap={8} alignItems="center" padding={8}>
            <Label bold size="lg">Request Queue</Label>
            <Subtitle>Async processing</Subtitle>
            <Label size="sm">Priority handling</Label>
          </Stack>
        </Card>
      </Row>
    </Stack>

    <Divider width={1500} />

    {/* Processing Layer */}
    <Stack gap={20}>
      <Row gap={12} alignItems="center">
        <Badge text="Processing Layer" variant="warning" />
        <Subtitle size="sm">Safety checks and model inference</Subtitle>
      </Row>
      
      <Row gap={50} justifyContent="center" alignItems="flex-start">
        {/* Safety Pipeline */}
        <Stack gap={16} width={350}>
          <Label bold>Safety Pipeline</Label>
          <Card id="input-safety" variant="warning" width={320} height={85}>
            <Stack gap={6} alignItems="center" padding={8}>
              <Label bold size="lg">Input Safety</Label>
              <Subtitle>Content filtering</Subtitle>
              <Label size="sm">Harmful content detection</Label>
            </Stack>
          </Card>
          <Card id="prompt-optimization" variant="warning" width={320} height={85}>
            <Stack gap={6} alignItems="center" padding={8}>
              <Label bold size="lg">Prompt Optimization</Label>
              <Subtitle>Context preparation</Subtitle>
              <Label size="sm">Token management</Label>
            </Stack>
          </Card>
        </Stack>

        {/* Model Inference */}
        <Stack gap={16} width={350}>
          <Label bold>Model Inference</Label>
          <Card id="claude-model" variant="accent" width={320} height={140}>
            <Stack gap={10} alignItems="center" padding={12}>
              <Label bold size="lg">Claude Model</Label>
              <Subtitle>Neural network inference</Subtitle>
              <Row gap={8} marginTop={6}>
                <Badge text="100B+ params" variant="info" />
                <Badge text="200K context" variant="success" />
              </Row>
              <Label size="sm">Constitutional AI safety</Label>
            </Stack>
          </Card>
        </Stack>

        {/* Output Processing */}
        <Stack gap={16} width={350}>
          <Label bold>Output Processing</Label>
          <Card id="output-safety" variant="warning" width={320} height={85}>
            <Stack gap={6} alignItems="center" padding={8}>
              <Label bold size="lg">Output Safety</Label>
              <Subtitle>Response filtering</Subtitle>
              <Label size="sm">Quality assurance</Label>
            </Stack>
          </Card>
          <Card id="formatting" variant="warning" width={320} height={85}>
            <Stack gap={6} alignItems="center" padding={8}>
              <Label bold size="lg">Response Formatting</Label>
              <Subtitle>Structure & metadata</Subtitle>
              <Label size="sm">Token counting</Label>
            </Stack>
          </Card>
        </Stack>
      </Row>
    </Stack>

    <Divider width={1500} />

    {/* Data & Monitoring Layer */}
    <Stack gap={20}>
      <Row gap={12} alignItems="center">
        <Badge text="Data & Monitoring" variant="secondary" />
        <Subtitle size="sm">Logging, analytics, and feedback</Subtitle>
      </Row>
      <Row gap={30} justifyContent="center">
        <Card id="analytics" variant="secondary" width={220} height={85}>
          <Stack gap={8} alignItems="center" padding={8}>
            <Label bold size="lg">Analytics</Label>
            <Subtitle>Usage metrics</Subtitle>
            <Label size="sm">Performance tracking</Label>
          </Stack>
        </Card>
        <Card id="logs" variant="secondary" width={220} height={85}>
          <Stack gap={8} alignItems="center" padding={8}>
            <Label bold size="lg">Log Storage</Label>
            <Subtitle>Request/response logs</Subtitle>
            <Label size="sm">Audit trail</Label>
          </Stack>
        </Card>
        <Card id="feedback" variant="secondary" width={220} height={85}>
          <Stack gap={8} alignItems="center" padding={8}>
            <Label bold size="lg">Feedback Loop</Label>
            <Subtitle>Model improvement</Subtitle>
            <Label size="sm">RLHF training data</Label>
          </Stack>
        </Card>
      </Row>
    </Stack>

    {/* Arrows - User to Gateway */}
    <Arrow from="web-app" to="auth" label="HTTPS" color="#1976d2" thickness="medium" />
    <Arrow from="mobile-app" to="auth" label="HTTPS" color="#1976d2" thickness="medium" />
    <Arrow from="backend" to="auth" label="HTTPS" color="#1976d2" thickness="medium" />

    {/* Arrows - Gateway flow */}
    <Arrow from="auth" to="load-balancer" label="validated" color="#388e3c" thickness="medium" />
    <Arrow from="load-balancer" to="queue" label="routed" color="#388e3c" thickness="medium" />

    {/* Arrows - Processing flow */}
    <Arrow from="queue" to="input-safety" label="request" color="#ff9800" thickness="medium" />
    <Arrow from="input-safety" to="prompt-optimization" label="safe" color="#66bb6a" thickness="medium" />
    <Arrow from="prompt-optimization" to="claude-model" label="optimized prompt" color="#1976d2" thickness="thick" />
    <Arrow from="claude-model" to="output-safety" label="raw response" color="#ab47bc" thickness="thick" />
    <Arrow from="output-safety" to="formatting" label="filtered" color="#66bb6a" thickness="medium" />

    {/* Arrows - Response back */}
    <Arrow from="formatting" to="web-app" label="JSON response" color="#4caf50" thickness="medium" curve="arc" />
    <Arrow from="formatting" to="mobile-app" label="JSON response" color="#4caf50" thickness="medium" curve="arc" />
    <Arrow from="formatting" to="backend" label="JSON response" color="#4caf50" thickness="medium" curve="arc" />

    {/* Arrows - Monitoring */}
    <Arrow from="load-balancer" to="analytics" label="metrics" color="#9e9e9e" style="dashed" />
    <Arrow from="queue" to="logs" label="log" color="#9e9e9e" style="dashed" />
    <Arrow from="claude-model" to="feedback" label="training data" color="#9e9e9e" style="dashed" />

    {/* Legend */}
    <Row gap={30} marginTop={30} justifyContent="center">
      <Badge text="🔒 Secure HTTPS transport" variant="primary" />
      <Badge text="🛡️ Multi-layer safety checks" variant="warning" />
      <Badge text="⚡ <100ms p50 latency" variant="success" />
      <Badge text="📊 Full observability" variant="secondary" />
    </Row>
  </Stack>
);

console.log('🎨 Generating Anthropic-style diagram...\n');

(async () => {
  try {
    const svg = await renderToSVG(<AnthropicStyleDiagram />, {
      width: 1600,
      height: 1200,
      backgroundColor: '#fafafa'
    });
    
    writeFileSync('examples/anthropic-style-diagram.svg', svg);
    console.log('✅ Generated anthropic-style-diagram.svg');
    console.log('\nThis diagram demonstrates:');
    console.log('  • Clear layer separation (User, Gateway, Processing, Data)');
    console.log('  • Professional card-based components with multiple text layers');
    console.log('  • Color-coded arrows showing data flow');
    console.log('  • Badges for metadata and status information');
    console.log('  • Clean typography hierarchy');
    console.log('  • Visual dividers between sections');
    console.log('\n✨ Ready for production documentation!');
  } catch (error) {
    console.error('Error generating diagram:', error);
    console.error('Stack:', error instanceof Error ? error.stack : String(error));
    process.exit(1);
  }
})();
