import React from 'react';
import { renderToSVG } from '../renderer';
import { writeFileSync, mkdirSync } from 'fs';
import { join } from 'path';
import { PresentationTheme, getCurrentTheme } from '../presentation-theme';

export interface SlideDefinition {
  name: string;
  component: React.ReactElement;
  skip?: boolean;
}

export interface SlideDeckOptions {
  outputDir?: string; // Made optional - if not provided, just returns SVGs
  width?: number;
  height?: number;
  backgroundColor?: string;
  createHTML?: boolean;
  htmlTitle?: string;
  theme?: PresentationTheme;
}

/**
 * Generate a complete slide deck from an array of slide definitions
 * Returns array of SVG strings. If outputDir is provided, also writes files.
 */
export async function generateSlideDeck(
  slides: SlideDefinition[],
  options: SlideDeckOptions = {}
): Promise<string[]> {
  const theme = options.theme || getCurrentTheme();
  const {
    outputDir,
    width = theme.slideWidth,
    height = theme.slideHeight,
    backgroundColor = theme.background,
    createHTML = true,
    htmlTitle = 'Presentation'
  } = options;

  // Filter out skipped slides
  const activeSlides = slides.filter(slide => !slide.skip);

  console.log(`Generating ${activeSlides.length} slides...\n`);

  // Generate each slide
  const svgs: string[] = [];
  
  for (let i = 0; i < activeSlides.length; i++) {
    const slide = activeSlides[i];
    const svg = await renderToSVG(slide.component, {
      width,
      height,
      backgroundColor,
    });
    
    svgs.push(svg);
    
    if (outputDir) {
      const filename = `${slide.name}.svg`;
      writeFileSync(join(outputDir, filename), svg);
      console.log(`✓ Generated ${filename} (${i + 1}/${activeSlides.length})`);
    }
  }

  // Generate HTML viewer if requested and outputDir provided
  if (outputDir) {
    mkdirSync(outputDir, { recursive: true });
    
    if (createHTML) {
      const htmlContent = createHTMLViewer(activeSlides, htmlTitle);
      writeFileSync(join(outputDir, 'index.html'), htmlContent);
      console.log('\n✓ Generated index.html viewer');
    }

    console.log(`\nPresentation complete! ${activeSlides.length} slides generated.`);
    if (createHTML) {
      console.log(`Open ${join(outputDir, 'index.html')} to view.`);
    }
  }
  
  return svgs;
}

/**
 * Create an HTML viewer for the slide deck
 */
function createHTMLViewer(slides: SlideDefinition[], title: string): string {
  return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>${title}</title>
  <style>
    * {
      box-sizing: border-box;
    }
    body {
      margin: 0;
      padding: 20px;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
      background: #f5f5f5;
    }
    .container {
      max-width: 1240px;
      margin: 0 auto;
      background: white;
      padding: 20px;
      border-radius: 8px;
      box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    }
    h1 {
      text-align: center;
      color: #333;
      margin-bottom: 30px;
    }
    .slide {
      margin-bottom: 40px;
      border: 1px solid #ddd;
      border-radius: 4px;
      overflow: hidden;
      display: none;
    }
    .slide.active {
      display: block;
    }
    .slide img {
      width: 100%;
      height: auto;
      display: block;
    }
    .controls {
      text-align: center;
      margin-top: 20px;
      position: sticky;
      top: 20px;
      background: white;
      padding: 15px;
      border-radius: 4px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.1);
      z-index: 100;
    }
    button {
      margin: 0 10px;
      padding: 10px 20px;
      font-size: 16px;
      cursor: pointer;
      background: #1976d2;
      color: white;
      border: none;
      border-radius: 4px;
      transition: background 0.2s;
    }
    button:hover:not(:disabled) {
      background: #1565c0;
    }
    button:disabled {
      background: #ccc;
      cursor: not-allowed;
    }
    .slide-number {
      display: inline-block;
      margin: 0 20px;
      font-size: 18px;
      font-weight: bold;
    }
    .keyboard-hint {
      margin-top: 10px;
      font-size: 12px;
      color: #666;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>${title}</h1>
    
    <div class="controls">
      <button id="prevBtn" onclick="previousSlide()">← Previous</button>
      <span class="slide-number" id="slideNumber">Slide 1 of ${slides.length}</span>
      <button id="nextBtn" onclick="nextSlide()">Next →</button>
      <div class="keyboard-hint">
        Use arrow keys or space to navigate
      </div>
    </div>

    <div id="slideContainer">
      ${slides.map((slide, i) => `
      <div class="slide ${i === 0 ? 'active' : ''}" id="slide-${i}">
        <img src="${slide.name}.svg" alt="Slide ${i + 1}">
      </div>
      `).join('')}
    </div>
  </div>

  <script>
    let currentSlide = 0;
    const totalSlides = ${slides.length};

    function updateButtons() {
      document.getElementById('prevBtn').disabled = currentSlide === 0;
      document.getElementById('nextBtn').disabled = currentSlide === totalSlides - 1;
    }

    function showSlide(n) {
      document.querySelectorAll('.slide').forEach(slide => {
        slide.classList.remove('active');
      });
      
      currentSlide = Math.max(0, Math.min(n, totalSlides - 1));
      document.getElementById('slide-' + currentSlide).classList.add('active');
      document.getElementById('slideNumber').textContent = 
        'Slide ' + (currentSlide + 1) + ' of ' + totalSlides;
      
      updateButtons();
      
      // Scroll to top
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }

    function nextSlide() {
      if (currentSlide < totalSlides - 1) {
        showSlide(currentSlide + 1);
      }
    }

    function previousSlide() {
      if (currentSlide > 0) {
        showSlide(currentSlide - 1);
      }
    }

    document.addEventListener('keydown', (e) => {
      if (e.key === 'ArrowRight' || e.key === ' ') {
        e.preventDefault();
        nextSlide();
      } else if (e.key === 'ArrowLeft') {
        e.preventDefault();
        previousSlide();
      } else if (e.key === 'Home') {
        e.preventDefault();
        showSlide(0);
      } else if (e.key === 'End') {
        e.preventDefault();
        showSlide(totalSlides - 1);
      }
    });

    // Initialize button states
    updateButtons();
  </script>
</body>
</html>`;
}

/**
 * Helper to number slide names automatically
 */
export function numberSlides(
  slides: Array<{ name: string; component: React.ReactElement; skip?: boolean }>
): SlideDefinition[] {
  let slideNumber = 1;
  return slides.map(slide => {
    if (slide.skip) {
      return slide;
    }
    const numberedName = `${slideNumber.toString().padStart(2, '0')}-${slide.name}`;
    slideNumber++;
    return {
      ...slide,
      name: numberedName
    };
  });
}
