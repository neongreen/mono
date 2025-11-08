import * as esbuild from 'esbuild';

await esbuild.build({
    entryPoints: ['extract.js'],
    bundle: true,
    platform: 'node',
    target: 'node18',
    outfile: 'dist/mozilla.bundle.js',
    format: 'esm',
    banner: {
        js: '#!/usr/bin/env node'
    }
});

console.log('Build complete: dist/mozilla.bundle.js');
