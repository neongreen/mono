This command is for performing the requests related to Mise and especially Mise tasks.

For example, if the user asks to add a new task:

> Add cloc --by-file --include-lang=Go .

You should add the task to the top-level `mise.toml` file in the repository root.

All project tasks should use the project-name prefix format: `project-name:task-name`.
For example:
- `conf:test` for running conf tests
- `tk:build` for building tk
- `mdbook-comments:dev` for running mdbook-comments dev server

You should guess the name of the task, or propose several options if not obvious.
If you propose options:
- Don't add the task until the user picks the name.
- After the user picks the name, update this file (mise.md) with the command, proposed names, and user's decision.
This way you can become smarter over time.

## Syntax Notes:
- Use `"""` syntax for multi-line commands instead of escaped quotes
- Example: `run = """command with "quotes" and && operators"""` instead of `run = "command with \"quotes\" and && operators"`
