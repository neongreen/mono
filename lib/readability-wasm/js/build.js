// Build script to bundle Postlight Parser for WASM compilation
const esbuild = require('esbuild');
const fs = require('fs');
const path = require('path');

async function build() {
  try {
    // Bundle the parser with all dependencies
    await esbuild.build({
      entryPoints: ['index.js'],
      bundle: true,
      platform: 'node',
      target: 'es2020',
      outfile: 'dist/bundle.js',
      format: 'cjs',
      external: [], // Bundle everything
      minify: false, // Keep readable for debugging
      sourcemap: false,
    });

    console.log('✓ Bundle created successfully');

    // Verify the bundle exists
    const bundlePath = path.join(__dirname, 'dist', 'bundle.js');
    if (fs.existsSync(bundlePath)) {
      const stats = fs.statSync(bundlePath);
      console.log(`✓ Bundle size: ${(stats.size / 1024).toFixed(2)} KB`);
    }
  } catch (error) {
    console.error('Build failed:', error);
    process.exit(1);
  }
}

build();
