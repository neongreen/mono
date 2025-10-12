# Fix: Comment Preservation When Extracting Functions

## Issue
The `dissect move` command was not preserving comments when extracting functions. This resulted in:
- Comments being left orphaned in the source file
- Functions losing their documentation in the target file
- Code quality degradation

## Root Cause
The original implementation was using AST node manipulation across different `token.FileSet` objects. When a function declaration (`*ast.FuncDecl`) was extracted from the temp file (created by gopls) and added to the target file's AST, the associated `Doc` comments existed but their position information was incorrect. This caused `go/format` to place comments in the wrong location (after the function instead of before).

## Solution
Changed the approach from AST node manipulation to text-based extraction:

1. **New helper function**: Created `goutils.ExtractFunctionText()` that:
   - Reads the source file as text
   - Parses it to find the function and its `Doc` comments
   - Uses position information to extract the exact text range (from `Doc.Pos()` to `funcDecl.End()`)
   - Returns the raw text including all comments

2. **Updated move.go**: Modified `moveIdentifier()` to:
   - Extract function text from the temp file created by gopls
   - Append the text directly to the target file
   - Let `goimports` handle import merging and formatting

## Benefits
- Comments are now properly preserved
- Simpler implementation (less AST manipulation)
- Works correctly with all comment styles (single-line, multi-line, doc comments)
- Gopls already handles the extraction correctly, we just needed to preserve its output

## Test Coverage
Added `TestMoveCommandWithComments` that verifies:
- Multi-line comments are preserved
- Single-line comments are preserved
- Comments move with their functions (not left orphaned)
- Comments for other functions remain in the source file
- Code still builds after extraction

All existing tests continue to pass.
