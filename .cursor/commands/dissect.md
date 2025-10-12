This command is to run the `dissect` tool on a Go project file or folder and move some of the parts of Go code to improve how AI agents work with the code. 

It should move code to have well-defined modules with a clear semantics.

It should focus on the largest modules first.

It should run a build to validate the changes. 
This should be done after *every* refactored module or significant change, not just at the end.

It should run tests after it's done.

It should run `mise run //project-name:fmt` after refactoring to fix any formatting issues.

## Usage Notes

- Use `dissect move` command for selective extraction of specific functions/types
- Analyze code manually first, then guide dissect with specific extraction targets
- Extract semantically related code together (e.g., all type definitions, all parsing helpers)
- Build validation after each extraction: `mise run //project-name:run --help`
- Focus on clear separation: types.go for definitions, dedicated files for specific functionality