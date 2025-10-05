# Text Annotator - Workflow

## Application Lifecycle

```
┌─────────────────────────────────────────────┐
│         Application Launch                  │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│     AppDelegate.applicationDidFinishLaunching│
├─────────────────────────────────────────────┤
│  • Create menu bar status item              │
│  • Register global hotkey (Cmd+Shift+A)     │
│  • Request accessibility permissions        │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│       Application Ready & Waiting           │
│  [Menu Bar Icon Visible]                    │
└─────────────────────────────────────────────┘
```

## Annotation Workflow

### 1. Trigger Annotation

```
User Selects Text in Any App
           │
           ├─────────────┬──────────────┐
           │             │              │
           ▼             ▼              ▼
    Press Cmd+Shift+A   OR   Click Menu Bar Icon
           │             │              │
           └─────────────┴──────────────┘
                        │
                        ▼
            showAnnotationWindow()
```

### 2. Text Capture Process

```
┌─────────────────────────────────────────────┐
│        getSelectedText()                     │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│  Save Current Clipboard                      │
│  oldContents = pasteboard.string()           │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│  Simulate Cmd+C                              │
│  CGEvent(keyDown: C, flags: .maskCommand)    │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│  Wait for Clipboard Update (0.1s)           │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│  Read New Clipboard                          │
│  selectedText = pasteboard.string()          │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│  Restore Old Clipboard (if different)       │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
          Return selectedText
```

### 3. Window Display

```
┌─────────────────────────────────────────────┐
│  AnnotationWindow(selectedText: text)        │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│         Setup Window UI                      │
├─────────────────────────────────────────────┤
│  ┌─────────────────────────────────────┐   │
│  │ Selected Text:                       │   │
│  │ ┌─────────────────────────────────┐ │   │
│  │ │  [Read-only text display]       │ │   │
│  │ └─────────────────────────────────┘ │   │
│  │                                      │   │
│  │ Your Annotation:                     │   │
│  │ ┌─────────────────────────────────┐ │   │
│  │ │  [Input field - FOCUSED]        │ │   │
│  │ └─────────────────────────────────┘ │   │
│  │ Press Cmd+Enter to save              │   │
│  │                                      │   │
│  │         [Save (⌘↩)]   [Cancel]      │   │
│  └─────────────────────────────────────┘   │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│  Window.makeKeyAndOrderFront()               │
│  [Floating above all windows]                │
└─────────────────────────────────────────────┘
```

### 4. Save Process

```
User Types Annotation
           │
           ▼
  Presses Cmd+Enter (or clicks Save)
           │
           ▼
┌─────────────────────────────────────────────┐
│   Validate Input (not empty)                 │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│   saveToJSON()                               │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│   Load Existing Annotations                  │
│   from ~/.text-annotations.json              │
│   (if file exists)                           │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│   Create New Annotation Object               │
│   {                                          │
│     timestamp: ISO8601 format,               │
│     selectedText: "...",                     │
│     annotation: "..."                        │
│   }                                          │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│   Append to Annotations Array                │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│   Serialize to JSON with Pretty Printing     │
│   JSONSerialization.data(prettyPrinted)      │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│   Write to ~/.text-annotations.json          │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│   Close Window                               │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
      Application Ready (back to waiting)
```

## State Diagram

```
                    ┌──────────────────┐
                    │   Not Running    │
                    └────────┬─────────┘
                             │ Launch
                             ▼
                    ┌──────────────────┐
                    │   Initializing   │
                    │  • Setup UI      │
                    │  • Register HK   │
                    │  • Request Perms │
                    └────────┬─────────┘
                             │
                             ▼
    ┌────────────────────────────────────────────┐
    │                  Idle                       │
    │  [Menu bar icon visible, waiting for input]│
    └───────────┬────────────────────────────────┘
                │                        ▲
                │ Trigger (hotkey/click) │
                ▼                        │
    ┌────────────────────────┐          │
    │  Capturing Text        │          │
    │  [Simulating Cmd+C]    │          │
    └───────────┬────────────┘          │
                │                        │
                ▼                        │
    ┌────────────────────────┐          │
    │  Window Displayed      │          │
    │  [User entering text]  │          │
    └───────────┬────────────┘          │
                │                        │
                ├─ Cancel ──────────────┘
                │
                │ Save
                ▼
    ┌────────────────────────┐
    │  Saving to JSON        │
    │  [Writing file]        │
    └───────────┬────────────┘
                │
                └────────────────────────┘
```

## Error Handling

```
┌─────────────────────────────────────────────┐
│          Error Scenarios                     │
├─────────────────────────────────────────────┤
│                                              │
│  1. No Accessibility Permission              │
│     → Show macOS permission dialog           │
│     → Hotkey registration fails silently     │
│                                              │
│  2. Empty Annotation                         │
│     → Show alert dialog                      │
│     → Don't close window                     │
│                                              │
│  3. No Text Selected                         │
│     → Clipboard remains same                 │
│     → Window shows empty text                │
│                                              │
│  4. JSON File Write Error                    │
│     → Try/catch ignored (silent fail)        │
│     → Print to console                       │
│                                              │
│  5. Clipboard Access Blocked                 │
│     → May get empty or old clipboard         │
│     → Window shows what clipboard had        │
│                                              │
└─────────────────────────────────────────────┘
```

## Performance Considerations

- **Hotkey Registration**: Once at startup, no runtime overhead
- **Text Capture**: ~100ms delay for clipboard to update
- **Window Creation**: Created fresh each time (not cached)
- **JSON File**: Loaded and parsed on each save
- **File Size**: Grows linearly with annotations (no cleanup)

## Thread Safety

- All UI operations on main thread (AppKit requirement)
- Carbon event handler runs on main thread
- File I/O is synchronous on main thread
- No background threads used
