import React from 'react';
import { Stack, Card, Label, Subtitle, Arrow, renderToSVGWithLayout } from '../index';
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

  console.log('\n' + '='.repeat(60));
  console.log('Lint Tests Complete');
  console.log('='.repeat(60) + '\n');
}

runLintTests().catch(console.error);
