import React from 'react';
import { renderToSVGWithLayout } from '../renderer';
import { LayoutLinter } from '../test/layout-lints';

// Import the examples
import { Stack, Row, Card, Title, Subtitle, Label, Arrow } from '../index';

// Simple flowchart from styled.tsx
const StyledFlowchart = () => (
  <Stack gap={40} padding={40} alignItems="center">
    <Stack gap={16} alignItems="center">
      <Title level={1}>Modern Flowchart</Title>
      <Subtitle>Using semantic components for clean, professional styling</Subtitle>
    </Stack>
    
    <Card
      id="start"
      variant="primary"
      width={200}
      height={80}
    >
      <Stack gap={8} alignItems="center">
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
      <Stack gap={8} alignItems="center">
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
      <Stack gap={8} alignItems="center">
        <Label bold>Complete</Label>
        <Subtitle>Return results</Subtitle>
      </Stack>
    </Card>

    <Arrow from="start" to="process" color="#1976d2" strokeWidth={2} />
    <Arrow from="process" to="end" color="#388e3c" strokeWidth={2} />
  </Stack>
);

// Architecture diagram from styled.tsx
const StyledArchitecture = () => (
  <Stack gap={32} padding={40}>
    <Stack gap={8} alignItems="center">
      <Title level={1}>Three-Tier Architecture</Title>
      <Subtitle size="base">Modern web application design</Subtitle>
    </Stack>

    {/* Presentation Tier */}
    <Stack gap={12}>
      <Title level={3}>Presentation Tier</Title>
    <Row gap={32} justifyContent="center">
        <Card
          id="web-app"
          variant="primary"
          width={180}
          height={90}
        >
          <Stack gap={10} alignItems="center">
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
          <Stack gap={10} alignItems="center">
            <Label bold size="lg">Mobile App</Label>
            <Subtitle>React Native</Subtitle>
          </Stack>
        </Card>
      </Row>
    </Stack>

    {/* Business Logic Tier */}
    <Stack gap={12}>
      <Title level={3}>Business Logic Tier</Title>
      <Stack gap={24} alignItems="center">
        <Card
          id="api-gateway"
          variant="success"
          width={180}
          height={90}
        >
          <Stack gap={10} alignItems="center">
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
          <Stack gap={10} alignItems="center">
            <Label bold size="lg">Microservices</Label>
            <Subtitle>Business logic</Subtitle>
          </Stack>
        </Card>
      </Stack>
    </Stack>

    {/* Data Tier */}
    <Stack gap={12}>
      <Title level={3}>Data Tier</Title>
      <Row gap={32} justifyContent="center">
        <Card
          id="database"
          variant="secondary"
          width={180}
          height={90}
        >
          <Stack gap={10} alignItems="center">
            <Label bold size="lg">Database</Label>
            <Subtitle>PostgreSQL</Subtitle>
          </Stack>
        </Card>
        <Card
          id="cache"
          variant="secondary"
          width={180}
          height={90}
          marginLeft={32}
        >
          <Stack gap={10} alignItems="center">
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

async function runLints() {
  console.log('\n' + '='.repeat(70));
  console.log('Running Layout Lints on Actual Examples');
  console.log('='.repeat(70));

  // Lint flowchart
  console.log('\n📋 Linting: Styled Flowchart');
  console.log('-'.repeat(70));
  const flowchart = await renderToSVGWithLayout(<StyledFlowchart />, { 
    width: 800, 
    height: 700 
  });
  const linter1 = new LayoutLinter(flowchart.layout);
  const lints1 = linter1.runAllLints();
  
  if (lints1.length > 0) {
    console.log(LayoutLinter.formatLints(lints1));
  } else {
    console.log('✅ No layout issues found!\n');
  }

  // Lint architecture
  console.log('\n📋 Linting: Three-Tier Architecture');
  console.log('-'.repeat(70));
  const architecture = await renderToSVGWithLayout(<StyledArchitecture />, { 
    width: 900, 
    height: 800 
  });
  const linter2 = new LayoutLinter(architecture.layout);
  const lints2 = linter2.runAllLints();
  
  if (lints2.length > 0) {
    console.log(LayoutLinter.formatLints(lints2));
  } else {
    console.log('✅ No layout issues found!\n');
  }

  console.log('\n' + '='.repeat(70));
  console.log('Layout Linting Complete');
  console.log('='.repeat(70));
  console.log('\n💡 Tip: These are suggestions, not errors. Review them to improve');
  console.log('   the visual hierarchy and spacing of your diagrams.\n');
}

runLints().catch(console.error);
