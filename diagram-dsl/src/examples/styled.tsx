import React from 'react';
import { Stack, Row, Card, Title, Subtitle, Label, Arrow, renderToSVG } from '../index';
import { writeFileSync } from 'fs';
import { join } from 'path';

/**
 * Example showcasing the new semantic/styled components
 * Demonstrates how to create beautiful diagrams with minimal code
 */

// Example 1: Simple flowchart using styled components
const StyledFlowchart = () => (
  <Stack gap={20} padding={40} alignItems="center">
    <Title level={1}>Modern Flowchart</Title>
    <Subtitle>Using semantic components for clean, professional styling</Subtitle>
    
    <Card
      id="start"
      variant="primary"
      width={200}
      height={80}
    >
      <Stack gap={4} alignItems="center">
        <Label bold>Start Process</Label>
        <Subtitle>Initialize system</Subtitle>
      </Stack>
    </Card>

    <Card
      id="process"
      variant="warning"
      width={200}
      height={80}
    >
      <Stack gap={4} alignItems="center">
        <Label bold>Process Data</Label>
        <Subtitle>Transform inputs</Subtitle>
      </Stack>
    </Card>

    <Card
      id="end"
      variant="success"
      width={200}
      height={80}
    >
      <Stack gap={4} alignItems="center">
        <Label bold>Complete</Label>
        <Subtitle>Return results</Subtitle>
      </Stack>
    </Card>

    <Arrow from="start" to="process" color="#1976d2" strokeWidth={2} />
    <Arrow from="process" to="end" color="#388e3c" strokeWidth={2} />
  </Stack>
);

// Example 2: Architecture diagram with semantic components
const StyledArchitecture = () => (
  <Stack gap={32} padding={40}>
    <Stack gap={8} alignItems="center">
      <Title level={1}>Three-Tier Architecture</Title>
      <Subtitle size="base">Modern web application design</Subtitle>
    </Stack>

    {/* Presentation Tier */}
    <Stack gap={12}>
      <Title level={3}>Presentation Tier</Title>
      <Row gap={16} justifyContent="center">
        <Card
          id="web-app"
          variant="primary"
          width={180}
          height={90}
        >
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Web Application</Label>
            <Subtitle>React + TypeScript</Subtitle>
          </Stack>
        </Card>
        <Card
          id="mobile-app"
          variant="primary"
          width={180}
          height={90}
        >
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Mobile App</Label>
            <Subtitle>React Native</Subtitle>
          </Stack>
        </Card>
      </Row>
    </Stack>

    {/* Business Logic Tier */}
    <Stack gap={12}>
      <Title level={3}>Business Logic Tier</Title>
      <Row gap={16} justifyContent="center">
        <Card
          id="api-gateway"
          variant="success"
          width={180}
          height={90}
        >
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">API Gateway</Label>
            <Subtitle>REST + GraphQL</Subtitle>
          </Stack>
        </Card>
        <Card
          id="services"
          variant="success"
          width={180}
          height={90}
        >
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Microservices</Label>
            <Subtitle>Business logic</Subtitle>
          </Stack>
        </Card>
      </Row>
    </Stack>

    {/* Data Tier */}
    <Stack gap={12}>
      <Title level={3}>Data Tier</Title>
      <Row gap={16} justifyContent="center">
        <Card
          id="database"
          variant="secondary"
          width={180}
          height={90}
        >
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Database</Label>
            <Subtitle>PostgreSQL</Subtitle>
          </Stack>
        </Card>
        <Card
          id="cache"
          variant="secondary"
          width={180}
          height={90}
        >
          <Stack gap={6} alignItems="center">
            <Label bold size="lg">Cache</Label>
            <Subtitle>Redis</Subtitle>
          </Stack>
        </Card>
      </Row>
    </Stack>

    <Arrow from="web-app" to="api-gateway" color="#1976d2" strokeWidth={2} label="HTTPS" />
    <Arrow from="mobile-app" to="api-gateway" color="#1976d2" strokeWidth={2} />
    <Arrow from="api-gateway" to="services" color="#388e3c" strokeWidth={2} />
    <Arrow from="services" to="database" color="#7b1fa2" strokeWidth={2} label="SQL" />
    <Arrow from="services" to="cache" color="#7b1fa2" strokeWidth={2} />
  </Stack>
);

// Example 3: Comparison - showing title hierarchy
const TitleHierarchy = () => (
  <Stack gap={24} padding={40} alignItems="center">
    <Title level={1}>Title Level 1</Title>
    <Subtitle>Main diagram title - 36px, bold, centered</Subtitle>

    <Title level={2}>Title Level 2</Title>
    <Subtitle>Section heading - 24px, bold, centered</Subtitle>

    <Title level={3}>Title Level 3</Title>
    <Subtitle>Subsection heading - 20px, bold, centered</Subtitle>

    <Label bold size="lg">Large Label (Bold)</Label>
    <Subtitle>For prominent information - 16px</Subtitle>

    <Label>Regular Label</Label>
    <Subtitle>Standard text - 14px</Subtitle>

    <Label size="sm">Small Label</Label>
    <Subtitle>Compact information - 12px</Subtitle>
  </Stack>
);

// Render examples
const outputDir = join(__dirname, '../../examples');

async function generateStyledExamples() {
  try {
    const svg1 = await renderToSVG(<StyledFlowchart />, { 
      width: 800, 
      height: 700, 
      backgroundColor: 'white' 
    });
    writeFileSync(join(outputDir, 'styled-flowchart.svg'), svg1);
    console.log('✓ Generated styled-flowchart.svg');

    const svg2 = await renderToSVG(<StyledArchitecture />, { 
      width: 900, 
      height: 800, 
      backgroundColor: 'white' 
    });
    writeFileSync(join(outputDir, 'styled-architecture.svg'), svg2);
    console.log('✓ Generated styled-architecture.svg');

    const svg3 = await renderToSVG(<TitleHierarchy />, { 
      width: 800, 
      height: 600, 
      backgroundColor: 'white' 
    });
    writeFileSync(join(outputDir, 'title-hierarchy.svg'), svg3);
    console.log('✓ Generated title-hierarchy.svg');

    console.log('\nStyled examples generated successfully!');
  } catch (error) {
    console.error('Error generating styled examples:', error);
    console.error('Stack:', error instanceof Error ? error.stack : String(error));
    process.exit(1);
  }
}

generateStyledExamples();
