# Plan for Implementing `internal/testutils` Solution

## Objective
Modify the Go Function Mover Tool to correctly place extracted test helper functions into an `internal/testutils` subdirectory and update their package declarations, as per the guidelines in `docs/logic/test-files.md`.
**Status**: Logic implemented, verification pending.

## Current State
- The `DetermineExtractionTarget` function in `pkg/refactor/target.go` already contains logic to identify test files using `fileutils.IsTestFile` and sets `targetDir` to `filepath.Join(filepath.Dir(sourceFilePath), "internal", "testutils")` and `targetPackage` to `"testutils"`.
- The `main.go` orchestrates the `gopls` extraction, followed by moving the file to `targetDir` and updating its package.
**Status**: Done.

## Proposed Changes

1.  **Verify `fileutils.IsTestFile`:**
    *   Ensure `pkg/fileutils/fileutils.go` correctly identifies test files (those ending with `_test.go`).
    *   If not, update `fileutils.IsTestFile` to accurately reflect this.
**Status**: Done.

2.  **Refine `DetermineExtractionTarget` (if necessary):**
    *   The current implementation in `pkg/refactor/target.go` seems to already implement the core logic for test files.
    *   Review the `fileutils.GetPackageDeclaration` call for non-test files to ensure robust error handling or fallback.
**Status**: Done.

3.  **Testing:**
    *   Add a new integration test case to verify the correct behavior for test helper function extraction.
    *   **Test Scenario:**
        *   Create a sample `my_package/my_package_test.go` file with a test helper function (e.g., `func setupTest()`).
        *   Run the `dissect` tool on `my_package_test.go` targeting `setupTest`.
        *   **Assertions:**
            *   Verify that `my_package/internal/testutils/setuptest.go` (or similar name chosen by `gopls`) is created.
            *   Verify that the content of the new file has `package testutils`.
            *   Verify that `setupTest` is removed from `my_package_test.go`.
            *   Verify that the overall project still builds and tests pass.
**Status**: Not Done.

## Implementation Steps

1.  **Read `pkg/fileutils/fileutils.go`**: Confirm the implementation of `IsTestFile` and `GetPackageDeclaration`.
**Status**: Done.

2.  **Create Test Case**: Add a new test case in `cmd/main_test.go` (or a dedicated integration test file) to specifically test the `internal/testutils` extraction.
**Status**: Not Done.

3.  **Run Tests**: Execute `go test ./...` to verify the new functionality and ensure no regressions.
**Status**: Not Done.
