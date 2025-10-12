# dissect Design Document

This document explains the design decisions, implementation approach, and technical details of the `dissect` tool, particularly for the `move` command's support of different Go declaration types.

## Table of Contents

- [Overview](#overview)
- [Implementation Approach](#implementation-approach)
- [Why the Dual Approach](#why-the-dual-approach)
- [gopls Implementation Analysis](#gopls-implementation-analysis)
- [Import Management](#import-management)
- [Known Limitations](#known-limitations)
- [Edge Cases](#edge-cases)
- [Future Improvements](#future-improvements)

## Overview

The `dissect move` command supports moving various Go declarations (functions, types, interfaces, constants, variables) from one file to another. The implementation uses two different approaches depending on the declaration type:

1. **Functions and Methods**: Uses gopls's `refactor.extract.toNewFile` code action
2. **Types, Interfaces, Constants, Variables**: Uses manual AST manipulation

## Implementation Approach

### Functions and Methods (via gopls)

For functions and methods, we leverage gopls's built-in refactoring capability:

```go
// Uses gopls codeaction -kind=refactor.extract.toNewFile
tempFile, err := gopls.ExtractToNewFile(sourceFile, identifier, moduleRoot)
```

**Advantages:**
- Robust, well-tested implementation maintained by the Go team
- Handles complex cases like import management automatically
- Preserves comments and formatting correctly
- Validates refactoring preconditions

**Files:** `pkg/gopls/extract_to_new_file.go`, `cmd/move.go` (moveFunctionWithGopls)

### Types, Interfaces, Constants, Variables (Manual AST)

For other declaration types, we use manual AST manipulation:

```go
// Parse source file, find declaration, serialize it, remove from source, add to target
func moveDeclarationManually(sourceFile, identifier, targetFile string, ...) error
```

**Process:**
1. Parse source file AST
2. Locate the declaration by identifier name
3. Serialize declaration with comments using `go/printer`
4. Remove declaration from source AST
5. Append serialized declaration to target file
6. Run `goimports` on both files to fix imports and formatting

**Files:** `cmd/move.go` (moveDeclarationManually)

## Why the Dual Approach

### gopls Cannot Move Types/Interfaces in Practice

While gopls's implementation **technically supports** extracting types, interfaces, constants, and variables (as seen in `gopls/internal/golang/extracttofile.go`'s `selectedToplevelDecls` function), it **does not offer the code action** for these declaration types in practice.

**Testing confirms this behavior:**

```bash
# Position cursor on a function name
$ gopls codeaction -kind=refactor.extract.toNewFile main.go:5:6
command "Extract declarations to new file" [refactor.extract.toNewFile]

# Position cursor on a type name  
$ gopls codeaction -kind=refactor.extract.toNewFile main.go:3:6
# No output - no code action available
```

**Why?** The likely reason is that gopls only triggers the code action when the cursor is positioned on a function's identifier, not on type/interface/const/var identifiers. This appears to be an intentional design choice in gopls's code action provider, possibly to avoid confusion or because the feature wasn't fully productionized for non-function declarations.

## gopls Implementation Analysis

By examining gopls's source code (`gopls/internal/golang/extracttofile.go`), we can see the sophisticated features it provides:

### 1. Import Management

gopls performs detailed import analysis:

```go
// From gopls source
func findImportEdits(file *ast.File, info *types.Info, start, end token.Pos) 
    (adds, deletes []*ast.ImportSpec, _ error)
```

- Tracks which imports are referenced in extracted vs remaining code
- Adds necessary imports to the new file
- Removes imports from source file that are only used in extracted code
- Handles unparenthesized imports specially (removes entire declaration)
- **Returns error** for dot imports (not supported)

### 2. Comment Association

```go
// Extends selection to include doc comments
if comment != nil && comment.Pos() < start {
    start = comment.Pos()
}
```

- Automatically includes doc comments (`decl.Doc`) with declarations
- Handles trailing whitespace after declarations
- Preserves comment positioning and formatting

### 3. Validation

```go
func selectedToplevelDecls(pgf *parsego.File, start, end token.Pos) 
    (token.Pos, token.Pos, string, bool)
```

- Checks that selection doesn't intersect package declarations
- Validates selection doesn't intersect import statements
- Ensures selection doesn't partially intersect declarations or comments
- Auto-expands selection when only keywords are selected (e.g., just "type" or "const")

### 4. Range Expansion

For better UX, gopls expands the selection intelligently:

```go
// If only selecting keyword "func" or function name, extend to whole function
if posRangeContains(decl.Pos(), decl.Name.End(), start, end) {
    start, end = decl.Pos(), decl.End()
}
```

## Import Management

### Current Implementation (via goimports)

Our manual approach relies on `goimports` for import management:

```go
// After moving declaration, run goimports on both files
commands.RunGoimportsOnFile(targetFile)
commands.RunGoimportsOnFile(sourceFile)
```

**How goimports works:**
- Scans code for unqualified identifiers
- Adds missing imports based on stdlib and local packages
- Removes unused imports
- Formats import blocks

**Advantages:**
- Simple, reliable, widely used
- No need to implement complex import resolution logic
- Handles most cases correctly

**Limitations compared to gopls:**
- Less precise - may add imports that could be avoided
- Doesn't validate dot imports upfront
- No explicit control over which imports move vs stay

## Known Limitations

### 1. Dot Imports

**Issue:** Files with dot imports (e.g., `import . "fmt"`) are processed without upfront validation.

**Behavior:** The move operation succeeds, but goimports won't add dot imports to the target file (it uses standard imports instead). This usually works but changes the code style.

**Example:**
```go
// Before move
import . "fmt"
Println("hello")  // Uses dot-imported identifier

// After move (if type is moved)
// Target file gets: import "fmt"
// Not: import . "fmt"
```

**gopls behavior:** Returns explicit error: `"extract to new file" does not support files containing dot imports`

### 2. Grouped Declarations

**Issue:** When moving one identifier from a grouped declaration, the entire group moves.

**Why:** In Go's AST, grouped declarations are a single `*ast.GenDecl` node:

```go
const (
    Red   = "red"   // \
    Blue  = "blue"  //  > Single GenDecl with multiple ValueSpecs
    Green = "green" // /
)
```

**Behavior:** Moving `Blue` moves all three constants. This is **by design** and documented in tests.

### 3. Type and Method Separation

**Issue:** Moving a type doesn't move its methods; they stay in the source file.

**Why:** Methods are separate `*ast.FuncDecl` nodes in the AST, not part of the type declaration.

**Behavior:** This is **valid Go** - methods can be in different files from their types. The code continues to build correctly.

**Example:**
```go
// Before
type User struct { Name string }
func (u User) SayHello() { ... }

// After moving User
// target.go: type User struct { Name string }
// source.go: func (u User) SayHello() { ... }  // Still valid!
```

### 4. No Partial Declaration Selection

**Issue:** Unlike gopls, we don't validate that operations aren't being performed on partial selections.

**Example:** If a user tries to move only part of a multi-declaration block, the entire block moves (as explained in #2).

**gopls behavior:** Validates and rejects partial selections.

### 5. Cross-file Type Dependencies

**Issue:** If Type A references Type B and both are in the same file, moving only Type A could require Type B in the target file.

**Mitigation:** `goimports` won't add Type B (it's not an import), so this will cause a build error, alerting the user to move both types.

## Edge Cases

The following edge cases are tested in `cmd/move_edge_cases_test.go`:

### 1. Grouped Constants/Variables

```go
const (
    A = 1
    B = 2
)
// Moving B moves entire const block
```

**Test:** `TestMoveEdgeCases/GroupedConstantsBehavior`

### 2. Type Without Methods

```go
type User struct { Name string }
func (u User) Method() {}
// Moving User leaves Method in source (valid Go)
```

**Test:** `TestMoveEdgeCases/TypeWithMethodsSeparation`

### 3. Import Management

```go
import "time"
type Config struct { Timeout time.Duration }
// After move: target gets time import, source loses it
```

**Test:** `TestMoveEdgeCases/ImportManagementWithTypes`

### 4. Dot Imports

```go
import . "fmt"
type T struct{}
// After move: target doesn't get dot import (uses standard import)
```

**Test:** `TestMoveEdgeCases/DotImportHandling`

## Future Improvements

### High Priority

1. **Dot Import Validation**
   - Add upfront check for dot imports
   - Return clear error message like gopls does
   - Prevents silent behavior changes

2. **Method Co-location Warning**
   - Detect when moving a type that has methods
   - Warn user that methods won't be moved
   - Suggest moving methods too

3. **Grouped Declaration Awareness**
   - Detect when identifier is part of grouped declaration
   - Warn user that entire group will move
   - Option to split group first?

### Medium Priority

4. **Partial Selection Validation**
   - Validate selection boundaries
   - Ensure complete declarations are selected
   - Return clear error for invalid selections

5. **Import Precision**
   - Track which imports are actually used in moved code
   - Only move necessary imports (like gopls does)
   - More precise than goimports

### Low Priority

6. **Comment Handling Enhancement**
   - Better handling of inline comments
   - Preserve comment positions more precisely
   - Handle comment groups better

7. **Range Expansion**
   - Auto-expand selection when only keyword is selected
   - Better UX for command-line usage
   - Match gopls behavior

## References

- gopls source: `golang.org/x/tools/gopls/internal/golang/extracttofile.go`
- Go AST documentation: `go/ast` package
- goimports: `golang.org/x/tools/cmd/goimports`

## See Also

- [README.md](README.md) - Usage documentation
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
- [TODO.md](TODO.md) - Known issues and future work
