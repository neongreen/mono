# Iframe Isolation Research: Protecting Parent Pages from Iframe Memory Issues

## Problem Statement

When embedding an iframe in a web page, poorly written or buggy code within the iframe can cause:
- High memory usage that affects the entire tab
- Performance degradation and sluggishness in the parent page
- Browser tab freezes
- Complete tab crashes

This is particularly problematic for:
- Website builders (Wix, Webflow, Squarespace)
- AI agents running user code
- Code playgrounds (CodeSandbox, StackBlitz)
- Preview environments (Storybook, design tools)

## Root Cause

Despite the iframe being a separate browsing context, it still shares:
1. **Process space** - Most browsers use a single process for same-origin content
2. **Memory heap** - JavaScript heap is shared within the same process
3. **Event loop** - Main thread is shared, blocking operations affect everything
4. **Browser tab resources** - GPU, compositing, rendering pipeline

## Industry Solutions

### 1. Process Isolation (Most Effective)

#### Site Isolation (Chrome/Chromium)
- **How it works**: Each cross-origin iframe gets its own renderer process
- **Requirements**: 
  - Iframe must be cross-origin (different domain)
  - Enabled by default in Chrome 67+
- **Effectiveness**: ★★★★★
- **Used by**: Google Docs, Gmail, Figma

**Implementation:**
```html
<!-- Parent: https://example.com -->
<iframe src="https://sandbox.example.com/preview"></iframe>
```

Key points:
- Different subdomain = different origin
- Browser automatically isolates
- No shared memory or process space
- Complete protection from crashes

**Real-world example - Figma:**
```
Main app:      https://www.figma.com
Preview frame: https://preview.figma.com
Plugin iframe: https://plugin-sandbox.figma.com
```

#### Cross-Origin-Opener-Policy (COOP) + Cross-Origin-Embedder-Policy (COEP)
- **How it works**: Forces stricter cross-origin isolation
- **Headers needed**:
  ```
  Cross-Origin-Opener-Policy: same-origin
  Cross-Origin-Embedder-Policy: require-corp
  ```
- **Effectiveness**: ★★★★☆
- **Used by**: Web-based IDEs, advanced web apps

**Real-world example - StackBlitz:**
- Uses COOP/COEP headers on preview domains
- WebContainer technology for full isolation
- Each preview gets isolated browsing context

### 2. Sandbox Attributes (Partial Protection)

The `sandbox` attribute restricts iframe capabilities but doesn't provide process isolation.

```html
<iframe 
  sandbox="allow-scripts allow-same-origin"
  src="...">
</iframe>
```

**Sandbox flags:**
- `allow-scripts` - Allow JavaScript execution
- `allow-same-origin` - Allow same-origin access (use with caution)
- `allow-forms` - Allow form submission
- `allow-popups` - Allow popups
- `allow-modals` - Allow modals (alert, confirm, prompt)

**Effectiveness**: ★★☆☆☆
- Does NOT prevent memory leaks
- Does NOT provide process isolation
- Only restricts capabilities

**Note**: `allow-same-origin` with `allow-scripts` is dangerous - effectively removes sandbox protection.

### 3. Web Workers (Code Execution Isolation)

Move heavy computation to Web Workers to keep main thread responsive.

```javascript
// parent.js
const worker = new Worker('heavy-computation.js');
worker.postMessage({ task: 'processData', data: largeDataset });

worker.onmessage = (e) => {
  // Worker completed without blocking main thread
  updateUI(e.data.result);
};
```

**Effectiveness**: ★★★☆☆
- Protects main thread from blocking
- Separate JavaScript heap
- Still in same process (can affect tab)
- Cannot access DOM directly

**Used by**: 
- CodeMirror (syntax highlighting)
- Monaco Editor (language services)
- Web-based image processors

### 4. Service Workers (Network-level Control)

Service Workers can intercept and modify network requests, enabling:
- Response caching to reduce iframe load
- Request throttling
- Content transformation
- Offline functionality

```javascript
// service-worker.js
self.addEventListener('fetch', (event) => {
  if (event.request.url.includes('/preview/')) {
    // Throttle or cache preview requests
    event.respondWith(
      caches.match(event.request)
        .then(response => response || fetch(event.request))
    );
  }
});
```

**Effectiveness**: ★★☆☆☆
- Does NOT prevent memory issues
- Helps with network-level problems
- Reduces redundant requests

### 5. Monitoring and Killing (Reactive Approach)

**Memory monitoring:**
```javascript
if (performance.memory) {
  const { usedJSHeapSize, totalJSHeapSize, jsHeapSizeLimit } = performance.memory;
  const usage = (usedJSHeapSize / jsHeapSizeLimit) * 100;
  
  if (usage > 80) {
    // Kill and restart iframe
    iframe.src = 'about:blank';
    setTimeout(() => iframe.src = originalSrc, 100);
  }
}
```

**Effectiveness**: ★★★☆☆
- Reactive, not preventive
- Can recover from issues
- May lose user work
- `performance.memory` is non-standard (Chrome only)

**Used by**: 
- CodeSandbox (automatic refresh on high memory)
- Replit (container restart)

### 6. Browser Context Separation (Desktop Apps)

For Electron/desktop apps, use separate BrowserWindow instances.

```javascript
// Electron
const { BrowserWindow } = require('electron');

const previewWindow = new BrowserWindow({
  webPreferences: {
    nodeIntegration: false,
    contextIsolation: true,
    sandbox: true,
    partition: 'preview-partition' // Separate session
  }
});
```

**Effectiveness**: ★★★★★
- Complete isolation
- Separate process guaranteed
- Only for desktop apps

**Used by**: 
- VS Code (webview isolation)
- Electron-based IDEs

## Industry Case Studies

### 1. **CodeSandbox**

**Approach**: Multi-layered isolation
- Cross-origin iframes (`sandbox.codesandbox.io`)
- Service Worker for bundling
- Memory monitoring with automatic refresh
- Web Workers for transpilation

**Architecture:**
```
Editor:          https://codesandbox.io
Preview:         https://xyz123.csb.app  (cross-origin)
Bundler:         Service Worker
Transpiler:      Web Worker
```

### 2. **StackBlitz (WebContainers)**

**Approach**: Revolutionary - entire Node.js runtime in browser
- WebContainer API (proprietary)
- COOP/COEP headers for SharedArrayBuffer
- Cross-origin preview domains
- Process-like isolation in browser

**Key innovation**: Runs Node.js in the browser with full isolation.

### 3. **Figma**

**Approach**: Multi-origin architecture
- Main app: `www.figma.com`
- Preview: `preview.figma.com`
- Plugins: `plugin-sandbox.figma.com`
- Each in separate process due to cross-origin

### 4. **Replit**

**Approach**: Server-side isolation
- Code runs in Docker containers on server
- Browser shows VNC-like view or proxied output
- Complete isolation from browser
- More infrastructure cost

### 5. **Observable**

**Approach**: Notebooks with iframe isolation
- Each cell can run in iframe
- Cross-origin for untrusted code
- Web Workers for computation
- Memory limits and timeouts

### 6. **Webflow/Wix/Squarespace**

**Approach**: Preview in cross-origin iframe
- User's site preview: `preview.webflow.io`
- Editor: `webflow.com`
- Automatic iframe refresh on errors
- Resource limits on preview

## Recommended Strategy (Best Practices)

### Tier 1: Must Have (Process Isolation)
1. **Use cross-origin iframes** for any untrusted or user-generated content
2. **Separate subdomain** for preview/sandbox: `sandbox.yourdomain.com`
3. **Enable COOP/COEP** headers if you need SharedArrayBuffer

### Tier 2: Should Have (Defense in Depth)
4. **Sandbox attributes** with minimal permissions
5. **Web Workers** for heavy computation
6. **Memory monitoring** with automatic recovery
7. **Resource limits** (timeouts, size limits)

### Tier 3: Nice to Have
8. **Service Workers** for caching and throttling
9. **Error boundaries** in React/framework
10. **Graceful degradation** with fallbacks

### Implementation Template

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Isolated Preview</title>
  <!-- Enable COOP/COEP for maximum isolation -->
  <meta http-equiv="Cross-Origin-Opener-Policy" content="same-origin">
  <meta http-equiv="Cross-Origin-Embedder-Policy" content="require-corp">
</head>
<body>
  <div id="app">
    <!-- Cross-origin iframe with sandbox -->
    <iframe 
      id="preview"
      src="https://sandbox.yourdomain.com/preview"
      sandbox="allow-scripts"
      style="width: 100%; height: 600px; border: 1px solid #ccc;">
    </iframe>
  </div>

  <script>
    const iframe = document.getElementById('preview');
    let restartCount = 0;
    const MAX_RESTARTS = 3;

    // Monitor memory (Chrome only)
    setInterval(() => {
      if (performance.memory) {
        const usage = (performance.memory.usedJSHeapSize / 
                      performance.memory.jsHeapSizeLimit) * 100;
        
        if (usage > 85 && restartCount < MAX_RESTARTS) {
          console.warn('High memory usage detected, restarting iframe');
          restartIframe();
        }
      }
    }, 5000);

    function restartIframe() {
      restartCount++;
      const src = iframe.src;
      iframe.src = 'about:blank';
      setTimeout(() => {
        iframe.src = src;
        // Reset counter after successful restart
        setTimeout(() => restartCount = 0, 60000);
      }, 100);
    }

    // Listen for unload events (iframe crash detection)
    iframe.addEventListener('load', () => {
      console.log('Iframe loaded successfully');
      restartCount = Math.max(0, restartCount - 1);
    });
  </script>
</body>
</html>
```

## Technical Limitations

### 1. Same-Origin Limitation
- Same-origin iframes share process space
- No way to force process isolation for same-origin
- **Solution**: Use different subdomains

### 2. Browser Support
- Site Isolation: Chrome 67+, Edge 79+, limited Safari support
- COOP/COEP: Modern browsers only
- SharedArrayBuffer: Requires COOP/COEP
- **Solution**: Feature detection and fallbacks

### 3. Cross-Origin Communication
- `postMessage` required for communication
- No direct DOM access
- More complex data passing
- **Solution**: Design clear message protocols

### 4. SEO and Crawlers
- Cross-origin content may not be indexed
- Crawlers may not execute JavaScript
- **Solution**: Server-side rendering for public content

## Measurement and Monitoring

### Key Metrics to Track

```javascript
// Memory usage (Chrome only)
const memoryInfo = performance.memory;
console.log('Used:', memoryInfo.usedJSHeapSize);
console.log('Total:', memoryInfo.totalJSHeapSize);
console.log('Limit:', memoryInfo.jsHeapSizeLimit);

// Performance timing
const perfData = performance.getEntriesByType('navigation')[0];
console.log('DOM Content Loaded:', perfData.domContentLoadedEventEnd);
console.log('Load Complete:', perfData.loadEventEnd);

// Long tasks (requires PerformanceObserver)
const observer = new PerformanceObserver((list) => {
  for (const entry of list.getEntries()) {
    if (entry.duration > 50) {
      console.warn('Long task detected:', entry.duration, 'ms');
    }
  }
});
observer.observe({ entryTypes: ['longtask'] });

// Frame rate monitoring
let lastTime = performance.now();
let frames = 0;
function checkFrameRate() {
  const now = performance.now();
  frames++;
  
  if (now >= lastTime + 1000) {
    const fps = Math.round((frames * 1000) / (now - lastTime));
    console.log('FPS:', fps);
    if (fps < 30) {
      console.warn('Low frame rate detected');
    }
    frames = 0;
    lastTime = now;
  }
  
  requestAnimationFrame(checkFrameRate);
}
checkFrameRate();
```

## Conclusion

**Most Effective Solution**: Cross-origin iframes with separate subdomains
- ✅ True process isolation
- ✅ Prevents memory leaks from spreading
- ✅ Protects against crashes
- ✅ Supported in modern browsers
- ✅ Used by industry leaders

**Key Takeaway**: Sandbox attributes alone are insufficient. Process isolation through cross-origin iframes is the only reliable solution for preventing iframe memory issues from affecting the parent page.

**Action Items**:
1. Set up separate subdomain (e.g., `sandbox.yourdomain.com`)
2. Configure CORS properly for cross-origin communication
3. Use `postMessage` for iframe-parent communication
4. Add memory monitoring for early warning
5. Implement automatic iframe restart on high memory
6. Consider COOP/COEP headers for additional isolation

## References

- [Chrome Site Isolation](https://www.chromium.org/Home/chromium-security/site-isolation/)
- [MDN: iframe sandbox](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/iframe#attr-sandbox)
- [COOP and COEP explained](https://web.dev/coop-coep/)
- [Web Workers API](https://developer.mozilla.org/en-US/docs/Web/API/Web_Workers_API)
- [StackBlitz WebContainers](https://blog.stackblitz.com/posts/introducing-webcontainers/)
- [CodeSandbox Architecture](https://codesandbox.io/docs/learn/sandboxes/overview)
