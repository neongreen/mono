# tk-vscode Agent Guidelines

## Notifications

**Never show success notifications in tk-vscode.** Only show warnings and errors.

- Success operations should complete silently
- Only show notifications for failures (errors) or partial success (warnings)
- This keeps the UI clean and reduces notification fatigue

Examples:
- Task created successfully → no notification
- Project created successfully → no notification
- Task moved successfully → no notification
- Task move failed → error notification
- Some tasks moved, some failed → warning notification
