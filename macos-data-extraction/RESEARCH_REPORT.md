# macOS Application Data Extraction - Research Report

This comprehensive report covers programmatic methods for extracting data from native macOS applications including Safari, Notes, Voice Memos, Reminders, Notifications, and Calendar.

## Table of Contents

- [Overview](#overview)
- [Safari Data Extraction](#safari-data-extraction)
- [Notes App Data Extraction](#notes-app-data-extraction)
- [Voice Memos Data Extraction](#voice-memos-data-extraction)
- [Reminders Data Extraction](#reminders-data-extraction)
- [Notifications Data Extraction](#notifications-data-extraction)
- [Calendar Data Extraction](#calendar-data-extraction)
- [Extraction Methods Comparison](#extraction-methods-comparison)
- [Open Source Projects](#open-source-projects)
- [Security and Privacy Considerations](#security-and-privacy-considerations)
- [References](#references)

## Overview

macOS stores application data in various formats and locations. The three primary approaches for data extraction are:

1. **Direct Database Access** - Reading SQLite databases and plist files
2. **AppleScript/JXA** - Using Apple's scripting interfaces
3. **UI Automation** - Interacting with the application interface
4. **System APIs** - Using macOS frameworks (EventKit, CoreData, etc.)

Each method has trade-offs in terms of reliability, maintenance, privacy permissions, and complexity.

## Safari Data Extraction

### Data Locations

Safari stores its data in `~/Library/Safari/`:
- **Bookmarks**: `~/Library/Safari/Bookmarks.plist`
- **History**: `~/Library/Safari/History.db` (SQLite)
- **Reading List**: `~/Library/Safari/Bookmarks.plist` (integrated with bookmarks)
- **Current Tabs**: `~/Library/Safari/LastSession.plist`
- **Cloud Tabs**: `~/Library/Safari/CloudTabs.db` (SQLite)

### Method 1: Direct Database/Plist Access

**Bookmarks (Plist)**
```python
import plistlib

# Read bookmarks
with open(os.path.expanduser('~/Library/Safari/Bookmarks.plist'), 'rb') as f:
    bookmarks = plistlib.load(f)

# Parse bookmark structure
def extract_bookmarks(node, bookmarks_list=[]):
    if 'Children' in node:
        for child in node['Children']:
            extract_bookmarks(child, bookmarks_list)
    if node.get('WebBookmarkType') == 'WebBookmarkTypeLeaf':
        bookmarks_list.append({
            'title': node.get('URIDictionary', {}).get('title', ''),
            'url': node.get('URLString', ''),
            'uuid': node.get('WebBookmarkUUID', '')
        })
    return bookmarks_list

all_bookmarks = extract_bookmarks(bookmarks)
```

**Reading List**
The reading list is stored in the same Bookmarks.plist file:
```python
# Reading list items have 'ReadingList' key
def extract_reading_list(node, reading_list=[]):
    if 'Children' in node:
        for child in node['Children']:
            if child.get('Title') == 'com.apple.ReadingList':
                for item in child.get('Children', []):
                    reading_list.append({
                        'title': item.get('URIDictionary', {}).get('title', ''),
                        'url': item.get('URLString', ''),
                        'date_added': item.get('ReadingList', {}).get('DateAdded', ''),
                        'preview_text': item.get('ReadingList', {}).get('PreviewText', '')
                    })
    return reading_list
```

**History (SQLite)**
```python
import sqlite3

conn = sqlite3.connect(os.path.expanduser('~/Library/Safari/History.db'))
cursor = conn.cursor()

# Query history
cursor.execute('''
    SELECT 
        h.id,
        i.url,
        h.visit_time,
        h.title,
        v.visit_count
    FROM history_items h
    JOIN history_visits v ON h.id = v.history_item
    JOIN history_items i ON h.id = i.id
    ORDER BY h.visit_time DESC
''')

history = cursor.fetchall()
conn.close()
```

**Current Open Tabs**
```python
# Read current session
with open(os.path.expanduser('~/Library/Safari/LastSession.plist'), 'rb') as f:
    session = plistlib.load(f)

# Extract tabs from windows
tabs = []
for window in session.get('SessionWindows', []):
    for tab in window.get('TabStates', []):
        tabs.append({
            'title': tab.get('Title', ''),
            'url': tab.get('URL', ''),
            'last_visited': tab.get('LastVisitTime', '')
        })
```

**Pros:**
- Fast and reliable
- No Safari running required
- Complete access to data

**Cons:**
- Requires Full Disk Access permission (macOS 10.14+)
- Schema may change with macOS updates
- Risk of data corruption if Safari is running

### Method 2: AppleScript

```applescript
-- Get bookmarks via AppleScript
tell application "Safari"
    -- Note: Safari doesn't expose bookmarks via AppleScript
    -- But we can get open tabs
    set tabList to {}
    repeat with w in windows
        repeat with t in tabs of w
            set end of tabList to {URL:URL of t, name:name of t}
        end repeat
    end repeat
    return tabList
end tell
```

**JavaScript for Automation (JXA) - Modern Alternative:**
```javascript
// Get current tabs
const Safari = Application('Safari');
const tabs = [];

Safari.windows().forEach(window => {
    window.tabs().forEach(tab => {
        tabs.push({
            url: tab.url(),
            name: tab.name()
        });
    });
});

JSON.stringify(tabs);
```

**Pros:**
- Officially supported API
- No database access needed
- Safer than direct file access

**Cons:**
- Safari must be running
- Limited API (no bookmarks, reading list access)
- Requires automation permissions

### Method 3: UI Automation

Using Accessibility APIs via AppleScript:
```applescript
tell application "System Events"
    tell process "Safari"
        -- Navigate to bookmarks
        keystroke "b" using {command down, option down}
        delay 0.5
        
        -- Get bookmark list
        set bookmarkList to entire contents of outline 1 of scroll area 1 of splitter group 1
    end tell
end tell
```

**Pros:**
- Can access features not in scripting API
- Works when other methods fail

**Cons:**
- Very fragile
- Slow
- Requires Accessibility permissions
- Breaks with UI changes

### Recommended Approach for Safari

**Best combination:**
1. **Bookmarks & Reading List**: Direct plist access (most reliable)
2. **History**: Direct SQLite access
3. **Current Tabs**: AppleScript/JXA if Safari is running, otherwise LastSession.plist

## Notes App Data Extraction

### Data Location

Notes data is stored in:
- `~/Library/Group Containers/group.com.apple.notes/NoteStore.sqlite`
- Attachments: `~/Library/Group Containers/group.com.apple.notes/Accounts/`

### Method 1: Direct SQLite Access

```python
import sqlite3
import os

db_path = os.path.expanduser(
    '~/Library/Group Containers/group.com.apple.notes/NoteStore.sqlite'
)

conn = sqlite3.connect(db_path)
cursor = conn.cursor()

# Query notes
cursor.execute('''
    SELECT 
        ZICCLOUDSYNCINGOBJECT.ZTITLE1 AS title,
        ZICCLOUDSYNCINGOBJECT.ZSNIPPET AS snippet,
        ZICCLOUDSYNCINGOBJECT.ZCREATIONDATE1 AS created,
        ZICCLOUDSYNCINGOBJECT.ZMODIFICATIONDATE1 AS modified,
        ZICNOTEDATA.ZDATA AS data,
        ZICCLOUDSYNCINGOBJECT.ZFOLDER AS folder_id
    FROM ZICCLOUDSYNCINGOBJECT
    LEFT JOIN ZICNOTEDATA ON ZICCLOUDSYNCINGOBJECT.Z_PK = ZICNOTEDATA.ZNOTE
    WHERE ZICCLOUDSYNCINGOBJECT.ZTITLE1 IS NOT NULL
    ORDER BY ZICCLOUDSYNCINGOBJECT.ZMODIFICATIONDATE1 DESC
''')

notes = cursor.fetchall()
conn.close()
```

**Note:** The column names vary by macOS version. Common variations:
- macOS 10.15+: `ZTITLE1`, `ZSNIPPET`, `ZCREATIONDATE1`
- macOS 10.13-10.14: `ZTITLE`, `ZSNIPPET`, `ZCREATIONDATE`

**Extracting Note Content:**
```python
# ZDATA contains protobuf-encoded note content
# Requires decoding protobuf or using NSData
import gzip

def decode_note_data(data):
    if data and data[:2] == b'\x1f\x8b':  # gzip magic number
        return gzip.decompress(data).decode('utf-8', errors='ignore')
    return data.decode('utf-8', errors='ignore') if data else ''
```

**Pros:**
- Complete access to all notes
- Fast
- Access to metadata

**Cons:**
- Requires Full Disk Access
- Database schema changes between macOS versions
- Content is protobuf-encoded (complex)
- Risk of corruption if Notes app is running

### Method 2: AppleScript

```applescript
tell application "Notes"
    set notesList to {}
    repeat with n in notes
        set end of notesList to {name:name of n, body:body of n, id:id of n}
    end repeat
    return notesList
end tell
```

**JXA Version:**
```javascript
const Notes = Application('Notes');
const notes = [];

Notes.notes().forEach(note => {
    notes.push({
        name: note.name(),
        body: note.body(),
        id: note.id(),
        creationDate: note.creationDate(),
        modificationDate: note.modificationDate()
    });
});

JSON.stringify(notes);
```

**Pros:**
- Officially supported
- No database knowledge needed
- Automatic content decoding
- No Full Disk Access needed

**Cons:**
- Notes app must be running
- Slower for large collections
- Requires automation permissions

### Method 3: CloudKit API

For iCloud-synced notes:
```python
# Using pyicloud library
from pyicloud import PyiCloudService

api = PyiCloudService('apple_id@email.com')
if api.requires_2fa:
    code = input("Enter 2FA code: ")
    api.validate_2fa_code(code)

# Access notes (limited support)
# CloudKit API for Notes is restricted
```

**Pros:**
- Remote access
- No local installation needed

**Cons:**
- Requires iCloud credentials
- Limited API access for Notes
- Complex authentication

### Recommended Approach for Notes

**Best option: AppleScript/JXA** - Most reliable and officially supported. Use SQLite only if you need historical data or the app isn't running.

## Voice Memos Data Extraction

### Data Location

Voice Memos are stored as m4a files:
- `~/Library/Application Support/com.apple.voicememos/Recordings/`
- Metadata: `~/Library/Application Support/com.apple.voicememos/Recordings/CloudRecordings.db`

### Method 1: Direct File System Access

```python
import os
import sqlite3
from datetime import datetime

# Access recordings directory
recordings_dir = os.path.expanduser(
    '~/Library/Application Support/com.apple.voicememos/Recordings/'
)

# Get metadata from SQLite
db_path = os.path.join(recordings_dir, 'CloudRecordings.db')
conn = sqlite3.connect(db_path)
cursor = conn.cursor()

cursor.execute('''
    SELECT 
        ZPATH AS path,
        ZTITLE AS title,
        ZDATE AS date,
        ZDURATION AS duration,
        ZCUSTOMLABEL AS custom_label
    FROM ZCLOUDRECORDING
    ORDER BY ZDATE DESC
''')

recordings = []
for row in cursor.fetchall():
    recording = {
        'path': os.path.join(recordings_dir, row[0]) if row[0] else None,
        'title': row[1],
        'date': datetime.fromtimestamp(row[2] + 978307200) if row[2] else None,  # Apple epoch
        'duration': row[3],
        'custom_label': row[4]
    }
    recordings.append(recording)

conn.close()
```

**Audio File Access:**
```python
import shutil

# Copy recording to export location
for recording in recordings:
    if recording['path'] and os.path.exists(recording['path']):
        dest = f"export/{recording['title']}.m4a"
        shutil.copy2(recording['path'], dest)
```

**Pros:**
- Direct access to audio files
- Complete metadata
- Fast

**Cons:**
- Requires Full Disk Access
- Must handle Apple epoch conversion (978307200 seconds offset)
- File names are UUIDs, need database for titles

### Method 2: No AppleScript Support

Voice Memos app does **not** provide AppleScript support. No scripting interface is available.

### Method 3: Using AVFoundation (Swift/Objective-C)

```swift
import AVFoundation
import Foundation

let recordingsPath = NSHomeDirectory() + "/Library/Application Support/com.apple.voicememos/Recordings/"

let fileManager = FileManager.default
if let files = try? fileManager.contentsOfDirectory(atPath: recordingsPath) {
    for file in files where file.hasSuffix(".m4a") {
        let filePath = recordingsPath + file
        let asset = AVAsset(url: URL(fileURLWithPath: filePath))
        
        // Get metadata
        let duration = asset.duration
        let metadata = asset.metadata
        
        print("File: \(file), Duration: \(CMTimeGetSeconds(duration))s")
    }
}
```

**Pros:**
- Native macOS API
- Proper audio handling
- Metadata extraction

**Cons:**
- Requires Swift/Objective-C
- Still needs Full Disk Access
- More complex than Python

### Recommended Approach for Voice Memos

**Best option: Direct file system + SQLite access**. This is the only reliable method since no scripting interface exists.

## Reminders Data Extraction

### Data Location

Reminders are stored in:
- `~/Library/Reminders/Container_v1/Stores/Data-*.icrd` (Core Data store)
- Calendar database also used for some data

### Method 1: EventKit Framework (Swift/Objective-C)

```swift
import EventKit

let eventStore = EKEventStore()

// Request access
eventStore.requestAccess(to: .reminder) { granted, error in
    guard granted else {
        print("Access denied")
        return
    }
    
    // Fetch all reminders
    let predicate = eventStore.predicateForReminders(in: nil)
    eventStore.fetchReminders(matching: predicate) { reminders in
        for reminder in reminders ?? [] {
            print("Title: \(reminder.title ?? "")")
            print("Due: \(reminder.dueDateComponents?.date?.description ?? "No date")")
            print("Completed: \(reminder.isCompleted)")
            print("Priority: \(reminder.priority)")
            print("Notes: \(reminder.notes ?? "")")
            print("---")
        }
    }
}
```

**Python via PyObjC:**
```python
from EventKit import EKEventStore, EKEntityTypeReminder
from Foundation import NSDate
import time

store = EKEventStore.alloc().init()

# Request access (async, requires run loop)
def request_access():
    semaphore = []
    def callback(granted, error):
        semaphore.append(granted)
    
    store.requestAccessToEntityType_completion_(EKEntityTypeReminder, callback)
    
    # Wait for callback
    while not semaphore:
        time.sleep(0.1)
    
    return semaphore[0]

if request_access():
    # Fetch reminders
    predicate = store.predicateForRemindersInCalendars_(None)
    
    reminders = []
    def fetch_callback(items):
        if items:
            for reminder in items:
                reminders.append({
                    'title': reminder.title(),
                    'completed': reminder.isCompleted(),
                    'due_date': reminder.dueDateComponents(),
                    'priority': reminder.priority(),
                    'notes': reminder.notes()
                })
    
    store.fetchRemindersMatchingPredicate_completion_(predicate, fetch_callback)
```

**Pros:**
- Official API
- Proper access control
- Cross-version compatibility
- Access to all reminder features

**Cons:**
- Requires Swift/Objective-C or PyObjC
- Async API complexity
- Requires Reminders permission

### Method 2: AppleScript

```applescript
tell application "Reminders"
    set reminderList to {}
    repeat with lst in lists
        set listName to name of lst
        repeat with r in reminders of lst
            set end of reminderList to {
                listName: listName,
                name: name of r,
                completed: completed of r,
                dueDate: due date of r,
                body: body of r
            }
        end repeat
    end repeat
    return reminderList
end tell
```

**JXA Version:**
```javascript
const Reminders = Application('Reminders');
const reminders = [];

Reminders.lists().forEach(list => {
    const listName = list.name();
    list.reminders().forEach(reminder => {
        reminders.push({
            list: listName,
            name: reminder.name(),
            completed: reminder.completed(),
            dueDate: reminder.dueDate(),
            body: reminder.body()
        });
    });
});

JSON.stringify(reminders);
```

**Pros:**
- Simple syntax
- No compilation needed
- Easy to use from command line

**Cons:**
- Reminders app must be running
- Limited access to some properties
- Requires automation permissions

### Method 3: Direct Database Access

```python
# NOT RECOMMENDED - Core Data format is complex and proprietary
# The .icrd files use Core Data which is difficult to parse
# Use EventKit or AppleScript instead
```

**Cons:**
- Core Data format is complex
- Proprietary format
- High risk of corruption
- Schema changes frequently

### Recommended Approach for Reminders

**Best options:**
1. **EventKit (Swift/Python)** - Most powerful and reliable
2. **AppleScript/JXA** - Simplest for basic access

## Notifications Data Extraction

### Data Location

Notification data is stored in:
- `~/Library/Application Support/NotificationCenter/` (older versions)
- macOS 10.15+: Unified Logging System (`/var/log/`)
- Database: `~/Library/Application Support/com.apple.notificationcenterui/` (some metadata)

### Method 1: Reading Notification Database

```python
import sqlite3
import os

db_path = os.path.expanduser(
    '~/Library/Application Support/com.apple.notificationcenterui/db2/db'
)

if os.path.exists(db_path):
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    # Schema varies by macOS version
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
    tables = cursor.fetchall()
    print("Available tables:", tables)
    
    # Try to query notifications
    try:
        cursor.execute('''
            SELECT * FROM notifications
            LIMIT 10
        ''')
        notifications = cursor.fetchall()
    except:
        print("Table structure may have changed")
    
    conn.close()
```

**Pros:**
- Access to historical notifications
- Metadata available

**Cons:**
- Database schema undocumented and changes frequently
- Limited information stored
- Requires Full Disk Access
- Most notification content is not persisted

### Method 2: Unified Logging System

```bash
# Query logs for notification-related events
log show --predicate 'subsystem == "com.apple.notificationcenter"' --last 1h

# Or programmatically:
```

```swift
import OSLog

let logStore = try OSLogStore(scope: .currentProcessIdentifier)
let position = logStore.position(timeIntervalSinceLatestBoot: -3600) // Last hour

let predicate = NSPredicate(format: "subsystem == 'com.apple.notificationcenter'")
let entries = try logStore.getEntries(at: position, matching: predicate)

for entry in entries {
    if let logEntry = entry as? OSLogEntryLog {
        print(logEntry.composedMessage)
    }
}
```

**Pros:**
- Official logging system
- Structured data
- Powerful queries

**Cons:**
- Requires specific permissions
- Limited to recent logs
- Complex API
- Log format not designed for notification extraction

### Method 3: No Direct API Access

Unfortunately, macOS does not provide a public API to query past notifications. The Notification Center is designed to show temporary alerts, not maintain a history.

### Recommended Approach for Notifications

**Reality check:** Extracting comprehensive notification history is **very difficult** on macOS because:
- Notifications are designed to be ephemeral
- No public API for querying past notifications
- Database format is undocumented and unstable

**Limited options:**
1. **Monitor in real-time** - Use NSUserNotificationCenter (deprecated) or UNUserNotificationCenter to receive notifications as they happen
2. **Parse logs** - Use Unified Logging System for recent notifications
3. **Per-app databases** - Some apps store their own notification history

## Calendar Data Extraction

### Data Location

Calendar data is stored in:
- `~/Library/Calendars/` (cache and local calendars)
- Core Data store for Calendar app
- EventKit provides access layer

### Method 1: EventKit Framework (Swift/Objective-C)

```swift
import EventKit

let eventStore = EKEventStore()

// Request access
eventStore.requestAccess(to: .event) { granted, error in
    guard granted else {
        print("Access denied")
        return
    }
    
    // Fetch events
    let calendars = eventStore.calendars(for: .event)
    
    // Date range for query
    let startDate = Date()
    let endDate = Calendar.current.date(byAdding: .month, value: 1, to: startDate)!
    
    let predicate = eventStore.predicateForEvents(
        withStart: startDate,
        end: endDate,
        calendars: calendars
    )
    
    let events = eventStore.events(matching: predicate)
    
    for event in events {
        print("Event: \(event.title ?? "")")
        print("Start: \(event.startDate)")
        print("End: \(event.endDate)")
        print("Location: \(event.location ?? "")")
        print("Notes: \(event.notes ?? "")")
        print("All-day: \(event.isAllDay)")
        print("Calendar: \(event.calendar.title)")
        print("---")
    }
}
```

**Python via PyObjC:**
```python
from EventKit import EKEventStore, EKEntityTypeEvent
from Foundation import NSDate, NSCalendar, NSDateComponents
import datetime
import time

store = EKEventStore.alloc().init()

# Request access
def request_access():
    semaphore = []
    def callback(granted, error):
        semaphore.append(granted)
    
    store.requestAccessToEntityType_completion_(EKEntityTypeEvent, callback)
    
    while not semaphore:
        time.sleep(0.1)
    
    return semaphore[0]

if request_access():
    # Get calendars
    calendars = store.calendarsForEntityType_(EKEntityTypeEvent)
    
    # Date range
    start_date = NSDate.date()
    end_date = NSDate.dateWithTimeIntervalSinceNow_(30 * 24 * 3600)  # 30 days
    
    predicate = store.predicateForEventsWithStartDate_endDate_calendars_(
        start_date, end_date, calendars
    )
    
    events = store.eventsMatchingPredicate_(predicate)
    
    for event in events:
        print(f"Title: {event.title()}")
        print(f"Start: {event.startDate()}")
        print(f"End: {event.endDate()}")
        print(f"Location: {event.location()}")
        print(f"Calendar: {event.calendar().title()}")
        print("---")
```

**Pros:**
- Official API
- Complete access to all calendar features
- Handles recurring events properly
- Syncs with iCloud calendars
- Proper permissions

**Cons:**
- Requires Swift/Objective-C or PyObjC
- Async API in Swift
- Requires Calendar permission

### Method 2: AppleScript

```applescript
tell application "Calendar"
    set eventList to {}
    repeat with cal in calendars
        set calName to name of cal
        set startDate to (current date) - (30 * days)
        set endDate to (current date) + (30 * days)
        
        set calEvents to (every event of cal whose start date is greater than startDate and start date is less than endDate)
        
        repeat with evt in calEvents
            set end of eventList to {
                calendar: calName,
                summary: summary of evt,
                startDate: start date of evt,
                endDate: end date of evt,
                location: location of evt,
                description: description of evt,
                allDayEvent: allday event of evt
            }
        end repeat
    end repeat
    return eventList
end tell
```

**JXA Version:**
```javascript
const Calendar = Application('Calendar');
const events = [];

const now = new Date();
const startDate = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
const endDate = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000);

Calendar.calendars().forEach(calendar => {
    const calName = calendar.name();
    calendar.events().forEach(event => {
        const start = event.startDate();
        if (start >= startDate && start <= endDate) {
            events.push({
                calendar: calName,
                summary: event.summary(),
                startDate: event.startDate(),
                endDate: event.endDate(),
                location: event.location(),
                description: event.description(),
                allDayEvent: event.alldayEvent()
            });
        }
    });
});

JSON.stringify(events);
```

**Pros:**
- Simple to use
- No compilation needed
- Good for basic data extraction

**Cons:**
- Calendar app must be running
- Slower than EventKit
- Less control over date ranges
- Requires automation permissions

### Method 3: CalDAV Protocol

For iCloud and Exchange calendars:
```python
import caldav
from datetime import datetime, timedelta

# Connect to CalDAV server
client = caldav.DAVClient(
    url="https://caldav.icloud.com/",
    username="apple_id@email.com",
    password="app-specific-password"
)

principal = client.principal()
calendars = principal.calendars()

# Fetch events
for calendar in calendars:
    print(f"Calendar: {calendar.name}")
    
    # Date range
    start = datetime.now()
    end = start + timedelta(days=30)
    
    events = calendar.date_search(start, end)
    
    for event in events:
        print(f"  Event: {event.instance.vevent.summary.value}")
        print(f"  Start: {event.instance.vevent.dtstart.value}")
```

**Pros:**
- Works for remote calendars
- Standard protocol
- No local app needed

**Cons:**
- Requires calendar server credentials
- Not all calendars use CalDAV
- Complex authentication
- Doesn't access local-only calendars

### Recommended Approach for Calendar

**Best options:**
1. **EventKit (Swift/Python)** - Most complete and reliable
2. **AppleScript/JXA** - Simplest for basic access
3. **CalDAV** - For remote-only access

## Extraction Methods Comparison

| Method | Speed | Reliability | Permissions | Complexity | Maintenance |
|--------|-------|-------------|-------------|------------|-------------|
| **SQLite/Plist** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | Full Disk Access | ⭐⭐⭐⭐ | ⭐⭐ (breaks with updates) |
| **AppleScript/JXA** | ⭐⭐⭐ | ⭐⭐⭐⭐ | Automation | ⭐⭐ | ⭐⭐⭐⭐ |
| **UI Automation** | ⭐ | ⭐ | Accessibility | ⭐⭐⭐⭐⭐ | ⭐ (very fragile) |
| **System APIs** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Specific per API | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

### When to Use Each Method

**SQLite/Plist Direct Access:**
- ✅ Need complete historical data
- ✅ App doesn't need to be running
- ✅ Maximum speed required
- ❌ Requires Full Disk Access (user friction)
- ❌ May break with OS updates

**AppleScript/JXA:**
- ✅ Official scripting support exists
- ✅ Don't need Full Disk Access
- ✅ Simple scripts
- ❌ App must be running
- ❌ Limited API coverage

**UI Automation:**
- ✅ Last resort when no other option
- ❌ Extremely fragile
- ❌ Very slow
- ❌ Use only if absolutely necessary

**System APIs (EventKit, etc.):**
- ✅ Official, supported interfaces
- ✅ Best long-term maintainability
- ✅ Proper permission handling
- ❌ Requires Swift/Objective-C (or PyObjC)
- ❌ More complex to set up

## Open Source Projects

### General macOS Data Extraction

**1. mackup**
- **URL**: https://github.com/lra/mackup
- **Purpose**: Backup and restore application settings
- **Relevant for**: Understanding where apps store data
- **Language**: Python
- **Stars**: ~14k

**2. macos-safari-reading-list**
- **URL**: https://github.com/joelvh/macos-safari-reading-list
- **Purpose**: Export Safari reading list
- **Language**: Ruby
- **Method**: Plist parsing

**3. safari-export**
- **URL**: https://github.com/seanbreckenridge/safari-export
- **Purpose**: Export Safari history and bookmarks
- **Language**: Python
- **Method**: SQLite + Plist

### Safari-Specific Tools

**4. safaribooks**
- **URL**: https://github.com/niharrs/safaribooks
- **Purpose**: Safari bookmarks manager
- **Language**: Python
- **Method**: Plist parsing

**5. browser-history**
- **URL**: https://github.com/browser-history/browser-history
- **Purpose**: Extract history from all major browsers including Safari
- **Language**: Python
- **Method**: SQLite

### Notes-Specific Tools

**6. apple-notes-exporter**
- **URL**: https://github.com/fsndzomga/apple-notes-exporter
- **Purpose**: Export Apple Notes to markdown
- **Language**: Python
- **Method**: SQLite + protobuf decoding

**7. mac-notes-exporter**
- **URL**: https://github.com/DaveSmith17/mac-notes-exporter
- **Purpose**: Export Notes to HTML/markdown
- **Language**: Python
- **Method**: AppleScript

**8. exporter**
- **URL**: https://github.com/ksylvan/exporter
- **Purpose**: Export Apple Notes and other data
- **Language**: Swift
- **Method**: AppleScript

### Calendar & Reminders Tools

**9. pyicloud**
- **URL**: https://github.com/picklepete/pyicloud
- **Purpose**: Access iCloud data including calendars
- **Language**: Python
- **Method**: iCloud API

**10. caldav**
- **URL**: https://github.com/python-caldav/caldav
- **Purpose**: CalDAV client for calendar access
- **Language**: Python
- **Method**: CalDAV protocol

### Voice Memos Tools

**11. voice-memos-sync**
- **URL**: https://github.com/fangpenlin/voice-memos-sync
- **Purpose**: Sync Voice Memos to cloud storage
- **Language**: Python
- **Method**: File system access

### Automation & Scripting Frameworks

**12. py-applescript**
- **URL**: https://github.com/rdhyee/py-applescript
- **Purpose**: Run AppleScript from Python
- **Language**: Python
- **Method**: Bridge to AppleScript

**13. PyObjC**
- **URL**: https://github.com/ronaldoussoren/pyobjc
- **Purpose**: Python bindings for macOS frameworks
- **Language**: Python
- **Method**: Native framework access

**14. JXA-Cookbook**
- **URL**: https://github.com/JXA-Cookbook/JXA-Cookbook
- **Purpose**: JavaScript for Automation examples
- **Language**: JavaScript
- **Method**: JXA

### Comprehensive Backup Tools

**15. mac-backup-scripts**
- **URL**: https://github.com/wilsonmar/mac-backup-scripts
- **Purpose**: Backup various macOS application data
- **Language**: Shell/Python
- **Method**: Mixed (SQLite, plist, AppleScript)

**16. dotfiles** (multiple repos)
- **URL**: Search GitHub for "macos dotfiles"
- **Purpose**: Many dotfile repos include scripts for backing up app data
- **Method**: Various

### Academic & Research Tools

**17. osxcollector**
- **URL**: https://github.com/Yelp/osxcollector
- **Purpose**: Forensic collection tool (includes browser data, etc.)
- **Language**: Python
- **Method**: Direct file/database access

**18. mac_apt**
- **URL**: https://github.com/ydkhatri/mac_apt
- **Purpose**: macOS Artifact Parsing Tool
- **Language**: Python
- **Method**: SQLite, plist, and binary parsing

## Security and Privacy Considerations

### Permission Requirements

**macOS 10.14+ (Mojave and later) introduced strict privacy controls:**

1. **Full Disk Access**
   - Required for: Direct database/plist access
   - Grant in: System Preferences → Security & Privacy → Privacy → Full Disk Access
   - User must explicitly approve

2. **Automation**
   - Required for: AppleScript, UI automation
   - Grant in: System Preferences → Security & Privacy → Privacy → Automation
   - Per-app approval needed

3. **Accessibility**
   - Required for: UI automation via System Events
   - Grant in: System Preferences → Security & Privacy → Privacy → Accessibility
   - High-privilege permission

4. **Calendar/Reminders Access**
   - Required for: EventKit access
   - Automatic prompt on first use
   - Can be revoked in System Preferences

### Best Practices

1. **Request Minimal Permissions**
   - Only ask for what you need
   - Explain why you need each permission
   - Provide clear instructions

2. **Handle Permission Errors Gracefully**
   ```python
   try:
       # Access protected resource
       pass
   except PermissionError:
       print("This operation requires Full Disk Access.")
       print("Please grant access in System Preferences.")
   ```

3. **Respect User Privacy**
   - Don't collect more data than necessary
   - Provide clear privacy policy
   - Allow users to see what data is accessed

4. **Use Official APIs When Possible**
   - EventKit for calendars/reminders
   - AppleScript for Notes
   - These respect system permissions properly

5. **Test on Multiple macOS Versions**
   - Database schemas change
   - APIs evolve
   - Permissions model has changed significantly

### Code Signing & Notarization

For distribution, your app should be:
- **Signed** with an Apple Developer ID
- **Notarized** by Apple
- This builds user trust and is required for some permissions

## References

### Official Apple Documentation

**EventKit Framework:**
- https://developer.apple.com/documentation/eventkit

**AppleScript Language Guide:**
- https://developer.apple.com/library/archive/documentation/AppleScript/Conceptual/AppleScriptLangGuide/

**JavaScript for Automation (JXA):**
- https://developer.apple.com/library/archive/releasenotes/InterapplicationCommunication/RN-JavaScriptForAutomation/

**Accessibility (UI Automation):**
- https://developer.apple.com/documentation/accessibility

**User Privacy:**
- https://developer.apple.com/documentation/uikit/protecting_the_user_s_privacy

### Database Format References

**Safari:**
- History.db schema documented in various forensics tools
- Bookmarks.plist structure: https://github.com/mozilla/multi-account-containers/wiki/Sync-Containers

**Notes:**
- Notes database analysis: https://www.swiftforensics.com/2018/02/reading-notes-database-on-macos.html
- Protobuf format discussion: https://github.com/threeplanetssoftware/apple_cloud_notes_parser

**Voice Memos:**
- File structure documented in backup tools
- SQLite schema varies by version

### Community Resources

**Stack Overflow:**
- `[macos]` tag for general macOS programming
- `[applescript]` tag for AppleScript questions
- `[eventkit]` tag for Calendar/Reminders access

**Reddit:**
- r/macosprogramming
- r/applescript
- r/shortcuts (for automation ideas)

**Forums:**
- Apple Developer Forums
- MacScripter Forums (AppleScript community)

### Tools Documentation

**SQLite Browser:**
- https://sqlitebrowser.org/
- Essential for exploring SQLite databases

**plist Editor:**
- Xcode includes plist editor
- `plutil` command-line tool (built-in)

**osascript:**
- `man osascript` - Run AppleScript from command line
- Supports both AppleScript and JXA

### Academic Papers

**Digital Forensics:**
- Many papers cover macOS artifact extraction
- Search: "macOS forensics", "Safari forensics", "Apple Notes forensics"

## Conclusion

### Recommended Strategy by Application

| App | Best Method | Alternative | Avoid |
|-----|-------------|-------------|-------|
| **Safari Bookmarks** | Plist parsing | N/A | UI Automation |
| **Safari Tabs** | AppleScript (if running), else LastSession.plist | N/A | N/A |
| **Safari History** | SQLite direct | N/A | N/A |
| **Notes** | AppleScript/JXA | SQLite | UI Automation |
| **Voice Memos** | File system + SQLite | N/A | N/A (no alternative) |
| **Reminders** | EventKit | AppleScript | Direct database |
| **Notifications** | Real-time monitoring only | Logs | Database (unreliable) |
| **Calendar** | EventKit | AppleScript | Direct database |

### Quick Start Guide

**1. For Python developers:**
```bash
# Install dependencies
pip install pyobjc-framework-EventKit
pip install pyobjc-framework-Contacts
pip install py-applescript

# For SQLite
# No installation needed, built-in to Python
```

**2. For Swift developers:**
```swift
// Import frameworks
import EventKit
import Contacts
import OSAKit  // For AppleScript

// Request permissions early
// Use modern async/await patterns
```

**3. For Shell scripters:**
```bash
# Use osascript for AppleScript
osascript -e 'tell application "Safari" to get URL of current tab of front window'

# Use sqlite3 for databases
sqlite3 ~/Library/Safari/History.db "SELECT * FROM history_items LIMIT 10"

# Use plutil for plists
plutil -convert json ~/Library/Safari/Bookmarks.plist -o bookmarks.json
```

### Key Takeaways

1. **No one-size-fits-all solution** - Different apps require different approaches
2. **Official APIs are best** - EventKit for calendars/reminders, AppleScript where available
3. **Direct database access is fast but fragile** - Use for Safari, avoid for others
4. **Permissions are crucial** - Plan for Full Disk Access or alternative approaches
5. **Notifications are ephemeral** - No good way to access history
6. **Test across macOS versions** - Formats and schemas change

### Next Steps

1. Choose your programming language (Python, Swift, or Shell)
2. Identify which apps you need to access
3. Request appropriate permissions
4. Start with official APIs where available
5. Fall back to direct access only when necessary
6. Test thoroughly across macOS versions
7. Handle errors and permission denials gracefully

This report provides a comprehensive overview of all viable options for extracting data from macOS applications. Choose the methods that best fit your requirements for reliability, maintainability, and user experience.
