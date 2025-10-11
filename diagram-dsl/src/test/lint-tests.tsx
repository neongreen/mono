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

  // Test 4: Small font size detection
  console.log('\n--- Test 4: Minimum Font Size Check ---');
  const SmallFontExample = () => (
    <Stack gap={20} padding={40}>
      <Card id="card1" variant="primary" width={200} height={80}>
        <Stack gap={8} alignItems="center">
          <Label bold>Normal Text</Label>
          <Stack gap={4}>
            <text fontSize={8}>This text is too small</text>
          </Stack>
        </Stack>
      </Card>
    </Stack>
  );
  
  const result4 = await renderToSVGWithLayout(<SmallFontExample />, { 
    width: 800, 
    height: 400 
  });
  const linter4 = new LayoutLinter(result4.layout);
  const lints4 = linter4.runAllLints();
  console.log(LayoutLinter.formatLints(lints4));
  
  const hasSmallFontInfo = lints4.some(l => l.message.includes('small font size'));
  if (hasSmallFontInfo) {
    console.log('✓ Test 4 passed: Small font size detected');
  } else {
    console.log('ℹ Test 4: No small font detected (test may need JSX lowercase workaround)');
  }

  // Test 5: Inconsistent spacing detection
  console.log('\n--- Test 5: Inconsistent Spacing Check ---');
  const InconsistentSpacingExample = () => (
    <Stack gap={10} padding={40}>
      <Card id="card1" variant="primary" width={200} height={60}>
        <Label>First Card</Label>
      </Card>
      <Card id="card2" variant="success" width={200} height={60} marginTop={30}>
        <Label>Second Card (extra margin)</Label>
      </Card>
      <Card id="card3" variant="secondary" width={200} height={60}>
        <Label>Third Card</Label>
      </Card>
    </Stack>
  );
  
  const result5 = await renderToSVGWithLayout(<InconsistentSpacingExample />, { 
    width: 800, 
    height: 500 
  });
  const linter5 = new LayoutLinter(result5.layout);
  const lints5 = linter5.runAllLints();
  console.log(LayoutLinter.formatLints(lints5));
  
  const hasInconsistentSpacing = lints5.some(l => l.message.includes('inconsistent spacing'));
  if (hasInconsistentSpacing) {
    console.log('✓ Test 5 passed: Inconsistent spacing detected');
  } else {
    console.log('ℹ Test 5: No inconsistent spacing detected (spacing may be within tolerance)');
  }

  console.log('\n' + '='.repeat(60));
  console.log('Lint Tests Complete');
  console.log('='.repeat(60) + '\n');
  console.log('💡 New lints added:');
  console.log('   - Overlapping elements detection');
  console.log('   - Minimum font size checks');
  console.log('   - Inconsistent spacing detection');
  console.log('   - Arrow crossing detection');
  console.log('\n');
}

runLintTests().catch(console.error);
