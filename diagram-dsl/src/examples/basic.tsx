import React from 'react';
import { Box, Stack, Text, Arrow, renderToSVG } from '../index';
import { writeFileSync } from 'fs';
import { join } from 'path';

// Example 1: Simple vertical stack with boxes
const Example1 = () => (
  <Stack gap={20} padding={40} alignItems="center" width={800} height={600}>
    <Text fontSize={32} fontWeight="bold">Simple Flowchart</Text>
    
    <Box
      id="start"
      width={200}
      height={80}
      backgroundColor="#e0f7fa"
      borderColor="#00acc1"
      borderWidth={2}
      borderRadius={8}
      padding={20}
      justifyContent="center"
      alignItems="center"
    >
      <Text fontSize={18}>Start Process</Text>
    </Box>

    <Box
      id="process"
      width={200}
      height={80}
      backgroundColor="#fff3e0"
      borderColor="#fb8c00"
      borderWidth={2}
      borderRadius={8}
      padding={20}
      justifyContent="center"
      alignItems="center"
    >
      <Text fontSize={18}>Process Data</Text>
    </Box>

    <Box
      id="end"
      width={200}
      height={80}
      backgroundColor="#f1f8e9"
      borderColor="#7cb342"
      borderWidth={2}
      borderRadius={8}
      padding={20}
      justifyContent="center"
      alignItems="center"
    >
      <Text fontSize={18}>End Process</Text>
    </Box>

    <Arrow from="start" to="process" color="#00acc1" strokeWidth={2} />
    <Arrow from="process" to="end" color="#fb8c00" strokeWidth={2} />
  </Stack>
);

// Example 2: Horizontal row with multiple boxes
const Example2 = () => (
  <Stack gap={30} padding={40} width={800} height={600}>
    <Text fontSize={32} fontWeight="bold">Architecture Diagram</Text>
    
    <Stack gap={20} alignItems="center">
      <Box
        id="frontend"
        width={180}
        height={100}
        backgroundColor="#e3f2fd"
        borderColor="#1976d2"
        borderWidth={2}
        borderRadius={8}
        padding={15}
        justifyContent="center"
        alignItems="center"
      >
        <Text fontSize={16} fontWeight="bold">Frontend</Text>
        <Text fontSize={12}>React App</Text>
      </Box>

      <Box
        id="api"
        width={180}
        height={100}
        backgroundColor="#fce4ec"
        borderColor="#c2185b"
        borderWidth={2}
        borderRadius={8}
        padding={15}
        justifyContent="center"
        alignItems="center"
      >
        <Text fontSize={16} fontWeight="bold">API Gateway</Text>
        <Text fontSize={12}>REST API</Text>
      </Box>

      <Box
        id="database"
        width={180}
        height={100}
        backgroundColor="#f3e5f5"
        borderColor="#7b1fa2"
        borderWidth={2}
        borderRadius={8}
        padding={15}
        justifyContent="center"
        alignItems="center"
      >
        <Text fontSize={16} fontWeight="bold">Database</Text>
        <Text fontSize={12}>PostgreSQL</Text>
      </Box>
    </Stack>

    <Arrow from="frontend" to="api" color="#1976d2" strokeWidth={2} label="HTTP" />
    <Arrow from="api" to="database" color="#c2185b" strokeWidth={2} label="SQL" />
  </Stack>
);

// Render examples
const outputDir = join(__dirname, '../../examples');

async function generateExamples() {
  try {
    const svg1 = await renderToSVG(<Example1 />, { width: 800, height: 600, backgroundColor: 'white' });
    writeFileSync(join(outputDir, 'basic-flowchart.svg'), svg1);
    console.log('✓ Generated basic-flowchart.svg');

    const svg2 = await renderToSVG(<Example2 />, { width: 800, height: 600, backgroundColor: 'white' });
    writeFileSync(join(outputDir, 'architecture-diagram.svg'), svg2);
    console.log('✓ Generated architecture-diagram.svg');

    console.log('\nExamples generated successfully!');
  } catch (error) {
    console.error('Error generating examples:', error);
    console.error('Stack:', error instanceof Error ? error.stack : String(error));
    process.exit(1);
  }
}

generateExamples();
