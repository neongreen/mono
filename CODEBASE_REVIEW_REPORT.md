# Monorepo Code Review Report

Date: 2025-10-13

## Executive Summary

This comprehensive review analyzes the neongreen/mono repository, which contains 8 independent projects (6 Go tools, 1 TypeScript library, 1 shared Go library). The codebase demonstrates solid engineering practices overall, with good separation of concerns, consistent tooling via mise, and comprehensive documentation. However, there are opportunities for reducing duplication, improving test coverage, and leveraging existing open-source solutions.

## A. Technical Debt

### 1. Code Duplication Across Projects

#### GitHub API Integration Duplication (HIGH PRIORITY)

**Issue:** Three projects (prrun, want, printpdf) implement their own GitHub API integration logic with overlapping functionality.

**Evidence:**
- `prrun/github.go`: 243 lines - PR release fetching, authentication
- `printpdf/pkg/fetcher/fetcher.go`: 175 lines - GitHub file fetching, authentication
- `want/main.go`: Partial implementation for release downloads
- `lib/ghrelease/ghrelease.go`: 236 lines - Shared library with similar code

**Problems:**
- Authentication logic duplicated 3+ times (GITHUB_TOKEN, MISE_GITHUB_TOKEN, gh CLI)
- Release fetching logic exists in both `prrun/github.go` and `lib/ghrelease`
- GitHub URL parsing scattered across multiple files
- Type conversions required between `prrun.GitHubRelease` and `ghrelease.Release`

**Impact:**
- Bug fixes need to be applied in multiple places
- Inconsistent error messages and behavior
- Maintenance burden increases with each new project

**Recommendation:**
- Consolidate all GitHub API functionality into `lib/ghrelease`
- Add PR release fetching to `lib/ghrelease` (move from prrun)
- Add file fetching capabilities to `lib/ghrelease` (move from printpdf)
- Make all projects depend on the shared library
- Estimated effort: 4-6 hours
- Expected benefit: 30% reduction in GitHub-related code

### 2. Regex Pattern Duplication

**Issue:** Similar regex patterns defined multiple times across projects.

**Evidence:**
```go
// prrun/github.go
re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/pull/(\d+)`)

// printpdf/pkg/fetcher/fetcher.go
var githubBlobRegex = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/blob/([^/]+)/(.+)`)
var githubRawRegex = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/raw/([^/]+)/(.+)`)
var githubPullFileRegex = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/pull/(\d+)/files`)
```

**Recommendation:**
- Create `lib/ghurl` package with all GitHub URL parsing logic
- Export parsed URL structures with clear fields
- Centralize regex compilation (compiled once, not per-use)

### 3. Error Handling Inconsistencies

**Issue:** Error wrapping patterns are inconsistent across projects.

**Evidence:**
- Some functions use `fmt.Errorf("message: %w", err)` (correct)
- Others use `fmt.Errorf("message: %v", err)` (loses error chain)
- No use of custom error types for better error handling
- Error messages inconsistently formatted

**Example from prrun/github.go:**
```go
return nil, fmt.Errorf("failed to create request: %w", err)  // Good
return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)  // Lost context
```

**Recommendation:**
- Establish error wrapping guidelines in AGENTS.md
- Always use `%w` for error wrapping
- Consider using custom error types for common error conditions
- Add context to all errors (e.g., which URL failed, which file)

### 4. Missing Go Context Usage

**Issue:** HTTP requests and long-running operations don't use Go's context package.

**Evidence:**
```go
// lib/ghrelease/ghrelease.go line 98
resp, err := client.Do(req)  // No context timeout
```

**Impact:**
- No request timeouts
- Cannot cancel long-running operations
- No propagation of cancellation signals

**Recommendation:**
- Add context parameters to all long-running functions
- Set reasonable timeouts on HTTP clients
- Propagate context through function call chains
- Estimated effort: 2-3 hours

### 5. Global State and Singletons

**Issue:** Some packages use package-level variables that could cause issues in testing or concurrent usage.

**Evidence:**
```go
// diagram-dsl/src/layout/yoga-engine.ts line 4
let yogaInstance: any = null;

// markdown-format/main.go line 13
var commonAbbreviations = []string{"e.g.", "i.e.", ...}
var sentenceBoundaryRegex = regexp.MustCompile(...)
```

**Recommendation:**
- Make `commonAbbreviations` a const or pass as parameter
- Document that yoga-engine is thread-safe (or make it not use globals)
- Pre-compile regexes at package init if they're truly constant

### 6. Type Safety Issues in TypeScript

**Issue:** Use of `any` type reduces type safety benefits.

**Evidence:**
```typescript
// diagram-dsl/src/layout/yoga-engine.ts
let yogaInstance: any = null;
private yoga: any;
private createYogaNode(node: LayoutNode): any {
```

**Recommendation:**
- Import proper types from yoga-layout package
- Use `unknown` instead of `any` where type is truly unknown
- Add type guards where necessary
- Estimated effort: 1-2 hours

### 7. Inconsistent Function Formatting in prrun

**Issue:** The prrun/github.go file has bizarre formatting that makes it hard to read.

**Evidence:**
```go
// prrun/github.go lines 31-37
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases",
		owner, repo)
	req,
		err := ghrelease.CreateAuthenticatedRequest("GET", apiURL)
	if err !=
		nil {
		return nil, fmt.Errorf("failed to create request: %w",
			err,
		)
```

**Recommendation:**
- Run `gofmt` or `goimports` on the file
- Add pre-commit hooks to enforce formatting
- This appears to be an editor/formatter misconfiguration

### 8. Test Coverage Gaps

**Issue:** Test coverage is inconsistent across projects.

**Evidence:**
- dissect: 17 test files (good coverage)
- prrun: 3 test files
- printpdf: 1 test file (only fetcher package)
- markdown-format: 1 test file + many testdata files (good)
- claude-trace: 0 test files
- want: 0 test files
- diagram-dsl: 3 test files

**Recommendation:**
- Add integration tests for prrun (test actual PR detection)
- Add unit tests for claude-trace storage and discovery logic
- Add tests for want's fulfillment planning
- Add tests for printpdf's converter selection logic

## B. Obvious Bugs

### 1. No Critical Bugs Found

The codebase review did not uncover any obvious critical bugs. The code is generally defensive with good error handling.

### 2. Potential Race Condition in yoga-engine.ts

**Location:** `diagram-dsl/src/layout/yoga-engine.ts` line 4-12

**Issue:** The `yogaInstance` singleton could theoretically have a race condition if `loadYoga()` is called concurrently.

**Code:**
```typescript
let yogaInstance: any = null;

async function loadYoga(): Promise<any> {
  if (!yogaInstance) {  // Not atomic
    const Yoga = await import('yoga-layout');
    yogaInstance = Yoga.default;
  }
  return yogaInstance;
}
```

**Impact:** Low (unlikely in typical usage since yoga is loaded once at startup)

**Fix:**
```typescript
let yogaInstance: any = null;
let yogaPromise: Promise<any> | null = null;

async function loadYoga(): Promise<any> {
  if (!yogaPromise) {
    yogaPromise = import('yoga-layout').then(module => {
      yogaInstance = module.default;
      return yogaInstance;
    });
  }
  return yogaPromise;
}
```

### 3. Missing Error Handling for File Operations

**Location:** Multiple places where file operations don't check for permission errors

**Example:** `diagram-dsl/src/examples/basic.tsx`
```typescript
writeFileSync(join(outputDir, 'basic-flowchart.svg'), svg1);
```

**Issue:** No check if directory exists or is writable before writing.

**Impact:** Low (fails loudly with exception)

**Recommendation:** Add try-catch blocks and provide helpful error messages.

### 4. Regex Compilation Inefficiency

**Location:** `prrun/github.go` line 15

**Code:**
```go
func parsePRURL(prURL string) (*PRInfo, error) {
	re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/pull/(\d+)`)
```

**Issue:** Regex is compiled on every function call instead of once at package init.

**Impact:** Small performance overhead (negligible in this context)

**Fix:** Move to package-level var:
```go
var prURLRegex = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/pull/(\d+)`)
```

## C. Reusable Components

### 1. Highly Reusable: lib/ghrelease

**Status:** Already shared library

**Current Users:** prrun, want

**Potential Users:** printpdf, dissect (for auto-updates)

**Capabilities:**
- GitHub release fetching
- Platform detection
- Asset downloading
- Authentication handling

**Recommendation:** Expand with PR and file fetching capabilities.

### 2. Should Be Extracted: HTTP Client Utilities

**Current Location:** Scattered across projects

**Should Include:**
- Authenticated request creation
- Retry logic
- Timeout configuration
- User-Agent setting
- Rate limit handling

**Suggested Location:** `lib/httpclient` or add to `lib/ghrelease`

### 3. Should Be Extracted: File System Utilities

**Current Duplication:**
- Directory creation with error handling
- Atomic file writes
- Path validation
- Home directory expansion

**Found in:** printpdf, prrun, want, claude-trace

**Suggested Location:** `lib/fsutil`

### 4. Should Be Extracted: Terminal UI Components (from claude-trace)

**Current Location:** `claude-trace/pkg/tui`

**Reusability:** High - other tools could benefit from similar TUI patterns

**Components:**
- Tag toggles
- List navigation
- Text editing
- Status displays

**Potential Users:** Want (for interactive selection), dissect (for preview mode)

### 5. Could Be Extracted: Markdown Processing

**Current Implementations:**
- printpdf: Markdown to PDF
- markdown-format: Markdown reformatting
- want: Uses markitdown

**Opportunity:** Create shared markdown utilities for common operations:
- Parsing with goldmark
- AST traversal helpers
- Common transformations

### 6. Could Be Extracted: Command Execution Helpers

**Found in:**
- dissect/pkg/commands
- printpdf (external tool invocation)
- want (command plan execution)

**Common Patterns:**
- Running external commands with output capture
- Working directory management
- Environment variable handling
- Exit code checking

## D. Opportunities for Open-Source Libraries

### 1. RECOMMENDED: Replace Homegrown GitHub Client with go-github

**Current State:** Custom GitHub API client code in multiple places

**Recommended Library:** [google/go-github](https://github.com/google/go-github)
- Mature, well-maintained
- Complete GitHub API coverage
- Handles pagination, rate limits, authentication
- Type-safe

**Benefits:**
- Remove 500+ lines of custom code
- Automatic handling of API changes
- Better error messages
- Reduced maintenance burden

**Effort:** Medium (3-4 hours to integrate, test)

**ROI:** High - eliminates entire category of bugs and reduces maintenance

### 2. RECOMMENDED: Use cobra more consistently

**Current State:** dissect, claude-trace, want use cobra. Others use custom CLI parsing.

**Library:** [spf13/cobra](https://github.com/spf13/cobra) - already used

**Recommendation:**
- Migrate prrun to cobra (currently uses manual flag parsing)
- Migrate markdown-format to cobra
- Standardize flag naming conventions across all tools

**Benefits:**
- Consistent CLI interface across tools
- Better help text
- Flag validation
- Subcommand support

**Effort:** Low (1-2 hours per tool)

### 3. CONSIDER: Replace Text Measurement with Existing Solution

**Current State:** diagram-dsl uses canvas package for text measurement

**Alternative:** The current approach is actually good. The canvas package is the de facto standard.

**Verdict:** Keep current implementation.

### 4. CONSIDER: Use a Logging Library

**Current State:** Mix of `log/slog`, `fmt.Println`, and `fmt.Fprintf(os.Stderr, ...)`

**Recommended:** Standardize on `log/slog` (already used in dissect)

**Benefits:**
- Structured logging
- Log levels
- Context propagation
- Better observability

**Projects needing migration:**
- prrun
- printpdf
- markdown-format
- claude-trace

**Effort:** Low (2-3 hours total)

### 5. NOT RECOMMENDED: Goldmark Replacement

**Current State:** markdown-format uses goldmark for parsing

**Evaluation:** Goldmark is the best choice. Research in LOSSLESS_ROUNDTRIP.md confirms this.

**Verdict:** Keep goldmark.

### 6. CONSIDER: Test Framework Standardization

**Current State:**
- Go projects: standard `testing` package (good)
- TypeScript: custom test runner

**Recommendation:**
- Consider using a proper test framework for diagram-dsl (Jest, Vitest, or node:test)
- Add test coverage reporting
- Keep Go tests as-is (standard library is sufficient)

### 7. RECOMMENDED: Add Dependency Scanning

**Current Tools:** None

**Recommended:**
- [dependabot](https://github.com/dependabot) - GitHub native, free
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) - Go official tool

**Benefits:**
- Automatic security updates
- Vulnerability notifications
- Dependency version tracking

**Effort:** Very low (1 hour setup)

### 8. CONSIDER: CLI Documentation Generator

**Current State:** README files manually maintained

**Recommendation:**
- For cobra-based tools: Use `cobra-cli` to generate docs
- Auto-generate CLI documentation from code
- Keep CLI help in sync with README

**Benefit:** Documentation stays in sync with code

**Effort:** Low (already using cobra)

### 9. NOT NEEDED: PDF Conversion Library

**Current State:** printpdf uses external tools (Typst, Prince, WeasyPrint)

**Evaluation:** This is the correct approach. PDF generation is complex and external tools are better.

**Verdict:** Keep current implementation.

### 10. CONSIDER: Use go-git for Repository Operations

**Current State:** want and prrun shell out to git commands

**Alternative:** [go-git](https://github.com/go-git/go-git)

**Trade-offs:**
- Pro: Pure Go, no external dependencies
- Con: More complex API
- Con: May not support all git features

**Recommendation:** Only if you need programmatic git operations beyond simple cloning.

**Current Verdict:** Keep current implementation (shelling out is fine for these use cases).

## E. Long-term Codebase Health Recommendations

### 1. Establish Testing Standards (HIGH PRIORITY)

**Problem:** Test coverage varies wildly (0% to good coverage).

**Recommendation:**
- Document testing standards in AGENTS.md
- Require tests for all new features
- Add CI check that fails if coverage drops
- Target: 70% coverage for critical paths

**Template for AGENTS.md:**
```markdown
## Testing Requirements

All projects must include:
- Unit tests for core logic
- Integration tests for external dependencies
- Error path testing
- Minimum 70% coverage for non-UI code
```

### 2. Dependency Management Strategy

**Current State:** No clear policy on dependency updates.

**Recommendation:**
- Enable Dependabot for all projects
- Create DEPENDENCIES.md documenting:
  - How to evaluate new dependencies
  - Security update policy
  - Major version upgrade process
- Review dependencies quarterly

### 3. Create Monorepo Shared Utilities

**Proposed Structure:**
```
lib/
├── ghrelease/    (exists - expand it)
├── githubapi/    (new - use go-github)
├── fsutil/       (new - file system helpers)
├── httputil/     (new - HTTP client helpers)
└── testutil/     (new - shared test helpers)
```

**Benefits:**
- Reduce duplication
- Easier to maintain
- Consistent behavior across tools
- Single place to fix bugs

### 4. Documentation Improvements

**Current State:** Good documentation overall, but could be better.

**Recommendations:**

a) Add CONTRIBUTING.md at root level:
- How to add a new project
- Testing requirements
- Code review process
- Release process

b) Standardize project documentation:
- Every project needs: README, ARCHITECTURE (if complex), TESTING
- Remove unnecessary docs (AUDIT.md, IMPROVEMENTS.md, SUMMARY.md files in diagram-dsl)
- Move implementation summaries to GitHub PR descriptions

c) Create ARCHITECTURE.md at root level:
- Explain monorepo structure
- Shared libraries
- Release workflow
- CI/CD pipeline

### 5. Improve Error Messages

**Current State:** Error messages vary in quality.

**Recommendation:** Establish error message guidelines:

```markdown
## Error Message Guidelines

Good error messages:
1. Say what went wrong
2. Say what was expected
3. Say how to fix it
4. Include relevant context (file path, URL, etc.)

Example:
❌ "failed to fetch release"
✅ "failed to fetch release 'v1.0.0' from github.com/owner/repo: API returned 404. Check that the release exists and you have access to the repository."
```

### 6. Consolidate Build and Run Commands

**Current State:** Using mise for Go projects (good), npm for TypeScript (standard).

**Recommendation:**
- Document that mise is the standard (already in AGENTS.md - good!)
- Ensure all projects have mise.toml with standard tasks
- Consider adding mise support for diagram-dsl (run npm through mise)

### 7. Version Management Strategy

**Current Issue:** No clear versioning strategy documented.

**Recommendation:**
- Document versioning approach in RELEASE_WORKFLOW.md
- Clarify relationship between Git commits and versions
- Explain PR vs main branch releases

### 8. Reduce Documentation Bloat

**Issue:** Some projects have excessive documentation files.

**Example - diagram-dsl has:**
- AUDIT.md (no longer needed - was for verification)
- IMPROVEMENTS.md (should be in git history)
- SEMANTIC_COMPONENTS_SUMMARY.md (should be in git history)
- IMPLEMENTATION_SUMMARY.md (overlaps with ARCHITECTURE.md)

**Recommendation:**
- Keep: README, ARCHITECTURE, DESIGN, TESTING, TROUBLESHOOTING
- Remove: AUDIT, IMPROVEMENTS, SUMMARY files
- Move historical information to git commit messages

### 9. CI/CD Improvements

**Current State:** Good CI setup with test and release workflows.

**Recommendations:**

a) Add dependency caching in all workflows:
```yaml
- uses: actions/setup-go@v5
  with:
    cache-dependency-path: ${{ matrix.project }}/go.sum
```

b) Add workflow status badges to project READMEs

c) Consider matrix builds for Go projects (currently separate workflows)

d) Add automatic changelog generation on releases

### 10. Code Quality Tools

**Recommended Additions:**

a) For Go projects:
- `golangci-lint` (comprehensive linter)
- `govulncheck` (security scanning)
- `gofumpt` (stricter formatting)

b) For TypeScript:
- ESLint (already using via lints?)
- Prettier (format consistency)
- TypeScript strict mode

c) Add pre-commit hooks:
- Format check
- Lint check
- Test run

### 11. Improve Project Status Tracking

**Current State:** README lists projects but doesn't clearly indicate maturity.

**Recommendation:** Add maturity badges to README:

```markdown
## Projects

- **dissect** - Go tool for code refactoring [Production Ready]
- **prrun** - PR binary runner [Production Ready]
- **markdown-format** - Markdown formatting tool [Production Ready]
- **printpdf** - PDF conversion tool [Beta]
- **diagram-dsl** - Diagram generation DSL [Beta]
- **claude-trace** - TUI for reviewing Claude logs [Beta]
- **want** - Interactive task fulfillment [Alpha - Incomplete]
```

### 12. Consider Monorepo-Specific Tools

**Current State:** Using separate CI workflows per project.

**Consider:**
- [Turborepo](https://turbo.build/) - Task caching and orchestration
- [Bazel](https://bazel.build/) - Build system (overkill for this size)
- [Nx](https://nx.dev/) - Monorepo tooling (overkill for this size)

**Verdict:** Current setup is fine. These tools are overkill for 8 small projects.

### 13. Establish Code Review Guidelines

**Current State:** No documented code review process.

**Recommendation:** Add to CONTRIBUTING.md:
- All changes via pull request
- At least one approval required
- Tests must pass
- Documentation must be updated
- Breaking changes require discussion

### 14. Create Issue Templates

**Recommendation:** Add GitHub issue templates:
- Bug report
- Feature request
- Project proposal (for new projects)
- Documentation improvement

### 15. Improve Cross-Project Consistency

**Create style guides:**

a) Go Style Guide:
- Error handling conventions
- Package naming
- Interface design
- Testing patterns

b) TypeScript Style Guide:
- No `any` types
- Explicit return types
- Error handling

c) Documentation Style Guide:
- README structure
- Code comment style
- Commit message format

## Summary of Priorities

### High Priority (Do First)
1. Consolidate GitHub API code into lib/ghrelease using go-github
2. Establish testing standards and add missing tests
3. Fix error handling inconsistencies (use %w everywhere)
4. Add context.Context to long-running operations
5. Enable Dependabot for security updates

### Medium Priority (Do Soon)
1. Extract HTTP client utilities to lib/
2. Migrate all tools to cobra CLI framework
3. Standardize on slog for logging
4. Add golangci-lint to all Go projects
5. Clean up excessive documentation files

### Low Priority (Nice to Have)
1. Add pre-commit hooks
2. Create CONTRIBUTING.md and ARCHITECTURE.md
3. Add workflow status badges
4. Improve error messages across all tools
5. Add changelog generation

## Estimated Effort

| Category | Effort | Impact |
|----------|--------|--------|
| GitHub API consolidation | 6-8 hours | High |
| Testing improvements | 10-15 hours | High |
| Error handling fixes | 3-4 hours | Medium |
| Context propagation | 2-3 hours | Medium |
| Documentation cleanup | 2-3 hours | Medium |
| CI/CD improvements | 4-6 hours | Medium |
| Code quality tools | 3-4 hours | Medium |
| **Total** | **30-43 hours** | - |

## Conclusion

This monorepo is well-structured with good separation of concerns and consistent tooling. The main areas for improvement are:

1. Reducing duplication (especially GitHub API code)
2. Improving test coverage
3. Standardizing error handling
4. Leveraging go-github library

The codebase does not suffer from major architectural issues or technical debt that would require large refactoring. The recommendations focus on incremental improvements that will reduce maintenance burden and improve reliability.

No bad habits are being entrenched - the existing patterns (mise for builds, separate projects, shared libraries) are sound. The main risk is letting duplication grow as new projects are added. Addressing the high-priority items (especially GitHub API consolidation) will establish patterns that prevent future duplication.
