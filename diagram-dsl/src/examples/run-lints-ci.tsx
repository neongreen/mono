import React from 'react';
import { renderToSVGWithLayout } from '../renderer';
import { LayoutLinter, LayoutLint } from '../test/layout-lints';

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
      <Row gap={16} justifyContent="center">
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
      <Stack gap={16} alignItems="center">
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
      <Row gap={16} justifyContent="center">
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

interface ExampleFile {
  name: string;
  file: string;
  component: () => JSX.Element;
  width: number;
  height: number;
}

const examples: ExampleFile[] = [
  {
    name: 'Styled Flowchart',
    file: 'src/examples/styled.tsx',
    component: StyledFlowchart,
    width: 800,
    height: 700,
  },
  {
    name: 'Three-Tier Architecture',
    file: 'src/examples/styled.tsx',
    component: StyledArchitecture,
    width: 900,
    height: 800,
  },
];

/**
 * Output lint in GitHub Actions format
 * https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions
 */
function outputGitHubAnnotation(lint: LayoutLint, file: string) {
  const level = lint.type === 'warning' ? 'warning' : 'notice';
  // GitHub Actions format: ::warning file={name},line={line},col={col}::{message}
  // Since we don't have line numbers, we'll just output the file
  console.log(`::${level} file=${file},title=Layout Lint::${lint.message}`);
}

async function runLints() {
  let totalLints = 0;
  let hasWarnings = false;

  console.log('\n' + '='.repeat(70));
  console.log('Running Layout Lints on Examples');
  console.log('='.repeat(70));

  for (const example of examples) {
    console.log(`\n📋 Linting: ${example.name}`);
    console.log('-'.repeat(70));

    try {
      const result = await renderToSVGWithLayout(<example.component />, {
        width: example.width,
        height: example.height,
      });
      const linter = new LayoutLinter(result.layout);
      const lints = linter.runAllLints();

      if (lints.length > 0) {
        console.log(LayoutLinter.formatLints(lints));
        totalLints += lints.length;
        
        // Output GitHub Actions annotations
        for (const lint of lints) {
          outputGitHubAnnotation(lint, example.file);
          if (lint.type === 'warning') {
            hasWarnings = true;
          }
        }
      } else {
        console.log('✅ No layout issues found!\n');
      }
    } catch (error) {
      console.error(`Error linting ${example.name}:`, error);
      process.exit(1);
    }
  }

  console.log('\n' + '='.repeat(70));
  console.log('Layout Linting Complete');
  console.log('='.repeat(70));
  console.log(`\nTotal issues found: ${totalLints}`);
  console.log('\n💡 Tip: These are suggestions, not errors. Review them to improve');
  console.log('   the visual hierarchy and spacing of your diagrams.\n');

  // Exit with non-zero if there are warnings (but workflow will be marked as optional)
  if (hasWarnings) {
    process.exit(1);
  }
}

runLints().catch((error) => {
  console.error('Fatal error:', error);
  process.exit(1);
});
