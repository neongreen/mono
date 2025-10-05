# Plan: Prevent Touching Files with `init` Functions and Other Declarations

## Problem Statement
The tool should skip files if they contain an `init` function or any declarations other than function, type, and import declarations. 

## user feedback

such a simple task. taking you so long. you are changing so many things. look at `jj show --git --no-pager`. so many things for such a simple task. stop. you had to implement ONE FUNCTION to 
  detect if a file must not be touched. you implemented it. then you should've ran the existing tests. made sure they still passed. then made a few test files. ran the tool and checked if the 
  result looks ok. then implemented a .toml file. that's all

## Proposed Solution

todo

## Testing
*   Create a test Go file with an `init` function and verify that the operation is skipped.
*   Create a test Go file with a type definition and verify that the extraction proceeds as expected.
*   Create a test Go file with a global variable and verify that the operation is skipped.
*   Create a test Go file with only function declarations and imports and verify that the extraction proceeds as expected.
* create a .toml test file
