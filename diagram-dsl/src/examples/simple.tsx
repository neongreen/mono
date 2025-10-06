import React from 'react';
import { Box, Text, renderToSVG } from '../index';
import { writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Simplest possible example - just a box with text
const SimpleBox = () => (
  <Box
    width={300}
    height={200}
    backgroundColor="#f0f0f0"
    borderColor="#333"
    borderWidth={2}
    borderRadius={10}
    padding={20}
    justifyContent="center"
    alignItems="center"
  >
    <Text fontSize={24} fontWeight="bold">
      Hello, diagram-dsl!
    </Text>
  </Box>
);

const outputDir = join(__dirname, '../../examples');

async function generateSimpleExample() {
  try {
    const svg = await renderToSVG(<SimpleBox />, { 
      width: 400, 
      height: 300, 
      backgroundColor: 'white' 
    });
    writeFileSync(join(outputDir, 'simple-box.svg'), svg);
    console.log('✓ Generated simple-box.svg');
  } catch (error) {
    console.error('Error generating simple example:', error);
    process.exit(1);
  }
}

generateSimpleExample();
