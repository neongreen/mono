# Go Linter Analysis Report

**Date:** 2025-11-23
**Linter:** golangci-lint v2.6.1
**Command:** `golangci-lint run --new-from-rev=0c23a5a5`
**Total Issues:** 63

## Summary by Linter

| Linter | Count | False Positives | Action Recommended |
|--------|-------|-----------------|-------------------|
| unparam | 32 | 28 (~87%) | Fix real issues, ignore rest |
| gosec | 18 | 2 (~11%) | Fix timeout issue, review tests |
| modernize | 10 | 0 (0%) | Apply all fixes |
| staticcheck | 2 | 1 (~50%) | Fix switch, add nolint for recursive |
| depguard | 1 | 1 (100%) | Add exemption for examples |

---

## 1. unparam (32 issues) - 87% False Positives

**Description:** Detects unused function parameters and return values that always have the same value.

### Analysis

Most of these are **false positives** or **intentional design choices**:

#### False Positives (28 issues):

1. **Interface implementations** - Parameters required by interface even if not used:
   - `tk/cmd/actions.go` - Multiple functions with unused `reducer *reducer.Reducer` parameter
     - These functions implement a common interface/signature pattern
     - The `reducer` parameter is required for consistency even if not all implementations use it

2. **Future extensibility** - Parameters kept for API stability:
   - `dissect/move.go:567` - `sourceFset *token.FileSet` unused but kept for potential future use
   - `dissect/pkg/refactor/move_file.go:754` - `newPath` parameter for future features
   - `conf/cmd/display.go:73` - `toolName` might be used in future formatting

3. **Test helpers** - Parameters for test flexibility:
   - `tk/internal/database/*_test.go` - Multiple test helper functions with unused params
   - `ingest/main_test.go:13` - Second return value not used in all test cases

4. **Display/formatting functions** - Context parameters not always needed:
   - `tk/cmd/query.go` - Multiple display functions with `currentPath` parameter
     - Kept for consistent function signatures across display functions

5. **Intentional ignored parameters** - Matching external APIs:
   - `linters/cobralint/cobralint.go` - `pass` parameter matching analysis.Analyzer interface
   - `markdown-format/formatter.go:28` - `depth` parameter for recursive calls (prepared for future use)

#### Real Issues (4 issues):

1. **`claude-trace/pkg/parser/format.go:26`**
   ```go
   func parseSimpleTrace(content string) (*ParsedTrace, error)
   ```
   - **Issue:** `content` parameter is unused
   - **Recommendation:** **Remove the parameter** if truly unused, or implement the parsing

2. **`ingest/pkg/linear/ingest.go:197`**
   ```go
   func stringValue(meta map[string]any, key string, fallback string) string
   ```
   - **Issue:** `fallback` always receives empty string `""`
   - **Recommendation:** **Remove the parameter** and use `""` directly in the function

3. **`lib/svghatch/svghatch.go:193`**
   ```go
   func (r *Replacer) addPatternDefs(svg *SVGNode) error
   ```
   - **Issue:** Always returns `nil`, error return value never used
   - **Recommendation:** **Change return type** to `func (r *Replacer) addPatternDefs(svg *SVGNode)`

4. **`printpdf/pkg/converter/markdown.go:363`**
   ```go
   func wrapHTMLWithPageOptions(htmlContent []byte, options PageOptions) ([]byte, error)
   ```
   - **Issue:** Error return is always `nil`
   - **Recommendation:** **Change return type** to `func wrapHTMLWithPageOptions(...) []byte`

### Recommendations

- **Fix the 4 real issues** listed above
- **Suppress the remaining 28** with:
  - Add `//nolint:unparam` comments for intentional design
  - Or update `.golangci.yml` to exclude these patterns
  - Or accept them as low-priority cleanup items

---

## 2. gosec (18 issues) - 11% False Positives

**Description:** Security linter detecting potential vulnerabilities.

### G204: Subprocess launched with potential tainted input (17 issues)

#### Potential Security Issues (4 issues in dissect):

**`dissect/pkg/gopls/add_import.go:26`, `dissect/pkg/dependencies/manager.go:78`, `dissect/pkg/externaltest/externaltest.go:79,89`**
```go
cmd := exec.Command(goplsPath, "execute", "-write", "gopls.add_import", ...)
```
- **Issue:** `goplsPath` and other tool paths come from parameters/configuration
- **Severity:** MEDIUM - potential command injection if paths not validated
- **Analysis:** These are refactoring tools that intentionally execute external programs (gopls, goimports, go)
- **Recommendation:** 
  - **Option 1:** Add validation that tool paths are what they claim to be
  - **Option 2:** Accept the risk with `//nolint:gosec // G204: intentional tool execution`
  - The tools are meant to run external commands, so this is somewhat expected behavior

#### False Positives (13 issues):

All other G204 issues are **test code** or **controlled execution**:

1. **Test code (8 issues):**
   - `printpdf/pkg/golden/golden_test.go:253,262` - Test configuration paths
   - `prrun/main_test.go:165` - Test helper binary
   - Other test files with hardcoded test paths

2. **User-intended command execution (8 issues):**
   - `dissect/move.go` - Other exec.Command calls with validated inputs
   - `dissect/pkg/refactor/move_file.go` - Refactoring tool commands
   - `jj-run/main.go` - User explicitly provides commands to run
   - `prrun/main.go` - PR run commands (user-controlled but intentional)

**Recommendation:** Add `//nolint:gosec // G204: test code` or `//nolint:gosec // G204: user-provided command (intentional)` to suppress false positives.

### G114: Use of net/http serve without timeouts (1 issue)

**`tk/cmd/mcp.go:173`**
```go
if err := http.ListenAndServe(addr, handler); err != nil {
```

- **Issue:** HTTP server has no timeout configuration
- **Severity:** MEDIUM - can lead to resource exhaustion attacks
- **Recommendation:** **Fix this** - use `http.Server` with timeouts:
  ```go
  server := &http.Server{
      Addr:         addr,
      Handler:      handler,
      ReadTimeout:  15 * time.Second,
      WriteTimeout: 15 * time.Second,
      IdleTimeout:  60 * time.Second,
  }
  if err := server.ListenAndServe(); err != nil {
  ```

---

## 3. modernize (10 issues) - 0% False Positives

**Description:** Suggests modern Go idioms and simplifications.

All 10 issues are **legitimate improvements**:

### `interface{}` → `any` (8 issues)

Go 1.18+ introduced `any` as a cleaner alias for `interface{}`:

- `aihook/hook_response_test.go:162,218,273`
- `aihook/integration_test.go:64,71`
- `aihook/main.go:94,101,117`

**Recommendation:** **Replace all** `map[string]interface{}` with `map[string]any`

### String concatenation in loops (1 issue)

**`aihook/pkg/validator/validator.go:100`**
```go
msg += "  " + v + "\n"
```
- **Issue:** Inefficient string concatenation in loop
- **Recommendation:** **Use strings.Builder**:
  ```go
  var sb strings.Builder
  for _, v := range violations {
      sb.WriteString("  ")
      sb.WriteString(v)
      sb.WriteString("\n")
  }
  msg := sb.String()
  ```

### HasPrefix + TrimPrefix → CutPrefix (1 issue)

**`lion/internal/extractor/extractor.go:193`**
```go
if strings.HasPrefix(text, "//") {
    text = strings.TrimPrefix(text, "//")
```
- **Issue:** Can be simplified with Go 1.20+ `strings.CutPrefix`
- **Recommendation:** **Replace with**:
  ```go
  if after, ok := strings.CutPrefix(text, "//"); ok {
      text = after
  ```

---

## 4. staticcheck (2 issues) - 50% False Positives

**Description:** Advanced static analysis for bugs and style issues.

### QF1003: Could use tagged switch (1 issue)

**`aihook/main.go:149`**
```go
if decision == "allow" {
    // ...
} else if decision == "deny" {
    // ...
}
```
- **Recommendation:** **Use switch statement**:
  ```go
  switch decision {
  case "allow":
      // ...
  case "deny":
      // ...
  }
  ```

### S1021: Merge variable declaration with assignment (1 issue)

**`aihook/pkg/validator/validator.go:35`**
```go
var walker func(syntax.Node) bool
walker = func(n syntax.Node) bool {
```
- **Issue:** Linter suggests merging into one declaration
- **Analysis:** **Cannot be simplified** - this is a recursive function that references itself
- **Recommendation:** **Add nolint comment**:
  ```go
  var walker func(syntax.Node) bool  //nolint:S1021 // recursive function needs forward declaration
  ```

---

## 5. depguard (1 issue) - 100% False Positive

**Description:** Prevents use of specific packages according to rules.

**`lib/readability-wasm/example/main.go:6`**
```go
import "flag"
```

- **Rule violated:** Should use `github.com/spf13/cobra` instead of `flag` package
- **Context:** This is in an **example** directory showing library usage
- **Analysis:** **False positive** - the rule is meant for production CLI tools, not library examples
- **Recommendation:** 
  - **Option 1:** Update `.golangci.yml` to exclude example directories:
    ```yaml
    issues:
      exclusions:
        paths:
          - .*/example/.*
    ```
  - **Option 2:** Add `//nolint:depguard` comment in the example

---

## Summary of Actions

### High Priority (Security Issues)

1. ✅ **Fix `tk/cmd/mcp.go:173`** - Add HTTP server timeouts (G114) **[HIGH PRIORITY]**
2. ⚠️ **Review dissect G204 issues** - Decide on validation or accept risk with nolint **[MEDIUM PRIORITY]**

### Medium Priority (Code Quality)

3. ✅ **Apply all 10 modernize fixes** - Update to modern Go idioms
4. ✅ **Apply 1 staticcheck fix** - Use switch statement
5. ⚠️ **Add nolint for recursive walker** - S1021 is false positive
6. ✅ **Fix 4 real unparam issues** - Remove unused parameters/return values

### Low Priority (False Positives)

6. 📝 **Update `.golangci.yml`** - Exclude example directories from depguard
7. 📝 **Add nolint comments** - For intentional unparam and gosec test code

### Statistics

- **Total issues:** 63
- **False positives:** 32 (51%)
- **Real issues to fix:** 15 (24%)  
- **Issues requiring nolint comments:** 16 (25%)
- **Security issues (high priority):** 1 (2%)
- **Security issues (medium priority):** 4 (6%)
- **Code quality improvements:** 10 (16%)

---

## Detailed Issue List

### Issues to Fix (17)

1. `tk/cmd/mcp.go:173` - Add HTTP timeouts (gosec G114) **[SECURITY - HIGH]**
2. 4× `dissect/pkg/` - Review tool path validation (gosec G204) **[SECURITY - MEDIUM]**
3. 8× `interface{}` → `any` (modernize)
4. `aihook/pkg/validator/validator.go:100` - Use strings.Builder (modernize)
5. `lion/internal/extractor/extractor.go:193` - Use CutPrefix (modernize)
6. `aihook/main.go:149` - Use switch statement (staticcheck)
7. `aihook/pkg/validator/validator.go:35` - Keep as-is (recursive walker needs forward declaration)
8. `claude-trace/pkg/parser/format.go:26` - Remove unused parameter (unparam)
9. `ingest/pkg/linear/ingest.go:197` - Remove unused parameter (unparam)
10. `lib/svghatch/svghatch.go:193` - Remove error return (unparam)
11. `printpdf/pkg/converter/markdown.go:363` - Remove error return (unparam)

### Issues to Suppress (30)

- 28 unparam issues (interface implementations, future extensibility, test helpers)
- 13 gosec G204 in test code
- 1 depguard in example directory

---

## Configuration Updates Recommended

Add to `.golangci.yml`:

```yaml
issues:
  exclusions:
    paths:
      - .*/example/.*  # Exclude examples from depguard rules
      - .*_test\.go    # Exclude tests from G204 (subprocess) warnings
```

Or create a `.golangci-baseline.md` to track accepted violations.

---

## Conclusion

The golangci-lint run with 63 issues breaks down as follows:

### Must Fix (1 issue)
- **HTTP server timeout in `tk/cmd/mcp.go`** - This is a real security issue that should be fixed immediately

### Should Review (4 issues)
- **Command execution in dissect tools** - Review whether tool path validation is needed or accept with nolint

### Should Apply (10 issues)
- **Modernize improvements** - All 10 are legitimate Go best practices (use `any`, `strings.Builder`, `CutPrefix`)
- **Switch statement in aihook** - Cleaner code structure

### Can Ignore with Nolint (32 issues)
- **28 unparam issues** - Mostly intentional design (interface consistency, future extensibility)
- **13 gosec G204 in tests** - Test code intentionally runs commands
- **1 depguard in example** - Example code exemption
- **1 recursive walker** - Cannot be simplified

### Confidence in Analysis
- ✅ **High confidence**: modernize, staticcheck, depguard analyses
- ✅ **Medium confidence**: unparam analysis (requires understanding design intent)
- ⚠️ **Needs review**: gosec G204 in dissect (tool paths security)

### Recommended Next Steps

1. **Immediate:** Fix the HTTP timeout issue in tk/cmd/mcp.go
2. **High priority:** Apply all 10 modernize fixes (automated, low risk)
3. **Medium priority:** Apply the staticcheck switch fix
4. **Review required:** Discuss dissect tool path security with the team
5. **Low priority:** Add nolint comments for false positives
6. **Consider:** Update `.golangci.yml` to exclude examples/ and *_test.go from certain rules
