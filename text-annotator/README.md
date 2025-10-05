# Text Annotator

A simple macOS utility that allows you to quickly annotate selected text from anywhere in the system.

**Note**: This is a macOS-only application that requires macOS 12 or later and can only be built on macOS.

## Features

- **Global Keyboard Shortcut**: Press `Cmd+Shift+A` to activate
- **System-wide**: Works with any application
- **Quick Annotation**: Add your notes to selected text
- **Persistent Storage**: All annotations saved to `~/.text-annotations.json`
- **Easy Save**: Press `Cmd+Enter` to save and close

## How It Works

1. Select any text in any application
2. Press `Cmd+Shift+A` (or click the menu bar icon)
3. A popup window appears showing the selected text
4. Type your annotation in the text field
5. Press `Cmd+Enter` (or click "Save") to save
6. The annotation is appended to `~/.text-annotations.json`

## Requirements

- macOS 12.0 or later
- Xcode or Swift command line tools installed
- This project can only be built on macOS (uses AppKit/Cocoa frameworks)

## Building

On macOS, run:

```bash
cd text-annotator
swift build
```

Or use the build script:

```bash
./build.sh
```

## Running

```bash
swift run
```

Or build and run the release version:

```bash
swift build -c release
.build/release/TextAnnotator
```

## Permissions

The app requires:
- **Accessibility permissions**: To capture selected text and register global hotkeys
- On first run, macOS will prompt you to grant these permissions in System Settings

## Data Storage

Annotations are stored in JSON format at `~/.text-annotations.json`:

```json
[
  {
    "timestamp": "2024-01-01T12:00:00Z",
    "selectedText": "The text you selected",
    "annotation": "Your annotation here"
  }
]
```

## Usage Tips

- The app runs in the menu bar (look for the note icon)
- Click the menu bar icon to manually trigger annotation
- Press `Escape` to cancel without saving
- The window floats above other windows for easy access
