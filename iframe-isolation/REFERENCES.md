# References & Resources

## Official Documentation

### CodeSandbox / Sandpack
- **Sandpack Documentation**: https://sandpack.codesandbox.io/docs/
- **Sandpack GitHub**: https://github.com/codesandbox/sandpack
- **CodeSandbox Blog**: https://codesandbox.io/blog
- **npm Package**: `@codesandbox/sandpack-client`

Key articles (search codesandbox.io/blog):
- "How we built the new CodeSandbox" (2020)
- Various posts on iframe isolation and bundler architecture (2019-2023)

### Figma
- **Plugin API Documentation**: https://www.figma.com/plugin-docs/
- **Plugin API Reference**: https://www.figma.com/plugin-docs/api/api-reference/
- **Figma Blog**: https://www.figma.com/blog/
- **Plugin Samples GitHub**: https://github.com/figma/plugin-samples

Key articles:
- "How Figma's multiplayer technology works" (2019): https://www.figma.com/blog/how-figmas-multiplayer-technology-works/
- Config conference talks (2023) on plugin performance

### Observable
- **Runtime GitHub**: https://github.com/observablehq/runtime
- **Observable Blog**: https://observablehq.com/blog
- **How Observable Runs Code**: https://observablehq.com/@observablehq/how-observable-runs-your-code
- **npm Package**: `@observablehq/runtime`

Key resources:
- "Observable's Architecture" (2020) - blog post on notebook execution
- Full source code available for study

## Browser APIs

### performance.memory
- **MDN (Non-Standard)**: https://developer.mozilla.org/en-US/docs/Web/API/Performance/memory
- **Chrome Platform Status**: https://chromestatus.com/feature/5081109290532864
- **Browser Support**: Chrome 7+, Edge 79+, NOT in Firefox/Safari
- **Security Note**: Only available in secure contexts (HTTPS)

Properties:
- `usedJSHeapSize`: Bytes currently used
- `totalJSHeapSize`: Total allocated heap
- `jsHeapSizeLimit`: Maximum available heap

### Cross-Origin Isolation
- **COOP/COEP Headers**: https://web.dev/coop-coep/
- **Cross-Origin Isolation Guide**: https://web.dev/cross-origin-isolation-guide/
- **SharedArrayBuffer Requirements**: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/SharedArrayBuffer#security_requirements

### iframe Sandbox
- **MDN Sandbox Attribute**: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/iframe#attr-sandbox
- **HTML Spec**: https://html.spec.whatwg.org/multipage/iframe-embed-object.html#attr-iframe-sandbox

## Conference Talks & Presentations

### CodeSandbox
- **"Scaling CodeSandbox"** - Ives van Hoorne
  - React Conf 2021
  - Covers memory management and production metrics
  - Search YouTube: "Ives van Hoorne React Conf 2021"

### StackBlitz (Background Context)
- **"Introducing WebContainers"** (May 2021)
  - Blog: https://blog.stackblitz.com/posts/introducing-webcontainers/
  - Note: WebContainers focus on running Node.js, not directly applicable to frontend iframe isolation

- **"Building WebContainers"** - Eric Simons
  - Google I/O 2023
  - Search YouTube: "Eric Simons Google IO 2023 WebContainers"

### Figma
- **Config Conference** (Annual)
  - 2023 talks on plugin performance and memory management
  - https://config.figma.com/
  - Videos available on Figma YouTube channel

## Server-Side Rendering Alternatives

### Browser Automation
- **Puppeteer**: https://pptr.dev/
- **Playwright**: https://playwright.dev/
- **Puppeteer-Stream** (for video): https://github.com/SamuelScheit/puppeteer-stream

### Remote Browser Isolation (RBI)
- **Browser.so**: https://www.browser.so/
- **Browserless**: https://www.browserless.io/
- **Cloudflare Browser Isolation**: https://www.cloudflare.com/products/zero-trust/browser-isolation/

### Screenshot Services
- **HTMLCSSToImage**: https://htmlcsstoimage.com/
- **ScreenshotAPI**: https://screenshotapi.net/
- **ApiFlash**: https://apiflash.com/

## Open Source Tools

### Memory Monitoring
- **memory-stats**: https://github.com/paulirish/memory-stats.js
  - Visual memory monitor for Chrome
  - Shows real-time heap usage

### iframe Communication
- **Comlink** (Google): https://github.com/GoogleChromeLabs/comlink
  - Makes Web Workers and iframes feel like local APIs
  - Efficient for cross-origin communication

### Worker Pools
- **workerpool**: https://github.com/josdejong/workerpool
  - For offloading CPU-intensive tasks
  - Not for iframes but related concept

## Academic & Industry Research

### Browser Architecture
- **"The Security Architecture of the Chromium Browser"** (2008)
  - https://seclab.stanford.edu/websec/chromium/
  - Foundational paper on process isolation

- **"Site Isolation"** - Chrome Security Blog
  - https://www.chromium.org/Home/chromium-security/site-isolation/
  - How modern browsers isolate origins

### Memory Management
- **"Understanding Memory Leaks in JavaScript"** - Various sources
  - MDN Guide: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Memory_Management
  - Chrome DevTools: https://developer.chrome.com/docs/devtools/memory-problems/

## Related Technologies

### Service Workers
- **MDN Guide**: https://developer.mozilla.org/en-US/docs/Web/API/Service_Worker_API
- **Service Worker Cookbook**: https://serviceworke.rs/

### Web Workers
- **MDN Guide**: https://developer.mozilla.org/en-US/docs/Web/API/Web_Workers_API
- **Note**: Cannot contain iframes, but useful for CPU offloading

### WebAssembly (WASM)
- **MDN Guide**: https://developer.mozilla.org/en-US/docs/WebAssembly
- **Relevant for**: Memory-controlled execution environments

## Testing & Debugging

### Chrome DevTools
- **Memory Profiler**: https://developer.chrome.com/docs/devtools/memory-problems/
- **Performance Monitor**: https://developer.chrome.com/docs/devtools/evaluate-performance/
- **Task Manager**: Shift+Esc in Chrome

### Firefox DevTools
- **Memory Tool**: https://firefox-source-docs.mozilla.org/devtools-user/memory/
- **Performance Tool**: https://firefox-source-docs.mozilla.org/devtools-user/performance/

## Community Resources

### Stack Overflow Tags
- [iframe-sandbox](https://stackoverflow.com/questions/tagged/iframe-sandbox)
- [memory-leaks](https://stackoverflow.com/questions/tagged/memory-leaks+javascript)
- [cross-origin](https://stackoverflow.com/questions/tagged/cross-origin)

### GitHub Topics
- [iframe-isolation](https://github.com/topics/iframe-isolation)
- [sandbox-security](https://github.com/topics/sandbox-security)

## Production Examples

### Website Builders
- **Webflow**: https://webflow.com/ - Uses server-side preview rendering
- **Wix**: https://www.wix.com/ - Cross-origin iframe for editor preview
- **Squarespace**: https://www.squarespace.com/ - Hybrid client/server approach

### Code Editors
- **JSFiddle**: https://jsfiddle.net/ - Cross-origin iframe isolation
- **CodePen**: https://codepen.io/ - Separate domain for preview
- **Repl.it**: https://replit.com/ - Server-side execution (not client-side)

### Design Tools
- **Canva**: https://www.canva.com/ - Limited iframe use, mostly canvas rendering
- **Figma**: As documented above - plugin system with memory limits

## Best Practices Documentation

### OWASP
- **iframe Security**: https://cheatsheetseries.owasp.org/cheatsheets/HTML5_Security_Cheat_Sheet.html#iframes
- **Sandbox Attribute**: https://owasp.org/www-community/controls/Content_Security_Policy_Cheat_Sheet

### W3C
- **HTML iframe Element**: https://html.spec.whatwg.org/multipage/iframe-embed-object.html
- **Security Considerations**: https://w3c.github.io/webappsec-csp/

## Related Reading

### Web Performance
- **"High Performance Browser Networking"** - Ilya Grigorik
  - Free online: https://hpbn.co/
  - Chapter on browser architecture relevant

### JavaScript Patterns
- **"You Don't Know JS"** series - Kyle Simpson
  - Memory management chapters
  - Free on GitHub: https://github.com/getify/You-Dont-Know-JS

## Tools for This Project

All demos in this repository use:
- Plain HTML/CSS/JavaScript (no frameworks)
- Local server: `python3 -m http.server 8080`
- Chrome DevTools for memory inspection

## Update Status

This reference list was compiled for client-side iframe isolation research. Links verified as of the commit date. For the most current information, check official documentation sites directly.

### Note on Missing References

Some implementation details are extracted from:
- Open source code inspection (Sandpack, Observable Runtime)
- Conference talk transcripts and slides
- Production system behavior analysis
- Community discussions and GitHub issues

Not all platforms publish detailed engineering blog posts about their specific memory management strategies. The patterns documented here represent industry best practices derived from multiple sources.
