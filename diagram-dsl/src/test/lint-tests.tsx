import React from 'react';
import { Stack, Row, Card, Label, Subtitle, Arrow, Text, renderToSVGWithLayout } from '../index';
import { LayoutLinter } from './layout-lints';

/**
 * Test cases for layout linting
 */

// Test 1: Short arrow - boxes are too close
const ShortArrowExample = () => (
  <Stack gap={8} padding={40} alignItems="center">
    <Card
      id="box1"
      variant="primary"
      width={200}
      height={80}
    >
      <Stack gap={8} alignItems="center">
        <Label bold>First Box</Label>
        <Subtitle>With label</Subtitle>
      </Stack>
    </Card>

    <Card
      id="box2"
      variant="success"
      width={200}
      height={80}
    >
      <Stack gap={8} alignItems="center">
        <Label bold>Second Box</Label>
        <Subtitle>Too close</Subtitle>
      </Stack>
    </Card>

    <Arrow from="box1" to="box2" color="#1976d2" strokeWidth={2} />
  </Stack>
);

// Test 2: Good arrow length - adequate spacing
const GoodArrowExample = () => (
  <Stack gap={40} padding={40} alignItems="center">
    <Card
      id="box1"
      variant="primary"
      width={200}
      height={80}
    >
      <Stack gap={8} alignItems="center">
        <Label bold>First Box</Label>
        <Subtitle>With label</Subtitle>
      </Stack>
    </Card>

    <Card
      id="box2"
      variant="success"
      width={200}
      height={80}
    >
      <Stack gap={8} alignItems="center">
        <Label bold>Second Box</Label>
        <Subtitle>Good spacing</Subtitle>
      </Stack>
    </Card>

    <Arrow from="box1" to="box2" color="#1976d2" strokeWidth={2} />
  </Stack>
);

// Test 3: Internal vs External spacing issue
const BadSpacingExample = () => (
  <Stack gap={10} padding={40} alignItems="center">
    <Card
      id="card1"
      variant="primary"
      width={200}
      height={80}
    >
      <Stack gap={8} alignItems="center">
        <Label bold>High Padding Card</Label>
        <Subtitle>Internal gap 16px default</Subtitle>
      </Stack>
    </Card>

    <Card
      id="card2"
      variant="success"
      width={200}
      height={80}
    >
      <Stack gap={8} alignItems="center">
        <Label bold>Another Card</Label>
        <Subtitle>Gap between cards: 10px</Subtitle>
      </Stack>
    </Card>
  </Stack>
);

// Test 4: Arrowheads too close to each other
const ArrowheadCollisionExample = () => (
  <Stack gap={24} padding={40} alignItems="center">
    <Row gap={16} justifyContent="center">
      <Card id="source-a" variant="primary" width={160} height={70}>
        <Stack gap={6} alignItems="center">
          <Label bold>Source A</Label>
          <Subtitle>Top arrow</Subtitle>
        </Stack>
      </Card>
      <Card id="source-b" variant="secondary" width={160} height={70}>
        <Stack gap={6} alignItems="center">
          <Label bold>Source B</Label>
          <Subtitle>Also top</Subtitle>
        </Stack>
      </Card>
    </Row>

    <Card id="target" variant="success" width={200} height={90}>
      <Stack gap={8} alignItems="center">
        <Label bold>Shared Target</Label>
        <Subtitle>Arrowheads overlap</Subtitle>
      </Stack>
    </Card>

    <Arrow from="source-a" to="target" color="#1976d2" strokeWidth={2} />
    <Arrow from="source-b" to="target" color="#7b1fa2" strokeWidth={2} />
  </Stack>
);

// Test 5: Arrow label overlap
const ArrowLabelOverlapExample = () => (
  <Stack width={700} height={260} padding={60} position="relative" alignItems="stretch">
    <Row id="row" gap={320} justifyContent="space-between">
      <Card id="left-node" variant="primary" width={160} height={80}>
        <Stack gap={6} alignItems="center">
          <Label bold>Left Node</Label>
          <Subtitle>Data Source</Subtitle>
        </Stack>
      </Card>
      <Card id="right-node" variant="success" width={160} height={80}>
        <Stack gap={6} alignItems="center">
          <Label bold>Right Node</Label>
          <Subtitle>Destination</Subtitle>
        </Stack>
      </Card>
    </Row>

    <Card
      id="blocker"
      variant="warning"
      width={220}
      height={70}
      position="absolute"
      top={110}
      left={240}
    >
      <Stack gap={4} alignItems="center">
        <Label bold>Analytics Panel</Label>
        <Subtitle>Label should not overlap</Subtitle>
      </Stack>
    </Card>

    <Arrow from="left-node" to="right-node" color="#1976d2" strokeWidth={2} label="Shared Telemetry" />
  </Stack>
);

// Test 6: Arrow crossings
const CrossingArrowsExample = () => (
  <Stack gap={60} padding={60}>
    <Row gap={120} justifyContent="center">
      <Card id="top-left" variant="primary" width={140} height={70}>
        <Stack gap={6} alignItems="center">
          <Label bold>Top Left</Label>
          <Subtitle>Start A</Subtitle>
        </Stack>
      </Card>
      <Card id="top-right" variant="secondary" width={140} height={70}>
        <Stack gap={6} alignItems="center">
          <Label bold>Top Right</Label>
          <Subtitle>Start B</Subtitle>
        </Stack>
      </Card>
    </Row>

    <Row gap={120} justifyContent="center">
      <Card id="bottom-left" variant="warning" width={140} height={70}>
        <Stack gap={6} alignItems="center">
          <Label bold>Bottom Left</Label>
          <Subtitle>End B</Subtitle>
        </Stack>
      </Card>
      <Card id="bottom-right" variant="success" width={140} height={70}>
        <Stack gap={6} alignItems="center">
          <Label bold>Bottom Right</Label>
          <Subtitle>End A</Subtitle>
        </Stack>
      </Card>
    </Row>

    <Arrow from="top-left" to="bottom-right" color="#1976d2" strokeWidth={2} />
    <Arrow from="top-right" to="bottom-left" color="#d32f2f" strokeWidth={2} />
  </Stack>
);

// Test 7: Text overflow inside a narrow card
const TextOverflowExample = () => (
  <Stack gap={24} padding={40} alignItems="center">
    <Card id="narrow-card" variant="secondary" width={140}>
      <Label id="narrow-label" size="lg" textAlign="center">
        This label text is intentionally too wide for the card
      </Label>
    </Card>
  </Stack>
);

// Test 8: Text spacing too tight
const TightTextSpacingExample = () => (
  <Card id="tight-card" variant="info" width={200} height={100} padding={20} alignItems="center" justifyContent="center">
    <Text id="tight-heading" fontSize={16}>Heading</Text>
    <Text id="tight-subtitle" fontSize={12}>Subtitle immediately below</Text>
  </Card>
);

async function runLintTests() {
  console.log('\n' + '='.repeat(60));
  console.log('Running Layout Lint Tests');
  console.log('='.repeat(60));

  // Test 1: Short arrow
  console.log('\n--- Test 1: Short Arrow Detection ---');
  const result1 = await renderToSVGWithLayout(<ShortArrowExample />, { 
    width: 800, 
    height: 400 
  });
  const linter1 = new LayoutLinter(result1.layout);
  const lints1 = linter1.runAllLints();
  console.log(LayoutLinter.formatLints(lints1));
  
  if (lints1.some(l => l.message.includes('Arrow') && l.message.includes('short'))) {
    console.log('✓ Test 1 passed: Short arrow detected');
  } else {
    console.log('✗ Test 1 failed: Short arrow not detected');
  }

  // Test 2: Good arrow length
  console.log('\n--- Test 2: Good Arrow Length (No Warning Expected) ---');
  const result2 = await renderToSVGWithLayout(<GoodArrowExample />, { 
    width: 800, 
    height: 400 
  });
  const linter2 = new LayoutLinter(result2.layout);
  const lints2 = linter2.runAllLints();
  console.log(LayoutLinter.formatLints(lints2));
  
  const hasShortArrowWarning = lints2.some(l => l.message.includes('Arrow') && l.message.includes('short'));
  if (!hasShortArrowWarning) {
    console.log('✓ Test 2 passed: No short arrow warning for good spacing');
  } else {
    console.log('✗ Test 2 failed: Unexpected short arrow warning');
  }

  // Test 3: Internal vs External spacing
  console.log('\n--- Test 3: Internal vs External Spacing Check ---');
  const result3 = await renderToSVGWithLayout(<BadSpacingExample />, { 
    width: 800, 
    height: 400 
  });
  const linter3 = new LayoutLinter(result3.layout);
  const lints3 = linter3.runAllLints();
  console.log(LayoutLinter.formatLints(lints3));
  
  const hasSpacingWarning = lints3.some(l => l.message.includes('internal spacing') || l.message.includes('external gap'));
  if (hasSpacingWarning) {
    console.log('✓ Test 3 passed: Internal vs external spacing issue detected');
  } else {
    console.log('✗ Test 3 failed: Spacing issue not detected');
  }

  // Test 4: Arrowhead proximity lint
  console.log('\n--- Test 4: Arrowhead Proximity Check ---');
  const result4 = await renderToSVGWithLayout(<ArrowheadCollisionExample />, {
    width: 800,
    height: 500
  });
  const linter4 = new LayoutLinter(result4.layout);
  const lints4 = linter4.runAllLints();
  console.log(LayoutLinter.formatLints(lints4));

  const hasArrowheadWarning = lints4.some(l => l.message.includes('Arrowheads') && l.message.includes('close'));
  if (hasArrowheadWarning) {
    console.log('✓ Test 4 passed: Arrowhead proximity issue detected');
  } else {
    console.log('✗ Test 4 failed: Arrowhead proximity issue not detected');
  }

  // Test 5: Arrow label overlap detection
  console.log('\n--- Test 5: Arrow Label Overlap Detection ---');
  const result5 = await renderToSVGWithLayout(<ArrowLabelOverlapExample />, {
    width: 800,
    height: 400
  });
  const linter5 = new LayoutLinter(result5.layout);
  const lints5 = linter5.runAllLints();
  console.log(LayoutLinter.formatLints(lints5));

  const hasLabelOverlap = lints5.some(l => l.message.includes('overlaps node'));
  if (hasLabelOverlap) {
    console.log('✓ Test 5 passed: Arrow label overlap detected');
  } else {
    console.log('✗ Test 5 failed: Arrow label overlap not detected');
  }

  // Test 6: Arrow crossing detection
  console.log('\n--- Test 6: Arrow Crossing Detection ---');
  const result6 = await renderToSVGWithLayout(<CrossingArrowsExample />, {
    width: 900,
    height: 700
  });
  const linter6 = new LayoutLinter(result6.layout);
  const lints6 = linter6.runAllLints();
  console.log(LayoutLinter.formatLints(lints6));

  const hasCrossingWarning = lints6.some(l => l.message.includes('cross each other'));
  if (hasCrossingWarning) {
    console.log('✓ Test 6 passed: Arrow crossing detected');
  } else {
    console.log('✗ Test 6 failed: Arrow crossing not detected');
  }

  // Test 7: Text overflow detection
  console.log('\n--- Test 7: Text Overflow Detection ---');
  const result7 = await renderToSVGWithLayout(<TextOverflowExample />, {
    width: 600,
    height: 300
  });
  const linter7 = new LayoutLinter(result7.layout);
  const lints7 = linter7.runAllLints();
  console.log(LayoutLinter.formatLints(lints7));

  const hasOverflowWarning = lints7.some(l => l.message.includes('exceeds available width'));
  if (hasOverflowWarning) {
    console.log('✓ Test 7 passed: Text overflow detected');
  } else {
    console.log('✗ Test 7 failed: Text overflow not detected');
  }

  // Test 8: Text spacing detection
  console.log('\n--- Test 8: Text Spacing Detection ---');
  const result8 = await renderToSVGWithLayout(<TightTextSpacingExample />, {
    width: 400,
    height: 300
  });
  const linter8 = new LayoutLinter(result8.layout);
  const lints8 = linter8.runAllLints();
  console.log(LayoutLinter.formatLints(lints8));

  const hasTightSpacingWarning = lints8.some(l => l.message.includes('separation'));
  if (hasTightSpacingWarning) {
    console.log('✓ Test 8 passed: Text spacing issue detected');
  } else {
    console.log('✗ Test 8 failed: Text spacing issue not detected');
  }

  console.log('\n' + '='.repeat(60));
  console.log('Lint Tests Complete');
  console.log('='.repeat(60) + '\n');
}

runLintTests().catch(console.error);
