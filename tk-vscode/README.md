# tk Tasks VS Code extension

This extension displays the tasks returned by `tk ls --json` inside the VS Code explorer.

## Features

- Runs `tk ls --json` in the selected workspace folder (or a configured directory).
- Shows the tasks grouped using the `--group` flag (prefix by default).
- Highlights blocked tasks with a dedicated icon.
- Provides a refresh command to re-run the query on demand.
- Edit task titles via context menu (uses `tk describe`).
- Rotate task status using the "Status" button on each task (cycles through: next → wip → done → unset).
- Create new tasks in a group using the add button on group headers.
- Task text color matches the status icon color for easy visual identification.
- **Drag and drop tasks between groups** to move them to different projects (uses `tk mv`).

## Usage

1. Install the extension locally by running `mise run //tk-vscode:install` followed by `mise run //tk-vscode:build`.
2. Use the VS Code `Developer: Install Extension from Location...` command to load the folder.
3. Open a workspace that contains a `tk` database and open the **tk Tasks** view in the Explorer.
4. Optionally adjust the `tk Tasks` settings to change the binary path, working directory, or grouping mode.
5. Click the **Status** button on any task to rotate its status (next → wip → done → unset).
6. Right-click on any task to edit its title.
7. Click the add button on a group header to create a new task in that group.
8. Hover over any task to see its full details in a tooltip, including title, status, and blockers.
9. **Drag and drop** a task onto a different group to move it to that project.

## Design Notes

This extension uses VS Code's native TreeView API, which provides excellent integration with themes, keyboard navigation, accessibility features, and context menus. Like other popular VS Code extensions (GitLens, GitHub Pull Requests, Todo Tree), we accept that TreeView truncates long text - full content is available in tooltips on hover. See [VS Code issue #68806](https://github.com/microsoft/vscode/issues/68806) for background on TreeView limitations.

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
