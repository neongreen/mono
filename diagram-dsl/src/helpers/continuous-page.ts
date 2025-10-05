/**
 * Continuous Page Mode (Pageless)
 * Like Google Docs pageless mode - content flows without slide boundaries
 */

import React from 'react';
import { renderToSVG } from '../renderer';
import { PresentationTheme, getCurrentTheme } from '../presentation-theme';

export interface ContentBlock {
  name: string;
  component: React.ReactElement;
  spacing?: number; // Extra spacing after this block
}

export interface ContinuousPageOptions {
  width?: number;
  backgroundColor?: string;
  padding?: number;
  gap?: number; // Gap between content blocks
  theme?: PresentationTheme;
}

/**
 * Generate a single continuous vertical page with no slide boundaries
 * Content flows seamlessly from top to bottom
 */
export async function generateContinuousPage(
  blocks: ContentBlock[],
  options: ContinuousPageOptions = {}
): Promise<string> {
  const theme = options.theme || getCurrentTheme();
  const width = options.width || theme.slideWidth;
  const backgroundColor = options.backgroundColor || theme.background;
  const padding = options.padding ?? theme.slidePadding;
  const gap = options.gap ?? theme.gapLarge;

  // Render each block to get dimensions
  const renderedBlocks: Array<{
    name: string;
    svg: string;
    height: number;
    spacing: number;
  }> = [];

  let totalHeight = padding; // Start with top padding

  for (const block of blocks) {
    const result = await renderToSVG(block.component, {
      width: width - 2 * padding,
      height: 10000, // Large initial height, will be trimmed
      backgroundColor: 'none',
    });

    // Extract actual height from SVG
    const heightMatch = result.match(/height="(\d+)"/);
    const blockHeight = heightMatch ? parseInt(heightMatch[1]) : 800;
    
    const blockSpacing = block.spacing ?? gap;
    
    renderedBlocks.push({
      name: block.name,
      svg: result,
      height: blockHeight,
      spacing: blockSpacing,
    });

    totalHeight += blockHeight + blockSpacing;
  }

  totalHeight += padding; // Add bottom padding

  // Combine all blocks into one continuous SVG
  let combinedSVG = `<svg width="${width}" height="${totalHeight}" xmlns="http://www.w3.org/2000/svg">`;
  
  // Background
  combinedSVG += `<rect width="${width}" height="${totalHeight}" fill="${backgroundColor}"/>`;
  
  let currentY = padding;
  
  for (const block of renderedBlocks) {
    // Extract the inner content from each SVG (skip the outer svg tag)
    const contentMatch = block.svg.match(/<svg[^>]*>([\s\S]*)<\/svg>/);
    const content = contentMatch ? contentMatch[1] : block.svg;
    
    // Wrap in a group and position it
    combinedSVG += `<g transform="translate(${padding}, ${currentY})">`;
    combinedSVG += content;
    combinedSVG += `</g>`;
    
    currentY += block.height + block.spacing;
  }
  
  combinedSVG += `</svg>`;
  
  return combinedSVG;
}

/**
 * Helper to create a themed continuous page with consistent styling
 */
export async function generateThemedContinuousPage(
  blocks: ContentBlock[],
  theme: PresentationTheme,
  options: Omit<ContinuousPageOptions, 'theme'> = {}
): Promise<string> {
  return generateContinuousPage(blocks, {
    ...options,
    theme,
    backgroundColor: options.backgroundColor || theme.background,
  });
}
