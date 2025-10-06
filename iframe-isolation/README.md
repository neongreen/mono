# Iframe Isolation Demos

This folder contains research and practical demonstrations of techniques to protect parent pages from memory issues and crashes in embedded iframes.

## Contents

- **RESEARCH.md** - Comprehensive research report on iframe isolation techniques and industry practices
- **INDUSTRY-OVERVIEW.md** - Compact overview of production systems (CodeSandbox, Figma, Observable) optimized for printing
- **IMPLEMENTATION-PATTERNS.md** - Code examples and patterns from production systems
- **REFERENCES.md** - Complete list of links, documentation, and resources
- **demos/** - Working HTML demonstrations of various isolation approaches

## The Problem

When you embed an iframe in a web page, buggy or poorly written code inside the iframe can cause:
- Memory leaks that affect the entire browser tab
- Performance issues and sluggishness in the parent page  
- Browser freezes
- Complete tab crashes

This happens because iframes often share the same browser process and memory space as the parent page.

## Quick Start

1. Open the demos in a web browser
2. Each demo is self-contained and can be opened directly
3. Use browser DevTools to monitor memory and performance

## Demos

### Demo 1: Problem Demonstration
**File**: `demos/01-problem.html`

Shows the actual problem - an iframe with a memory leak that impacts the parent page.
- Infinite loop creating objects
- Watch memory grow in DevTools
- Parent page becomes sluggish

### Demo 2: Sandbox Attributes
**File**: `demos/02-sandbox.html`

Demonstrates iframe `sandbox` attribute (provides limited protection).
- Various sandbox flags
- Shows what sandbox CAN and CANNOT prevent
- Conclusion: Sandbox alone is insufficient for memory isolation

### Demo 3: Cross-Origin Isolation (Recommended)
**File**: `demos/03-cross-origin.html`

The industry-standard solution using cross-origin iframes.
- Separate subdomain for iframe content
- True process isolation
- Complete protection from memory issues
- **Note**: Requires web server with multiple origins to test properly

### Demo 4: COOP/COEP Headers
**File**: `demos/04-coop-coep.html`

Advanced isolation using Cross-Origin-Opener-Policy and Cross-Origin-Embedder-Policy headers.
- Stricter isolation boundaries
- Required for SharedArrayBuffer
- **Note**: Requires server configuration

### Demo 5: Web Workers
**File**: `demos/05-web-workers.html`

Moving heavy computation to Web Workers to protect the main thread.
- Separate JavaScript heap
- Non-blocking computation
- Still shares process space

### Demo 6: Memory Monitoring
**File**: `demos/06-memory-monitor.html`

Practical implementation of memory monitoring and automatic iframe restart.
- Detects high memory usage
- Automatic recovery
- Chrome-only (uses `performance.memory`)

## Key Findings

### ✅ What Works

1. **Cross-origin iframes** (separate subdomain)
   - True process isolation in modern browsers
   - Best protection against memory leaks and crashes
   - Used by: CodeSandbox, Figma, StackBlitz

2. **Memory monitoring + automatic restart**
   - Reactive but effective
   - Requires Chrome's non-standard `performance.memory`
   - Good backup solution

3. **Web Workers for computation**
   - Protects main thread
   - Separate heap
   - Good for CPU-intensive tasks

### ❌ What Doesn't Work

1. **Sandbox attribute alone**
   - Does NOT provide process isolation
   - Does NOT prevent memory leaks
   - Only restricts capabilities

2. **Same-origin iframes**
   - Always share process space
   - No isolation possible

## Recommended Solution

For production applications, use a **multi-layered approach**:

```html
<!-- Cross-origin iframe (MOST IMPORTANT) -->
<iframe 
  src="https://sandbox.yourdomain.com/preview"
  sandbox="allow-scripts"
  style="width: 100%; height: 600px;">
</iframe>

<script>
// Add memory monitoring as backup
setInterval(() => {
  if (performance.memory) {
    const usage = (performance.memory.usedJSHeapSize / 
                  performance.memory.jsHeapSizeLimit) * 100;
    if (usage > 85) {
      restartIframe();
    }
  }
}, 5000);
</script>
```

## Testing the Demos

### Local Testing (Single Origin)
Most demos work when opened directly in browser:
```bash
cd demos
# Open any demo file in your browser
open 01-problem.html
```

### Full Testing (Cross-Origin)
For demos requiring multiple origins (Demo 3, 4), set up a local server:

```bash
# Using Python
python3 -m http.server 8000

# Then edit your /etc/hosts to add:
# 127.0.0.1 sandbox.localhost
# 127.0.0.1 localhost

# Access at:
# http://localhost:8000/demos/03-cross-origin.html
```

Or use a cloud deployment for true cross-origin testing.

## Browser DevTools Tips

### Monitor Memory
1. Open DevTools (F12)
2. Go to Performance tab
3. Click "Record" 
4. Interact with page
5. Stop recording and analyze memory usage

Or use Memory tab:
1. Take heap snapshot
2. Interact with iframe
3. Take another snapshot
4. Compare to find leaks

### Monitor Process Isolation
Chrome Task Manager (Shift+Esc):
- Shows separate processes
- Cross-origin iframes appear as separate processes
- Monitor memory per process

## Industry Examples

- **CodeSandbox**: `codesandbox.io` → `xyz.csb.app` (cross-origin)
- **Figma**: `figma.com` → `preview.figma.com` (cross-origin)
- **StackBlitz**: `stackblitz.com` + WebContainers (cross-origin + COOP/COEP)
- **Webflow**: `webflow.com` → `preview.webflow.io` (cross-origin)

## Further Reading

**For Technical Analysis**:
- **RESEARCH.md** - Detailed technical analysis, implementation guidelines, browser compatibility

**For Production Insights** (optimized for printing):
- **INDUSTRY-OVERVIEW.md** - How CodeSandbox, Figma, Observable handle iframe isolation in production
- **IMPLEMENTATION-PATTERNS.md** - Complete code examples and patterns
- **REFERENCES.md** - Links to documentation, GitHub repos, blog posts, conference talks

## License

This research and demos are part of the neongreen/mono repository.
