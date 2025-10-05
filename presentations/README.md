# Presentations

This directory contains presentations built using diagram-dsl.

## Available Presentations

### LLM Context Management

A comprehensive presentation about context management strategies in LLM agent implementations.

**Location:** `llm-context-management/`

**Topics covered:**
- The Context Challenge
- Sliding Window strategy
- Hierarchical Summarization
- Vector Memory (RAG)
- Hybrid Approach
- Practical Considerations
- Summary & Recommendations

**Versions:**
- `src/presentation.tsx` - Original version (778 lines)
- `src/presentation-v2.tsx` - Refactored with new components (684 lines, -12%)

**To generate:**
```bash
cd llm-context-management
pnpm generate
```

**To view:**
```bash
open output/index.html
# or for v2:
open output-v2/index.html
```

## Creating a New Presentation

1. Create a new directory under `presentations/`
2. Initialize with package.json:
```json
{
  "name": "your-presentation-name",
  "version": "1.0.0",
  "description": "Your presentation description",
  "type": "module",
  "scripts": {
    "generate": "tsx src/presentation.tsx"
  },
  "dependencies": {
    "diagram-dsl": "workspace:*",
    "react": "^19.2.0",
    "react-dom": "^19.2.0"
  },
  "devDependencies": {
    "@types/node": "^24.6.2",
    "@types/react": "^19.2.0",
    "tsx": "^4.20.6",
    "typescript": "^5.9.3"
  }
}
```

3. Create `src/presentation.tsx`:
```tsx
import React from 'react';
import { 
  Slide, Title, Subtitle, Section, List, ProsCons, Callout,
  renderToSVG 
} from 'diagram-dsl';
import { writeFileSync, mkdirSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const outputDir = join(__dirname, '../output');

// Define your slides
const Slide1 = () => (
  <Slide>
    <Title level={1}>Your Title</Title>
    {/* your content */}
  </Slide>
);

// Generate function
async function generatePresentation() {
  mkdirSync(outputDir, { recursive: true });

  const slides = [
    { name: '01-slide', component: <Slide1 /> },
    // add more slides
  ];

  for (const slide of slides) {
    const svg = await renderToSVG(slide.component, {
      width: 1200,
      height: 800,
      backgroundColor: 'white',
    });
    
    writeFileSync(join(outputDir, `${slide.name}.svg`), svg);
    console.log(`✓ Generated ${slide.name}.svg`);
  }
}

generatePresentation();
```

4. Install dependencies:
```bash
cd your-presentation-name
pnpm install
```

5. Generate:
```bash
pnpm generate
```

## Component Guide

See [diagram-dsl/PRESENTATION_COMPONENTS.md](../diagram-dsl/PRESENTATION_COMPONENTS.md) for a comprehensive guide to all presentation components.

### Quick Reference

**Layout:**
- `Slide` - Standard slide container (1200x800)
- `Grid` - Multi-column layouts
- `Spacer` - Flexible spacing

**Content:**
- `Section` - Titled content sections
- `List` - Bullet point lists
- `ProsCons` - Side-by-side pros/cons
- `Callout` - Highlighted important info
- `RichText` - Mixed text formatting

**Typography:**
- `Title` - Main headings (level 1-3)
- `Subtitle` - Secondary headings
- `Label` - Small labels
- `Text` - Body text

**Containers:**
- `Card` - Styled boxes with variants
- `Box`, `Stack`, `Row` - Basic layout

## Tips

1. **Start with the refactored example** (`llm-context-management/src/presentation-v2.tsx`) as a template
2. **Use components** instead of manual layout for consistency
3. **Define slides as separate components** for better organization
4. **Keep slide count manageable** (8-12 slides is ideal)
5. **Use variants** for consistent theming
6. **Test individual slides** before generating the full deck

## HTML Viewer Template

Generate an HTML viewer for your presentation:

```javascript
const htmlContent = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Your Presentation Title</title>
  <style>
    /* Copy styles from llm-context-management example */
  </style>
</head>
<body>
  <div class="container">
    <h1>Your Presentation Title</h1>
    <div class="controls">
      <button onclick="previousSlide()">← Previous</button>
      <span class="slide-number" id="slideNumber">Slide 1 of N</span>
      <button onclick="nextSlide()">Next →</button>
    </div>
    <div id="slideContainer">
      ${slides.map((slide, i) => \`
      <div class="slide" id="slide-\${i}" style="display: \${i === 0 ? 'block' : 'none'}">
        <img src="\${slide.name}.svg" alt="Slide \${i + 1}">
      </div>
      \`).join('')}
    </div>
  </div>
  <script>/* Copy script from example */</script>
</body>
</html>`;

writeFileSync(join(outputDir, 'index.html'), htmlContent);
```

## Publishing Presentations

To share your presentations:

1. **SVG files** - Individual slides can be embedded in documents
2. **HTML viewer** - Self-contained viewer works in any browser
3. **PDF export** - Print from browser or use tools like wkhtmltopdf
4. **Image export** - Convert SVGs to PNG for embedding

## Examples

See `llm-context-management/src/presentation-v2.tsx` for a complete example showcasing all components and best practices.
