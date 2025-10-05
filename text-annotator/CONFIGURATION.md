# Text Annotator - Configuration Guide

The text-annotator utility is highly customizable through a JSON configuration file. This allows you to adjust behavior, appearance, and UI without modifying code.

## Configuration File Location

The configuration file is automatically created at:
```
~/.text-annotator/config.json
```

On first launch, the application creates this file with default settings if it doesn't exist.

## Quick Start

1. **Copy the example config**:
   ```bash
   cp config.example.json ~/.text-annotator/config.json
   ```

2. **Edit the configuration**:
   ```bash
   nano ~/.text-annotator/config.json
   ```

3. **Restart the application** to apply changes

## Configuration Sections

### Hotkey Settings

Control the global keyboard shortcut:

```json
{
  "hotkeyEnabled": true,
  "hotkeyKeyCode": 0,
  "hotkeyModifiers": ["command", "shift"]
}
```

- **hotkeyEnabled**: Enable/disable global hotkey (default: `true`)
- **hotkeyKeyCode**: Virtual key code (default: `0` = A key)
- **hotkeyModifiers**: Array of modifiers (options: `"command"`, `"shift"`, `"option"`, `"control"`)

**Common Key Codes**:
- A: `0` (0x00)
- S: `1` (0x01)
- D: `2` (0x02)
- Space: `49` (0x31)
- See [Apple's Key Codes](https://eastmanreference.com/complete-list-of-applescript-key-codes) for more

**Example Hotkey Combinations**:
```json
// Cmd+Shift+A (default)
{"hotkeyKeyCode": 0, "hotkeyModifiers": ["command", "shift"]}

// Ctrl+Alt+N
{"hotkeyKeyCode": 45, "hotkeyModifiers": ["control", "option"]}

// Cmd+Space
{"hotkeyKeyCode": 49, "hotkeyModifiers": ["command"]}
```

### Menu Bar Settings

Customize the menu bar icon:

```json
{
  "showMenuBarIcon": true,
  "menuBarIconName": "note.text"
}
```

- **showMenuBarIcon**: Show/hide menu bar status item (default: `true`)
- **menuBarIconName**: SF Symbol name (default: `"note.text"`)

**Popular SF Symbol Icons**:
- `"note.text"` - Note with text
- `"text.bubble"` - Chat bubble
- `"pencil"` - Pencil
- `"square.and.pencil"` - Square with pencil
- `"doc.text"` - Document
- `"bookmark"` - Bookmark

Browse all at: [SF Symbols App](https://developer.apple.com/sf-symbols/)

### Window Settings

Control window appearance and behavior:

```json
{
  "windowWidth": 500,
  "windowHeight": 300,
  "windowTitle": "Add Annotation",
  "windowLevel": "floating",
  "windowCentered": true,
  "windowOpacity": 1.0
}
```

- **windowWidth**: Window width in pixels (default: `500`)
- **windowHeight**: Window height in pixels (default: `300`)
- **windowTitle**: Title bar text (default: `"Add Annotation"`)
- **windowLevel**: Window stacking level (options: `"normal"`, `"floating"`, `"modalPanel"`)
- **windowCentered**: Center window on screen (default: `true`)
- **windowOpacity**: Transparency 0.0-1.0 (default: `1.0`)

### UI Behavior Settings

Control how the UI behaves:

```json
{
  "autoFocusAnnotationField": true,
  "closeOnSave": true,
  "closeOnCancel": true,
  "validateEmptyAnnotation": true,
  "restoreClipboard": true,
  "clipboardCaptureDelay": 0.1
}
```

- **autoFocusAnnotationField**: Auto-focus input field (default: `true`)
- **closeOnSave**: Close window after saving (default: `true`)
- **closeOnCancel**: Close window on cancel (default: `true`)
- **validateEmptyAnnotation**: Prevent saving empty annotations (default: `true`)
- **restoreClipboard**: Restore clipboard after capture (default: `true`)
- **clipboardCaptureDelay**: Wait time for clipboard update in seconds (default: `0.1`)

### Text Display Settings

Customize how selected text is displayed:

```json
{
  "selectedTextLabel": "Selected Text:",
  "selectedTextFontSize": 11,
  "selectedTextScrollHeight": 80,
  "selectedTextEditable": false,
  "selectedTextSelectable": true
}
```

- **selectedTextLabel**: Header text (default: `"Selected Text:"`)
- **selectedTextFontSize**: Font size in points (default: `11`)
- **selectedTextScrollHeight**: Scroll area height in pixels (default: `80`)
- **selectedTextEditable**: Allow editing selected text (default: `false`)
- **selectedTextSelectable**: Allow selecting text (default: `true`)

### Annotation Field Settings

Customize the annotation input field:

```json
{
  "annotationLabel": "Your Annotation:",
  "annotationPlaceholder": "Enter your annotation here...",
  "annotationFieldHeight": 24,
  "annotationFontSize": 13,
  "showHintLabel": true,
  "hintText": "Press Cmd+Enter to save"
}
```

- **annotationLabel**: Field label (default: `"Your Annotation:"`)
- **annotationPlaceholder**: Placeholder text (default: `"Enter your annotation here..."`)
- **annotationFieldHeight**: Input field height in pixels (default: `24`)
- **annotationFontSize**: Font size in points (default: `13`)
- **showHintLabel**: Show keyboard shortcut hint (default: `true`)
- **hintText**: Hint text content (default: `"Press Cmd+Enter to save"`)

### Button Settings

Customize buttons:

```json
{
  "showSaveButton": true,
  "saveButtonTitle": "Save (⌘↩)",
  "saveButtonWidth": 100,
  "showCancelButton": true,
  "cancelButtonTitle": "Cancel",
  "cancelButtonWidth": 80
}
```

- **showSaveButton**: Display save button (default: `true`)
- **saveButtonTitle**: Save button text (default: `"Save (⌘↩)"`)
- **saveButtonWidth**: Button width in pixels (default: `100`)
- **showCancelButton**: Display cancel button (default: `true`)
- **cancelButtonTitle**: Cancel button text (default: `"Cancel"`)
- **cancelButtonWidth**: Button width in pixels (default: `80`)

### Color Settings

Customize colors (use hex format like `"#FF0000"`):

```json
{
  "backgroundColor": null,
  "textColor": null,
  "labelColor": null,
  "hintColor": null
}
```

- **backgroundColor**: Window background (default: `null` = system default)
- **textColor**: Text color (default: `null` = system default)
- **labelColor**: Label color (default: `null` = system default)
- **hintColor**: Hint text color (default: `null` = secondary label color)

**Example Custom Colors**:
```json
{
  "backgroundColor": "#2D2D2D",
  "textColor": "#FFFFFF",
  "labelColor": "#FFCC00",
  "hintColor": "#999999"
}
```

### Spacing Settings

Fine-tune spacing and padding:

```json
{
  "windowPadding": 20,
  "sectionSpacing": 20,
  "elementSpacing": 8,
  "hintSpacing": 4,
  "buttonSpacing": 12
}
```

- **windowPadding**: Padding around window edges (default: `20`)
- **sectionSpacing**: Space between major sections (default: `20`)
- **elementSpacing**: Space between related elements (default: `8`)
- **hintSpacing**: Space above hint text (default: `4`)
- **buttonSpacing**: Space between buttons (default: `12`)

### Font Settings

Customize typography:

```json
{
  "headerFontSize": 12,
  "headerFontWeight": "bold",
  "hintFontSize": 10
}
```

- **headerFontSize**: Header font size in points (default: `12`)
- **headerFontWeight**: Font weight (options: `"regular"`, `"bold"`, `"semibold"`, `"medium"`, `"light"`, `"thin"`, `"heavy"`, `"black"`)
- **hintFontSize**: Hint text font size (default: `10`)

### Storage Settings

Control how annotations are saved:

```json
{
  "storageFilePath": "~/.text-annotations.json",
  "prettyPrintJSON": true,
  "includeTimestamp": true,
  "timestampFormat": "iso8601",
  "customTimestampFormat": null
}
```

- **storageFilePath**: JSON file path (default: `"~/.text-annotations.json"`)
- **prettyPrintJSON**: Format JSON nicely (default: `true`)
- **includeTimestamp**: Add timestamps to annotations (default: `true`)
- **timestampFormat**: Format type (options: `"iso8601"`, `"unix"`, `"custom"`)
- **customTimestampFormat**: Custom format string for `"custom"` type

**Timestamp Examples**:
```json
// ISO 8601 format: "2024-01-15T14:30:00Z"
{"timestampFormat": "iso8601"}

// Unix timestamp: 1705327800
{"timestampFormat": "unix"}

// Custom format: "2024-01-15 14:30"
{"timestampFormat": "custom", "customTimestampFormat": "yyyy-MM-dd HH:mm"}
```

### Permissions Settings

Control permission requests:

```json
{
  "requestAccessibilityOnLaunch": true,
  "showAccessibilityPrompt": true
}
```

- **requestAccessibilityOnLaunch**: Check permissions on startup (default: `true`)
- **showAccessibilityPrompt**: Show macOS permission dialog (default: `true`)

### Advanced Settings

Debug and logging options:

```json
{
  "debugMode": false,
  "logToConsole": true
}
```

- **debugMode**: Enable debug logging (default: `false`)
- **logToConsole**: Print log messages (default: `true`)

## Example Configurations

### Minimal UI

A streamlined interface with no buttons:

```json
{
  "showSaveButton": false,
  "showCancelButton": false,
  "showHintLabel": false,
  "windowHeight": 250,
  "closeOnSave": true
}
```

### Dark Theme

A dark color scheme:

```json
{
  "backgroundColor": "#1E1E1E",
  "textColor": "#FFFFFF",
  "labelColor": "#E0E0E0",
  "hintColor": "#808080"
}
```

### Compact Layout

Smaller window with reduced spacing:

```json
{
  "windowWidth": 400,
  "windowHeight": 250,
  "windowPadding": 12,
  "sectionSpacing": 12,
  "elementSpacing": 6,
  "selectedTextScrollHeight": 60,
  "annotationFieldHeight": 20
}
```

### Alternative Hotkey

Use Ctrl+Alt+A instead:

```json
{
  "hotkeyKeyCode": 0,
  "hotkeyModifiers": ["control", "option"]
}
```

### Custom Storage

Save to a different location:

```json
{
  "storageFilePath": "~/Documents/my-annotations.json"
}
```

## Tips

1. **Test incrementally**: Change one setting at a time
2. **Keep a backup**: Copy your working config before experimenting
3. **Use debug mode**: Enable `debugMode` to see what's happening
4. **Validate JSON**: Use a JSON validator if settings don't apply
5. **Restart required**: Changes require restarting the app

## Troubleshooting

### Settings not applying
- Ensure JSON is valid (no trailing commas, proper quotes)
- Check file location: `~/.text-annotator/config.json`
- Restart the application

### Hotkey not working
- Verify key code is correct
- Check no other app uses the same combination
- Ensure Accessibility permissions are granted

### Colors not showing
- Use hex format: `"#RRGGBB"`
- Set to `null` to use system defaults

### Window too small/large
- Adjust `windowWidth` and `windowHeight`
- Modify spacing values if content doesn't fit

## Resetting to Defaults

Delete the config file to reset:

```bash
rm ~/.text-annotator/config.json
```

On next launch, default settings will be recreated.
