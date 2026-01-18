# Duplicate Code Detection Research - Complete Index

## 📋 Overview

This research provides a comprehensive analysis of tools and techniques for detecting duplicate code in software projects, with specific focus on Go and TypeScript, and special consideration for AI-assisted development workflows.

**Total Research:** 2,900+ lines across 6 documents  
**Tools Evaluated:** 20+  
**Real Projects Tested:** 2 (dissect, diagram-dsl)  
**Custom Rules Created:** 21 Semgrep patterns  

## 🚀 Quick Navigation

### Start Here
If you're new to this research, start with these documents in order:

1. **[Executive Summary](DUPLICATE_CODE_DETECTION_EXECUTIVE_SUMMARY.md)** (9.4K) - 5 min read
   - TL;DR and immediate actions
   - Best tool recommendations
   - Quick start commands

2. **[README](DUPLICATE_CODE_DETECTION_README.md)** (7.1K) - 5 min read
   - Overview and quick start
   - Installation instructions
   - Success metrics

3. **[Tool Comparison](DUPLICATE_CODE_DETECTION_TOOL_COMPARISON.md)** (11K) - 10 min read
   - Feature comparison matrix
   - Decision trees
   - Performance data

### Deep Dive
For comprehensive information:

4. **[Full Research](DUPLICATE_CODE_DETECTION_RESEARCH.md)** (31K) - 30 min read
   - Complete tool analysis
   - AST-based approaches
   - Code embeddings
   - MCP server concepts
   - Gap analysis

5. **[Practical Examples](DUPLICATE_CODE_DETECTION_EXAMPLES.md)** (15K) - 15 min read
   - Real command outputs
   - Integration examples
   - Before/after refactoring
   - CI/CD setup

### Resources

6. **[Semgrep Rules](semgrep-ai-patterns.yaml)** (7.5K)
   - 21 custom patterns
   - Ready to use
   - AI-specific patterns

## 📚 Document Summaries

### Executive Summary
**Best for:** Busy developers who want actionable recommendations

**Contains:**
- Tool selection guide (dupl, jscpd, Semgrep)
- Installation commands
- Test results summary
- Immediate actions
- Workflow recommendations

**Key Takeaway:** Use dupl for Go, jscpd for TypeScript, Semgrep for patterns.

---

### README
**Best for:** First-time users getting started

**Contains:**
- Quick start guide
- Tool installation
- Basic usage examples
- Success metrics
- Future work roadmap

**Key Takeaway:** Install and run tools in 5 minutes.

---

### Tool Comparison
**Best for:** Deciding which tool to use

**Contains:**
- Feature comparison matrices
- Performance benchmarks
- Cost analysis
- Decision trees
- Use case scenarios
- Integration examples

**Key Sections:**
- Summary comparison table
- Detection capabilities matrix
- Integration & workflow comparison
- Real-world performance data
- Quick decision tree

**Key Takeaway:** Choose based on your specific needs and constraints.

---

### Full Research
**Best for:** Understanding the complete landscape

**Contains:**
- 20+ tool evaluations
- Go-specific tools: dupl, goclone, goreporter
- TypeScript tools: jscpd, eslint-plugin-sonarjs, ts-morph
- Language-agnostic: Semgrep, SonarQube, Code Climate, Sourcegraph
- AST-based analysis
- Code embedding models
- MCP server concepts
- Gap analysis
- Future directions

**Key Sections:**
- Types of duplication (exact, structural, semantic, functional)
- Tool-by-tool analysis with pros/cons
- Comparative analysis matrix
- Testing results on real code
- Recommendations for different scenarios
- Gap analysis
- Proposed custom solutions

**Key Takeaway:** Token-based tools work for exact/structural duplicates; semantic duplicates need ML.

---

### Practical Examples
**Best for:** Learning by example

**Contains:**
- Real command line examples
- Actual tool outputs
- Before/after refactoring code
- Integration snippets (pre-commit, CI/CD, Makefile, VSCode)
- Troubleshooting guide

**Key Sections:**
- dupl examples with output
- jscpd examples with reports
- Semgrep pattern examples
- Integration examples
- Comparison of outputs
- Best practices

**Key Takeaway:** Copy-paste ready examples for immediate use.

---

### Semgrep Rules
**Best for:** Preventing AI-generated duplicates

**Contains:**
- 21 custom Semgrep rules
- Go patterns (14 rules)
- TypeScript/JavaScript patterns (7 rules)
- Common patterns (both languages)
- AI-specific patterns

**Pattern Categories:**
- Error handling
- Validation
- String checks
- Array operations
- HTTP handlers
- JSON encoding
- Database queries
- Test patterns

**Key Takeaway:** Drop-in patterns to detect common AI duplicates.

## 🎯 Use Case Navigation

### "I just want to check my code for duplicates"

**Read:**
1. Executive Summary → Quick Start section
2. Run: `dupl .` (Go) or `jscpd .` (TypeScript)

**Time:** 5 minutes

---

### "I need to set up duplicate detection in CI/CD"

**Read:**
1. Practical Examples → CI/CD Integration
2. Tool Comparison → Integration & Workflow

**Copy:** GitHub Actions workflow from examples

**Time:** 15 minutes

---

### "I want to understand all available tools"

**Read:**
1. Full Research (complete read)
2. Tool Comparison → Feature matrices

**Time:** 45 minutes

---

### "AI agents keep duplicating code patterns"

**Read:**
1. Executive Summary → For AI Agent Workflows
2. Practical Examples → Semgrep Examples

**Use:** semgrep-ai-patterns.yaml

**Time:** 20 minutes

---

### "I need to present this to my team"

**Use:**
1. Executive Summary (overview)
2. Tool Comparison → Decision trees
3. Practical Examples → Show tool outputs

**Time:** 30 minutes for presentation prep

---

### "I want to build a custom solution"

**Read:**
1. Full Research → AST-Based Analysis
2. Full Research → Code Embedding Models
3. Full Research → Custom MCP Server concepts

**Time:** 1-2 hours

## 📊 Key Findings Summary

### Tools That Work Well

| Tool | Strength | Speed | Accuracy |
|------|----------|-------|----------|
| dupl | Go exact/structural | ⚡⚡⚡⚡⚡ | ⭐⭐⭐⭐ |
| jscpd | Multi-language, reports | ⚡⚡⚡⚡ | ⭐⭐⭐⭐ |
| Semgrep | Custom patterns | ⚡⚡⚡⚡ | ⭐⭐⭐⭐⭐ |
| SonarQube | Enterprise, comprehensive | ⚡⚡ | ⭐⭐⭐⭐⭐ |

### What Works, What Doesn't

✅ **Works Well:**
- Exact copy-paste detection
- Structural duplicates (same code, different names)
- Pattern-based detection (with Semgrep)

❌ **Doesn't Work:**
- Semantic duplicates (different code, same function)
- Cross-language duplicates
- Intent-based detection

⚠️ **Partially Works:**
- Real-time AI integration (needs custom MCP server)
- Incremental analysis (SonarQube, Semgrep)
- Context awareness (requires custom rules)

### Test Results

**Synthetic Go Code:**
- dupl: 3 clone groups (exact + structural) ✅
- jscpd: 1 clone group (exact) ✅
- Semgrep: 6 pattern matches ✅

**Real Go Project (dissect):**
- dupl: 3 clone groups at threshold 30
- Mostly in test setup and command processing

**Real TypeScript Project (diagram-dsl):**
- jscpd: 14.99% duplication (638 lines)
- 14 clone groups found
- Mostly in examples and test utilities

## 🛠️ Recommended Toolchain

### For Most Projects

```bash
# Install
go install github.com/mibk/dupl@latest
npm install -g jscpd
pip install semgrep

# Run
dupl -threshold 30 .                    # Go
jscpd . --format typescript --min-lines 5  # TypeScript
semgrep --config=semgrep-ai-patterns.yaml .  # Patterns
```

### For CI/CD

```yaml
# GitHub Actions
- name: Check Duplicates
  run: |
    dupl -threshold 50 . || echo "Warning"
    jscpd . --threshold 15
    semgrep --config=semgrep-ai-patterns.yaml .
```

### For AI Agents

**Workflow:**
1. Before writing: Search for similar code
2. After writing: Run dupl/jscpd
3. Before commit: Check Semgrep patterns

## 📈 Implementation Roadmap

### Phase 1: Immediate (Today)
- [ ] Install dupl and jscpd
- [ ] Run on current codebase
- [ ] Review findings
- [ ] Set baseline thresholds

### Phase 2: This Week
- [ ] Add pre-commit hooks
- [ ] Configure CI/CD checks
- [ ] Create custom Semgrep rules
- [ ] Document team workflow

### Phase 3: This Month
- [ ] Set up SonarQube (if needed)
- [ ] Track duplication metrics
- [ ] Refactor high-duplication areas
- [ ] Train team on tools

### Phase 4: This Quarter
- [ ] Experiment with code embeddings
- [ ] Build custom MCP server (if needed)
- [ ] Fine-tune thresholds based on data
- [ ] Automate refactoring suggestions

## 🔗 External Resources

### Tool Documentation
- [dupl GitHub](https://github.com/mibk/dupl)
- [jscpd Documentation](https://github.com/kucherenko/jscpd)
- [Semgrep Docs](https://semgrep.dev/docs/)
- [SonarQube](https://www.sonarqube.org/)

### Related Research
- [CodeBERT Paper](https://arxiv.org/abs/2002.08155)
- [Tree-sitter](https://tree-sitter.github.io/)
- [AST Analysis](https://en.wikipedia.org/wiki/Abstract_syntax_tree)
- [Model Context Protocol](https://modelcontextprotocol.io/)

## 🤝 Contributing

Found a useful tool or technique? Add it to the research!

1. Test the tool on real code
2. Document findings
3. Update relevant sections
4. Share Semgrep rules if applicable

## 📞 Questions?

Each document addresses different aspects:
- **Quick answer?** → Executive Summary
- **How to use?** → Practical Examples
- **Which tool?** → Tool Comparison
- **Deep understanding?** → Full Research
- **Custom rules?** → Semgrep Rules file

## 📝 Document Statistics

| Document | Size | Lines | Reading Time | Audience |
|----------|------|-------|--------------|----------|
| Executive Summary | 9.4K | 420 | 5 min | Everyone |
| README | 7.1K | 320 | 5 min | New users |
| Tool Comparison | 11K | 580 | 10 min | Decision makers |
| Full Research | 31K | 1,040 | 30 min | Technical deep dive |
| Practical Examples | 15K | 660 | 15 min | Implementers |
| Semgrep Rules | 7.5K | 260 | Reference | Developers |
| **Total** | **81K** | **2,917** | **~60 min** | All |

## 🎓 Learning Path

**Beginner (30 min):**
1. Executive Summary
2. README
3. Run one tool on your code

**Intermediate (1 hour):**
1. Tool Comparison
2. Practical Examples
3. Set up CI/CD integration

**Advanced (2 hours):**
1. Full Research
2. Custom Semgrep rules
3. AST-based approaches

**Expert (Ongoing):**
1. Build custom MCP server
2. Experiment with code embeddings
3. Contribute improvements

## 🏆 Success Stories

### What You Can Achieve

**Week 1:**
- Detect duplicates in existing codebase
- Set up automated checks
- Baseline metrics established

**Month 1:**
- CI/CD integration working
- Team trained on tools
- Custom rules created
- 20% duplication reduction

**Quarter 1:**
- Duplication under threshold
- Refactoring patterns established
- Metrics tracked over time
- Team habits changed

## 🔮 Future Directions

### Short Term
- Additional Semgrep rules
- More integration examples
- Language-specific guides

### Medium Term
- MCP server implementation
- Code embedding experiments
- Automated refactoring tools

### Long Term
- AI-native duplicate detection
- Real-time agent integration
- Semantic similarity search

## 📅 Maintenance

**This research is current as of:** January 2025

**To update:**
- Check tool versions
- Re-run tests on latest code
- Update benchmarks
- Add new tools
- Refine recommendations

## 💡 Final Thoughts

**The Bottom Line:**
- Duplicate code detection is a solved problem for exact/structural duplicates
- Use dupl (Go) + jscpd (TypeScript) + Semgrep (patterns)
- Takes 5 minutes to start, lifetime to master
- AI integration is the frontier - build custom solutions for best results

**Start simple, iterate often, measure impact.**

---

*Research conducted by reviewing 20+ tools, testing on 2 real projects, and creating 2,900+ lines of documentation.*
