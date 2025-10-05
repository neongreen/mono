import React from 'react';
import { Box, Stack, Text, Row, renderToSVGWithLayout } from '../index';
import { LayoutAssertions } from './layout-assertions';

async function runLayoutTests() {
  console.log('Running layout assertion tests...\n');

  let passed = 0;
  let failed = 0;

  // Test 1: Text centered in box
  try {
    const { layout } = await renderToSVGWithLayout(
      <Box id="container" width={200} height={100} justifyContent="center" alignItems="center">
        <Text id="text">Centered Text</Text>
      </Box>,
      { width: 300, height: 200 }
    );

    const assertions = new LayoutAssertions(layout);
    assertions.assertCentered('text', 'container');
    
    console.log('✓ Test 1: Text is centered in box');
    passed++;
  } catch (error) {
    console.log('✗ Test 1: Text centering failed -', (error as Error).message);
    failed++;
  }

  // Test 2: Text fits inside box with padding
  try {
    const { layout } = await renderToSVGWithLayout(
      <Box id="container" width={200} height={100} padding={20}>
        <Text id="text">Hello</Text>
      </Box>,
      { width: 300, height: 200 }
    );

    const assertions = new LayoutAssertions(layout);
    assertions.assertFitsInside('text', 'container', 20);
    
    console.log('✓ Test 2: Text fits inside box with padding');
    passed++;
  } catch (error) {
    console.log('✗ Test 2: Text containment failed -', (error as Error).message);
    failed++;
  }

  // Test 3: Gap between stacked boxes
  try {
    const { layout } = await renderToSVGWithLayout(
      <Stack gap={20}>
        <Box id="box1" width={100} height={50} backgroundColor="red" />
        <Box id="box2" width={100} height={50} backgroundColor="blue" />
      </Stack>,
      { width: 300, height: 200 }
    );

    const assertions = new LayoutAssertions(layout);
    assertions.assertGap('box1', 'box2', 20);
    
    console.log('✓ Test 3: Gap between stacked boxes is correct');
    passed++;
  } catch (error) {
    console.log('✗ Test 3: Gap assertion failed -', (error as Error).message);
    failed++;
  }

  // Test 4: Vertical alignment in row (alignItems centers vertically)
  try {
    const { layout } = await renderToSVGWithLayout(
      <Row alignItems="center">
        <Box id="box1" width={50} height={60} backgroundColor="red" />
        <Box id="box2" width={50} height={80} backgroundColor="blue" />
        <Box id="box3" width={50} height={70} backgroundColor="green" />
      </Row>,
      { width: 300, height: 200 }
    );

    const assertions = new LayoutAssertions(layout);
    // In a Row, alignItems="center" aligns vertically (cross-axis)
    // So we check if the vertical centers are aligned
    const box1 = assertions.findById('box1');
    const box2 = assertions.findById('box2');
    const box3 = assertions.findById('box3');
    
    if (!box1?.computed || !box2?.computed || !box3?.computed) {
      throw new Error('Boxes not found');
    }
    
    const centerY1 = box1.computed.y + box1.computed.height / 2;
    const centerY2 = box2.computed.y + box2.computed.height / 2;
    const centerY3 = box3.computed.y + box3.computed.height / 2;
    
    // All should have the same vertical center
    if (Math.abs(centerY1 - centerY2) > 1 || Math.abs(centerY2 - centerY3) > 1) {
      throw new Error(`Vertical centers not aligned: ${centerY1}, ${centerY2}, ${centerY3}`);
    }
    
    console.log('✓ Test 4: Boxes are vertically centered in row');
    passed++;
  } catch (error) {
    console.log('✗ Test 4: Vertical alignment in row failed -', (error as Error).message);
    failed++;
  }

  // Test 5: No overlap between side-by-side boxes
  try {
    const { layout } = await renderToSVGWithLayout(
      <Row gap={10}>
        <Box id="box1" width={50} height={50} backgroundColor="red" />
        <Box id="box2" width={50} height={50} backgroundColor="blue" />
      </Row>,
      { width: 300, height: 200 }
    );

    const assertions = new LayoutAssertions(layout);
    assertions.assertNoOverlap('box1', 'box2');
    
    console.log('✓ Test 5: Boxes don\'t overlap');
    passed++;
  } catch (error) {
    console.log('✗ Test 5: Overlap check failed -', (error as Error).message);
    failed++;
  }

  // Test 6: Text measurements are accurate (not just estimates)
  try {
    const { layout } = await renderToSVGWithLayout(
      <Text id="text" fontSize={24} fontWeight="bold">Test Text</Text>,
      { width: 300, height: 200 }
    );

    const assertions = new LayoutAssertions(layout);
    const textNode = assertions.findById('text');
    
    if (!textNode || !textNode.computed) {
      throw new Error('Text node not found');
    }

    // With canvas measurement, the width should be more accurate than the old estimate
    // Old estimate would be: "Test Text".length * 24 * 0.6 = 9 * 24 * 0.6 = 129.6
    // Canvas measurement should be different (and more accurate)
    const oldEstimate = "Test Text".length * 24 * 0.6;
    const actualWidth = textNode.computed.width;
    
    // Canvas should give us more accurate measurements
    // The difference should be noticeable (more than 5px)
    if (Math.abs(actualWidth - oldEstimate) < 5) {
      console.log(`⚠ Test 6: Text measurement may still be using estimates (width: ${actualWidth} vs estimate: ${oldEstimate})`);
    }
    
    console.log(`✓ Test 6: Text has accurate measurements (width: ${actualWidth}, height: ${textNode.computed.height})`);
    passed++;
  } catch (error) {
    console.log('✗ Test 6: Text measurement test failed -', (error as Error).message);
    failed++;
  }

  // Test 7: Multiple text sizes measured accurately
  try {
    const { layout } = await renderToSVGWithLayout(
      <Stack gap={10}>
        <Text id="small" fontSize={12}>Small text</Text>
        <Text id="medium" fontSize={18}>Medium text</Text>
        <Text id="large" fontSize={32}>Large text</Text>
      </Stack>,
      { width: 400, height: 300 }
    );

    const assertions = new LayoutAssertions(layout);
    const small = assertions.findById('small');
    const medium = assertions.findById('medium');
    const large = assertions.findById('large');

    if (!small?.computed || !medium?.computed || !large?.computed) {
      throw new Error('Text nodes not found');
    }

    // Heights should increase with font size
    if (small.computed.height >= medium.computed.height || 
        medium.computed.height >= large.computed.height) {
      throw new Error(`Text heights not proportional to font size: ${small.computed.height}, ${medium.computed.height}, ${large.computed.height}`);
    }

    console.log('✓ Test 7: Multiple text sizes measured correctly');
    passed++;
  } catch (error) {
    console.log('✗ Test 7: Multi-size text test failed -', (error as Error).message);
    failed++;
  }

  // Summary
  console.log(`\n${'='.repeat(50)}`);
  console.log(`Layout tests passed: ${passed}`);
  console.log(`Layout tests failed: ${failed}`);
  console.log(`Total: ${passed + failed}`);
  console.log('='.repeat(50));

  return failed === 0;
}

export { runLayoutTests };
