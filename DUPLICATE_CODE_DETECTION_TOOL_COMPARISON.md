# Duplicate Code Detection - Tool Comparison Matrix

Quick reference guide for choosing the right tool for your needs.

## Summary Comparison

| Tool | Languages | Speed | Accuracy | Ease of Use | Cost | Best For |
|------|-----------|-------|----------|-------------|------|----------|
| **dupl** | Go | ⚡⚡⚡⚡⚡ | ⭐⭐⭐⭐ | ✅✅✅ | Free | Go projects, CI/CD |
| **jscpd** | 150+ | ⚡⚡⚡⚡ | ⭐⭐⭐⭐ | ✅✅✅ | Free | Multi-language, reports |
| **Semgrep** | 30+ | ⚡⚡⚡⚡ | ⭐⭐⭐⭐⭐ | ✅✅ | Free tier | Custom patterns |
| **SonarQube** | 30+ | ⚡⚡ | ⭐⭐⭐⭐⭐ | ✅ | $$$ | Enterprise, teams |
| **Code Climate** | 10+ | ⚡⚡⚡ | ⭐⭐⭐⭐ | ✅✅ | $$$ | Cloud, GitHub |
| **PMD CPD** | 20+ | ⚡⚡⚡ | ⭐⭐⭐ | ✅✅ | Free | Java ecosystem |
| **Sourcegraph** | All | ⚡⚡ | ⭐⭐⭐⭐ | ✅✅ | $$$ | Code search, discovery |
| **Custom AST** | Any | ⚡ | ⭐⭐⭐⭐⭐ | ❌ | Dev time | Full control |

Legend:
- Speed: ⚡ = very slow to ⚡⚡⚡⚡⚡ = instant
- Accuracy: ⭐ = poor to ⭐⭐⭐⭐⭐ = excellent
- Ease of Use: ❌ = difficult, ✅ = easy, ✅✅✅ = very easy
- Cost: Free, $ = cheap, $$$ = expensive, Dev time = requires development

## Detailed Feature Comparison

### Detection Capabilities

| Feature | dupl | jscpd | Semgrep | SonarQube | AST-based |
|---------|------|-------|---------|-----------|-----------|
| **Exact duplicates** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Structural duplicates** | ⚠️ | ⚠️ | ✅ | ✅ | ✅ |
| **Semantic duplicates** | ❌ | ❌ | ⚠️ | ⚠️ | ✅ |
| **Pattern matching** | ❌ | ❌ | ✅ | ✅ | ✅ |
| **Cross-file** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Cross-language** | ❌ | ⚠️ | ⚠️ | ⚠️ | ❌ |
| **Custom rules** | ❌ | ⚠️ | ✅ | ✅ | ✅ |
| **Ignores formatting** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Ignores naming** | ⚠️ | ⚠️ | ⚠️ | ⚠️ | ✅ |

Legend: ✅ = Yes, ⚠️ = Partial, ❌ = No

### Integration & Workflow

| Feature | dupl | jscpd | Semgrep | SonarQube | Code Climate |
|---------|------|-------|---------|-----------|--------------|
| **CLI** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **CI/CD** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Pre-commit hooks** | ✅ | ✅ | ✅ | ⚠️ | ⚠️ |
| **IDE integration** | ❌ | ❌ | ✅ | ✅ | ⚠️ |
| **Git integration** | ❌ | ⚠️ | ✅ | ✅ | ✅ |
| **Web dashboard** | ❌ | ⚠️ | ✅ | ✅ | ✅ |
| **API access** | ❌ | ❌ | ✅ | ✅ | ✅ |
| **Webhook support** | ❌ | ❌ | ⚠️ | ✅ | ✅ |

### Reporting & Output

| Feature | dupl | jscpd | Semgrep | SonarQube | Code Climate |
|---------|------|-------|---------|-----------|--------------|
| **Console output** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **JSON export** | ❌ | ✅ | ✅ | ✅ | ✅ |
| **HTML report** | ❌ | ✅ | ✅ | ✅ | ❌ |
| **XML export** | ❌ | ✅ | ❌ | ✅ | ❌ |
| **CSV export** | ❌ | ✅ | ❌ | ✅ | ❌ |
| **Markdown** | ❌ | ✅ | ⚠️ | ❌ | ❌ |
| **Historical trends** | ❌ | ❌ | ⚠️ | ✅ | ✅ |
| **Blame info** | ❌ | ✅ | ❌ | ✅ | ✅ |
| **Code snippets** | ✅ | ✅ | ✅ | ✅ | ✅ |

### Performance Characteristics

| Tool | Small Projects (<1K files) | Medium (1K-10K) | Large (>10K) | Incremental |
|------|---------------------------|----------------|--------------|-------------|
| **dupl** | <1s | <5s | <30s | ❌ |
| **jscpd** | <2s | <20s | 1-5min | ❌ |
| **Semgrep** | <5s | <30s | 2-10min | ✅ |
| **SonarQube** | <10s | 1-5min | 10-30min | ✅ |
| **Code Climate** | <10s | 1-5min | 10-30min | ✅ |

### Setup & Maintenance

| Aspect | dupl | jscpd | Semgrep | SonarQube | Code Climate |
|--------|------|-------|---------|-----------|--------------|
| **Installation time** | 1 min | 2 min | 3 min | 30 min | 5 min |
| **Configuration** | None | Optional | Required | Complex | Simple |
| **Learning curve** | 10 min | 15 min | 1 hour | 1 day | 30 min |
| **Maintenance** | None | Low | Low | High | Low |
| **Dependencies** | Go | Node.js | Python | Java/Docker | Cloud |
| **Self-hosted** | N/A | N/A | N/A | ✅ | ❌ |
| **Cloud option** | N/A | N/A | ✅ | ✅ | ✅ |

## Use Case Decision Matrix

### Scenario: Small Go Project

**Best Choice:** dupl

**Why:**
- ✅ Single command install
- ✅ No configuration needed
- ✅ Very fast
- ✅ Accurate for Go

**Command:**
```bash
go install github.com/mibk/dupl@latest
dupl -threshold 30 .
```

### Scenario: Multi-Language Project (Go + TypeScript)

**Best Choice:** jscpd

**Why:**
- ✅ Supports both languages
- ✅ Single tool for all code
- ✅ Great HTML reports
- ✅ Easy to configure

**Command:**
```bash
npm install -g jscpd
jscpd . --format go,typescript --reporters html,console
```

### Scenario: AI Agent Writing Code

**Best Choice:** Semgrep

**Why:**
- ✅ Custom pattern matching
- ✅ Can detect "AI patterns"
- ✅ Finds variations
- ✅ Real-time feedback possible

**Setup:**
```bash
pip install semgrep
semgrep --config=semgrep-ai-patterns.yaml .
```

### Scenario: Enterprise Team

**Best Choice:** SonarQube

**Why:**
- ✅ Comprehensive analysis
- ✅ Historical tracking
- ✅ Team collaboration
- ✅ Quality gates
- ✅ Beautiful dashboards

**Note:** Requires infrastructure investment

### Scenario: Open Source Project

**Best Choice:** Code Climate (free tier)

**Why:**
- ✅ Free for open source
- ✅ GitHub integration
- ✅ Badge support
- ✅ No maintenance

### Scenario: Custom Requirements

**Best Choice:** Custom AST-based solution

**Why:**
- ✅ Full control over logic
- ✅ Can detect semantic duplicates
- ✅ Domain-specific rules
- ✅ Exact fit for needs

**Note:** Significant development effort

## Quick Decision Tree

```
Start
  ↓
┌────────────────────┐
│ What language?     │
└────────┬───────────┘
         ↓
    ┌────┴────┐
    │         │
   Go      TypeScript/Multiple
    ↓         ↓
  dupl      jscpd
    ↓         ↓
┌──────────────────────┐
│ Need custom patterns?│
└──────┬───────────────┘
       ↓
   ┌───┴───┐
   │       │
  Yes     No
   ↓       ↓
 Semgrep  Continue with dupl/jscpd
   ↓
┌──────────────────────┐
│ Need dashboard/team? │
└──────┬───────────────┘
       ↓
   ┌───┴───┐
   │       │
  Yes     No
   ↓       ↓
SonarQube  Done!
```

## Real-World Performance Data

Based on testing performed on this repository:

### dissect (Go project)
- **Files:** ~30 Go files
- **Lines:** ~3,000
- **dupl:** <100ms, found 3 clone groups
- **jscpd:** ~200ms, found similar results
- **Recommendation:** dupl for speed

### diagram-dsl (TypeScript project)
- **Files:** ~40 TypeScript/TSX files
- **Lines:** ~4,200
- **jscpd:** ~400ms, found 14 clones (14.99% duplication)
- **Semgrep:** ~2s, found 24 pattern matches
- **Recommendation:** jscpd for overview, Semgrep for patterns

## Cost Comparison

### Free Options
- **dupl:** Free, open source
- **jscpd:** Free, open source
- **Semgrep:** Free tier available (community rules)
- **SonarQube:** Free community edition (self-hosted)

### Paid Options
- **SonarQube Developer:** $150/year per developer
- **SonarQube Enterprise:** Custom pricing
- **Code Climate:** $500/month for teams
- **Semgrep Pro:** $50/month per developer

### Hidden Costs
- **Self-hosted:** Server costs, maintenance time
- **Custom AST:** Development time (1-4 weeks)
- **Integration:** CI/CD setup, workflow changes
- **Training:** Team learning curve

## Recommendation Summary

### For Most Projects
Start with **dupl** (Go) or **jscpd** (TypeScript/multi-language). Free, fast, easy.

### Add Custom Patterns
Once you see common duplicates, create **Semgrep rules** to prevent them.

### Scale to Enterprise
When you need dashboards and history, migrate to **SonarQube** or **Code Climate**.

### Go Custom
Only build custom AST-based solution if you have unique needs that no tool satisfies.

## Integration Examples

### Pre-commit Hook
```bash
#!/bin/bash
# .git/hooks/pre-commit

# Go projects
if [ -f go.mod ]; then
    dupl -threshold 30 . || exit 1
fi

# TypeScript projects
if [ -f package.json ]; then
    jscpd . --threshold 15 || exit 1
fi
```

### GitHub Actions
```yaml
name: Code Quality

on: [push, pull_request]

jobs:
  duplicates:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Check Go duplicates
        run: |
          go install github.com/mibk/dupl@latest
          dupl -threshold 50 .
      
      - name: Check TypeScript duplicates
        run: |
          npm install -g jscpd
          jscpd . --threshold 15
      
      - name: Check patterns
        run: |
          pip install semgrep
          semgrep --config=semgrep-ai-patterns.yaml .
```

### Makefile
```makefile
.PHONY: check-duplicates
check-duplicates:
	@echo "Checking for duplicate code..."
	@dupl -threshold 30 . || echo "Duplicates found"
	@jscpd . --threshold 15 || echo "Duplicates found"
	@semgrep --config=semgrep-ai-patterns.yaml . || echo "Patterns found"
```

## Common Pitfalls

### ❌ Don't
- Set thresholds too low (too many false positives)
- Run on generated code
- Ignore all duplicates in tests
- Only check before release

### ✅ Do
- Start with high thresholds, lower gradually
- Exclude build artifacts and dependencies
- Review test duplicates (might indicate common setup)
- Run on every commit (CI/CD)
- Create custom rules for your patterns
- Track trends over time

## Further Resources

- **dupl:** https://github.com/mibk/dupl
- **jscpd:** https://github.com/kucherenko/jscpd
- **Semgrep:** https://semgrep.dev/
- **SonarQube:** https://www.sonarqube.org/
- **Code Climate:** https://codeclimate.com/

## Updates

This comparison was last updated based on:
- Tool versions as of January 2025
- Real-world testing on this repository
- Community feedback and documentation

Check tool websites for latest features and pricing.
