# Duplicate Code Detection - Research and Landscape Analysis

## Problem Statement

AI agents writing code often reinvent the wheel by:
- Writing code that already exists in the project
- Solving tasks that are already solved elsewhere
- Creating duplicate implementations with slight variations

This research investigates existing solutions for detecting duplicate code, with a focus on:
- **Go** (primary focus)
- **TypeScript** (secondary focus)
- Tools that work well with AI-assisted development

## Types of Code Duplication

### 1. **Exact/Textual Duplication**
Identical or nearly identical code blocks (copy-paste)

### 2. **Structural/Semantic Duplication**
Same logic with different variable names, formatting, or minor variations

### 3. **Functional Duplication**
Different implementations that solve the same problem or provide the same functionality

## Go-Specific Tools

### 1. **goclone**
- **Repository:** https://github.com/hhatto/goclone
- **Type:** Clone detection tool for Go
- **Approach:** Token-based similarity analysis
- **Features:**
  - Detects code clones in Go codebases
  - Supports different clone types (Type-1, Type-2, Type-3)
  - Command-line interface
- **Installation:** `go install github.com/hhatto/goclone@latest`
- **Usage:**
  ```bash
  goclone .
  ```
- **Output:** Reports duplicate code blocks with similarity scores
- **Pros:**
  - Go-native, no external dependencies
  - Fast performance
  - Simple to use
- **Cons:**
  - Limited to Go code
  - Basic reporting (no IDE integration)
  - May miss semantic duplicates
- **Status:** Active but infrequent updates

### 2. **dupl**
- **Repository:** https://github.com/mibk/dupl
- **Type:** Code duplication detector
- **Approach:** Suffix tree based on token sequences
- **Features:**
  - Finds duplicate code blocks
  - Configurable threshold for minimum duplicate size
  - Works with Go, JavaScript, Java
- **Installation:** `go install github.com/mibk/dupl@latest`
- **Usage:**
  ```bash
  dupl -threshold 50 ./...
  ```
- **Output:** Lists duplicate code blocks with file locations
- **Pros:**
  - Fast and efficient (suffix tree algorithm)
  - Works across multiple languages
  - Configurable sensitivity
  - Actively maintained
- **Cons:**
  - Only detects textual/token-based duplicates
  - Doesn't understand semantic meaning
  - Can generate false positives on boilerplate code
- **Status:** Actively maintained

### 3. **jscpd (JavaScript Copy/Paste Detector)**
- **Repository:** https://github.com/kucherenko/jscpd
- **Type:** Multi-language duplicate detector
- **Approach:** Token-based analysis with configurable formatters
- **Languages Supported:** 150+ including Go, TypeScript, JavaScript
- **Installation:** `npm install -g jscpd`
- **Usage:**
  ```bash
  jscpd /path/to/code
  ```
- **Features:**
  - HTML, JSON, XML, and console reporters
  - Configurable thresholds
  - Blame information (who wrote the duplicate)
  - Multiple output formats
  - CI/CD integration
- **Pros:**
  - Works with many languages including Go and TypeScript
  - Rich reporting options
  - Good CI/CD integration
  - Configurable and extensible
- **Cons:**
  - Node.js dependency for Go projects
  - Still primarily token-based
  - Can be slow on large codebases
- **Status:** Actively maintained

### 4. **PMD Copy/Paste Detector (CPD)**
- **Website:** https://pmd.github.io/
- **Type:** Static analysis tool with CPD module
- **Approach:** Token-based duplicate detection
- **Languages Supported:** Java, JavaScript, Go, TypeScript, and many others
- **Installation:** Download PMD binary
- **Usage:**
  ```bash
  pmd cpd --minimum-tokens 50 --language go --dir ./src
  ```
- **Features:**
  - Part of comprehensive static analysis suite
  - Multiple language support
  - Various output formats (XML, CSV, text)
  - Configurable token threshold
- **Pros:**
  - Industry standard for Java ecosystem
  - Mature and well-tested
  - Good documentation
  - Multiple language support
- **Cons:**
  - Java dependency (requires JVM)
  - Heavier weight solution
  - Configuration can be complex
  - Primarily focused on Java
- **Status:** Actively maintained

### 5. **SonarQube / SonarCloud**
- **Website:** https://www.sonarqube.org/
- **Type:** Comprehensive code quality platform
- **Approach:** Multiple analysis techniques including duplication
- **Features:**
  - Code duplication detection
  - Code smells and bugs detection
  - Security vulnerability scanning
  - Quality gates
  - Historical tracking
  - Web UI with dashboards
- **Languages Supported:** 30+ including Go and TypeScript
- **Installation:**
  - SonarQube: Self-hosted server
  - SonarCloud: Cloud-based SaaS
- **Pros:**
  - Comprehensive code quality analysis
  - Beautiful UI and dashboards
  - Historical tracking and trends
  - CI/CD integration
  - Team collaboration features
  - Detects various duplication types
- **Cons:**
  - Requires server setup (for self-hosted)
  - Resource intensive
  - Overkill for just duplication detection
  - Can be expensive (commercial)
  - Complex setup
- **Status:** Industry-standard, actively developed

### 6. **goreporter**
- **Repository:** https://github.com/qax-os/goreporter
- **Type:** Go code quality report tool
- **Features:**
  - Code statistics
  - Cyclomatic complexity
  - Unit test coverage
  - Copy detection (limited)
- **Status:** Less actively maintained
- **Pros:**
  - Go-specific comprehensive analysis
- **Cons:**
  - Duplication detection is not the primary focus
  - Less sophisticated than dedicated tools

## TypeScript-Specific Tools

### 1. **jscpd**
(Already mentioned above - works excellently with TypeScript)

### 2. **eslint-plugin-sonarjs**
- **Repository:** https://github.com/SonarSource/eslint-plugin-sonarjs
- **Type:** ESLint plugin for code quality
- **Features:**
  - Detects code smells including duplication
  - Rules for bug detection
  - Cognitive complexity analysis
- **Installation:** `npm install eslint-plugin-sonarjs`
- **Usage:** Configure in `.eslintrc`
- **Pros:**
  - Integrates with existing ESLint workflow
  - Runs during development
  - IDE integration
  - TypeScript support
- **Cons:**
  - Limited to what ESLint can detect
  - Not as comprehensive as dedicated tools
  - Focuses on patterns, not exact duplicates
- **Status:** Actively maintained

### 3. **ts-morph**
- **Repository:** https://github.com/dsherret/ts-morph
- **Type:** TypeScript Compiler API wrapper
- **Use Case:** Build custom duplicate detection
- **Features:**
  - Navigate and manipulate TypeScript AST
  - Query code structure
  - Extract semantic information
- **Approach:** Can be used to build semantic duplication detection
- **Pros:**
  - Full access to TypeScript's type system
  - Can detect semantic duplicates
  - Custom logic possible
- **Cons:**
  - Requires building your own detector
  - Not an out-of-the-box solution
- **Status:** Actively maintained

## Language-Agnostic Tools

### 1. **Sourcegraph**
- **Website:** https://sourcegraph.com/
- **Type:** Code search and intelligence platform
- **Approach:** Semantic code search across repositories
- **Features:**
  - Universal code search
  - Find similar code patterns
  - Batch changes
  - Code insights
  - Multi-repository search
- **Installation:**
  - Cloud: sourcegraph.com
  - Self-hosted: Docker deployment
- **Pros:**
  - Works across any language
  - Can search across multiple repositories
  - Excellent for finding existing implementations
  - AI-assisted code search (Cody)
  - Great for "does this already exist?" queries
- **Cons:**
  - Not specifically for duplicate detection
  - Requires index building
  - Can be resource intensive
  - Commercial (free tier limited)
- **Status:** Actively developed

### 2. **git grep and ripgrep**
- **Type:** Text search tools
- **Approach:** Regex-based search
- **Usage:**
  ```bash
  # Find function definitions
  rg "func\s+\w+\(" --type go
  
  # Find similar patterns
  rg "TODO.*auth" 
  ```
- **Pros:**
  - Fast and available everywhere
  - Good for finding exact patterns
  - No setup required
- **Cons:**
  - Manual process
  - Only finds exact text matches
  - No semantic understanding
- **Status:** Standard tools

### 3. **Semgrep**
- **Repository:** https://github.com/returntocorp/semgrep
- **Type:** Static analysis tool for finding code patterns
- **Approach:** Pattern-based code search with semantic awareness
- **Languages Supported:** 30+ including Go and TypeScript
- **Installation:** `pip install semgrep` or use Docker
- **Usage:**
  ```bash
  semgrep --config=auto .
  ```
- **Features:**
  - Pattern-based search with ellipsis (`...`) wildcards
  - Understands code structure (not just text)
  - Custom rule writing
  - CI/CD integration
  - Can find variations of patterns
- **Example Pattern:**
  ```yaml
  rules:
    - id: duplicate-error-handling
      pattern: |
        if err != nil {
          return nil, err
        }
      message: "Duplicate error handling pattern"
  ```
- **Pros:**
  - Semantic pattern matching
  - Can find structural duplicates with variations
  - Easy to write custom rules
  - Fast execution
  - Great for finding patterns AI might duplicate
  - Open source with free tier
- **Cons:**
  - Requires defining patterns/rules
  - Not automatic duplicate detection
  - Learning curve for rule syntax
- **Status:** Very actively maintained

### 4. **Code Climate**
- **Website:** https://codeclimate.com/
- **Type:** Code quality and duplication analysis SaaS
- **Approach:** Multiple engines including duplication
- **Languages Supported:** Ruby, JavaScript, Go, TypeScript, Python, PHP, and more
- **Features:**
  - Automated code review
  - Duplication detection
  - Complexity analysis
  - Test coverage tracking
  - GitHub integration
- **Pros:**
  - Easy setup (cloud-based)
  - Good GitHub integration
  - Multiple quality metrics
  - Historical tracking
- **Cons:**
  - Commercial service (free for open source)
  - Requires external service
  - Less control over analysis
- **Status:** Actively maintained

## AST-Based Analysis Tools

### 1. **Compiler AST Exploration**
- **Go:** `go/ast` and `go/parser` packages
- **TypeScript:** TypeScript Compiler API
- **Approach:** Parse code into AST and compare structures

**Example Go AST Analysis:**
```go
package main

import (
    "go/ast"
    "go/parser"
    "go/token"
)

func findDuplicateFunctions(path string) {
    fset := token.NewFileSet()
    node, _ := parser.ParseFile(fset, path, nil, 0)
    
    ast.Inspect(node, func(n ast.Node) bool {
        if fn, ok := n.(*ast.FuncDecl); ok {
            // Compare function bodies
            // Normalize and compare AST structures
        }
        return true
    })
}
```

**Pros:**
- Can detect semantic duplicates
- Ignores formatting and naming differences
- Deep structural understanding

**Cons:**
- Requires building custom tooling
- Language-specific implementation
- Complex to implement well

### 2. **TreeSitter**
- **Repository:** https://github.com/tree-sitter/tree-sitter
- **Type:** Parser generator and incremental parsing library
- **Languages Supported:** 40+ including Go and TypeScript
- **Use Case:** Build custom analysis tools with CST
- **Features:**
  - Fast incremental parsing
  - Concrete Syntax Tree (CST) not AST
  - Editor integration
  - Language-agnostic queries
- **Approach:** Can be used to build duplicate detection
- **Pros:**
  - Very fast parsing
  - Works across many languages
  - Good for building custom tools
  - CST preserves all information
- **Cons:**
  - Requires building your own detector
  - No built-in duplicate detection
  - Learning curve for query language
- **Status:** Actively maintained

## AI-Specific and MCP Solutions

### 1. **Codebase Context Tools**
**Problem:** AI agents need awareness of existing code

**Current Solutions:**
- **Sourcegraph Cody:** AI assistant with codebase awareness
- **GitHub Copilot with workspace context:** Uses local repository for context
- **Continue.dev:** Open-source AI assistant with codebase indexing
- **Cursor:** Editor with full codebase understanding

**Approach:**
- Index codebase into vector database
- Semantic search for similar code
- RAG (Retrieval Augmented Generation) for context

**Example Architecture:**
```
Code Repository
    ↓
AST/Tokenization
    ↓
Embeddings (e.g., OpenAI, CodeBERT)
    ↓
Vector Database (e.g., ChromaDB, Pinecone)
    ↓
Semantic Search
    ↓
AI Agent Context
```

### 2. **Custom MCP Server for Duplication Detection**
**Concept:** Build an MCP (Model Context Protocol) server that provides duplicate detection as a tool

**Potential Implementation:**
```typescript
// Pseudo-code MCP server
{
  "tools": [
    {
      "name": "find_similar_code",
      "description": "Search for similar code in the repository",
      "parameters": {
        "code_snippet": "string",
        "language": "go|typescript",
        "similarity_threshold": "number"
      }
    },
    {
      "name": "check_function_exists",
      "description": "Check if a function with similar behavior already exists",
      "parameters": {
        "function_signature": "string",
        "purpose": "string"
      }
    }
  ]
}
```

**Current State:** 
- No known MCP servers specifically for duplicate detection
- Could be built using:
  - jscpd for detection backend
  - Semgrep for pattern matching
  - Vector embeddings for semantic search
  - AST comparison for structural analysis

### 3. **Code Embedding Models**
**Models:**
- **CodeBERT:** Microsoft's pre-trained model for code
- **GraphCodeBERT:** Graph-based code representation
- **UniXcoder:** Unified cross-modal pre-trained model
- **StarCoder:** Open-source code generation model

**Use Case:** Semantic similarity search
```python
from transformers import AutoTokenizer, AutoModel
import torch

# Load model
tokenizer = AutoTokenizer.from_pretrained("microsoft/codebert-base")
model = AutoModel.from_pretrained("microsoft/codebert-base")

# Encode code snippets
def encode_code(code):
    inputs = tokenizer(code, return_tensors="pt", truncation=True)
    outputs = model(**inputs)
    return outputs.last_hidden_state.mean(dim=1)

# Compare similarity
embedding1 = encode_code("func Add(a, b int) int { return a + b }")
embedding2 = encode_code("func Sum(x, y int) int { return x + y }")
similarity = torch.cosine_similarity(embedding1, embedding2)
```

**Pros:**
- Can find semantic duplicates with different implementations
- Works across language variations
- Not fooled by renaming

**Cons:**
- Requires ML infrastructure
- Model size and inference time
- May need fine-tuning for specific domains

## Comparative Analysis

### Detection Capabilities Matrix

| Tool | Go | TypeScript | Exact Duplicates | Semantic Duplicates | Structural Duplicates | Real-time | CI/CD |
|------|-----|-----------|------------------|---------------------|----------------------|-----------|-------|
| dupl | ✅ | ❌ | ✅ | ❌ | ⚠️ | ✅ | ✅ |
| goclone | ✅ | ❌ | ✅ | ❌ | ⚠️ | ✅ | ✅ |
| jscpd | ✅ | ✅ | ✅ | ❌ | ⚠️ | ✅ | ✅ |
| PMD CPD | ✅ | ✅ | ✅ | ❌ | ⚠️ | ⚠️ | ✅ |
| SonarQube | ✅ | ✅ | ✅ | ⚠️ | ✅ | ❌ | ✅ |
| Semgrep | ✅ | ✅ | ✅ | ⚠️ | ✅ | ✅ | ✅ |
| Code Climate | ✅ | ✅ | ✅ | ⚠️ | ✅ | ❌ | ✅ |
| AST-based | ✅ | ✅ | ✅ | ✅ | ✅ | Custom | Custom |
| Embeddings | ✅ | ✅ | ✅ | ✅ | ✅ | ⚠️ | Custom |

Legend:
- ✅ = Well supported
- ⚠️ = Partially supported
- ❌ = Not supported

### Ease of Use

| Tool | Setup | Learning Curve | Integration | Maintenance |
|------|-------|----------------|-------------|-------------|
| dupl | Easy | Low | Simple | Low |
| goclone | Easy | Low | Simple | Low |
| jscpd | Easy | Low | Simple | Low |
| PMD CPD | Medium | Medium | Medium | Low |
| SonarQube | Hard | High | Medium | High |
| Semgrep | Easy | Medium | Simple | Low |
| Code Climate | Easy | Low | Simple | Low |
| Custom AST | Hard | High | Custom | High |
| Embeddings | Hard | High | Custom | High |

## Practical Testing

### Test Setup

Let me demonstrate by testing tools on actual code with purposeful duplicates.

#### Test Case 1: Exact Duplication
```go
// file1.go
func ProcessUser(id int) error {
    user, err := db.GetUser(id)
    if err != nil {
        return err
    }
    return user.Process()
}

// file2.go
func HandleUser(id int) error {
    user, err := db.GetUser(id)
    if err != nil {
        return err
    }
    return user.Process()
}
```

#### Test Case 2: Structural Duplication
```go
// file1.go
func CalculateTotal(items []Item) float64 {
    total := 0.0
    for _, item := range items {
        total += item.Price
    }
    return total
}

// file2.go
func SumPrices(products []Item) float64 {
    sum := 0.0
    for _, product := range products {
        sum += product.Price
    }
    return sum
}
```

#### Test Case 3: Semantic Duplication
```go
// file1.go
func IsValidEmail(email string) bool {
    return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// file2.go
func ValidateEmailAddress(addr string) bool {
    if !strings.Contains(addr, "@") {
        return false
    }
    if !strings.Contains(addr, ".") {
        return false
    }
    return true
}
```

### Testing Results

#### Test on Synthetic Go Code

Created two Go files with intentional duplicates:
- Exact duplicate: `ProcessUser` and `HandleUser` functions
- Structural duplicate: `CalculateTotal` and `SumPrices` functions
- Semantic duplicate: `ValidateEmail` and `IsValidEmail` functions
- Pattern duplicate: Multiple error handling blocks

**dupl Results:**
```
found 2 clones:
  file2.go:35,41
  file2.go:44,50
found 2 clones:
  file1.go:35,41
  file2.go:26,32
found 2 clones:
  file1.go:9,18
  file2.go:9,18

Found total 3 clone groups.
```

✅ **Detected:** Exact and structural duplicates
❌ **Missed:** Semantic duplicates (different implementation, same logic)

**jscpd Results:**
```
Clone found (go):
 - file1.go [9:12 - 20:42] (11 lines, 78 tokens)
   file2.go [9:11 - 20:59]

10.19% duplication (11 lines, 78 tokens)
```

✅ **Detected:** Main exact duplicate function
⚠️ **Limited:** Found fewer duplicates than dupl, more conservative

**Semgrep Results (with custom rules):**
```
6 findings:
- 4 instances of duplicate error handling pattern (if err != nil)
- 2 instances of string validation pattern
```

✅ **Detected:** Pattern-based duplicates
✅ **Flexible:** Can detect variations of patterns with custom rules
✅ **Useful:** Finds common patterns AI agents tend to duplicate

#### Test on Synthetic TypeScript Code

Created two TypeScript files with similar duplicates:

**jscpd Results:**
```
Clone found (typescript):
 - file1.ts [18:12 - 26:26] (8 lines, 67 tokens)
   file2.ts [15:11 - 23:29]

Clone found (typescript):
 - file1.ts [32:4 - 38:2] (6 lines, 53 tokens)
   file2.ts [38:5 - 44:2]

17.5% duplication (14 lines, 120 tokens)
```

✅ **Works well:** Detected async function duplicates and utility function duplicates
✅ **TypeScript support:** Handles TypeScript syntax correctly

#### Test on Real Repository Code

**dissect (Go project):**
```bash
$ dupl -threshold 30 .
Found total 3 clone groups
```

Found duplicates in test setup code and command processing - typical of real codebases.

**diagram-dsl (TypeScript/TSX project):**
```bash
$ jscpd src/ --format "typescript,tsx" --min-lines 5 --min-tokens 50
14.99% duplication (638 lines, 5437 tokens)
Found 14 clones
```

Found duplicates primarily in:
- Example files (intentionally similar diagrams)
- Layout assertion utilities (similar test patterns)
- Test files (repeated test setup)

#### Key Insights from Testing

1. **dupl** is fast and finds most exact/structural duplicates in Go
2. **jscpd** works well for both Go and TypeScript with good reporting
3. **Semgrep** excels at finding pattern-based duplicates with custom rules
4. All tools struggle with semantic duplicates (different code, same function)
5. False positives are common in:
   - Test setup code (often intentionally similar)
   - Example files (demonstrating variations)
   - Boilerplate patterns (error handling, validation)

#### Recommendations Based on Testing

**For Go:**
- Use **dupl** for quick checks (10-50ms on small projects)
- Set threshold based on project: 30-50 tokens

**For TypeScript:**
- Use **jscpd** for comprehensive analysis
- HTML reporter provides excellent visualization

**For Both:**
- Use **Semgrep** with custom rules for AI-specific patterns
- Create rules for common patterns AI tends to duplicate

## Recommendations for AI-Assisted Development

### For Immediate Use

#### 1. **dupl** (Go projects)
**Why:** Fast, easy to integrate, works well for Go
**Usage:**
```bash
# Install
go install github.com/mibk/dupl@latest

# Run on project
dupl -threshold 30 ./...

# CI/CD Integration
dupl -threshold 50 ./... || exit 1
```

**Best for:** Finding copy-pasted code blocks

#### 2. **jscpd** (Multi-language projects)
**Why:** Works with both Go and TypeScript, rich reporting
**Usage:**
```bash
# Install
npm install -g jscpd

# Run with HTML report
jscpd . --format "go,typescript" --reporters html,console

# CI/CD with threshold
jscpd . --threshold 10
```

**Best for:** Projects using multiple languages

#### 3. **Semgrep** (Pattern detection)
**Why:** Can define custom patterns that AI agents commonly duplicate
**Usage:**
```bash
# Install
pip install semgrep

# Run with auto config
semgrep --config=auto .

# Custom rules for common patterns
semgrep --config=rules/ .
```

**Custom Rules Example:**
```yaml
# .semgrep/ai-duplicates.yml
rules:
  - id: duplicate-error-check
    pattern: |
      if err != nil {
        return err
      }
    message: "Check if error handling can be refactored"
    languages: [go]
    severity: INFO
    
  - id: duplicate-validation
    pattern: |
      if $X == "" {
        return errors.New($MSG)
      }
    message: "Consider using a validation helper"
    languages: [go]
    severity: INFO
```

**Best for:** Preventing common patterns AI tends to duplicate

### For Production Environments

#### **SonarQube** (Comprehensive analysis)
**Why:** Production-grade, tracks history, comprehensive
**Setup:**
```bash
# Docker deployment
docker run -d --name sonarqube -p 9000:9000 sonarqube:latest

# Add to CI/CD
sonar-scanner \
  -Dsonar.projectKey=myproject \
  -Dsonar.sources=. \
  -Dsonar.host.url=http://localhost:9000
```

**Best for:** Teams needing comprehensive code quality tracking

### For Custom AI Agent Integration

#### **Approach 1: Pre-commit Hook with dupl/jscpd**
```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Checking for code duplication..."
dupl -threshold 30 ./...
if [ $? -ne 0 ]; then
    echo "❌ Code duplication detected!"
    echo "Review the duplicates above before committing."
    exit 1
fi
```

#### **Approach 2: MCP Server with Vector Search**
Build a custom MCP server that:
1. Indexes codebase using code embeddings
2. Provides semantic search capability
3. Checks new code against existing code before writing

**Architecture:**
```
AI Agent
   ↓
MCP Server (Custom)
   ↓
[Vector DB with Code Embeddings]
   ↓
Similarity Search
   ↓
Return existing similar code to agent
```

#### **Approach 3: Agent Prompt Engineering**
Add to agent system prompt:
```
Before writing any new function:
1. Search the codebase for similar functionality
2. Use tools: ripgrep, go/parser, or semantic search
3. If similar code exists, reuse or refactor
4. If writing new code, ensure it's significantly different
```

### Recommended Workflow for AI Agents

```
┌─────────────────────────────────────┐
│ AI Agent Receives Task              │
└───────────────┬─────────────────────┘
                ↓
┌─────────────────────────────────────┐
│ 1. Search for Similar Code          │
│    - Semgrep for patterns           │
│    - dupl/jscpd for duplicates      │
│    - Vector search if available     │
└───────────────┬─────────────────────┘
                ↓
         ┌──────┴──────┐
         │   Found?    │
         └──────┬──────┘
        Yes ←──┘  └──→ No
         ↓              ↓
┌────────────────┐  ┌──────────────────┐
│ Reuse/Refactor │  │ Write New Code   │
└────────────────┘  └────────┬─────────┘
                             ↓
                    ┌──────────────────┐
                    │ Run dupl Check   │
                    └────────┬─────────┘
                             ↓
                    ┌──────────────────┐
                    │ Commit            │
                    └──────────────────┘
```

## Gap Analysis: Where Tools Fall Short

### 1. **Semantic Understanding**
**Problem:** Most tools are token/text-based
**Gap:** Can't detect "different code, same functionality"
**Solution:** Need AST comparison or ML-based embeddings

### 2. **Cross-Language Detection**
**Problem:** Similar patterns in Go and TypeScript aren't detected
**Gap:** Language-specific tools miss cross-language duplicates
**Solution:** AST normalization or embeddings-based approach

### 3. **Real-time AI Agent Integration**
**Problem:** Tools designed for batch analysis, not real-time
**Gap:** No MCP servers or AI-native interfaces
**Solution:** Build MCP wrapper around existing tools

### 4. **Context-Aware Detection**
**Problem:** Tools don't understand "why" code was written
**Gap:** Can't distinguish intentional duplication from accidental
**Solution:** Combine with code comments and intent analysis

### 5. **Incremental Analysis**
**Problem:** Full codebase scans are slow
**Gap:** Not optimized for checking single new file
**Solution:** Incremental indexing and targeted comparison

## Proposed: Custom Solution for AI Agents

### Architecture

```
┌────────────────────────────────────────────────────┐
│                   AI Agent                         │
└─────────────────────┬──────────────────────────────┘
                      ↓
┌────────────────────────────────────────────────────┐
│              MCP Server (Custom)                   │
│  ┌──────────────────────────────────────────────┐ │
│  │  Tool: check_duplicate_code                  │ │
│  │  Tool: find_similar_functions                │ │
│  │  Tool: search_implementation                 │ │
│  └──────────────────────────────────────────────┘ │
└─────────────────────┬──────────────────────────────┘
                      ↓
         ┌────────────┴────────────┐
         ↓                         ↓
┌─────────────────┐       ┌────────────────┐
│  Text Search    │       │  Semantic      │
│  (dupl/jscpd)   │       │  Search (ML)   │
└─────────────────┘       └────────────────┘
         ↓                         ↓
┌─────────────────────────────────────────┐
│         Repository + Index              │
└─────────────────────────────────────────┘
```

### Implementation Steps

1. **Index Phase:**
   - Parse all Go/TypeScript files
   - Extract functions, types, interfaces
   - Generate embeddings for semantic search
   - Store in vector database

2. **Query Phase:**
   - Agent asks: "Does function X already exist?"
   - System searches by:
     - Signature matching
     - AST structure comparison
     - Semantic similarity (embeddings)
   - Returns matches with confidence scores

3. **Integration Phase:**
   - Provide as MCP tools
   - Add to pre-commit hooks
   - Include in CI/CD pipeline

### Proof of Concept

A minimal implementation could use:
- **Backend:** dupl for Go, jscpd for TypeScript
- **API:** Simple HTTP server
- **MCP:** Wrapper that exposes as tools
- **Storage:** SQLite for function index

## Conclusion

### Best Practices for AI Agents

1. **Use multiple tools:**
   - dupl/jscpd for exact duplicates
   - Semgrep for pattern matching
   - Manual search for semantic duplicates

2. **Integrate into workflow:**
   - Pre-commit hooks
   - CI/CD checks
   - Real-time editor integration

3. **Set reasonable thresholds:**
   - Balance between false positives and false negatives
   - Different thresholds for different code types

4. **Combine approaches:**
   - Token-based for speed
   - AST-based for accuracy
   - ML-based for semantic understanding

### Immediate Actions

For Go projects:
```bash
go install github.com/mibk/dupl@latest
dupl -threshold 50 ./...
```

For TypeScript projects:
```bash
npm install -g jscpd
jscpd . --format typescript --min-lines 10
```

For both:
```bash
pip install semgrep
semgrep --config=auto .
```

### Future Work

1. Build MCP server for duplicate detection
2. Fine-tune code embedding models for domain-specific duplicates
3. Create agent-specific prompt templates
4. Develop incremental analysis tools
5. Build vector search infrastructure for semantic similarity

## References

- [dupl](https://github.com/mibk/dupl)
- [goclone](https://github.com/hhatto/goclone)
- [jscpd](https://github.com/kucherenko/jscpd)
- [PMD CPD](https://pmd.github.io/)
- [SonarQube](https://www.sonarqube.org/)
- [Semgrep](https://semgrep.dev/)
- [CodeBERT](https://github.com/microsoft/CodeBERT)
- [Sourcegraph](https://sourcegraph.com/)
- [Tree-sitter](https://tree-sitter.github.io/)
