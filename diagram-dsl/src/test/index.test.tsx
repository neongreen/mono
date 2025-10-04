import React from 'react';
import { Box, Stack, Text, Row, Column, Arrow, renderToSVG } from '../index';
import { runLayoutTests } from './layout-tests';

async function runTests() {
  console.log('Running diagram-dsl tests...\n');

  let passed = 0;
  let failed = 0;

  // Test 1: Simple Box renders
  try {
    const svg = await renderToSVG(
      <Box width={100} height={100} backgroundColor="red" />,
      { width: 200, height: 200 }
    );
    
    if (svg.includes('fill="red"') && svg.includes('width="100"') && svg.includes('height="100"')) {
      console.log('✓ Test 1: Simple Box renders correctly');
      passed++;
    } else {
      console.log('✗ Test 1: Box rendering failed');
      failed++;
    }
  } catch (error) {
    console.log('✗ Test 1: Error -', error);
    failed++;
  }

  // Test 2: Text renders with content
  try {
    const svg = await renderToSVG(
      <Text fontSize={24}>Hello World</Text>,
      { width: 200, height: 200 }
    );
    
    if (svg.includes('Hello World') && svg.includes('font-size="24"')) {
      console.log('✓ Test 2: Text renders with content and font size');
      passed++;
    } else {
      console.log('✗ Test 2: Text rendering failed');
      failed++;
    }
  } catch (error) {
    console.log('✗ Test 2: Error -', error);
    failed++;
  }

  // Test 3: Stack with children
  try {
    const svg = await renderToSVG(
      <Stack gap={10} padding={20}>
        <Box width={50} height={50} backgroundColor="red" />
        <Box width={50} height={50} backgroundColor="blue" />
      </Stack>,
      { width: 200, height: 200 }
    );
    
    if (svg.includes('fill="red"') && svg.includes('fill="blue"')) {
      console.log('✓ Test 3: Stack renders multiple children');
      passed++;
    } else {
      console.log('✗ Test 3: Stack rendering failed');
      failed++;
    }
  } catch (error) {
    console.log('✗ Test 3: Error -', error);
    failed++;
  }

  // Test 4: Row arranges children horizontally
  try {
    const svg = await renderToSVG(
      <Row gap={10}>
        <Box width={50} height={50} backgroundColor="red" id="box1" />
        <Box width={50} height={50} backgroundColor="blue" id="box2" />
      </Row>,
      { width: 200, height: 100 }
    );
    
    // Both boxes should be present
    if (svg.includes('fill="red"') && svg.includes('fill="blue"')) {
      console.log('✓ Test 4: Row renders children');
      passed++;
    } else {
      console.log('✗ Test 4: Row rendering failed');
      failed++;
    }
  } catch (error) {
    console.log('✗ Test 4: Error -', error);
    failed++;
  }

  // Test 5: Arrow between boxes
  try {
    const svg = await renderToSVG(
      <Stack>
        <Box id="start" width={100} height={50} backgroundColor="red" />
        <Box id="end" width={100} height={50} backgroundColor="blue" />
        <Arrow from="start" to="end" color="black" />
      </Stack>,
      { width: 200, height: 200 }
    );
    
    if (svg.includes('<line') && svg.includes('marker-end')) {
      console.log('✓ Test 5: Arrow renders between boxes');
      passed++;
    } else {
      console.log('✗ Test 5: Arrow rendering failed');
      failed++;
    }
  } catch (error) {
    console.log('✗ Test 5: Error -', error);
    failed++;
  }

  // Test 6: Box with border and border radius
  try {
    const svg = await renderToSVG(
      <Box
        width={100}
        height={100}
        backgroundColor="white"
        borderColor="black"
        borderWidth={2}
        borderRadius={10}
      />,
      { width: 200, height: 200 }
    );
    
    if (svg.includes('stroke="black"') && svg.includes('stroke-width="2"') && svg.includes('rx="10"')) {
      console.log('✓ Test 6: Box with border and border radius renders');
      passed++;
    } else {
      console.log('✗ Test 6: Border rendering failed');
      failed++;
    }
  } catch (error) {
    console.log('✗ Test 6: Error -', error);
    failed++;
  }

  // Test 7: Text alignment
  try {
    const svg = await renderToSVG(
      <Box width={200} height={50} alignItems="center" justifyContent="center">
        <Text textAlign="center">Centered Text</Text>
      </Box>,
      { width: 300, height: 100 }
    );
    
    if (svg.includes('Centered Text') && svg.includes('text-anchor="middle"')) {
      console.log('✓ Test 7: Text alignment works');
      passed++;
    } else {
      console.log('✗ Test 7: Text alignment failed');
      failed++;
    }
  } catch (error) {
    console.log('✗ Test 7: Error -', error);
    failed++;
  }

  // Summary
  console.log(`\n${'='.repeat(50)}`);
  console.log(`SVG rendering tests passed: ${passed}`);
  console.log(`SVG rendering tests failed: ${failed}`);
  console.log(`Total: ${passed + failed}`);
  console.log('='.repeat(50));

  return failed === 0;
}

async function runAllTests() {
  const svgTestsPass = await runTests();
  console.log('\n');
  const layoutTestsPass = await runLayoutTests();
  
  if (!svgTestsPass || !layoutTestsPass) {
    process.exit(1);
  }
}

runAllTests();
