import { execSync } from 'child_process';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/**
 * Master script to generate all examples (both diagram-dsl and D2)
 * This allows easy comparison between the two tools
 */

const rootDir = join(__dirname, '../..');

console.log('='.repeat(50));
console.log('Generating all diagram examples');
console.log('='.repeat(50));
console.log();

// Step 1: Generate diagram-dsl examples
console.log('Step 1: Generating diagram-dsl examples...');
console.log('-'.repeat(50));
try {
  execSync('npm run examples', { 
    cwd: rootDir, 
    stdio: 'inherit' 
  });
  console.log();
} catch (error) {
  console.error('Failed to generate diagram-dsl examples');
  process.exit(1);
}

// Step 2: Generate D2 examples
console.log('Step 2: Generating D2 examples...');
console.log('-'.repeat(50));
try {
  execSync('tsx src/examples/generate-d2.ts', { 
    cwd: rootDir, 
    stdio: 'inherit' 
  });
} catch (error) {
  console.error('Failed to generate D2 examples');
  process.exit(1);
}

console.log('='.repeat(50));
console.log('All examples generated successfully!');
console.log('='.repeat(50));
console.log();
console.log('diagram-dsl outputs: examples/*.svg');
console.log('D2 outputs:          examples/d2-output/*.svg');
console.log();
