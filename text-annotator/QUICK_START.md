# Text Annotator - Quick Start Guide

Get started in 3 steps:

## 1. Build and Run

```bash
cd text-annotator
swift build -c release
.build/release/TextAnnotator
```

## 2. Grant Permissions

When prompted, grant Accessibility permissions in System Settings.

## 3. Start Annotating

1. Select text anywhere
2. Press **Cmd+Shift+A**
3. Type your annotation
4. Press **Cmd+Enter** to save

Your annotations are saved to `~/.text-annotations.json`

---

## Customization in 60 Seconds

### Step 1: Copy example config

```bash
cp config.example.json ~/.text-annotator/config.json
```

### Step 2: Edit settings

```bash
nano ~/.text-annotator/config.json
```

### Step 3: Restart app

Changes take effect immediately on restart.

---

## Quick Customizations

### Change Hotkey to Ctrl+Alt+A

Edit `~/.text-annotator/config.json`:

```json
{
  "hotkeyKeyCode": 0,
  "hotkeyModifiers": ["control", "option"]
}
```

### Enable Dark Mode

```json
{
  "backgroundColor": "#1E1E1E",
  "textColor": "#E0E0E0"
}
```

### Make Window Larger

```json
{
  "windowWidth": 700,
  "windowHeight": 400
}
```

### Hide Buttons (Keyboard Only)

```json
{
  "showSaveButton": false,
  "showCancelButton": false
}
```

### Change Storage Location

```json
{
  "storageFilePath": "~/Documents/my-notes.json"
}
```

---

## Troubleshooting

### Hotkey Not Working?

1. Check Accessibility permissions in System Settings
2. Verify no other app uses the same hotkey
3. Enable debug mode: `{"debugMode": true}`

### Window Not Appearing?

1. Look for menu bar icon (note symbol)
2. Click it to trigger manually
3. Check Console.app for errors

### Settings Not Applying?

1. Verify JSON is valid (use jsonlint.com)
2. Check file location: `~/.text-annotator/config.json`
3. Restart the application
4. Enable debug mode to see loading messages

---

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| **Cmd+Shift+A** | Open annotation window (global) |
| **Cmd+Enter** | Save annotation and close |
| **Escape** | Cancel without saving |

---

## View Your Annotations

```bash
# View all annotations
cat ~/.text-annotations.json

# Pretty print
cat ~/.text-annotations.json | python3 -m json.tool

# Search for keyword
cat ~/.text-annotations.json | jq '.[] | select(.annotation | contains("important"))'

# Count annotations
cat ~/.text-annotations.json | jq '. | length'
```

---

## Next Steps

📖 **Read Documentation**:
- [CONFIGURATION.md](CONFIGURATION.md) - All 60+ settings explained
- [PRESETS.md](PRESETS.md) - Copy-paste configurations
- [CUSTOMIZATION_EXAMPLES.md](CUSTOMIZATION_EXAMPLES.md) - Visual examples

🎨 **Try Presets**:
- Minimal mode (no buttons)
- Dark theme
- Compact layout
- Quick capture mode

⚙️ **Advanced**:
- Custom timestamp formats
- Project-specific storage
- Keyboard-only workflow
- Color themes

---

## Tips

✅ Keep a backup of working config before experimenting  
✅ Change one setting at a time when testing  
✅ Use debug mode to understand behavior  
✅ Check Console.app if something breaks  
✅ SF Symbols app has 5000+ icons for menu bar  

---

## Quick Reference

**Config file**: `~/.text-annotator/config.json`  
**Annotations**: `~/.text-annotations.json` (customizable)  
**Default hotkey**: Cmd+Shift+A (customizable)  
**Debug logs**: Console.app → search "TextAnnotator"

---

## Examples

### Professional Setup

```json
{
  "windowTitle": "Add Note",
  "selectedTextLabel": "Reference:",
  "annotationLabel": "Note:",
  "windowWidth": 600,
  "headerFontWeight": "semibold"
}
```

### Speed Setup

```json
{
  "clipboardCaptureDelay": 0.05,
  "closeOnSave": true,
  "validateEmptyAnnotation": false,
  "showHintLabel": false
}
```

### Minimal Setup

```json
{
  "showSaveButton": false,
  "showCancelButton": false,
  "showMenuBarIcon": false,
  "windowHeight": 220
}
```

---

## Getting Help

🐛 **Issues**: Check Console.app with debug mode enabled  
📚 **Documentation**: Read CONFIGURATION.md for all options  
💡 **Examples**: See CUSTOMIZATION_EXAMPLES.md for ideas  

---

**You're all set! Happy annotating! 📝**
