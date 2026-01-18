# Quick Reference - macOS Data Extraction

Fast lookup guide for extracting data from macOS applications.

## Safari

| Data Type | Location | Best Method | Permission |
|-----------|----------|-------------|------------|
| **Bookmarks** | `~/Library/Safari/Bookmarks.plist` | Direct plist | Full Disk Access |
| **Reading List** | `~/Library/Safari/Bookmarks.plist` | Direct plist | Full Disk Access |
| **History** | `~/Library/Safari/History.db` | SQLite | Full Disk Access |
| **Current Tabs** | AppleScript or `~/Library/Safari/LastSession.plist` | AppleScript/plist | Automation/Full Disk |
| **Cloud Tabs** | `~/Library/Safari/CloudTabs.db` | SQLite | Full Disk Access |

**Quick Commands:**
```bash
# Export bookmarks to JSON
plutil -convert json ~/Library/Safari/Bookmarks.plist -o bookmarks.json

# Query history
sqlite3 ~/Library/Safari/History.db "SELECT url, title FROM history_items LIMIT 10"

# Get current tabs (AppleScript)
osascript -e 'tell application "Safari" to get URL of current tab of front window'
```

## Notes

| Data Type | Location | Best Method | Permission |
|-----------|----------|-------------|------------|
| **All Notes** | `~/Library/Group Containers/group.com.apple.notes/NoteStore.sqlite` | AppleScript/JXA | Automation |
| **Note Content** | Same database (protobuf) | AppleScript/JXA | Automation |

**Quick Commands:**
```bash
# Get all notes (JXA)
osascript -l JavaScript -e 'const Notes = Application("Notes"); JSON.stringify(Notes.notes().map(n => ({name: n.name(), body: n.body()})))'
```

**Notes:**
- AppleScript is easier than SQLite parsing
- SQLite stores content as gzip-compressed protobuf
- Database schema changes between macOS versions

## Voice Memos

| Data Type | Location | Best Method | Permission |
|-----------|----------|-------------|------------|
| **Recordings** | `~/Library/Application Support/com.apple.voicememos/Recordings/*.m4a` | File system | Full Disk Access |
| **Metadata** | `~/Library/Application Support/com.apple.voicememos/Recordings/CloudRecordings.db` | SQLite | Full Disk Access |

**Quick Commands:**
```bash
# List recordings directory
ls -lh ~/Library/Application\ Support/com.apple.voicememos/Recordings/

# Query metadata
sqlite3 ~/Library/Application\ Support/com.apple.voicememos/Recordings/CloudRecordings.db "SELECT ZTITLE, ZDURATION FROM ZCLOUDRECORDING"
```

**Notes:**
- No AppleScript support
- Files are named with UUIDs - need database for titles
- Use Apple epoch (2001-01-01) for timestamp conversion

## Reminders

| Data Type | Location | Best Method | Permission |
|-----------|----------|-------------|------------|
| **All Reminders** | Core Data | EventKit or AppleScript | Reminders Access |

**Quick Commands:**
```applescript
# Get all reminders (AppleScript)
osascript -e 'tell application "Reminders" to return name of every reminder'

# Get reminders with details
osascript -e 'tell application "Reminders" to get {name, completed, body} of every reminder of list "Personal"'
```

**Python (EventKit):**
```python
from EventKit import EKEventStore, EKEntityTypeReminder
# See CODE_EXAMPLES.md for full example
```

**Notes:**
- Don't access Core Data directly - use EventKit or AppleScript
- EventKit is more powerful but requires PyObjC

## Calendar

| Data Type | Location | Best Method | Permission |
|-----------|----------|-------------|------------|
| **Events** | Core Data | EventKit or AppleScript | Calendar Access |
| **Calendars** | Core Data | EventKit or AppleScript | Calendar Access |

**Quick Commands:**
```applescript
# Get today's events (AppleScript)
osascript -e 'tell application "Calendar" to return summary of every event whose start date is greater than (current date)'

# Get calendars
osascript -e 'tell application "Calendar" to return name of every calendar'
```

**Python (EventKit):**
```python
from EventKit import EKEventStore, EKEntityTypeEvent
# See CODE_EXAMPLES.md for full example
```

**Notes:**
- Don't access Core Data directly - use EventKit or AppleScript
- EventKit handles recurring events properly
- AppleScript is simpler for basic queries

## Notifications

| Data Type | Location | Best Method | Permission |
|-----------|----------|-------------|------------|
| **Recent** | Unified Logging System | `log show` command | None |
| **History** | Not reliably available | N/A | N/A |

**Quick Commands:**
```bash
# Query notification logs (last hour)
log show --predicate 'subsystem == "com.apple.notificationcenter"' --last 1h

# Stream live notifications
log stream --predicate 'subsystem == "com.apple.notificationcenter"'
```

**Reality Check:**
- ⚠️ **No reliable way to get notification history**
- Notifications are designed to be ephemeral
- Some apps store their own notification data
- Consider real-time monitoring instead

## Extraction Methods Comparison

### Direct Database/Plist Access

**Use when:**
- ✅ Need complete historical data
- ✅ App doesn't need to be running
- ✅ Maximum speed required

**Pros:** Fast, complete access
**Cons:** Requires Full Disk Access, fragile (breaks with updates)

**Example:**
```python
import sqlite3
conn = sqlite3.connect(db_path)
cursor = conn.cursor()
cursor.execute("SELECT * FROM table")
```

### AppleScript/JXA

**Use when:**
- ✅ Official scripting support exists
- ✅ Don't want to deal with Full Disk Access
- ✅ Need simple data extraction

**Pros:** Official API, maintained by Apple, simpler permissions
**Cons:** App must be running, slower, limited API coverage

**Example:**
```bash
osascript -e 'tell application "App" to ...'
osascript -l JavaScript -e 'Application("App")...'
```

### UI Automation

**Use when:**
- ⚠️ Absolute last resort
- ⚠️ No other method works

**Pros:** Can access any UI element
**Cons:** Extremely fragile, slow, breaks easily, requires Accessibility permissions

**Example:**
```applescript
tell application "System Events"
    tell process "App"
        click button "X"
    end tell
end tell
```

### System APIs (EventKit, etc.)

**Use when:**
- ✅ Available for the app (Calendar, Reminders, Contacts)
- ✅ Need robust, maintained solution
- ✅ Can use Swift/Objective-C/PyObjC

**Pros:** Official, most reliable, proper permissions, cross-version compatible
**Cons:** Requires Swift/ObjC or PyObjC, more complex setup

**Example:**
```python
from EventKit import EKEventStore
store = EKEventStore.alloc().init()
```

## Permission Requirements

### Full Disk Access
**Required for:** Direct database/plist access
**Grant in:** System Preferences → Security & Privacy → Privacy → Full Disk Access

### Automation
**Required for:** AppleScript, UI automation
**Grant in:** System Preferences → Security & Privacy → Privacy → Automation

### Accessibility
**Required for:** UI automation via System Events
**Grant in:** System Preferences → Security & Privacy → Privacy → Accessibility

### Calendar/Reminders
**Required for:** EventKit access
**Grant in:** Automatic prompt, or System Preferences → Security & Privacy → Privacy

## Common Gotchas

### Apple Epoch
**Problem:** Timestamps are wrong
**Solution:** Apple uses epoch of 2001-01-01, not Unix epoch (1970-01-01)
```python
from datetime import datetime
apple_epoch = datetime(2001, 1, 1).timestamp()
real_timestamp = apple_epoch + apple_timestamp
```

### Database Schema Changes
**Problem:** SQL queries break after macOS update
**Solution:** Query `sqlite_master` first to check schema
```bash
sqlite3 database.db ".schema"
```

### Protobuf Content
**Problem:** Note content is unreadable binary
**Solution:** Use AppleScript instead, or decode protobuf (complex)

### Empty AppleScript Results
**Problem:** AppleScript returns nothing
**Solution:**
1. Check if app is running
2. Check automation permissions
3. Check if data exists

### Permission Errors
**Problem:** "Operation not permitted"
**Solution:** Grant Full Disk Access or use alternative method (AppleScript)

## Decision Tree

```
Need to extract macOS app data?
│
├─ Is it Safari bookmarks/history?
│  └─ Use direct plist/SQLite access
│     Requires: Full Disk Access
│
├─ Is it Notes?
│  └─ Use AppleScript/JXA
│     Requires: Automation permission
│     Alternative: SQLite (harder)
│
├─ Is it Voice Memos?
│  └─ Use file system + SQLite
│     Requires: Full Disk Access
│     No alternative available
│
├─ Is it Reminders or Calendar?
│  ├─ Can you use Python/Swift?
│  │  └─ Use EventKit (best)
│  │     Requires: App-specific permission
│  └─ Need simple solution?
│     └─ Use AppleScript
│        Requires: Automation permission
│
└─ Is it Notifications?
   └─ Use real-time monitoring
      Historical data not available
```

## One-Liners

**Safari bookmarks count:**
```bash
plutil -p ~/Library/Safari/Bookmarks.plist | grep -c URLString
```

**Last 5 Safari history items:**
```bash
sqlite3 ~/Library/Safari/History.db "SELECT url FROM history_items ORDER BY ROWID DESC LIMIT 5"
```

**Count of notes:**
```bash
osascript -e 'tell application "Notes" to count notes'
```

**Voice memo count:**
```bash
ls ~/Library/Application\ Support/com.apple.voicememos/Recordings/*.m4a 2>/dev/null | wc -l
```

**Today's calendar events:**
```bash
osascript -e 'tell application "Calendar" to return summary of every event of calendar "Work" whose start date is greater than (current date)'
```

**Incomplete reminders:**
```bash
osascript -e 'tell application "Reminders" to return name of every reminder whose completed is false'
```

## Resource Links

**Official Apple Documentation:**
- EventKit: https://developer.apple.com/documentation/eventkit
- AppleScript: https://developer.apple.com/library/archive/documentation/AppleScript/Conceptual/AppleScriptLangGuide/
- JXA: https://developer.apple.com/library/archive/releasenotes/InterapplicationCommunication/RN-JavaScriptForAutomation/

**Tools:**
- DB Browser for SQLite: https://sqlitebrowser.org/
- PyObjC: https://pypi.org/project/pyobjc/

**For More Details:**
- See [RESEARCH_REPORT.md](RESEARCH_REPORT.md) for comprehensive documentation
- See [CODE_EXAMPLES.md](CODE_EXAMPLES.md) for complete, working code
- See [README.md](README.md) for overview and structure

## Quick Setup

**Python Environment:**
```bash
pip install pyobjc-framework-EventKit pyobjc-framework-Cocoa
```

**Test Permissions:**
```bash
# Test Full Disk Access
ls ~/Library/Safari/History.db

# Test Automation
osascript -e 'tell application "System Events" to return name'
```

**First Script:**
```python
#!/usr/bin/env python3
import plistlib
import os

# Read Safari bookmarks
with open(os.path.expanduser('~/Library/Safari/Bookmarks.plist'), 'rb') as f:
    bookmarks = plistlib.load(f)
    print(f"Loaded {len(bookmarks)} bookmark entries")
```

## Need Help?

1. Check if you have the right permissions
2. Verify the app is running (for AppleScript)
3. Check macOS version compatibility
4. Review the full documentation in RESEARCH_REPORT.md
5. Look at working examples in CODE_EXAMPLES.md
