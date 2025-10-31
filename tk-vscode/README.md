# tk Tasks VS Code extension

This extension displays the tasks returned by `tk ls --json` inside the VS Code explorer.

## Features

- Runs `tk ls --json` in the selected workspace folder (or a configured directory).
- Shows the tasks grouped using the `--group` flag (prefix by default).
- Highlights blocked tasks with a dedicated icon.
- Provides a refresh command to re-run the query on demand.

## Usage

1. Install the extension locally by running `npm install` followed by `npm run compile`.
2. Use the VS Code `Developer: Install Extension from Location...` command to load the folder.
3. Open a workspace that contains a `tk` database and open the **tk Tasks** view in the Explorer.
4. Optionally adjust the `tk Tasks` settings to change the binary path, working directory, or grouping mode.

## Development

```bash
npm install
npm run compile
```

The compiled files are written to the `out` directory. Use `npm run watch` for incremental builds during development.

## Testing

Run the build locally to ensure the extension compiles without network failures:

```bash
npm install
npm run compile
```
