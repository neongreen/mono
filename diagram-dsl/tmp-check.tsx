import React from 'react';
import { Stack, Text, Box, Arrow, renderToSVGWithLayout, LayoutLinter } from './src/index';

const ArchitectureDiagram = () => (
  <Stack gap={30} padding={40} width={800} height={600}>
    <Text fontSize={32} fontWeight="bold">Architecture Diagram</Text>
    
    <Stack gap={20} alignItems="center">
      <Box id="frontend" width={180} height={100} backgroundColor="#e3f2fd" borderColor="#1976d2" borderWidth={2} borderRadius={8} padding={15} justifyContent="center" alignItems="center">
        <Text fontSize={16} fontWeight="bold">Frontend</Text>
        <Text fontSize={12}>React App</Text>
      </Box>
      <Box id="api" width={180} height={100} backgroundColor="#fce4ec" borderColor="#c2185b" borderWidth={2} borderRadius={8} padding={15} justifyContent="center" alignItems="center">
        <Text fontSize={16} fontWeight="bold">API Gateway</Text>
        <Text fontSize={12}>REST API</Text>
      </Box>
      <Box id="database" width={180} height={100} backgroundColor="#f3e5f5" borderColor="#7b1fa2" borderWidth={2} borderRadius={8} padding={15} justifyContent="center" alignItems="center">
        <Text fontSize={16} fontWeight="bold">Database</Text>
        <Text fontSize={12}>PostgreSQL</Text>
      </Box>
    </Stack>

    <Arrow from="frontend" to="api" color="#1976d2" strokeWidth={2} label="HTTP" />
    <Arrow from="api" to="database" color="#c2185b" strokeWidth={2} label="SQL" />
  </Stack>
);

(async () => {
  const { layout } = await renderToSVGWithLayout(<ArchitectureDiagram />, {
    width: 800,
    height: 600,
    backgroundColor: 'white',
  });
  const lints = new LayoutLinter(layout).runAllLints();
  console.log(LayoutLinter.formatLints(lints));
})();
