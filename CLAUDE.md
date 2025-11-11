# Claude rules

## Task tracking

Use `tk` for task tracking.

- Create tasks for all work you do
- Always keep status up to date
- Break big tasks into subtasks
- Search for related tasks and mark them as related
- Add notes to tasks as you go
- In commit messages mention which tasks they are related to

There are two versions:
- `tk` is the globally installed binary
- `tk-dev` is an alias that automatically builds and runs tk from the local checkout

**Important:** `tk-dev` Just Works™ - you don't need to run `go build` or anything. Just use `tk-dev` directly as a command, and it will build from source if needed.

When working on `tk` itself, use `tk-dev` to test your changes. Use `tk` for normal task tracking.
