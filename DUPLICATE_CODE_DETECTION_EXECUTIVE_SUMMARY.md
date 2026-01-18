# Duplicate Code Detection - Executive Summary

## TL;DR

**Best Tools for Detecting Duplicate Code:**
- **Go:** `dupl` (fast, accurate, easy to use)
- **TypeScript:** `jscpd` (multi-language, great reporting)
- **Pattern Detection:** `Semgrep` (finds variations, custom rules)

**For AI Agents:** Combine dupl/jscpd for exact duplicates + Semgrep for patterns + pre-commit hooks

## Quick Start

### Install Tools
```bash
# Go
go install github.com/mibk/dupl@latest

# TypeScript/Multi-language
npm install -g jscpd

# Pattern matching
pip install semgrep
```

### Run Analysis
```bash
# Go projects
dupl -threshold 30 .

# TypeScript projects
jscpd . --format typescript --min-lines 5

# Pattern detection (custom rules)
semgrep --config=rules/ .
```

## Key Findings

### What Tools Can Detect

| Type | Example | dupl | jscpd | Semgrep | Semantic (ML) |
|------|---------|------|-------|---------|---------------|
| **Exact Copy-Paste** | Identical code blocks | ✅ | ✅ | ✅ | ✅ |
| **Structural Duplicate** | Same logic, different names | ✅ | ✅ | ⚠️ | ✅ |
| **Pattern Duplicate** | Common patterns (error handling) | ❌ | ❌ | ✅ | ✅ |
| **Semantic Duplicate** | Different code, same function | ❌ | ❌ | ❌ | ✅ |

### Real-World Test Results

**Synthetic Go Code (2 files, ~50 lines each):**
- dupl: Found 3 clone groups (exact + structural)
- jscpd: Found 1 clone group (most obvious)
- Semgrep: Found 6 pattern matches (error handling, validation)

**Real Repository (dissect - Go project):**
- dupl: 3 clone groups with threshold 30
- Mostly in test code and command processing

**Real Repository (diagram-dsl - TypeScript project):**
- jscpd: 14.99% duplication (638 lines)
- Mostly in examples and test utilities

### What Tools Miss

1. **Semantic Duplicates:** All text/token-based tools miss functionally equivalent code with different implementations
2. **Cross-Language Duplicates:** Similar logic in Go and TypeScript isn't detected
3. **Intent-Based Duplicates:** Can't distinguish accidental from intentional duplication

## Recommendations

### For Immediate Use

**1. Quick Check Before Commit**
```bash
#!/bin/bash
# .git/hooks/pre-commit
dupl -threshold 30 . && jscpd . --threshold 15
```

**2. CI/CD Integration**
```yaml
# .github/workflows/code-quality.yml
- name: Check duplicates
  run: |
    go install github.com/mibk/dupl@latest
    dupl -threshold 50 . || echo "Warning: duplicates found"
```

**3. Periodic Review**
```bash
# Weekly/monthly analysis
jscpd . --format typescript,go --reporters html
# Opens HTML report showing all duplicates
```

### For AI Agent Workflows

**Approach 1: Pre-emptive Search**
Before writing new code, AI agent should:
1. Search for similar function names: `rg "func.*User"`
2. Run pattern matching: `semgrep --config=patterns.yml`
3. Check semantic similarity (if vector DB available)

**Approach 2: Post-write Validation**
After writing code:
1. Run dupl/jscpd on new file
2. If duplicates found, refactor or extract common code
3. Commit only if no duplicates above threshold

**Approach 3: Custom MCP Server**
Build an MCP server that:
- Indexes codebase (AST + embeddings)
- Provides "check_duplicate" tool
- Returns existing similar code
- Agent decides: reuse, refactor, or write new

### Best Practices

1. **Set Reasonable Thresholds**
   - dupl: 30-50 tokens (lower = more sensitive)
   - jscpd: 5-10 lines minimum
   - Adjust based on false positive rate

2. **Ignore Intentional Duplicates**
   - Test fixtures and setup code
   - Example/demo files
   - Generated code

3. **Focus on Business Logic**
   - Prioritize domain code over infrastructure
   - Error handling patterns can be similar
   - Boilerplate is often unavoidable

4. **Combine Multiple Tools**
   - dupl/jscpd for speed
   - Semgrep for patterns
   - Manual review for semantic issues

## Tools Landscape

### Recommended Tools

| Tool | Languages | Speed | Accuracy | Integration | Cost |
|------|-----------|-------|----------|-------------|------|
| **dupl** | Go | ⚡⚡⚡ | ⭐⭐⭐⭐ | Easy | Free |
| **jscpd** | 150+ | ⚡⚡ | ⭐⭐⭐⭐ | Easy | Free |
| **Semgrep** | 30+ | ⚡⚡⚡ | ⭐⭐⭐⭐⭐ | Medium | Free tier |
| **SonarQube** | 30+ | ⚡ | ⭐⭐⭐⭐⭐ | Complex | $$$$ |

### Tool Selection Guide

**Choose dupl if:**
- Working only with Go
- Need fast results
- Want simple CLI integration

**Choose jscpd if:**
- Multi-language project
- Want detailed HTML reports
- Need TypeScript support

**Choose Semgrep if:**
- Want to define custom patterns
- AI agents keep duplicating specific patterns
- Need semantic pattern matching

**Choose SonarQube if:**
- Enterprise environment
- Need historical tracking
- Want comprehensive code quality metrics

## Gap Analysis

### Current Limitations

1. **No Real-Time AI Integration**
   - Tools designed for batch analysis
   - No MCP servers available
   - Manual workflow integration needed

2. **Limited Semantic Understanding**
   - Token/text-based only
   - Can't detect "different code, same purpose"
   - No understanding of intent

3. **No Cross-Language Detection**
   - Go and TypeScript duplicates not detected
   - Each language analyzed separately

4. **No Context Awareness**
   - Can't distinguish intentional from accidental
   - No understanding of "why" code exists

### Future Solutions

1. **MCP Server for Duplicate Detection**
   - Wrap dupl/jscpd in MCP interface
   - Provide real-time feedback to AI agents
   - Integrate with code editors

2. **Code Embedding Models**
   - Use CodeBERT or similar
   - Semantic similarity search
   - Find functionally equivalent code

3. **Agent Prompt Engineering**
   - Add "check for duplicates" to system prompt
   - Require search before writing
   - Enforce refactoring when duplicates found

## Sample Workflow for AI Agents

```
┌─────────────────────────────────────┐
│ AI Agent Receives Task              │
│ "Add user validation function"      │
└───────────────┬─────────────────────┘
                ↓
┌─────────────────────────────────────┐
│ Step 1: Search Existing Code        │
│ $ rg "validate.*user"               │
│ $ rg "func.*Valid"                  │
└───────────────┬─────────────────────┘
                ↓
         ┌──────┴──────┐
         │ Found?      │
         └──────┬──────┘
        Yes ←──┘  └──→ No
         ↓              ↓
┌────────────────┐  ┌──────────────────┐
│ Use Existing   │  │ Step 2: Write    │
│ or Refactor    │  │ New Function     │
└────────────────┘  └────────┬─────────┘
                             ↓
                    ┌──────────────────┐
                    │ Step 3: Check    │
                    │ $ dupl .         │
                    │ $ semgrep .      │
                    └────────┬─────────┘
                             ↓
                    ┌──────────────────┐
                    │ Duplicates?      │
                    └────────┬─────────┘
                    Yes ←───┘  └──→ No
                     ↓              ↓
            ┌────────────────┐  ┌──────────────┐
            │ Refactor       │  │ Commit       │
            └────────────────┘  └──────────────┘
```

## Conclusion

**Duplicate code detection is a solved problem for exact/structural duplicates.** Tools like dupl and jscpd work well and are easy to integrate.

**The challenge is semantic duplicates and AI-specific workflows.** This requires:
- Custom MCP server integration
- Code embedding models for semantic search
- Better agent prompting and workflow design

**Immediate action:** Install dupl and jscpd, add to pre-commit hooks or CI/CD. This will catch 80% of duplicate code issues with minimal effort.

**Future work:** Build MCP server wrapper, fine-tune code embedding models for domain-specific code, and develop agent-specific workflow patterns.

## Next Steps

1. **Today:** Install and run dupl/jscpd on your projects
2. **This week:** Add to CI/CD pipeline, create Semgrep custom rules
3. **This month:** Experiment with code embeddings for semantic search
4. **This quarter:** Build custom MCP server for real-time duplicate detection

## Resources

- Full research document: `DUPLICATE_CODE_DETECTION_RESEARCH.md`
- Test examples: `/tmp/duplicate-test/`
- Tool repositories:
  - dupl: https://github.com/mibk/dupl
  - jscpd: https://github.com/kucherenko/jscpd
  - Semgrep: https://semgrep.dev/

---

**Questions or need help?** The full research document contains detailed information on:
- 20+ tools evaluated
- AST-based analysis approaches
- Code embedding models
- Building custom solutions
- Production deployment strategies
