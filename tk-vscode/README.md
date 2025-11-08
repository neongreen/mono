# tk Tasks VS Code extension

This extension displays the tasks returned by `tk ls --json` inside the VS Code explorer.

## Features

- Runs `tk ls --json` in the selected workspace folder (or a configured directory).
- Shows the tasks grouped using the `--group` flag (prefix by default).
- Highlights blocked tasks with a dedicated icon.
- **Refresh button** in the view toolbar to reload tasks from tk on demand.
- Edit task titles via context menu (uses `tk describe`).
- Rotate task status using the "Status" button on each task (cycles through: next → wip → done → unset).
- Create new tasks in a group using the add button on group headers.
- **Create new projects** using the add button in the view toolbar (uses `tk project create`).
- Task text color matches the status icon color for easy visual identification.
- **Drag and drop tasks between groups** to move them to different projects (uses `tk mv`).
- **Click on any task** to view its details in the Task Details panel, including the full title and notes.
- **Add new notes** to tasks from the Task Details panel with markdown support.
- **Notes are rendered as markdown** with proper formatting for headers, lists, code blocks, links, and more.
- **Search tasks** using Cmd+F (Mac) or Ctrl+F (Windows/Linux) to filter tasks by title or ID. Press Escape to clear the search.
- **Quick task creation** with Cmd+R Cmd+R (Mac) or Ctrl+R Ctrl+R (Windows/Linux). Auto-detects the project from your current selection or prompts you to choose one.

## Usage

1. Install the extension locally by running `mise run tk-vscode:install` followed by `mise run tk-vscode:build`.
2. Use the VS Code `Developer: Install Extension from Location...` command to load the folder.
3. Open a workspace that contains a `tk` database and open the **tk Tasks** view in the Explorer.
4. Optionally adjust the `tk Tasks` settings to change the binary path, working directory, or grouping mode.
5. Click the **Status** button on any task to rotate its status (next → wip → done → unset).
6. Right-click on any task to edit its title.
7. Click the add button on a group header to create a new task in that group.
8. **Click the "Create Project" button** in the view toolbar to create a new project.
9. **Click on any task** to view its details in the **Task Details** panel below the task list. The panel shows the task ID, title (which may be multiline), status, and all notes associated with the task.
10. **Add notes** to tasks directly from the Task Details panel using the "Add Note" button. Notes support full markdown formatting including headers, lists, code blocks, and links.
11. **Notes are rendered as markdown** for rich formatting and better readability.
12. Hover over any task to see its full details in a tooltip, including title, status, and blockers.
13. **Drag and drop** a task onto a different group to move it to that project.
14. **Search tasks** by pressing Cmd+F (Mac) or Ctrl+F (Windows/Linux), or click the search icon in the toolbar. Type to filter tasks by title or ID. The search term is displayed at the top of the tree. Press Escape or click the clear search icon to show all tasks again.
15. **Quick create tasks** by pressing Cmd+R Cmd+R (Mac) or Ctrl+R Ctrl+R (Windows/Linux). The extension will use the project of the currently selected task or group. If nothing is selected, it will show a list of projects to choose from.

## Design Notes

This extension uses VS Code's native TreeView API for the task list, which provides excellent integration with themes, keyboard navigation, accessibility features, and context menus. The Task Details panel uses a WebView to provide a clean, readable display of task information without the truncation limitations of TreeView. This design is similar to other popular VS Code extensions like GitHub Pull Requests and Review.

## Development

```bash
mise run tk-vscode:install
mise run tk-vscode:build
```

The compiled files are written to the `out` directory. Use `mise run tk-vscode:watch` for incremental builds during development.

## Testing

Run the build locally to ensure the extension compiles without network failures:

```bash
mise run tk-vscode:install
mise run tk-vscode:build
```
