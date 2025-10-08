# Duplicate Code Detection Research

This directory contains comprehensive research on tools and techniques for detecting duplicate code, with a focus on Go and TypeScript projects and AI-assisted development workflows.

## 📚 Documentation

### Quick Start
- **[Executive Summary](DUPLICATE_CODE_DETECTION_EXECUTIVE_SUMMARY.md)** - TL;DR with immediate actions
- **[Full Research](DUPLICATE_CODE_DETECTION_RESEARCH.md)** - Comprehensive analysis of 20+ tools
- **[Semgrep Rules](semgrep-ai-patterns.yaml)** - Custom rules for AI-generated patterns

### What's Inside

**Executive Summary** covers:
- Best tools for Go and TypeScript
- Quick installation and usage
- Test results from real projects
- Recommendations for different use cases
- AI-specific workflow patterns

**Full Research** covers:
- Go-specific tools (dupl, goclone, goreporter)
- TypeScript tools (jscpd, eslint-plugin-sonarjs, ts-morph)
- Language-agnostic tools (Semgrep, SonarQube, Sourcegraph)
- AST-based analysis approaches
- Code embedding models for semantic search
- MCP server concepts
- Gap analysis and future directions

**Semgrep Rules** include:
- 21 custom patterns for common duplicates
- Error handling patterns
- Validation patterns
- Database query patterns
- API response patterns
- AI-generated code patterns

## 🚀 Quick Start

### Install Tools

```bash
# Go duplicate detection
go install github.com/mibk/dupl@latest

# Multi-language duplicate detection
npm install -g jscpd

# Pattern-based detection
pip install semgrep
```

### Run Analysis

```bash
# Check Go code
dupl -threshold 30 .

# Check TypeScript code
jscpd . --format typescript --min-lines 5

# Check with custom patterns
semgrep --config=semgrep-ai-patterns.yaml .
```

## 📊 Key Findings

### What Works Well

1. **dupl** - Excellent for Go, fast (10-50ms), finds exact and structural duplicates
2. **jscpd** - Great for TypeScript, multi-language support, beautiful HTML reports
3. **Semgrep** - Best for pattern detection, highly customizable, finds variations

### What Doesn't Work

1. **Semantic duplicates** - All token-based tools miss functionally equivalent code
2. **Cross-language** - Similar logic in Go and TypeScript isn't detected
3. **AI integration** - No real-time MCP servers, batch analysis only

### Test Results

Tested on synthetic and real code:
- **Synthetic Go (2 files):** Found 3 clone groups with dupl
- **Real Go project (dissect):** Found 3 clone groups at threshold 30
- **Real TypeScript project (diagram-dsl):** 14.99% duplication, 14 clones

## 🎯 Recommendations

### For Quick Checks

```bash
# Pre-commit hook
#!/bin/bash
dupl -threshold 30 . || echo "Warning: duplicates found"
```

### For CI/CD

```yaml
# GitHub Actions
- name: Check duplicates
  run: |
    npm install -g jscpd
    jscpd . --threshold 15
```

### For AI Agents

1. **Before writing code:** Search for similar functions
2. **After writing code:** Run dupl/jscpd on new files
3. **Before commit:** Check with Semgrep patterns

## 🔧 Tool Selection

**Choose dupl if:**
- ✅ Working only with Go
- ✅ Need fast results
- ✅ Want simple CLI

**Choose jscpd if:**
- ✅ Multi-language project
- ✅ Want HTML reports
- ✅ Need TypeScript support

**Choose Semgrep if:**
- ✅ Want custom patterns
- ✅ Need semantic matching
- ✅ AI keeps duplicating specific patterns

**Choose SonarQube if:**
- ✅ Enterprise environment
- ✅ Need historical tracking
- ✅ Want comprehensive metrics

## 📝 Example: Using Semgrep Rules

The included `semgrep-ai-patterns.yaml` detects common patterns AI tends to duplicate:

```bash
# Run all rules
semgrep --config=semgrep-ai-patterns.yaml .

# Focus on high-severity only
semgrep --config=semgrep-ai-patterns.yaml . --severity=ERROR

# Exclude test files
semgrep --config=semgrep-ai-patterns.yaml . --exclude='*_test.go'

# Generate JSON report
semgrep --config=semgrep-ai-patterns.yaml . --json > duplicates.json
```

### Example Output

```
file1.go:
  ❱ duplicate-error-check
    Common error handling pattern detected.
    
    11┆ if err != nil {
    12┆   return err
    13┆ }

  ❱ duplicate-slice-iteration
    Sum calculation over slice. Might already exist.
    
    28┆ for _, item := range items {
    29┆   sum += item.Price
    30┆ }
```

## 🧪 Testing

All tools were tested on:
1. Synthetic Go code with intentional duplicates
2. Synthetic TypeScript code with similar patterns
3. Real projects: dissect (Go) and diagram-dsl (TypeScript)

Results documented in full research document.

## 🔮 Future Work

### Short Term
- Add dupl/jscpd to pre-commit hooks
- Create custom Semgrep rules for project-specific patterns
- Set up CI/CD integration

### Medium Term
- Experiment with code embeddings (CodeBERT)
- Build semantic search for functionally equivalent code
- Integrate with code review tools

### Long Term
- Build MCP server for real-time duplicate detection
- Fine-tune embedding models for domain-specific code
- Develop agent-specific workflow automation

## 🤝 Contributing

Found a useful tool or pattern? Contributions welcome!

1. Test the tool on real code
2. Document results
3. Add to research document
4. Share Semgrep rules if applicable

## 📖 Further Reading

### Tools
- [dupl](https://github.com/mibk/dupl) - Go duplicate detector
- [jscpd](https://github.com/kucherenko/jscpd) - Multi-language detector
- [Semgrep](https://semgrep.dev/) - Pattern-based analysis
- [SonarQube](https://www.sonarqube.org/) - Comprehensive platform

### Concepts
- [AST-based analysis](https://en.wikipedia.org/wiki/Abstract_syntax_tree)
- [Code embeddings](https://github.com/microsoft/CodeBERT)
- [Tree-sitter](https://tree-sitter.github.io/)
- [Model Context Protocol (MCP)](https://modelcontextprotocol.io/)

## 🐛 Known Limitations

1. **Token-based tools miss semantic duplicates**
   - Different code, same purpose
   - Need ML-based semantic search

2. **No cross-language detection**
   - Similar Go and TypeScript code not detected
   - Need unified representation

3. **False positives common in:**
   - Test setup code
   - Example files
   - Boilerplate patterns

4. **No real-time AI integration**
   - Batch analysis only
   - No MCP servers available
   - Manual workflow integration

## 📈 Success Metrics

Track these to measure effectiveness:

1. **Duplicate reduction:** % of code marked as duplicate
2. **False positive rate:** How often tools flag acceptable duplicates
3. **Time saved:** Fewer reimplementations
4. **Code quality:** Improved maintainability scores

## 💡 Tips

1. **Start with low thresholds** and increase if too many false positives
2. **Focus on business logic** not infrastructure code
3. **Combine multiple tools** for better coverage
4. **Review regularly** as codebase evolves
5. **Customize rules** for your specific patterns

## 📧 Questions?

See the full research document for detailed information on:
- Specific tool comparisons
- AST-based implementation details
- Building custom solutions
- Production deployment strategies
- AI agent integration patterns

---

**Last Updated:** 2025-01-08  
**Research Status:** Complete  
**Tools Tested:** 10+  
**Lines of Research:** 1000+
