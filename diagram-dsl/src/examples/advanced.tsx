import React from 'react';
import { Box, Stack, Row, Text, Arrow, renderToSVG } from '../index';
import { writeFileSync } from 'fs';
import { join } from 'path';

// Example: Multi-tier application architecture
const MultiTierArchitecture = () => (
  <Stack gap={40} padding={40} width={1000} height={700}>
    <Text fontSize={36} fontWeight="bold" textAlign="center">
      Multi-Tier Web Application Architecture
    </Text>

    {/* Client Tier */}
    <Stack gap={15}>
      <Text fontSize={20} fontWeight="bold" color="#1976d2">Client Tier</Text>
      <Row gap={15} justifyContent="center">
        <Box
          id="web"
          width={150}
          height={80}
          backgroundColor="#e3f2fd"
          borderColor="#1976d2"
          borderWidth={2}
          borderRadius={8}
          padding={15}
          justifyContent="center"
          alignItems="center"
        >
          <Text fontSize={14} fontWeight="bold">Web Browser</Text>
          <Text fontSize={11}>HTML/CSS/JS</Text>
        </Box>
        <Box
          id="mobile"
          width={150}
          height={80}
          backgroundColor="#e3f2fd"
          borderColor="#1976d2"
          borderWidth={2}
          borderRadius={8}
          padding={15}
          justifyContent="center"
          alignItems="center"
        >
          <Text fontSize={14} fontWeight="bold">Mobile App</Text>
          <Text fontSize={11}>iOS/Android</Text>
        </Box>
      </Row>
    </Stack>

    {/* Application Tier */}
    <Stack gap={15}>
      <Text fontSize={20} fontWeight="bold" color="#388e3c">Application Tier</Text>
      <Row gap={15} justifyContent="center">
        <Box
          id="api"
          width={150}
          height={80}
          backgroundColor="#e8f5e9"
          borderColor="#388e3c"
          borderWidth={2}
          borderRadius={8}
          padding={15}
          justifyContent="center"
          alignItems="center"
        >
          <Text fontSize={14} fontWeight="bold">API Gateway</Text>
          <Text fontSize={11}>REST/GraphQL</Text>
        </Box>
        <Box
          id="auth"
          width={150}
          height={80}
          backgroundColor="#e8f5e9"
          borderColor="#388e3c"
          borderWidth={2}
          borderRadius={8}
          padding={15}
          justifyContent="center"
          alignItems="center"
        >
          <Text fontSize={14} fontWeight="bold">Auth Service</Text>
          <Text fontSize={11}>OAuth2/JWT</Text>
        </Box>
        <Box
          id="business"
          width={150}
          height={80}
          backgroundColor="#e8f5e9"
          borderColor="#388e3c"
          borderWidth={2}
          borderRadius={8}
          padding={15}
          justifyContent="center"
          alignItems="center"
        >
          <Text fontSize={14} fontWeight="bold">Business Logic</Text>
          <Text fontSize={11}>Services</Text>
        </Box>
      </Row>
    </Stack>

    {/* Data Tier */}
    <Stack gap={15}>
      <Text fontSize={20} fontWeight="bold" color="#7b1fa2">Data Tier</Text>
      <Row gap={15} justifyContent="center">
        <Box
          id="db"
          width={150}
          height={80}
          backgroundColor="#f3e5f5"
          borderColor="#7b1fa2"
          borderWidth={2}
          borderRadius={8}
          padding={15}
          justifyContent="center"
          alignItems="center"
        >
          <Text fontSize={14} fontWeight="bold">Database</Text>
          <Text fontSize={11}>PostgreSQL</Text>
        </Box>
        <Box
          id="cache"
          width={150}
          height={80}
          backgroundColor="#f3e5f5"
          borderColor="#7b1fa2"
          borderWidth={2}
          borderRadius={8}
          padding={15}
          justifyContent="center"
          alignItems="center"
        >
          <Text fontSize={14} fontWeight="bold">Cache</Text>
          <Text fontSize={11}>Redis</Text>
        </Box>
        <Box
          id="storage"
          width={150}
          height={80}
          backgroundColor="#f3e5f5"
          borderColor="#7b1fa2"
          borderWidth={2}
          borderRadius={8}
          padding={15}
          justifyContent="center"
          alignItems="center"
        >
          <Text fontSize={14} fontWeight="bold">File Storage</Text>
          <Text fontSize={11}>S3/Azure Blob</Text>
        </Box>
      </Row>
    </Stack>

    {/* Arrows showing connections */}
    <Arrow from="web" to="api" color="#1976d2" strokeWidth={2} label="HTTPS" />
    <Arrow from="mobile" to="api" color="#1976d2" strokeWidth={2} label="HTTPS" />
    <Arrow from="api" to="auth" color="#388e3c" strokeWidth={2} />
    <Arrow from="api" to="business" color="#388e3c" strokeWidth={2} />
    <Arrow from="business" to="db" color="#7b1fa2" strokeWidth={2} label="SQL" />
    <Arrow from="business" to="cache" color="#7b1fa2" strokeWidth={2} />
  </Stack>
);

// Example: Decision flowchart with conditional paths
const DecisionFlowchart = () => (
  <Stack gap={30} padding={40} width={800} height={700} alignItems="center">
    <Text fontSize={32} fontWeight="bold">
      User Authentication Flow
    </Text>

    <Box
      id="start"
      width={180}
      height={70}
      backgroundColor="#fff3e0"
      borderColor="#f57c00"
      borderWidth={2}
      borderRadius={35}
      padding={15}
      justifyContent="center"
      alignItems="center"
    >
      <Text fontSize={16}>Start: Login Request</Text>
    </Box>

    <Box
      id="validate"
      width={200}
      height={80}
      backgroundColor="#e1f5fe"
      borderColor="#0288d1"
      borderWidth={2}
      borderRadius={8}
      padding={15}
      justifyContent="center"
      alignItems="center"
    >
      <Text fontSize={15}>Validate Credentials</Text>
    </Box>

    <Row gap={60} justifyContent="center" width={700}>
      <Box
        id="valid"
        width={180}
        height={80}
        backgroundColor="#e8f5e9"
        borderColor="#43a047"
        borderWidth={2}
        borderRadius={8}
        padding={15}
        justifyContent="center"
        alignItems="center"
      >
        <Text fontSize={14}>Valid ✓</Text>
        <Text fontSize={12}>Generate Token</Text>
      </Box>

      <Box
        id="invalid"
        width={180}
        height={80}
        backgroundColor="#ffebee"
        borderColor="#e53935"
        borderWidth={2}
        borderRadius={8}
        padding={15}
        justifyContent="center"
        alignItems="center"
      >
        <Text fontSize={14}>Invalid ✗</Text>
        <Text fontSize={12}>Show Error</Text>
      </Box>
    </Row>

    <Row gap={60} justifyContent="center" width={700}>
      <Box
        id="success"
        width={180}
        height={70}
        backgroundColor="#c8e6c9"
        borderColor="#43a047"
        borderWidth={2}
        borderRadius={35}
        padding={15}
        justifyContent="center"
        alignItems="center"
      >
        <Text fontSize={15}>Success: Logged In</Text>
      </Box>

      <Box
        id="retry"
        width={180}
        height={70}
        backgroundColor="#ffcdd2"
        borderColor="#e53935"
        borderWidth={2}
        borderRadius={35}
        padding={15}
        justifyContent="center"
        alignItems="center"
      >
        <Text fontSize={15}>Retry Login</Text>
      </Box>
    </Row>

    <Arrow from="start" to="validate" color="#0288d1" strokeWidth={2} />
    <Arrow from="validate" to="valid" color="#43a047" strokeWidth={2} label="Yes" />
    <Arrow from="validate" to="invalid" color="#e53935" strokeWidth={2} label="No" />
    <Arrow from="valid" to="success" color="#43a047" strokeWidth={2} />
    <Arrow from="invalid" to="retry" color="#e53935" strokeWidth={2} />
  </Stack>
);

// Render examples
const outputDir = join(__dirname, '../../examples');

async function generateAdvancedExamples() {
  try {
    const svg1 = await renderToSVG(<MultiTierArchitecture />, { 
      width: 1000, 
      height: 700, 
      backgroundColor: 'white' 
    });
    writeFileSync(join(outputDir, 'multi-tier-architecture.svg'), svg1);
    console.log('✓ Generated multi-tier-architecture.svg');

    const svg2 = await renderToSVG(<DecisionFlowchart />, { 
      width: 800, 
      height: 700, 
      backgroundColor: 'white' 
    });
    writeFileSync(join(outputDir, 'decision-flowchart.svg'), svg2);
    console.log('✓ Generated decision-flowchart.svg');

    console.log('\nAdvanced examples generated successfully!');
  } catch (error) {
    console.error('Error generating advanced examples:', error);
    console.error('Stack:', error instanceof Error ? error.stack : String(error));
    process.exit(1);
  }
}

generateAdvancedExamples();
