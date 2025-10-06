# Implementation Patterns for Client-Side Iframe Protection

## Sandpack Integration Example

The easiest production-ready approach uses CodeSandbox's open-source Sandpack:

```javascript
import { SandpackClient } from '@codesandbox/sandpack-client';

const client = new SandpackClient('#preview', {
  files: { '/index.html': '<h1>Hello</h1>' },
  entry: '/index.html'
});

// Monitor and restart on high memory
setInterval(async () => {
  const mem = await client.getMemoryUsage();
  if (mem.percentage > 85) await client.restartRuntime();
}, 5000);
```

**Installation**: `npm install @codesandbox/sandpack-client`  
**Docs**: https://sandpack.codesandbox.io/docs/

## Complete Production Monitor

Condensed implementation covering all patterns from CodeSandbox, Figma, and Observable:

```javascript
class IframeMonitor {
  constructor(iframe, config = {}) {
    this.iframe = iframe;
    this.thresholds = { warning: 70, critical: 85, emergency: 95 };
    this.measurements = [];
    this.restarts = 0;
    this.maxRestarts = 3;
    this.lastRestart = 0;
    this.checkInterval = 5000;
    Object.assign(this, config);
  }
  
  start() {
    this.interval = setInterval(() => this.check(), this.checkInterval);
  }
  
  check() {
    const pct = this.getMemoryPercentage();
    this.measurements.push({ time: Date.now(), pct });
    if (this.measurements.length > 20) this.measurements.shift();
    
    if (pct > this.thresholds.emergency) this.emergencyRestart();
    else if (pct > this.thresholds.critical) this.scheduledRestart(10000);
    else if (pct > this.thresholds.warning) this.warn();
    
    if (this.detectLeak()) this.scheduledRestart(2000);
  }
  
  getMemoryPercentage() {
    if (!performance.memory) return this.heuristicCheck();
    return (performance.memory.usedJSHeapSize / 
            performance.memory.jsHeapSizeLimit) * 100;
  }
  
  heuristicCheck() {
    // Firefox/Safari fallback - check load time
    const avgLoadTime = this.getAverageLoadTime();
    return avgLoadTime > 5000 ? 90 : 50; // Simplified
  }
  
  detectLeak() {
    if (this.measurements.length < 10) return false;
    const recent = this.measurements.slice(-10);
    const increasing = recent.filter((m, i) => 
      i > 0 && m.pct > recent[i-1].pct
    ).length;
    return increasing > 8 && recent[recent.length-1].pct > 50;
  }
  
  warn() {
    this.showBanner('Memory usage high', 'warning');
  }
  
  scheduledRestart(delay) {
    this.showBanner(`Restarting in ${delay/1000}s`, 'critical');
    setTimeout(() => this.restart(), delay);
  }
  
  emergencyRestart() {
    if (!this.canRestart()) {
      this.showBanner('Too many restarts', 'error');
      return;
    }
    this.restart();
  }
  
  canRestart() {
    const now = Date.now();
    if (now - this.lastRestart < 60000) return false; // 1min cooldown
    return this.restarts < this.maxRestarts;
  }
  
  async restart() {
    const state = await this.saveState();
    this.iframe.src = 'about:blank';
    setTimeout(() => {
      this.iframe.src = this.originalSrc;
      this.restarts++;
      this.lastRestart = Date.now();
      this.iframe.onload = () => this.restoreState(state);
    }, 100);
  }
  
  async saveState() {
    return new Promise(resolve => {
      this.iframe.contentWindow.postMessage({type: 'SAVE'}, '*');
      const handler = e => {
        if (e.data.type === 'STATE') {
          window.removeEventListener('message', handler);
          resolve(e.data.state);
        }
      };
      window.addEventListener('message', handler);
      setTimeout(() => resolve(null), 1000); // Timeout
    });
  }
  
  restoreState(state) {
    if (state) {
      this.iframe.contentWindow.postMessage({
        type: 'RESTORE', state
      }, '*');
    }
  }
  
  showBanner(msg, level) {
    console.log(`[${level}] ${msg}`);
    // Implement UI notification
  }
  
  stop() {
    clearInterval(this.interval);
  }
}

// Usage
const monitor = new IframeMonitor(document.querySelector('iframe'), {
  warning: 70, critical: 85, emergency: 95, maxRestarts: 3
});
monitor.start();
```

## Figma-Style Plugin Limits

Enforce memory limits declaratively like Figma:

```javascript
class PluginManager {
  constructor() {
    this.plugins = new Map();
  }
  
  load(pluginId, config) {
    const iframe = document.createElement('iframe');
    iframe.src = `https://${pluginId}.plugins.app`;
    iframe.sandbox = 'allow-scripts';
    
    const monitor = new IframeMonitor(iframe, {
      critical: 85,
      maxRestarts: 2,
      onExhausted: () => this.terminate(pluginId)
    });
    
    this.plugins.set(pluginId, { iframe, monitor, limit: config.memoryLimit });
    monitor.start();
    
    return iframe;
  }
  
  terminate(pluginId) {
    const plugin = this.plugins.get(pluginId);
    if (plugin) {
      plugin.monitor.stop();
      plugin.iframe.remove();
      this.plugins.delete(pluginId);
      this.notify(`Plugin ${pluginId} terminated due to memory`);
    }
  }
}
```

## Observable-Style Selective Isolation

Only isolate problematic cells like Observable:

```javascript
class SmartIsolation {
  needsIsolation(code) {
    return code.length > 10000 ||
           /while\s*\(\s*true\s*\)/.test(code) ||
           /new\s+Array\s*\(\s*\d{7,}\s*\)/.test(code) ||
           this.hasLeakedBefore(code);
  }
  
  async run(code, id) {
    if (this.needsIsolation(code)) {
      console.log(`Isolating ${id}`);
      return this.runInIframe(code, id);
    }
    return this.runInline(code);
  }
  
  runInIframe(code, id) {
    const iframe = document.createElement('iframe');
    iframe.srcdoc = `<script>${code}</script>`;
    iframe.sandbox = 'allow-scripts';
    
    const monitor = new IframeMonitor(iframe);
    monitor.start();
    
    return { iframe, monitor };
  }
  
  runInline(code) {
    return eval(code);
  }
  
  hasLeakedBefore(code) {
    const hash = this.hash(code);
    return localStorage.getItem(`leak:${hash}`) === 'true';
  }
  
  markAsLeaky(code) {
    const hash = this.hash(code);
    localStorage.setItem(`leak:${hash}`, 'true');
  }
  
  hash(str) {
    return str.split('').reduce((a,b) => {
      a = ((a << 5) - a) + b.charCodeAt(0);
      return a & a;
    }, 0);
  }
}
```

## Browser Compatibility Handler

Handle Chrome-only `performance.memory`:

```javascript
class MemoryDetector {
  constructor() {
    this.hasAPI = !!performance.memory;
    this.loadTimes = [];
  }
  
  getUsage() {
    if (this.hasAPI) {
      return {
        percentage: (performance.memory.usedJSHeapSize / 
                    performance.memory.jsHeapSizeLimit) * 100,
        absolute: performance.memory.usedJSHeapSize,
        method: 'api'
      };
    }
    return this.heuristicUsage();
  }
  
  heuristicUsage() {
    // Track iframe load performance
    const avgLoadTime = this.loadTimes.slice(-5)
      .reduce((a, b) => a + b, 0) / 5;
    
    // Longer load times suggest memory pressure
    const percentage = Math.min(100, (avgLoadTime / 50)); // 50ms = 1%
    
    return {
      percentage,
      absolute: null,
      method: 'heuristic',
      confidence: this.loadTimes.length >= 5 ? 'medium' : 'low'
    };
  }
  
  recordLoad(ms) {
    this.loadTimes.push(ms);
    if (this.loadTimes.length > 10) this.loadTimes.shift();
  }
}
```

## Server-Side Fallback Pattern

Hybrid approach - detect dangerous code and route to server:

```javascript
class HybridRenderer {
  isDangerous(code) {
    const patterns = [
      /while\s*\(\s*true\s*\)/,
      /for\s*\([^)]*;;[^)]*\)/,
      /new\s+Array\s*\(\s*\d{7,}\s*\)/,
      /(\w+)\s*=\s*\1/  // Variable assigned to itself
    ];
    return patterns.some(p => p.test(code));
  }
  
  async render(code) {
    if (this.isDangerous(code)) {
      console.log('Routing to server-side rendering');
      return this.serverRender(code);
    }
    return this.clientRender(code);
  }
  
  async serverRender(code) {
    const response = await fetch('/api/render', {
      method: 'POST',
      body: JSON.stringify({ code })
    });
    const blob = await response.blob();
    const img = document.createElement('img');
    img.src = URL.createObjectURL(blob);
    return img;
  }
  
  clientRender(code) {
    const iframe = document.createElement('iframe');
    iframe.srcdoc = code;
    new IframeMonitor(iframe).start();
    return iframe;
  }
}
```

## Key Implementation Notes

**Cross-Origin Setup**: Use separate subdomain (`sandbox.example.com`) not just different path. This ensures true process isolation in modern browsers.

**State Persistence**: Store in localStorage or IndexedDB, not in-memory. This survives iframe reloads and even page refreshes.

**Restart Limits**: Always implement cooldown (1 minute) and max attempts (3) to prevent infinite restart loops.

**User Communication**: Show clear messages about what's happening. Users accept restarts when they understand why.

**Progressive Enhancement**: Start with Chrome-only `performance.memory`, add heuristics for Firefox/Safari, gracefully degrade features.

**Memory Thresholds**: Industry standard is 70% warning, 85% critical, 95% emergency. Adjust based on your user base and typical workloads.

## Testing Strategies

Create intentional memory leaks for testing:

```javascript
// Test: Gradual leak
let leak = [];
setInterval(() => {
  leak.push(new Array(100000).fill('x'));
}, 100);

// Test: Sudden spike
let spike = new Array(10000000).fill('data');

// Test: Infinite loop (will freeze)
while(true) { Math.random(); }
```

Monitor your iframe monitor with these patterns to verify thresholds and restart behavior work correctly.
