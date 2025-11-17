// Build script for Readability test
const esbuild = require('esbuild');
const path = require('path');

async function build() {
  try {
    await esbuild.build({
      entryPoints: [path.join(__dirname, 'test-readability.js')],
      bundle: true,
      platform: 'node',
      target: 'es2020',
      outfile: path.join(__dirname, 'dist/readability-test-bundle.js'),
      format: 'cjs',
      external: [], // Bundle everything
      minify: false,
      sourcemap: false,
    });

    console.log('✓ Readability test bundle created');
    const fs = require('fs');
    const stats = fs.statSync(path.join(__dirname, 'dist/readability-test-bundle.js'));
    console.log(`✓ Bundle size: ${(stats.size / 1024).toFixed(2)} KB`);
  } catch (error) {
    console.error('Build failed:', error);
    process.exit(1);
  }
}

build();
