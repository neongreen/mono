# Text Annotator - Configuration Presets

Quick configuration presets you can copy to `~/.text-annotator/config.json` to try different styles and behaviors.

## Minimal Mode

Keyboard-only, no buttons, minimal UI:

```json
{
  "showSaveButton": false,
  "showCancelButton": false,
  "showHintLabel": false,
  "showMenuBarIcon": false,
  "windowHeight": 220,
  "windowPadding": 15,
  "sectionSpacing": 15,
  "closeOnSave": true
}
```

**Usage**: Just type and press Cmd+Enter to save, Escape to cancel.

## Dark Mode

Dark theme for night work:

```json
{
  "backgroundColor": "#1E1E1E",
  "textColor": "#E0E0E0",
  "labelColor": "#FFFFFF",
  "hintColor": "#808080",
  "windowOpacity": 0.95
}
```

## Compact

Smaller window, tighter spacing:

```json
{
  "windowWidth": 400,
  "windowHeight": 250,
  "windowPadding": 12,
  "sectionSpacing": 12,
  "elementSpacing": 6,
  "selectedTextScrollHeight": 50,
  "annotationFieldHeight": 20,
  "headerFontSize": 11,
  "selectedTextFontSize": 10,
  "annotationFontSize": 11,
  "saveButtonWidth": 80,
  "cancelButtonWidth": 70
}
```

## Spacious

Larger window, more breathing room:

```json
{
  "windowWidth": 700,
  "windowHeight": 400,
  "windowPadding": 30,
  "sectionSpacing": 30,
  "elementSpacing": 12,
  "selectedTextScrollHeight": 120,
  "annotationFieldHeight": 30,
  "headerFontSize": 14,
  "selectedTextFontSize": 13,
  "annotationFontSize": 15,
  "hintFontSize": 11
}
```

## Quick Capture

Fast workflow, auto-close disabled for rapid annotations:

```json
{
  "closeOnSave": false,
  "clipboardCaptureDelay": 0.05,
  "autoFocusAnnotationField": true,
  "validateEmptyAnnotation": false
}
```

**Usage**: Window stays open after saving, allowing you to quickly add multiple annotations.

## Translucent Overlay

Semi-transparent floating window:

```json
{
  "windowOpacity": 0.85,
  "windowLevel": "floating",
  "backgroundColor": "#000000"
}
```

## No Menu Bar

Hotkey-only mode:

```json
{
  "showMenuBarIcon": false,
  "hotkeyEnabled": true
}
```

## Alternative Hotkeys

### Ctrl+Alt+N
```json
{
  "hotkeyKeyCode": 45,
  "hotkeyModifiers": ["control", "option"]
}
```

### Cmd+Space (if not used by Spotlight)
```json
{
  "hotkeyKeyCode": 49,
  "hotkeyModifiers": ["command"]
}
```

### Shift+Cmd+S
```json
{
  "hotkeyKeyCode": 1,
  "hotkeyModifiers": ["command", "shift"]
}
```

## Unix Timestamps

For programmatic processing:

```json
{
  "timestampFormat": "unix",
  "prettyPrintJSON": false
}
```

**Output**:
```json
[{"timestamp":1705327800,"selectedText":"...","annotation":"..."}]
```

## Custom Storage Location

Save to Documents folder:

```json
{
  "storageFilePath": "~/Documents/annotations.json"
}
```

Or a project-specific location:

```json
{
  "storageFilePath": "~/Projects/myproject/.annotations.json"
}
```

## Editable Selected Text

Allow modifying the captured text:

```json
{
  "selectedTextEditable": true,
  "selectedTextScrollHeight": 100
}
```

## Professional

Clean, business-like appearance:

```json
{
  "windowTitle": "Add Note",
  "selectedTextLabel": "Reference Text:",
  "annotationLabel": "Note:",
  "annotationPlaceholder": "Add your note...",
  "saveButtonTitle": "Save",
  "headerFontWeight": "semibold",
  "windowWidth": 600,
  "windowHeight": 320
}
```

## Casual

Friendly, informal labels:

```json
{
  "windowTitle": "Quick Note",
  "selectedTextLabel": "You selected:",
  "annotationLabel": "Your thoughts:",
  "annotationPlaceholder": "What do you think about this?",
  "hintText": "Hit Cmd+Enter when done!",
  "saveButtonTitle": "Got it! ✓"
}
```

## Debug Mode

For troubleshooting:

```json
{
  "debugMode": true,
  "logToConsole": true
}
```

Check Console.app for detailed logs.

## Merge Multiple Presets

You can combine presets. For example, Dark Mode + Compact:

```json
{
  "backgroundColor": "#1E1E1E",
  "textColor": "#E0E0E0",
  "labelColor": "#FFFFFF",
  "hintColor": "#808080",
  "windowWidth": 400,
  "windowHeight": 250,
  "windowPadding": 12,
  "sectionSpacing": 12,
  "elementSpacing": 6,
  "selectedTextScrollHeight": 50
}
```

## Custom Icon

Use different SF Symbols for menu bar:

```json
{
  "menuBarIconName": "text.bubble"
}
```

**Popular choices**:
- `"text.bubble"` - Chat bubble
- `"pencil"` - Pencil
- `"doc.text"` - Document
- `"square.and.pencil"` - Edit icon
- `"bookmark"` - Bookmark

## Experimenting

1. Start with default config
2. Copy a preset to `~/.text-annotator/config.json`
3. Restart the app
4. Adjust individual values as needed
5. Keep your favorite configuration

## Creating Your Own Preset

1. Start with `config.example.json`
2. Modify settings to your liking
3. Test thoroughly
4. Document what makes it special
5. Share with others!

## Tips

- **Backup first**: Keep a copy of your working config
- **Test one change**: Modify one setting at a time
- **Check validity**: Use a JSON validator
- **Mix and match**: Combine elements from different presets
- **Restart required**: Changes take effect after restart
