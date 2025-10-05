# Text Annotator - Architecture

## Overview

Text Annotator is a native macOS utility built with Swift and AppKit. It uses system-level APIs to provide global keyboard shortcut functionality and text capture capabilities.

## Components

### 1. AppDelegate.swift

The main application controller that manages:

#### Status Bar Integration
- Creates a menu bar status item with a note icon
- Provides manual trigger point for annotation

#### Global Hotkey Registration
- Uses Carbon Events API to register `Cmd+Shift+A`
- Installs event handler to respond to hotkey presses
- Uses `EventHotKeyID` with signature "TXAN" (Text Annotator)

#### Text Capture
- Implements system-wide text selection capture
- Saves and restores clipboard to avoid interference
- Uses `CGEvent` to simulate `Cmd+C` keyboard event

#### Accessibility Permissions
- Requests permissions on first launch using `AXIsProcessTrustedWithOptions`
- Required for both hotkey registration and clipboard simulation

### 2. AnnotationWindow.swift

The popup window interface that provides:

#### Window Configuration
- Modal-style window with `.floating` level (stays on top)
- Fixed size (500x300) with centered positioning
- Standard window controls (close, minimize)

#### UI Components
- **Text Display**: Read-only scrollable view for selected text
- **Annotation Field**: Single-line text input with auto-focus
- **Buttons**: Save and Cancel with keyboard shortcuts
- **Helper Text**: Shows keyboard shortcut hint

#### Data Persistence
- Implements JSON file operations
- Loads existing annotations from `~/.text-annotations.json`
- Appends new annotations to array
- Uses `JSONSerialization` for encoding/decoding

#### Keyboard Handling
- Implements `NSTextFieldDelegate` to intercept `Cmd+Enter`
- Maps `Escape` key to cancel action
- Handles validation before saving

## Technologies Used

### Apple Frameworks

- **Cocoa/AppKit**: Native macOS UI framework
- **Carbon Events**: Legacy API for global hotkey registration
- **Accessibility API**: For checking/requesting permissions
- **Core Graphics**: For simulating keyboard events

### Swift Language Features

- **@main**: Application entry point
- **@objc**: Objective-C interoperability for selectors
- **Delegation**: NSWindowDelegate, NSTextFieldDelegate patterns
- **Auto Layout**: Constraint-based UI layout

## Data Flow

```
User Action (Cmd+Shift+A)
    ↓
Carbon Event Handler
    ↓
AppDelegate.showAnnotationWindow()
    ↓
AppDelegate.getSelectedText()
    ├─ Save clipboard
    ├─ Simulate Cmd+C
    ├─ Read clipboard
    └─ Restore clipboard
    ↓
AnnotationWindow(selectedText: ...)
    ↓
User enters annotation + Cmd+Enter
    ↓
AnnotationWindow.saveAnnotation()
    ↓
AnnotationWindow.saveToJSON()
    ├─ Load existing JSON
    ├─ Append new entry
    └─ Write to file
    ↓
Window closes
```

## Security Considerations

### Permissions Required

1. **Accessibility**: Required for:
   - Global hotkey registration
   - Keyboard event simulation
   - System-wide text capture

2. **No Network Access**: App is fully local
3. **No Analytics**: No data collection or telemetry

### Privacy

- All data stored locally in user's home directory
- No cloud sync or external transmission
- User has full control over JSON file

## File System

```
~/.text-annotations.json          # User data (annotations)
TextAnnotator.app/                # Application bundle
  ├── Contents/
  │   ├── MacOS/
  │   │   └── TextAnnotator       # Executable
  │   ├── Info.plist              # App metadata
  │   └── Resources/              # (none currently)
```

## Build System

- **Swift Package Manager**: Used for build configuration
- **Package.swift**: Defines executable target
- **Minimum Platform**: macOS 12.0
- **Build Output**: `.build/release/TextAnnotator`

## Limitations

1. **macOS Only**: Uses platform-specific APIs (Cocoa, Carbon)
2. **Text Selection**: Only works with selectable text
3. **Clipboard-based**: May not work with clipboard-blocking apps
4. **Single Instance**: No mechanism to prevent multiple instances

## Future Enhancements

Potential improvements:

- Search/filter annotations within the app
- Export to different formats (CSV, Markdown)
- Tags or categories for annotations
- Sync with cloud storage
- Configurable hotkey
- Rich text support
- Image capture support
- Custom JSON file location
- Import/merge annotations

## Testing

Since this is a GUI application with system-level integration:

- **Manual Testing**: Required on macOS
- **Permissions**: Test with and without Accessibility access
- **Edge Cases**: 
  - Empty selections
  - Very long text
  - Special characters
  - Clipboard conflicts
  - Multiple rapid activations

## Dependencies

- **No External Dependencies**: Pure Swift + system frameworks
- **Swift 5.9+**: Language requirement
- **macOS 12+**: Platform requirement
