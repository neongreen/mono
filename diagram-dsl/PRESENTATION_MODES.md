# Presentation Modes Guide

diagram-dsl now supports three distinct modes for presenting your content, each optimized for different use cases.

## The Three Modes

### 1. 📊 Slides Mode (Traditional)

Classic presentation format with discrete slides. Perfect for presentations with distinct topics and traditional slide navigation.

**Use cases:**
- Conference talks
- Business presentations  
- Educational lectures
- Pitch decks

**How to use:**
```typescript
import { generateSlideDeck, Slide, Title } from 'diagram-dsl';

const slides = [
  { name: 'intro', component: <Slide>...</Slide> },
  { name: 'content', component: <Slide>...</Slide> },
];

const svgs = await generateSlideDeck(slides, {
  width: 1200,
  height: 800,
  theme: darkTheme, // Optional theme
});

// Each SVG is a separate slide
svgs.forEach((svg, i) => {
  writeFileSync(`slide-${i + 1}.svg`, svg);
});
```

**Features:**
- Each slide is a separate SVG file
- Fixed dimensions (1200x800 default)
- Clean separation between topics
- Optional HTML viewer with navigation

---

### 2. 📜 Scrolling Slides Mode

Vertical scrolling with visible slide boundaries and gaps. Great for web-based documentation that needs clear section separation.

**Use cases:**
- Technical documentation
- Long-form articles with sections
- Tutorial series
- Product documentation

**How to use:**
```typescript
import { generateScrollingPage, Slide } from 'diagram-dsl';

const sections = [
  { name: 'intro', component: <Slide>...</Slide> },
  { name: 'usage', component: <Slide>...</Slide> },
];

const svg = await generateScrollingPage(sections, {
  width: 1200,
  slideHeight: 800,
  gap: 60, // Gap between slides
  theme: nordTheme,
});

writeFileSync('document.svg', svg);
```

**Features:**
- Single SVG with all slides
- Visible gaps between sections
- Maintains slide structure
- Perfect for scrolling web pages

---

### 3. 📄 Continuous Page Mode (Pageless)

Seamless content flow without slide boundaries, inspired by Google Docs pageless mode. Content flows naturally from top to bottom.

**Use cases:**
- Research papers
- Long-form technical reports
- Books and manuscripts
- Content that doesn't fit into discrete sections

**How to use:**
```typescript
import { generateContinuousPage, Stack, Title, Text } from 'diagram-dsl';

const blocks = [
  { 
    name: 'intro', 
    component: <Stack>...</Stack>,
    spacing: 40 // Optional spacing after this block
  },
  { 
    name: 'content', 
    component: <Stack>...</Stack> 
  },
];

const svg = await generateContinuousPage(blocks, {
  width: 1200,
  gap: 40, // Smaller gap between blocks
  padding: 60,
  theme: githubTheme,
});

writeFileSync('document.svg', svg);
```

**Features:**
- No visible boundaries between content
- Seamless vertical flow
- Variable spacing between blocks
- Ideal for narrative content

---

## Theme System

Choose from 10 built-in themes or create your own!

### Available Themes

1. **Default** - Clean, professional blue theme
2. **Dark** - Modern dark mode with high contrast
3. **Professional** - Muted, business-appropriate colors
4. **Minimal** - Black and white, zero borders
5. **Vibrant** - Bright, energetic colors
6. **Nord** - Popular Nord color palette
7. **Dracula** - Beloved Dracula theme
8. **Solarized Light** - Easy-on-eyes Solarized
9. **Solarized Dark** - Dark version of Solarized
10. **GitHub** - GitHub's design system colors

### Using Themes

**Option 1: Set globally**
```typescript
import { setCurrentTheme, darkTheme } from 'diagram-dsl';

setCurrentTheme(darkTheme);
// All subsequent renders use this theme
```

**Option 2: Pass to generation functions**
```typescript
await generateSlideDeck(slides, {
  theme: nordTheme,
  // ... other options
});
```

**Option 3: Create custom theme**
```typescript
import { createCustomTheme, defaultTheme } from 'diagram-dsl';

const myTheme = createCustomTheme({
  primary: '#ff6b6b',
  secondary: '#4ecdc4',
  background: '#f7f7f7',
  text: '#2c3e50',
});

setCurrentTheme(myTheme);
```

### Theme Properties

Each theme can customize:

- **Colors:** primary, secondary, accent, success, warning, danger, info
- **Text colors:** text, textSecondary, textMuted
- **Backgrounds:** background, backgroundSecondary
- **Typography:** fontFamily, fontFamilyMono, sizes, lineHeight
- **Spacing:** slideWidth, slideHeight, padding, gaps
- **Borders:** borderRadius, borderWidth
- **Shadows:** shadowLight, shadowMedium, shadowHeavy

---

## Image Support

Embed images in your presentations (currently shows placeholders, full rendering coming soon).

```typescript
import { Image } from 'diagram-dsl';

<Image
  src="https://example.com/diagram.png"
  alt="Architecture diagram"
  width={600}
  height={400}
  borderRadius={8}
  fit="contain" // or 'cover', 'fill', 'none'
/>
```

**Image component features:**
- Support for URLs and data URLs
- Configurable dimensions and fit modes
- Border radius control
- Alt text for accessibility
- Currently renders placeholders (full image support in progress)

---

## Comparison: Which Mode to Choose?

| Feature | Slides | Scrolling Slides | Continuous Page |
|---------|--------|------------------|-----------------|
| **Best for** | Presentations | Documentation | Reports |
| **Navigation** | Slide-by-slide | Vertical scroll | Vertical scroll |
| **Boundaries** | Clear slides | Visible gaps | No boundaries |
| **File output** | Multiple SVGs | Single SVG | Single SVG |
| **Content flow** | Discrete | Sectioned | Seamless |
| **Use HTML viewer** | ✅ Yes | ✅ Yes | ❌ Not needed |

---

## Complete Example

See `src/examples/all-presentation-modes.tsx` for a comprehensive demonstration of all three modes with multiple themes.

Run it:
```bash
cd diagram-dsl
npx tsx src/examples/all-presentation-modes.tsx
```

This generates examples in all three modes using 4 different themes, creating:
- 28 individual slide files (7 slides × 4 themes)
- 4 scrolling page files (1 per theme)
- 4 continuous page files (1 per theme)

---

## Advanced Usage

### Mixing Content Types

All modes support the full range of diagram-dsl components:

```typescript
<Slide>
  <Title level={1}>System Architecture</Title>
  
  <DataFlow nodes={...} connections={...} />
  
  <Image src="photo.png" width={400} />
  
  <CodeBlock language="Python" code={...} />
  
  <ComparisonTable columns={...} rows={...} />
</Slide>
```

### Content Blocks vs Slides

- **Slides mode & Scrolling slides:** Use `<Slide>` components
- **Continuous page:** Can use any component (`<Stack>`, `<Box>`, etc.)

### Spacing Control

```typescript
// Continuous page - fine-grained control
const blocks = [
  { component: <Title>...</Title>, spacing: 20 },
  { component: <Text>...</Text>, spacing: 40 },
  { component: <Image>...</Image>, spacing: 60 },
];

// Scrolling slides - consistent gaps
await generateScrollingPage(sections, {
  gap: 80, // Larger gaps for clear separation
});
```

---

## Tips & Best Practices

1. **Choose the right mode** based on your content structure and delivery method
2. **Use themes consistently** - set once at the beginning
3. **Test with multiple themes** to ensure your content works across color schemes
4. **Consider accessibility** - use high contrast themes when needed
5. **Mix and match components** - combine diagrams, code, text, and images freely
6. **Keep slide content focused** in slides mode (one main point per slide)
7. **Use continuous mode** when your narrative doesn't fit into discrete sections
8. **Leverage spacing** to create visual rhythm and improve readability



## Migration from Older Versions

If you're using older versions of the slide helpers:

**Before:**
```typescript
await generateSlideDeck(slides, {
  outputDir: './output', // Required
});
```

**Now:**
```typescript
// Get SVGs as strings
const svgs = await generateSlideDeck(slides, {
  // outputDir is optional
});

// Or still write to files
await generateSlideDeck(slides, {
  outputDir: './output', // Still works!
});
```

The new helpers return SVG strings for maximum flexibility while maintaining backward compatibility when `outputDir` is provided.
