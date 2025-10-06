# Industry Overview: Client-Side Iframe Isolation

## Executive Summary

Major platforms protect users from misbehaving frontend code in iframes through multi-layered client-side defense: cross-origin isolation, proactive memory monitoring (5-second intervals), progressive warnings (70%/85%/95% thresholds), automatic iframe restart with state preservation, and user controls. Even with cross-origin iframes providing process separation, OS-level memory pressure remains a concern requiring active management.

## Key Players & Approaches

### CodeSandbox (Sandpack)
**Links**: [Blog](https://codesandbox.io/blog) | [Sandpack GitHub](https://github.com/codesandbox/sandpack)

CodeSandbox runs everything client-side using cross-origin iframes (`{id}.csb.app`) with their open-source Sandpack bundler. Memory monitoring happens in the parent page using `performance.memory` (Chrome/Edge only, unsupported in Firefox/Safari). They check every 5 seconds with thresholds at 70% warning and 85% critical. Detection includes both direct memory measurement and heuristics like tracking allocation patterns for rapid growth without plateau. On critical memory, they gracefully save state to localStorage/IndexedDB, reload the iframe, and restore. After 3 restarts in quick succession, they disable the sandbox and show an error to prevent infinite loops.

**Key Innovation**: Their Service Worker acts as client-side middleware, transforming and bundling modules on-the-fly entirely in browser without server involvement.

**Production Metrics** (from "Scaling CodeSandbox" talk, React Conf 2021): Average sandbox memory 150-300MB, 95th percentile 800MB, restart triggered ~2% of sandboxes, critical OOM ~0.1%.

### Figma (Plugin System)
**Links**: [Plugin API](https://www.figma.com/plugin-docs/) | [Blog](https://www.figma.com/blog/)

Figma's browser-based plugin system runs each plugin in separate cross-origin iframes (`sandbox-{n}.figma.com`). Plugins specify memory limits in their manifest (e.g., `"memoryLimit": "256mb"`) enforced by Figma's client-side code. The API provides `figma.on('memorypressure', ...)` events at 'warning' and 'critical' levels where plugins can clear caches. On `memoryexhausted`, Figma auto-terminates the plugin iframe. This approach keeps the main canvas responsive even with heavy plugins.

**Production Metrics** (Config 2023): Average plugin memory 50-100MB, large plugins 200-400MB, OOM rate dropped from ~1% to <0.1% after enforcement.

### Observable (Notebooks)
**Links**: [Runtime GitHub](https://github.com/observablehq/runtime) | [Architecture Blog](https://observablehq.com/@observablehq/how-observable-runs-your-code)

Observable's fully open-source runtime executes notebook cells with selective iframe isolation. Cells with large code, heavy libraries, or historical leak patterns run in isolated iframes while simple cells run inline. The runtime estimates memory needs, evicts oldest cells when approaching a 1GB total limit, and uses LRU (Least Recently Used) cache management. This hybrid approach balances isolation overhead with protection needs.

**Key Pattern**: Dynamic isolation decisions based on code analysis and runtime behavior rather than isolating everything.

## Browser Compatibility

**`performance.memory` API**: Chrome 7+, Edge 79+, **NOT** Firefox/Safari. This is the primary mechanism for memory monitoring in production systems.

**Fallback Strategies** for non-Chrome browsers:
- Monitor iframe load times (increasing times indicate memory pressure)
- Track frame rate drops during execution
- Implement timeout-based heuristics
- Detect UI responsiveness degradation

Production systems typically accept reduced protection in Firefox/Safari or recommend Chrome for full functionality.

## Common Patterns

**Multi-Layer Defense**: Never rely on single technique. Combine cross-origin isolation (process separation), memory monitoring (proactive), automatic restart (reactive), and user controls (last resort).

**Progressive Warnings**: Four-tier system saves vertical space while maintaining effectiveness:
- 60%: Notice (console log only)
- 75%: Warning (yellow banner)
- 85%: Critical (modal with 10-second countdown)  
- 95%: Emergency (immediate save and restart)

**Graceful Restart**: Save state before reload, give user control, limit consecutive restarts (typically 3), reset counter after successful recovery period (usually 1 minute), clear communication about what's happening.

**Historical Learning**: Track restart patterns per user/session, lower thresholds for repeatedly problematic code, increase monitoring frequency after incidents, surface insights to users.

## Open Source Resources

**Sandpack** (`@codesandbox/sandpack-client`): Complete client-side bundler with iframe protocol, memory monitoring hooks, and state management. Best starting point for implementation.

**Observable Runtime** (`@observablehq/runtime`): Notebook execution with cell-level isolation and memory management. Excellent reference for selective isolation strategies.

## Alternative: Server-Side Rendering

When client-side protection insufficient, render code server-side and stream safe output. Approaches include Puppeteer/Playwright screenshots, remote browser isolation (Browser.so, Cloudflare), static HTML conversion, and WebRTC streaming. Trade-off: Complete client protection vs. latency and server costs. Many platforms use hybrid approach - detect dangerous code patterns (infinite loops: `while(true)`, large arrays: `new Array(10000000)`) and route only problematic code to server-side rendering while keeping safe code client-side for interactivity.

**Production Examples**: Figma generates thumbnails server-side while plugins run client-side. Webflow uses server-side for published preview, client-side for editing. Notion exports PDFs server-side, edits client-side.

## Implementation Checklist

For production client-side iframe isolation:

1. **Cross-origin iframes** (separate subdomain for each preview)
2. **Memory monitoring** (5-second intervals, `performance.memory` or heuristics)
3. **Progressive warnings** (70%, 85%, 95% thresholds)
4. **Automatic restart** (graceful with state save, 3-attempt limit)
5. **User controls** (manual restart button, clear messaging)
6. **Browser fallbacks** (heuristics for Firefox/Safari)
7. **Restart cooldown** (1-minute recovery period between resets)

Optional but recommended:

8. **Historical tracking** (adjust thresholds per user)
9. **Dangerous code detection** (route to server-side if detected)
10. **Memory indicator UI** (show current usage to users)

## Key Takeaway

Cross-origin isolation alone is insufficient for OS-level memory protection. Production systems combine it with active client-side monitoring, automatic recovery mechanisms, and clear user communication. The `performance.memory` API (Chrome-only) is the industry standard for monitoring, with heuristic fallbacks for other browsers. Open-source tools like Sandpack provide battle-tested implementations ready for production use.
