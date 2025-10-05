# Text Annotator - Customization Examples

This guide shows concrete examples of what you can customize and how the settings affect the app.

## Quick Reference Table

| Category | What You Can Change | Setting Name |
|----------|-------------------|--------------|
| **Hotkey** | Key combination | `hotkeyKeyCode`, `hotkeyModifiers` |
| | Enable/disable | `hotkeyEnabled` |
| **Window** | Size | `windowWidth`, `windowHeight` |
| | Title | `windowTitle` |
| | Transparency | `windowOpacity` (0.0-1.0) |
| | Floating level | `windowLevel` |
| | Position | `windowCentered` |
| | Background color | `backgroundColor` |
| **Menu Bar** | Show/hide icon | `showMenuBarIcon` |
| | Icon symbol | `menuBarIconName` |
| **Behavior** | Close on save | `closeOnSave` |
| | Close on cancel | `closeOnCancel` |
| | Auto-focus input | `autoFocusAnnotationField` |
| | Validate empty | `validateEmptyAnnotation` |
| | Restore clipboard | `restoreClipboard` |
| | Capture delay | `clipboardCaptureDelay` |
| **Text Display** | Label text | `selectedTextLabel` |
| | Font size | `selectedTextFontSize` |
| | Height | `selectedTextScrollHeight` |
| | Editable | `selectedTextEditable` |
| | Selectable | `selectedTextSelectable` |
| | Color | `textColor` |
| **Annotation** | Label | `annotationLabel` |
| | Placeholder | `annotationPlaceholder` |
| | Font size | `annotationFontSize` |
| | Field height | `annotationFieldHeight` |
| | Hint text | `hintText` |
| | Show hint | `showHintLabel` |
| **Buttons** | Show save button | `showSaveButton` |
| | Save button text | `saveButtonTitle` |
| | Save button width | `saveButtonWidth` |
| | Show cancel button | `showCancelButton` |
| | Cancel button text | `cancelButtonTitle` |
| | Cancel button width | `cancelButtonWidth` |
| **Colors** | Window background | `backgroundColor` |
| | Text color | `textColor` |
| | Label color | `labelColor` |
| | Hint color | `hintColor` |
| **Spacing** | Window padding | `windowPadding` |
| | Section spacing | `sectionSpacing` |
| | Element spacing | `elementSpacing` |
| | Hint spacing | `hintSpacing` |
| | Button spacing | `buttonSpacing` |
| **Storage** | File location | `storageFilePath` |
| | Pretty print | `prettyPrintJSON` |
| | Include timestamp | `includeTimestamp` |
| | Timestamp format | `timestampFormat` |
| **Debug** | Debug mode | `debugMode` |
| | Console logging | `logToConsole` |

## Visual Examples

### Default Configuration

```
┌─────────────────────────────────────────┐
│ Add Annotation                          │
├─────────────────────────────────────────┤
│ Selected Text:                          │
│ ┌─────────────────────────────────────┐ │
│ │ The quick brown fox jumps over...  │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ Your Annotation:                        │
│ ┌─────────────────────────────────────┐ │
│ │ [Input field]                       │ │
│ └─────────────────────────────────────┘ │
│ Press Cmd+Enter to save                 │
│                                         │
│                    [Save (⌘↩)] [Cancel] │
└─────────────────────────────────────────┘
```

Size: 500x300
Hotkey: Cmd+Shift+A
Colors: System defaults

### Minimal Mode

```
┌────────────────────────────────┐
│ Quick Note                     │
├────────────────────────────────┤
│ Selected Text:                 │
│ ┌────────────────────────────┐ │
│ │ The quick brown fox...     │ │
│ └────────────────────────────┘ │
│                                │
│ Your Annotation:               │
│ ┌────────────────────────────┐ │
│ │ [Input field]              │ │
│ └────────────────────────────┘ │
│                                │
└────────────────────────────────┘
```

Settings:
```json
{
  "showSaveButton": false,
  "showCancelButton": false,
  "showHintLabel": false,
  "windowHeight": 220
}
```

### Dark Mode

```
┌─────────────────────────────────────────┐
│ Add Annotation                    [○][□]│
├─────────────────────────────────────────┤
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░ Selected Text:                        ░│
│░ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ░│
│░ ┃ The quick brown fox jumps...      ┃ ░│
│░ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ░│
│░                                       ░│
│░ Your Annotation:                      ░│
│░ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ░│
│░ ┃ [Input field]                     ┃ ░│
│░ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ░│
│░ Press Cmd+Enter to save               ░│
│░                                       ░│
│░                  [Save (⌘↩)] [Cancel] ░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
└─────────────────────────────────────────┘
```

Settings:
```json
{
  "backgroundColor": "#1E1E1E",
  "textColor": "#E0E0E0",
  "labelColor": "#FFFFFF",
  "hintColor": "#808080"
}
```

### Compact Mode

```
┌──────────────────────────┐
│ Note                     │
├──────────────────────────┤
│ Selected:                │
│ ┌──────────────────────┐ │
│ │ The quick brown...   │ │
│ └──────────────────────┘ │
│                          │
│ Annotation:              │
│ ┌──────────────────────┐ │
│ │ [Input]              │ │
│ └──────────────────────┘ │
│                          │
│         [Save] [Cancel]  │
└──────────────────────────┘
```

Settings:
```json
{
  "windowWidth": 400,
  "windowHeight": 250,
  "windowPadding": 12,
  "sectionSpacing": 12
}
```

## Hotkey Customization Examples

### Change to Ctrl+Alt+A

```json
{
  "hotkeyKeyCode": 0,
  "hotkeyModifiers": ["control", "option"]
}
```

### Change to Cmd+Space

```json
{
  "hotkeyKeyCode": 49,
  "hotkeyModifiers": ["command"]
}
```

### Change to Shift+Cmd+N

```json
{
  "hotkeyKeyCode": 45,
  "hotkeyModifiers": ["command", "shift"]
}
```

## UI Text Customization

### Professional Style

```json
{
  "windowTitle": "Add Note",
  "selectedTextLabel": "Reference Text:",
  "annotationLabel": "Note:",
  "annotationPlaceholder": "Add your note...",
  "saveButtonTitle": "Save",
  "cancelButtonTitle": "Discard"
}
```

### Casual Style

```json
{
  "windowTitle": "Quick Thought",
  "selectedTextLabel": "You selected:",
  "annotationLabel": "Your thoughts:",
  "annotationPlaceholder": "What do you think?",
  "hintText": "Hit Cmd+Enter when done!",
  "saveButtonTitle": "Got it! ✓"
}
```

## Behavior Customization

### Quick Capture Mode

Stay open for multiple annotations:

```json
{
  "closeOnSave": false,
  "validateEmptyAnnotation": false,
  "clipboardCaptureDelay": 0.05
}
```

### Minimal Interruption

Quick in and out:

```json
{
  "closeOnSave": true,
  "closeOnCancel": true,
  "autoFocusAnnotationField": true,
  "restoreClipboard": true
}
```

## Storage Customization

### Project-Specific Annotations

```json
{
  "storageFilePath": "~/Projects/myproject/.annotations.json"
}
```

### Programmatic Format

```json
{
  "prettyPrintJSON": false,
  "timestampFormat": "unix",
  "includeTimestamp": true
}
```

Output:
```json
[{"timestamp":1705327800,"selectedText":"...","annotation":"..."}]
```

### No Timestamps

```json
{
  "includeTimestamp": false
}
```

Output:
```json
[{"selectedText":"...","annotation":"..."}]
```

## Color Theme Examples

### Solarized Dark

```json
{
  "backgroundColor": "#002B36",
  "textColor": "#839496",
  "labelColor": "#93A1A1",
  "hintColor": "#586E75"
}
```

### Dracula

```json
{
  "backgroundColor": "#282A36",
  "textColor": "#F8F8F2",
  "labelColor": "#BD93F9",
  "hintColor": "#6272A4"
}
```

### Gruvbox Light

```json
{
  "backgroundColor": "#FBF1C7",
  "textColor": "#3C3836",
  "labelColor": "#282828",
  "hintColor": "#7C6F64"
}
```

### Monokai

```json
{
  "backgroundColor": "#272822",
  "textColor": "#F8F8F2",
  "labelColor": "#F92672",
  "hintColor": "#75715E"
}
```

## Font Customization

### Large and Bold

```json
{
  "headerFontSize": 14,
  "headerFontWeight": "bold",
  "selectedTextFontSize": 13,
  "annotationFontSize": 14,
  "hintFontSize": 11
}
```

### Small and Light

```json
{
  "headerFontSize": 11,
  "headerFontWeight": "regular",
  "selectedTextFontSize": 10,
  "annotationFontSize": 11,
  "hintFontSize": 9
}
```

## Window Customization

### Always on Top

```json
{
  "windowLevel": "floating"
}
```

### Normal Window

```json
{
  "windowLevel": "normal"
}
```

### Translucent

```json
{
  "windowOpacity": 0.85,
  "backgroundColor": "#000000"
}
```

### Large and Spacious

```json
{
  "windowWidth": 700,
  "windowHeight": 400,
  "windowPadding": 30,
  "sectionSpacing": 30,
  "selectedTextScrollHeight": 120
}
```

## Menu Bar Icon Options

```json
{"menuBarIconName": "note.text"}        // 📝 Note with lines
{"menuBarIconName": "text.bubble"}      // 💬 Chat bubble
{"menuBarIconName": "pencil"}           // ✏️ Pencil
{"menuBarIconName": "doc.text"}         // 📄 Document
{"menuBarIconName": "bookmark"}         // 🔖 Bookmark
{"menuBarIconName": "square.and.pencil"} // □✏️ Edit icon
```

## Testing Your Configuration

1. Edit `~/.text-annotator/config.json`
2. Quit and restart the app
3. Press your hotkey to test
4. Check Console.app for debug output (if `debugMode: true`)

## Tips for Customization

1. **Start simple**: Change one thing at a time
2. **Keep backups**: Copy working configs before experimenting
3. **Use debug mode**: Enable for troubleshooting
4. **Check validity**: Invalid JSON = defaults used
5. **Mix presets**: Combine elements from PRESETS.md
6. **Measure first**: Note default sizes before changing
7. **Test thoroughly**: Try all workflows with changes

## Common Customization Goals

| Goal | Key Settings |
|------|-------------|
| Faster workflow | `clipboardCaptureDelay: 0.05`, `closeOnSave: true` |
| Less intrusive | `windowOpacity: 0.8`, `showMenuBarIcon: false` |
| More visible | `windowLevel: "floating"`, `windowOpacity: 1.0` |
| Bigger text | `selectedTextFontSize: 14`, `annotationFontSize: 15` |
| Dark theme | Set all color values to dark hex codes |
| Minimal UI | Hide buttons, hint label, reduce spacing |
| Keyboard only | Hide all buttons, rely on Cmd+Enter |
| Project specific | Change `storageFilePath` per project |
