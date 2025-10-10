import { renderToSVG } from '../renderer';
import { writeFileSync, mkdirSync } from 'fs';
import { join } from 'path';
import type { ReactElement } from 'react';
import { PresentationTheme, getCurrentTheme } from '../presentation-theme';

export interface PageSection {
  name: string;
  component: ReactElement;
  backgroundColor?: string;
}

export interface ScrollingPageOptions {
  outputDir?: string; // Made optional
  htmlTitle?: string;
  width?: number;
  slideHeight?: number;
  gap?: number; // Renamed from sectionGap for consistency
  backgroundColor?: string;
  createHTML?: boolean;
  theme?: PresentationTheme;
}

/**
 * Generate a single scrolling page document from multiple sections
 * Perfect for technical documentation, long-form content, and reports
 * Returns SVG string. If outputDir provided, also writes files.
 */
export async function generateScrollingPage(
  sections: PageSection[],
  options: ScrollingPageOptions = {}
): Promise<string> {
  const theme = options.theme || getCurrentTheme();
  const {
    outputDir,
    htmlTitle = 'Technical Documentation',
    width = theme.slideWidth,
    slideHeight = theme.slideHeight,
    gap = 60,
    backgroundColor = theme.background,
    createHTML = true
  } = options;

  console.log(`\nGenerating scrolling page with ${sections.length} sections...\n`);

  // Render each section to SVG
  const renderedSections: Array<{ name: string; svg: string; height: number }> = [];

  for (let i = 0; i < sections.length; i++) {
    const section = sections[i];
    
    // Render the section
    const result = await renderToSVG(section.component, {
      width,
      height: slideHeight,
      backgroundColor: section.backgroundColor || 'none',
    });
    
    // Use fixed slide height
    const height = slideHeight;
    
    renderedSections.push({
      name: section.name,
      svg: result,
      height
    });
    
    console.log(`✓ Rendered section: ${section.name} (${i + 1}/${sections.length})`);
  }

  // Combine into one scrolling SVG
  const totalHeight = renderedSections.reduce((sum, s) => sum + s.height + gap, 0) + gap;
  
  let combinedSVG = `<svg width="${width}" height="${totalHeight}" xmlns="http://www.w3.org/2000/svg">`;
  combinedSVG += `<rect width="${width}" height="${totalHeight}" fill="${backgroundColor}"/>`;
  
  let currentY = gap;
  for (const section of renderedSections) {
    const contentMatch = section.svg.match(/<svg[^>]*>([\s\S]*)<\/svg>/);
    const content = contentMatch ? contentMatch[1] : section.svg;
    
    combinedSVG += `<g transform="translate(0, ${currentY})">`;
    combinedSVG += content;
    combinedSVG += `</g>`;
    
    currentY += section.height + gap;
  }
  
  combinedSVG += `</svg>`;

  // Write files if outputDir provided
  if (outputDir) {
    mkdirSync(outputDir, { recursive: true });
    
    // Save individual SVGs
    renderedSections.forEach(({ name, svg }) => {
      const filename = `section-${name}.svg`;
      writeFileSync(join(outputDir, filename), svg);
    });

    if (createHTML) {
      // Generate HTML with all sections
      const html = generateScrollingHTML(renderedSections, {
        title: htmlTitle,
        width,
        sectionGap: gap
      });
      
      const htmlPath = join(outputDir, 'index.html');
      writeFileSync(htmlPath, html);
      console.log('\n✓ Generated index.html viewer');
    }

    console.log(`\nScrolling page complete! ${sections.length} sections generated.`);
    console.log(`Open ${join(outputDir, 'index.html')} to view.\n`);
  }
  
  return combinedSVG;
}

function generateScrollingHTML(
  sections: Array<{ name: string; svg: string; height: number }>,
  options: { title: string; width: number; sectionGap: number }
): string {
  const { title, width, sectionGap } = options;
  
  const sectionsHTML = sections.map((section, index) => {
    // Inline the SVG content
    const svgContent = section.svg
      .replace('<?xml version="1.0" encoding="UTF-8"?>', '')
      .trim();
    
    return `
    <section id="section-${section.name}" class="content-section" style="margin-bottom: ${sectionGap}px;">
      <div class="section-content">
        ${svgContent}
      </div>
    </section>`;
  }).join('\n');

  // Generate table of contents
  const tocHTML = sections.map(section => `
    <li><a href="#section-${section.name}">${section.name.replace(/-/g, ' ')}</a></li>
  `).join('\n');

  return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>${title}</title>
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      background: #f5f5f5;
      color: #333;
    }
    
    .container {
      display: flex;
      min-height: 100vh;
    }
    
    .sidebar {
      width: 250px;
      background: white;
      border-right: 1px solid #e0e0e0;
      padding: 24px;
      position: fixed;
      height: 100vh;
      overflow-y: auto;
      box-shadow: 2px 0 8px rgba(0,0,0,0.05);
    }
    
    .sidebar h1 {
      font-size: 20px;
      margin-bottom: 24px;
      color: #1976d2;
      font-weight: 600;
    }
    
    .toc {
      list-style: none;
    }
    
    .toc li {
      margin-bottom: 12px;
    }
    
    .toc a {
      color: #666;
      text-decoration: none;
      font-size: 14px;
      transition: color 0.2s;
      text-transform: capitalize;
      display: block;
      padding: 6px 12px;
      border-radius: 4px;
      transition: all 0.2s;
    }
    
    .toc a:hover {
      color: #1976d2;
      background: #f0f7ff;
    }
    
    .toc a.active {
      color: #1976d2;
      background: #e3f2fd;
      font-weight: 600;
    }
    
    .main-content {
      margin-left: 250px;
      flex: 1;
      padding: 40px;
      max-width: ${width + 80}px;
    }
    
    .content-section {
      background: white;
      border-radius: 8px;
      box-shadow: 0 2px 8px rgba(0,0,0,0.1);
      overflow: hidden;
    }
    
    .section-content {
      display: flex;
      justify-content: center;
      width: 100%;
    }
    
    .section-content svg {
      max-width: 100%;
      height: auto;
      display: block;
    }
    
    .back-to-top {
      position: fixed;
      bottom: 24px;
      right: 24px;
      background: #1976d2;
      color: white;
      border: none;
      border-radius: 50%;
      width: 48px;
      height: 48px;
      font-size: 20px;
      cursor: pointer;
      box-shadow: 0 4px 12px rgba(0,0,0,0.15);
      opacity: 0;
      transition: opacity 0.3s, transform 0.2s;
      z-index: 1000;
    }
    
    .back-to-top.visible {
      opacity: 1;
    }
    
    .back-to-top:hover {
      transform: translateY(-2px);
      box-shadow: 0 6px 16px rgba(0,0,0,0.2);
    }
    
    @media (max-width: 768px) {
      .sidebar {
        display: none;
      }
      
      .main-content {
        margin-left: 0;
        padding: 20px;
      }
    }
    
    @media print {
      .sidebar,
      .back-to-top {
        display: none;
      }
      
      .main-content {
        margin-left: 0;
      }
      
      .content-section {
        page-break-inside: avoid;
        box-shadow: none;
      }
    }
  </style>
</head>
<body>
  <div class="container">
    <aside class="sidebar">
      <h1>${title}</h1>
      <ul class="toc">
        ${tocHTML}
      </ul>
    </aside>
    
    <main class="main-content">
      ${sectionsHTML}
    </main>
  </div>
  
  <button class="back-to-top" id="backToTop" title="Back to top">↑</button>
  
  <script>
    // Smooth scrolling for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
      anchor.addEventListener('click', function (e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute('href'));
        if (target) {
          target.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
      });
    });
    
    // Back to top button
    const backToTop = document.getElementById('backToTop');
    
    window.addEventListener('scroll', () => {
      if (window.scrollY > 300) {
        backToTop.classList.add('visible');
      } else {
        backToTop.classList.remove('visible');
      }
    });
    
    backToTop.addEventListener('click', () => {
      window.scrollTo({ top: 0, behavior: 'smooth' });
    });
    
    // Active section highlighting
    const sections = document.querySelectorAll('.content-section');
    const tocLinks = document.querySelectorAll('.toc a');
    
    function updateActiveSection() {
      const scrollPos = window.scrollY + 100;
      
      sections.forEach((section, index) => {
        const top = section.offsetTop;
        const bottom = top + section.offsetHeight;
        
        if (scrollPos >= top && scrollPos < bottom) {
          tocLinks.forEach(link => link.classList.remove('active'));
          tocLinks[index]?.classList.add('active');
        }
      });
    }
    
    window.addEventListener('scroll', updateActiveSection);
    updateActiveSection();
    
    // Keyboard shortcuts
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Home') {
        e.preventDefault();
        window.scrollTo({ top: 0, behavior: 'smooth' });
      } else if (e.key === 'End') {
        e.preventDefault();
        window.scrollTo({ top: document.body.scrollHeight, behavior: 'smooth' });
      }
    });
  </script>
</body>
</html>`;
}
