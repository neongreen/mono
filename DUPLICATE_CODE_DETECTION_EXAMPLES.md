# Duplicate Code Detection - Practical Examples

Real examples showing how to use each tool and what output to expect.

## Table of Contents

1. [dupl Examples](#dupl-examples)
2. [jscpd Examples](#jscpd-examples)
3. [Semgrep Examples](#semgrep-examples)
4. [Integration Examples](#integration-examples)

## dupl Examples

### Basic Usage

```bash
# Install
go install github.com/mibk/dupl@latest

# Run on current directory
dupl .

# Set custom threshold (minimum tokens)
dupl -threshold 30 .

# Verbose output
dupl -verbose .

# Only check specific directories
dupl ./src ./pkg
```

### Example Output

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

**Interpretation:**
- 3 clone groups found
- First group: Lines 35-41 in file2.go duplicated on lines 44-50
- Second group: Lines 35-41 in file1.go duplicated on lines 26-32 in file2.go
- Third group: Lines 9-18 duplicated between both files

### Real-World Example

**Before running dupl:**
```go
// file1.go
func ProcessUser(id int) error {
    user, err := GetUser(id)
    if err != nil {
        return err
    }
    if user == nil {
        return errors.New("user not found")
    }
    return user.Process()
}

// file2.go
func HandleUser(id int) error {
    user, err := GetUser(id)
    if err != nil {
        return err
    }
    if user == nil {
        return errors.New("user not found")
    }
    return user.Process()
}
```

**dupl output:**
```
found 2 clones:
  file1.go:9,18
  file2.go:9,18
```

**After refactoring:**
```go
// common.go
func processUserCommon(id int) error {
    user, err := GetUser(id)
    if err != nil {
        return err
    }
    if user == nil {
        return errors.New("user not found")
    }
    return user.Process()
}

// file1.go
func ProcessUser(id int) error {
    return processUserCommon(id)
}

// file2.go
func HandleUser(id int) error {
    return processUserCommon(id)
}
```

### CI/CD Integration

```yaml
# .github/workflows/code-quality.yml
name: Code Quality

on: [push, pull_request]

jobs:
  check-duplicates:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dupl
        run: go install github.com/mibk/dupl@latest
      
      - name: Check for duplicates
        run: |
          RESULT=$(dupl -threshold 50 . | grep "Found total")
          echo "$RESULT"
          if echo "$RESULT" | grep -q "Found total [1-9]"; then
            echo "Warning: Duplicates found!"
            exit 1
          fi
```

## jscpd Examples

### Basic Usage

```bash
# Install
npm install -g jscpd

# Run on current directory
jscpd .

# Specific format
jscpd . --format typescript

# Multiple formats
jscpd . --format typescript,go,javascript

# Custom thresholds
jscpd . --min-lines 5 --min-tokens 50

# Generate HTML report
jscpd . --reporters html,console

# Output to specific directory
jscpd . --output ./reports
```

### Example Output

**Console:**
```
Clone found (typescript):
 - file1.ts [18:12 - 26:26] (8 lines, 67 tokens)
   file2.ts [15:11 - 23:29]

Clone found (typescript):
 - file1.ts [32:4 - 38:2] (6 lines, 53 tokens)
   file2.ts [38:5 - 44:2]

┌────────────┬────────────────┬─────────────┬──────────────┬──────────────┬──────────────────┬───────────────────┐
│ Format     │ Files analyzed │ Total lines │ Total tokens │ Clones found │ Duplicated lines │ Duplicated tokens │
├────────────┼────────────────┼─────────────┼──────────────┼──────────────┼──────────────────┼───────────────────┤
│ typescript │ 2              │ 80          │ 629          │ 2            │ 14 (17.5%)       │ 120 (19.08%)      │
└────────────┴────────────────┴─────────────┴──────────────┴──────────────┴──────────────────┴───────────────────┘
Found 2 clones.
Detection time: 26.74ms
```

**HTML Report:**
The HTML report provides:
- Interactive clone explorer
- Side-by-side comparison
- Syntax highlighting
- Statistics and charts
- Filterable results

### Configuration File

Create `.jscpd.json`:

```json
{
  "threshold": 10,
  "reporters": ["html", "console", "json"],
  "ignore": [
    "**/*.test.ts",
    "**/*.spec.ts",
    "**/node_modules/**",
    "**/dist/**",
    "**/coverage/**"
  ],
  "format": ["typescript", "javascript", "tsx"],
  "minLines": 5,
  "minTokens": 50,
  "output": "./reports/jscpd"
}
```

### Real-World TypeScript Example

**Before:**
```typescript
// file1.ts
async function processUser(userId: number): Promise<User | null> {
  const user = await fetchUser(userId);
  if (!user) {
    throw new Error('User not found');
  }
  return user;
}

// file2.ts
async function handleUser(userId: number): Promise<User | null> {
  const user = await fetchUser(userId);
  if (!user) {
    throw new Error('User not found');
  }
  return user;
}
```

**jscpd finds:**
```
Clone found (typescript):
 - file1.ts [2:1 - 7:2] (5 lines, 45 tokens)
   file2.ts [2:1 - 7:2]

17.5% duplication detected
```

**After refactoring:**
```typescript
// common.ts
async function getUserOrThrow(userId: number): Promise<User> {
  const user = await fetchUser(userId);
  if (!user) {
    throw new Error('User not found');
  }
  return user;
}

// file1.ts
async function processUser(userId: number): Promise<User> {
  return getUserOrThrow(userId);
}

// file2.ts
async function handleUser(userId: number): Promise<User> {
  return getUserOrThrow(userId);
}
```

## Semgrep Examples

### Basic Usage

```bash
# Install
pip install semgrep

# Run with auto config (fetches rules from registry)
# Note: Requires internet access
semgrep --config=auto .

# Run with local config
semgrep --config=semgrep-rules.yaml .

# Run specific rule
semgrep --config=semgrep-rules.yaml --include=duplicate-error-check .

# Different output formats
semgrep --config=semgrep-rules.yaml --json .
semgrep --config=semgrep-rules.yaml --sarif .

# Exclude paths
semgrep --config=semgrep-rules.yaml --exclude='*_test.go' .
```

### Example Rule

```yaml
# semgrep-rules.yaml
rules:
  - id: duplicate-error-handling
    pattern: |
      if err != nil {
        return err
      }
    message: |
      Common error handling pattern. Consider refactoring into helper.
    languages: [go]
    severity: INFO
    
  - id: duplicate-validation
    pattern: |
      if $X == "" {
        return errors.New($MSG)
      }
    message: |
      Empty string validation. Consider validation helper.
    languages: [go]
    severity: WARNING
```

### Example Output

```
┌──────────────────┐
│ 6 Code Findings  │
└──────────────────┘

file1.go
  ❱ duplicate-error-handling
    Common error handling pattern. Consider refactoring into helper.
    
    11┆ if err != nil {
    12┆   return err
    13┆ }

file2.go
  ❱ duplicate-error-handling
    Common error handling pattern. Consider refactoring into helper.
    
    11┆ if err != nil {
    12┆   return err
    13┆ }
    ⋮┆----------------------------------------
    37┆ if err != nil {
    38┆   return err
    39┆ }
    ⋮┆----------------------------------------
    46┆ if err != nil {
    47┆   return err
    48┆ }

┌──────────────┐
│ Scan Summary │
└──────────────┘
✅ Scan completed successfully.
 • Findings: 6 (6 blocking)
 • Rules run: 2
 • Targets scanned: 2
Ran 2 rules on 2 files: 6 findings.
```

### Advanced Pattern Matching

**Detect variations of same pattern:**

```yaml
rules:
  - id: any-error-check
    pattern-either:
      - pattern: |
          if err != nil {
            return err
          }
      - pattern: |
          if err != nil {
            return nil, err
          }
      - pattern: |
          if $ERR != nil {
            log.Error($MSG)
            return $ERR
          }
    message: "Error handling pattern with variations"
    languages: [go]
    severity: INFO
```

**With metavariable constraints:**

```yaml
rules:
  - id: duplicate-string-length-check
    pattern: |
      if len($STR) == 0 {
        $ACTION
      }
    metavariable-pattern:
      metavariable: $ACTION
      patterns:
        - pattern-either:
          - pattern: return $ERR
          - pattern: return errors.New($MSG)
    message: "String length check with error. Consider helper."
    languages: [go]
    severity: INFO
```

### Testing Rules

```bash
# Test a rule before using it
echo 'package main
func test() error {
    if err != nil {
        return err
    }
    return nil
}' | semgrep --config=rule.yaml -
```

## Integration Examples

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Checking for code duplicates..."

# Check Go files
if git diff --cached --name-only | grep -q '\.go$'; then
    if command -v dupl &> /dev/null; then
        echo "Running dupl on Go files..."
        DUPL_OUTPUT=$(dupl -threshold 30 .)
        if echo "$DUPL_OUTPUT" | grep -q "Found total [1-9]"; then
            echo "❌ Duplicates found:"
            echo "$DUPL_OUTPUT"
            echo ""
            echo "Consider refactoring before committing."
            exit 1
        fi
    fi
fi

# Check TypeScript files
if git diff --cached --name-only | grep -q '\.ts$\|\.tsx$'; then
    if command -v jscpd &> /dev/null; then
        echo "Running jscpd on TypeScript files..."
        jscpd . --threshold 20 --format typescript,tsx
        if [ $? -ne 0 ]; then
            echo "❌ High duplication detected in TypeScript"
            exit 1
        fi
    fi
fi

echo "✅ No problematic duplicates found"
exit 0
```

### Makefile

```makefile
.PHONY: check-duplicates check-duplicates-go check-duplicates-ts

check-duplicates: check-duplicates-go check-duplicates-ts

check-duplicates-go:
	@echo "Checking Go code for duplicates..."
	@dupl -threshold 30 . || (echo "Warning: duplicates found" && exit 0)

check-duplicates-ts:
	@echo "Checking TypeScript code for duplicates..."
	@jscpd . --format typescript,tsx --threshold 15 || (echo "Warning: duplicates found" && exit 0)

check-patterns:
	@echo "Checking for duplicate patterns..."
	@semgrep --config=semgrep-ai-patterns.yaml .
```

### Package.json Script

```json
{
  "scripts": {
    "check:duplicates": "jscpd . --format typescript,tsx --threshold 15",
    "check:duplicates:report": "jscpd . --format typescript,tsx --reporters html,console --output ./reports",
    "precommit": "npm run check:duplicates"
  }
}
```

### VSCode Task

Create `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Check Duplicates (Go)",
      "type": "shell",
      "command": "dupl -threshold 30 .",
      "problemMatcher": [],
      "group": {
        "kind": "test",
        "isDefault": false
      }
    },
    {
      "label": "Check Duplicates (TypeScript)",
      "type": "shell",
      "command": "jscpd . --format typescript,tsx",
      "problemMatcher": [],
      "group": {
        "kind": "test",
        "isDefault": false
      }
    },
    {
      "label": "Check All Duplicates",
      "dependsOn": [
        "Check Duplicates (Go)",
        "Check Duplicates (TypeScript)"
      ],
      "problemMatcher": []
    }
  ]
}
```

## Comparison of Outputs

### Finding the Same Duplicate

**Code:**
```go
// file1.go and file2.go both have:
func calculateSum(items []Item) float64 {
    total := 0.0
    for _, item := range items {
        total += item.Price
    }
    return total
}
```

**dupl output:**
```
found 2 clones:
  file1.go:35,41
  file2.go:26,32
```
✅ Found exact duplicate
❌ Doesn't show the actual code
❌ Just line numbers

**jscpd output:**
```
Clone found (go):
 - file1.go [35:1 - 41:2] (6 lines, 45 tokens)
   file2.go [26:1 - 32:2]

10% duplication (6 lines)
```
✅ Found exact duplicate
✅ Shows statistics
⚠️ Still just line numbers in console
✅ HTML report shows actual code

**Semgrep output:**
```
file1.go
  ❱ duplicate-slice-iteration
    Sum calculation over slice.
    
    37┆ for _, item := range items {
    38┆   total += item.Price
    39┆ }
    
file2.go
  ❱ duplicate-slice-iteration
    Sum calculation over slice.
    
    28┆ for _, item := range items {
    29┆   total += item.Price
    30┆ }
```
✅ Shows actual code snippet
✅ Explains what pattern was found
✅ Can detect variations (different variable names)
❌ Requires writing custom rule first

## Best Practices Summary

### Use dupl when:
- ✅ Working with Go exclusively
- ✅ Need fast feedback (CI/CD)
- ✅ Simple command-line usage

### Use jscpd when:
- ✅ Multi-language project
- ✅ Want detailed reports
- ✅ Need to share results with team (HTML)

### Use Semgrep when:
- ✅ Want to prevent specific patterns
- ✅ Need to detect variations
- ✅ Building custom rules for your domain

### Combine them:
```bash
# Quick check
dupl -threshold 30 .

# Detailed analysis
jscpd . --reporters html,console

# Pattern detection
semgrep --config=custom-rules.yaml .
```

## Troubleshooting

### dupl reports too many false positives

**Solution:** Increase threshold
```bash
# Instead of default (15 tokens)
dupl -threshold 30 .

# Or even higher for less sensitive check
dupl -threshold 50 .
```

### jscpd is slow

**Solution:** Limit scope and formats
```bash
# Only check specific formats
jscpd . --format typescript

# Exclude large directories
jscpd . --ignore "**/node_modules/**,**/dist/**"

# Increase minimum tokens
jscpd . --min-tokens 100
```

### Semgrep rule doesn't match

**Solution:** Test the pattern
```bash
# Create test file
cat > test.go << 'EOF'
package main
func test() error {
    if err != nil {
        return err
    }
    return nil
}
EOF

# Test rule
semgrep --config=rule.yaml test.go
```

## Conclusion

All three tools have their place:
- **dupl:** Fast, simple, Go-focused
- **jscpd:** Comprehensive, multi-language, great reports
- **Semgrep:** Powerful, customizable, pattern-focused

Use them together for best results!
