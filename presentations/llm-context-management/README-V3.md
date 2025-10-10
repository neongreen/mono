# LLM Context Management Presentation - Version 3

This comprehensive presentation demonstrates all three presentation modes available in diagram-dsl, showcasing best practices for creating technical content for software engineers.

## What's Inside

### 9 Information-Rich Slides

1. **Title Slide** - Introduction with topics overview
2. **The Problem** - Context challenges and constraints
3. **Strategy Overview** - Four core approaches
4. **Sliding Window** - Simple recent-message retention
5. **Summarization** - Compression strategies
6. **RAG** - Retrieval Augmented Generation
7. **Hybrid Approach** - Combining multiple strategies
8. **Implementation** - Practical considerations
9. **Summary** - Key takeaways

### Advanced Components Used

- **DataFlow** - System architecture diagrams
- **SequenceDiagram** - Actor interactions and message flows
- **ComparisonTable** - Before/after comparisons
- **CodeBlock** - Python implementation examples
- **Steps** - Sequential processes
- **Badge, Card, Section** - Content organization
- **Highlight, Well, Callout** - Emphasis and notes

### Three Presentation Modes

#### Mode 1: Traditional Slides
- 9 separate SVG files per theme
- Perfect for classic presentations
- Files: `{theme}-slide-01.svg` through `{theme}-slide-09.svg`

#### Mode 2: Scrolling Slides
- Single SVG with all slides and gaps
- Vertical scrolling with slide boundaries
- Files: `{theme}-scrolling.svg`

#### Mode 3: Continuous Page
- Seamless vertical flow without gaps
- Like Google Docs pageless mode
- Files: `{theme}-continuous.svg`

### Four Themes Demonstrated

- **Default** - Professional blue theme
- **Dark** - Modern dark mode
- **Nord** - Popular developer theme
- **Professional** - Muted business colors

## Quick Start

```bash
# Install dependencies (from monorepo root)
pnpm install

# Build the library
cd diagram-dsl
pnpm build

# Generate all presentations
cd ../presentations/llm-context-management
npx tsx src/presentation-v3-all-modes.tsx
```

## Output

Generated files go to `output-v3/` directory:

```
output-v3/
├── default-slide-01.svg ... default-slide-09.svg    (9 files)
├── default-scrolling.svg
├── default-continuous.svg
├── dark-slide-01.svg ... dark-slide-09.svg          (9 files)
├── dark-scrolling.svg
├── dark-continuous.svg
├── nord-slide-01.svg ... nord-slide-09.svg          (9 files)
├── nord-scrolling.svg
├── nord-continuous.svg
├── professional-slide-01.svg ... professional-slide-09.svg  (9 files)
├── professional-scrolling.svg
└── professional-continuous.svg

Total: 44 files (36 slides + 4 scrolling + 4 continuous)
```

## Key Features Demonstrated

### Content Organization
- Clear hierarchy with titles and subtitles
- Sectioned content for easy navigation
- Color-coded cards for visual distinction
- Progressive disclosure of information

### Technical Accuracy
- Real implementation examples
- Accurate system diagrams
- Proper technical terminology
- Practical considerations included

### Visual Design
- Consistent spacing and alignment
- Strategic use of color for emphasis
- Readable code blocks with line numbers
- Clear data flow visualizations

### Accessibility
- Alt text for images (where applicable)
- Sufficient color contrast
- Clear hierarchy
- Readable font sizes

## Learning From This Example

This presentation serves as a template for creating technical content. Key patterns:

1. **Start with a clear overview** - Tell them what you'll tell them
2. **Use diagrams for complex concepts** - DataFlow for architecture, SequenceDiagram for interactions
3. **Show code when relevant** - Real implementation examples
4. **Compare and contrast** - Use tables for direct comparisons
5. **Provide actionable takeaways** - Summarize key points

## Customization

### Change Themes
```typescript
const themes = [
  { name: 'dark', theme: darkTheme },
  { name: 'solarized', theme: solarizedLightTheme },
  // Add more themes
];
```

### Adjust Spacing
```typescript
// Scrolling mode
{ gap: 60 }  // Gap between slides

// Continuous mode
{ gap: 40 }  // Smaller gap for seamless flow
```

### Add More Slides
```typescript
const slides = [
  // ... existing slides
  { 
    name: '10-new-topic', 
    component: <NewTopicSlide /> 
  },
];
```

## Comparison: v1 vs v2 vs v3

| Feature | v1 | v2 | v3 |
|---------|----|----|-----|
| Slides | ✅ Basic | ✅ Enhanced | ✅ Advanced |
| Components | Basic | Improved | Full suite |
| Themes | 1 (default) | 2 (default, dark) | 4 themes |
| Modes | 1 (slides) | 1 (slides) | 3 (slides, scrolling, continuous) |
| Diagrams | Text-based | Basic | Advanced |
| Code blocks | ❌ | ✅ | ✅ Enhanced |
| Tables | ❌ | ❌ | ✅ |
| Sequence diagrams | ❌ | ❌ | ✅ |

## Tips for Technical Presentations

1. **Keep code snippets short** - 10-15 lines max per block
2. **Use diagrams liberally** - A picture is worth a thousand words
3. **Show, don't just tell** - Include real examples
4. **Layer information** - Start simple, add complexity
5. **Provide context** - Why does this matter?
6. **Include practical advice** - What should developers do?

## Performance

- **Generation time:** ~10-15 seconds for all 44 files
- **File sizes:** 2-5 KB per slide, 20-25 KB for scrolling/continuous
- **Memory usage:** Minimal - handles large presentations easily

## Next Steps

1. View the generated SVGs in your browser
2. Experiment with different themes
3. Try different spacing configurations
4. Add your own slides
5. Create your own presentation from this template

## Related Documentation

- `../../diagram-dsl/PRESENTATION_MODES.md` - Complete guide to presentation modes
- `../../diagram-dsl/PRESENTATION_EVOLUTION.md` - Technical evolution details
- `../../FINAL_ENHANCEMENT_SUMMARY.md` - Comprehensive feature summary

## Questions?

This example demonstrates the full power of diagram-dsl for creating technical presentations. Use it as a starting point for your own presentations, documentation, or reports.

The combination of three modes, multiple themes, and rich technical components makes it easy to create professional content that can be version-controlled, programmatically generated, and easily updated.
