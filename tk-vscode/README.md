# tk Tasks VS Code extension

This extension displays the tasks returned by `tk ls --json` inside the VS Code explorer.

## Features

- Runs `tk ls --json` in the selected workspace folder (or a configured directory).
- Shows the tasks grouped using the `--group` flag (prefix by default).
- Highlights blocked tasks with a dedicated icon.
- Provides a refresh command to re-run the query on demand.
- Edit task titles via context menu (uses `tk describe`).
- Rotate task status by clicking the status icon on the left (cycles through: next → wip → done → unset).
- Task text color matches the status icon color for easy visual identification.

## Usage

1. Install the extension locally by running `mise run //tk-vscode:install` followed by `mise run //tk-vscode:build`.
2. Use the VS Code `Developer: Install Extension from Location...` command to load the folder.
3. Open a workspace that contains a `tk` database and open the **tk Tasks** view in the Explorer.
4. Optionally adjust the `tk Tasks` settings to change the binary path, working directory, or grouping mode.
5. Right-click on any task to edit its title.
6. Click the status icon on any task to rotate its status.

**Note**: VS Code TreeView does not support word wrapping. Long task titles are truncated with ellipsis; hover over a task to see the full text in the tooltip.

## Development

```bash
mise run //tk-vscode:install
mise run //tk-vscode:build
```

The compiled files are written to the `out` directory. Use `mise run //tk-vscode:watch` for incremental builds during development.

## Testing

Run the build locally to ensure the extension compiles without network failures:

```bash
mise run //tk-vscode:install
mise run //tk-vscode:build
```
