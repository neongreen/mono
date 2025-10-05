# Go Test Files: Package Naming and Location

This document clarifies the rules and conventions for naming packages and locating test files in Go projects.

## Package Naming for Test Files

In Go, test files (those ending with `_test.go`) typically use one of two package naming conventions:

### 1. Same Package Name (White-Box Testing)

Most test files use the same package name as the code they are testing. This approach is used for "white-box" testing, allowing tests to access unexported (private) functions and variables within the package.

**Example:**

```go
// main.go
package main

// main_test.go
package main
```

### 2. Package Name with `_test` Suffix (Black-Box Testing)

For "black-box" testing, where you only want to test the public API of a package, you can use the package name with a `_test` suffix. This simulates how an external package would interact with your code, ensuring your public API is sufficient and well-designed.

**Example:**

```go
// math.go
package math

// math_test.go
package math_test
```

Both approaches are valid and commonly used depending on your testing strategy.

## Location of Test Files

Regardless of the package name used, **test files must always reside in the same directory as the code they are testing.** This is a fundamental requirement of Go's testing framework.

**Incorrect Structure (Will Not Work):**

```
project/
├── math/
│   ├── math.go
│   └── math_test.go     // Correct location
└── other/
    └── math_test.go     // ❌ Incorrect location - go test will not find this
```

**Correct Structure:**

```
math/
├── math.go       // package math
└── math_test.go  // package math_test or package math
```

The `go test` command discovers `*_test.go` files within the current directory, and Go treats each directory as a single package scope for build constraints and import resolution.

## Multiple Test Files and Package Names in the Same Directory

Go has a special rule that allows for exactly two package names within the same directory: the main package name (e.g., `math`) and its corresponding test package name (e.g., `math_test`). This means you can have multiple test files in the same directory, some using the main package name and others using the `_test` suffix.

**Example Structure:**

```
math/
├── math.go            // package math
├── operations.go      // package math
├── math_test.go       // package math_test (or package math)
├── operations_test.go // package math_test (or package math)
└── helpers_test.go    // package math_test (or package math)
```

Go will only complain if you have more than two distinct package names or if non-test files have different package names within the same directory.

## Extracting Test Helper Functions

When extracting test helper functions, you have several options:

### Option 1: Internal Test Package (Recommended)

Create a subdirectory within your package, typically named `internal/testutils` (or similar), to house test utility functions. This approach is clean, follows Go conventions, and clearly separates test utilities from production code.

**Structure:**

```
math/
├── math.go
├── math_test.go
└── internal/
    └── testutils/
        └── helpers.go    // package testutils
```

**Example `helpers.go`:**

```go
// math/internal/testutils/helpers.go
package testutils

import "testing"

func SetupTestData() { /* ... */ }
func AssertEqual(t *testing.T, expected, actual int) { /* ... */ }
```

**Example `math_test.go` using helpers:**

```go
// math_test.go
package math_test

import (
    "testing"
    "your-module/math/internal/testutils" // Adjust import path as needed
)

func TestSomething(t *testing.T) {
    testutils.SetupTestData()
    // ...
    testutils.AssertEqual(t, 1, 1)
}
```

### Option 2: Keep Helpers in Same Test Package

For simpler cases, you can place helper functions directly into separate `.go` files within the same directory as your main test files, ensuring they also use the `_test` package name (or the main package name if doing white-box testing). These files should also end with `_test.go`.

**Structure:**

```
math/
├── math.go
├── math_test.go          // package math_test (main test file)
└── helpers_for_test.go   // package math_test (helper file, note _test.go suffix)
```

**Example `helpers_for_test.go`:**

```go
// math/helpers_for_test.go
package math_test // Or 'package math' if white-box testing

import "testing"

func CommonSetup() { /* ... */ }
func CompareResults(t *testing.T, got, want interface{}) { /* ... */ }
```

This option is simpler for smaller sets of helpers that are tightly coupled to the tests in that specific package.

## Example

Assume two files `cmd/main.go` and `cmd/main_test.go`.

When extracting functions from these files, we would get:

```go
// cmd/validateargs.go
package main

func validateArgs(args []string) error {
    ...
}
```

```go
// cmd/processfile.go
package main

func processFile(filePath string) error {
    ...
}
```

```go
// cmd/findreporoot.go
package main_test

func findRepoRoot() (string, error) {
    ...
}
```

```go
// cmd/verifyfilecontains.go
package main_test
func verifyFileContains(t *testing.T, filePath, content string) error {
    ...
}
```

This means we will have two packages in the `cmd` directory, and both use files without the `_test` suffix, which is not allowed. This violates Go's package rules, as non-test files in the same directory must belong to the same package.

### Our solution

```
cmd/
├── main.go
├── main_test.go
├── validateargs.go      // package main
├── processfile.go       // package main
└── internal/
    └── testutils/
        ├── findreporoot.go      // package testutils
        └── verifyfilecontains.go // package testutils
```
