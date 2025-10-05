# Text Annotator - Usage Guide

## Quick Start

1. Build and run the application on macOS
2. Grant Accessibility permissions when prompted
3. Select any text in any application
4. Press `Cmd+Shift+A` to open the annotation window
5. Type your annotation
6. Press `Cmd+Enter` to save

## Detailed Features

### Global Hotkey

The utility registers a system-wide keyboard shortcut `Cmd+Shift+A` that works in any application. When pressed:
1. The currently selected text is captured
2. A floating window appears with the text and an annotation field
3. Focus is automatically set to the annotation input

### Menu Bar Icon

A note icon (📝) appears in your menu bar. Click it to manually trigger the annotation window with currently selected text.

### Annotation Window

The window shows:
- **Selected Text**: Read-only display of the captured text (scrollable)
- **Annotation Field**: Where you type your notes
- **Save Button**: Click or press `Cmd+Enter` to save
- **Cancel Button**: Click or press `Escape` to close without saving

### Keyboard Shortcuts

- `Cmd+Shift+A` - Open annotation window (global)
- `Cmd+Enter` - Save annotation and close window
- `Escape` - Cancel and close window

### Data Storage

All annotations are saved to `~/.text-annotations.json` in your home directory.

#### JSON Format

```json
[
  {
    "timestamp": "2024-01-15T14:30:00Z",
    "selectedText": "The text you selected",
    "annotation": "Your annotation here"
  }
]
```

Each entry contains:
- `timestamp`: ISO 8601 formatted date/time when annotation was created
- `selectedText`: The exact text that was selected
- `annotation`: Your notes about the text

### How Text Capture Works

When you trigger the annotation window:
1. The app saves your current clipboard contents
2. Simulates pressing `Cmd+C` to copy selected text
3. Reads the new clipboard contents
4. Restores your original clipboard (if different)

This ensures the annotation doesn't interfere with your workflow.

## Permissions

### Accessibility Access

The app needs Accessibility permissions to:
- Register and listen for global keyboard shortcuts
- Simulate keyboard events to capture selected text

On first launch, macOS will show a dialog asking you to grant these permissions in System Settings.

To manually check or change permissions:
1. Open **System Settings**
2. Go to **Privacy & Security** → **Accessibility**
3. Ensure **TextAnnotator** is listed and enabled

## Troubleshooting

### Hotkey Not Working

- Check that Accessibility permissions are granted
- Verify no other app is using `Cmd+Shift+A`
- Try clicking the menu bar icon instead

### Selected Text Not Captured

- Ensure the text is actually selectable (not an image)
- Grant Accessibility permissions
- Some apps may block clipboard access

### Window Not Appearing

- Check that the app is running (look for menu bar icon)
- Try using the menu bar icon instead of the hotkey
- Restart the application

## Tips

- The window floats above other windows for easy access
- You can have multiple annotations of the same text with different notes
- Annotations are appended, never overwritten
- Use a text editor or `jq` to view/search your annotations file
- Back up `~/.text-annotations.json` regularly if you collect important notes

## Advanced Usage

### Viewing Annotations

Use `jq` to format and search annotations:

```bash
# Pretty print all annotations
cat ~/.text-annotations.json | jq .

# Search for annotations containing a keyword
cat ~/.text-annotations.json | jq '.[] | select(.annotation | contains("important"))'

# Get annotations from today
today=$(date -u +%Y-%m-%d)
cat ~/.text-annotations.json | jq ".[] | select(.timestamp | startswith(\"$today\"))"

# Count total annotations
cat ~/.text-annotations.json | jq '. | length'
```

### Backing Up

```bash
# Create a timestamped backup
cp ~/.text-annotations.json ~/.text-annotations-$(date +%Y%m%d).json
```

### Integration with Other Tools

Since annotations are stored in JSON, you can easily:
- Import into databases
- Convert to other formats
- Process with scripts
- Sync with cloud storage
- Use with note-taking apps

## Uninstalling

1. Quit the application
2. Remove from your Applications folder
3. Delete `~/.text-annotations.json` if you don't want to keep your annotations
4. Remove Accessibility permissions in System Settings (optional)
