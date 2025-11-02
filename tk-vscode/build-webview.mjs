import * as esbuild from 'esbuild';
import * as fs from 'fs';

const production = process.argv.includes('--production');
const watch = process.argv.includes('--watch');

async function main() {
  const ctx = await esbuild.context({
    entryPoints: ['src/webview/index.tsx'],
    bundle: true,
    outfile: 'out/webview.js',
    platform: 'browser',
    format: 'iife',
    minify: production,
    sourcemap: !production,
    jsxFactory: 'h',
    jsxFragment: 'Fragment',
    jsx: 'automatic',
    jsxImportSource: 'preact',
    external: [],
  });

  if (watch) {
    await ctx.watch();
    console.log('Watching webview...');
  } else {
    await ctx.rebuild();
    await ctx.dispose();
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
