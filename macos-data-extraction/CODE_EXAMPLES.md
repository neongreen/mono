# Code Examples - macOS Data Extraction

Practical, ready-to-use code examples for extracting data from macOS applications.

## Table of Contents

- [Safari Examples](#safari-examples)
- [Notes Examples](#notes-examples)
- [Voice Memos Examples](#voice-memos-examples)
- [Reminders Examples](#reminders-examples)
- [Calendar Examples](#calendar-examples)
- [Utility Functions](#utility-functions)

## Safari Examples

### Extract Bookmarks (Python)

```python
#!/usr/bin/env python3
import plistlib
import os
from typing import List, Dict

def extract_bookmarks(bookmarks_path: str = None) -> List[Dict]:
    """Extract all bookmarks from Safari."""
    if bookmarks_path is None:
        bookmarks_path = os.path.expanduser('~/Library/Safari/Bookmarks.plist')
    
    with open(bookmarks_path, 'rb') as f:
        bookmarks_data = plistlib.load(f)
    
    def parse_node(node, bookmarks_list=None):
        if bookmarks_list is None:
            bookmarks_list = []
        
        if 'Children' in node:
            for child in node['Children']:
                parse_node(child, bookmarks_list)
        
        if node.get('WebBookmarkType') == 'WebBookmarkTypeLeaf':
            uri_dict = node.get('URIDictionary', {})
            bookmarks_list.append({
                'title': uri_dict.get('title', ''),
                'url': node.get('URLString', ''),
                'uuid': node.get('WebBookmarkUUID', ''),
                'date_added': node.get('DateAdded', None)
            })
        
        return bookmarks_list
    
    return parse_node(bookmarks_data)

# Usage
if __name__ == '__main__':
    bookmarks = extract_bookmarks()
    for bookmark in bookmarks[:10]:  # Print first 10
        print(f"{bookmark['title']}: {bookmark['url']}")
```

### Extract Reading List (Python)

```python
#!/usr/bin/env python3
import plistlib
import os
from typing import List, Dict

def extract_reading_list(bookmarks_path: str = None) -> List[Dict]:
    """Extract reading list from Safari."""
    if bookmarks_path is None:
        bookmarks_path = os.path.expanduser('~/Library/Safari/Bookmarks.plist')
    
    with open(bookmarks_path, 'rb') as f:
        bookmarks_data = plistlib.load(f)
    
    def find_reading_list(node, reading_list=None):
        if reading_list is None:
            reading_list = []
        
        if node.get('Title') == 'com.apple.ReadingList':
            for item in node.get('Children', []):
                reading_list_data = item.get('ReadingList', {})
                uri_dict = item.get('URIDictionary', {})
                reading_list.append({
                    'title': uri_dict.get('title', ''),
                    'url': item.get('URLString', ''),
                    'date_added': reading_list_data.get('DateAdded', None),
                    'date_last_viewed': reading_list_data.get('DateLastViewed', None),
                    'preview_text': reading_list_data.get('PreviewText', ''),
                })
            return reading_list
        
        if 'Children' in node:
            for child in node['Children']:
                result = find_reading_list(child, reading_list)
                if result:
                    return result
        
        return reading_list
    
    return find_reading_list(bookmarks_data)

# Usage
if __name__ == '__main__':
    reading_list = extract_reading_list()
    for item in reading_list:
        print(f"{item['title']}: {item['url']}")
```

### Extract History (Python)

```python
#!/usr/bin/env python3
import sqlite3
import os
from datetime import datetime
from typing import List, Dict

def extract_history(limit: int = 100) -> List[Dict]:
    """Extract Safari browsing history."""
    history_db = os.path.expanduser('~/Library/Safari/History.db')
    
    conn = sqlite3.connect(history_db)
    cursor = conn.cursor()
    
    query = '''
        SELECT 
            history_items.id,
            history_items.url,
            history_visits.visit_time,
            history_items.title,
            history_visits.visit_count
        FROM history_items
        JOIN history_visits ON history_items.id = history_visits.history_item
        ORDER BY history_visits.visit_time DESC
        LIMIT ?
    '''
    
    cursor.execute(query, (limit,))
    rows = cursor.fetchall()
    conn.close()
    
    history = []
    for row in rows:
        # Safari stores timestamps as seconds since 2001-01-01
        safari_epoch = datetime(2001, 1, 1)
        if row[2]:
            visit_time = safari_epoch.timestamp() + row[2]
            visit_datetime = datetime.fromtimestamp(visit_time)
        else:
            visit_datetime = None
        
        history.append({
            'id': row[0],
            'url': row[1],
            'visit_time': visit_datetime,
            'title': row[3],
            'visit_count': row[4]
        })
    
    return history

# Usage
if __name__ == '__main__':
    history = extract_history(limit=20)
    for item in history:
        print(f"{item['visit_time']}: {item['title']} - {item['url']}")
```

### Get Current Tabs (AppleScript/JXA)

**AppleScript:**
```applescript
#!/usr/bin/osascript

tell application "Safari"
    set tabList to {}
    repeat with w in windows
        repeat with t in tabs of w
            set end of tabList to {url:(URL of t), title:(name of t)}
        end repeat
    end repeat
    return tabList
end tell
```

**JavaScript for Automation (JXA):**
```javascript
#!/usr/bin/osascript -l JavaScript

const Safari = Application('Safari');
const tabs = [];

Safari.windows().forEach(window => {
    window.tabs().forEach(tab => {
        tabs.push({
            url: tab.url(),
            title: tab.name()
        });
    });
});

JSON.stringify(tabs);
```

**Run from Python:**
```python
import subprocess
import json

def get_safari_tabs():
    """Get currently open Safari tabs using JXA."""
    script = '''
    const Safari = Application('Safari');
    const tabs = [];
    Safari.windows().forEach(window => {
        window.tabs().forEach(tab => {
            tabs.push({
                url: tab.url(),
                title: tab.name()
            });
        });
    });
    JSON.stringify(tabs);
    '''
    
    result = subprocess.run(
        ['osascript', '-l', 'JavaScript', '-e', script],
        capture_output=True,
        text=True
    )
    
    if result.returncode == 0:
        return json.loads(result.stdout.strip())
    return []

# Usage
tabs = get_safari_tabs()
for tab in tabs:
    print(f"{tab['title']}: {tab['url']}")
```

## Notes Examples

### Extract Notes (AppleScript/JXA)

**JavaScript for Automation:**
```javascript
#!/usr/bin/osascript -l JavaScript

const Notes = Application('Notes');
const notes = [];

Notes.notes().forEach(note => {
    notes.push({
        name: note.name(),
        body: note.body(),
        id: note.id(),
        creationDate: note.creationDate(),
        modificationDate: note.modificationDate(),
        container: note.container.name()
    });
});

JSON.stringify(notes);
```

**Python wrapper:**
```python
#!/usr/bin/env python3
import subprocess
import json
from typing import List, Dict

def extract_notes() -> List[Dict]:
    """Extract all notes using JXA."""
    script = '''
    const Notes = Application('Notes');
    const notes = [];
    
    Notes.notes().forEach(note => {
        notes.push({
            name: note.name(),
            body: note.body(),
            id: note.id(),
            creationDate: note.creationDate().toString(),
            modificationDate: note.modificationDate().toString()
        });
    });
    
    JSON.stringify(notes);
    '''
    
    result = subprocess.run(
        ['osascript', '-l', 'JavaScript', '-e', script],
        capture_output=True,
        text=True
    )
    
    if result.returncode == 0:
        return json.loads(result.stdout.strip())
    return []

# Usage
if __name__ == '__main__':
    notes = extract_notes()
    for note in notes[:5]:  # Print first 5
        print(f"Title: {note['name']}")
        print(f"Modified: {note['modificationDate']}")
        print(f"Preview: {note['body'][:100]}...")
        print("---")
```

### Search Notes (JXA)

```javascript
#!/usr/bin/osascript -l JavaScript

function searchNotes(searchTerm) {
    const Notes = Application('Notes');
    const results = [];
    
    Notes.notes().forEach(note => {
        const body = note.body();
        const name = note.name();
        
        if (body.toLowerCase().includes(searchTerm.toLowerCase()) ||
            name.toLowerCase().includes(searchTerm.toLowerCase())) {
            results.push({
                name: name,
                body: body.substring(0, 200),
                modificationDate: note.modificationDate()
            });
        }
    });
    
    return JSON.stringify(results);
}

function run(argv) {
    if (argv.length === 0) {
        return "Usage: search-notes.js <search-term>";
    }
    return searchNotes(argv[0]);
}
```

## Voice Memos Examples

### Extract Voice Memos Metadata (Python)

```python
#!/usr/bin/env python3
import sqlite3
import os
from datetime import datetime
from typing import List, Dict
import shutil

def extract_voice_memos() -> List[Dict]:
    """Extract Voice Memos metadata and file locations."""
    recordings_dir = os.path.expanduser(
        '~/Library/Application Support/com.apple.voicememos/Recordings/'
    )
    db_path = os.path.join(recordings_dir, 'CloudRecordings.db')
    
    if not os.path.exists(db_path):
        print("Voice Memos database not found")
        return []
    
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    query = '''
        SELECT 
            ZPATH,
            ZTITLE,
            ZDATE,
            ZDURATION,
            ZCUSTOMLABEL
        FROM ZCLOUDRECORDING
        ORDER BY ZDATE DESC
    '''
    
    cursor.execute(query)
    rows = cursor.fetchall()
    conn.close()
    
    recordings = []
    # Apple epoch: seconds since 2001-01-01
    apple_epoch = datetime(2001, 1, 1).timestamp()
    
    for row in rows:
        path = row[0]
        full_path = os.path.join(recordings_dir, path) if path else None
        
        date = None
        if row[2]:
            timestamp = apple_epoch + row[2]
            date = datetime.fromtimestamp(timestamp)
        
        recordings.append({
            'path': full_path,
            'title': row[1] or 'Untitled',
            'date': date,
            'duration': row[3],  # in seconds
            'label': row[4],
            'exists': os.path.exists(full_path) if full_path else False
        })
    
    return recordings

def export_voice_memos(output_dir: str) -> None:
    """Export voice memos to a directory."""
    os.makedirs(output_dir, exist_ok=True)
    
    recordings = extract_voice_memos()
    
    for recording in recordings:
        if recording['exists']:
            filename = f"{recording['title']}.m4a"
            # Sanitize filename
            filename = "".join(c for c in filename if c.isalnum() or c in (' ', '-', '_', '.')).strip()
            
            dest = os.path.join(output_dir, filename)
            print(f"Exporting: {filename}")
            shutil.copy2(recording['path'], dest)

# Usage
if __name__ == '__main__':
    recordings = extract_voice_memos()
    
    for recording in recordings:
        duration_min = int(recording['duration'] // 60) if recording['duration'] else 0
        duration_sec = int(recording['duration'] % 60) if recording['duration'] else 0
        print(f"{recording['title']} - {duration_min}:{duration_sec:02d}")
        print(f"  Date: {recording['date']}")
        print(f"  Exists: {recording['exists']}")
        print()
    
    # Uncomment to export all recordings:
    # export_voice_memos('./exported_voice_memos')
```

## Reminders Examples

### Extract Reminders (Python with PyObjC)

```python
#!/usr/bin/env python3
from EventKit import EKEventStore, EKEntityTypeReminder
import time
from typing import List, Dict

def extract_reminders() -> List[Dict]:
    """Extract all reminders using EventKit."""
    store = EKEventStore.alloc().init()
    
    # Request access
    access_granted = []
    
    def completion_handler(granted, error):
        access_granted.append(granted)
    
    store.requestAccessToEntityType_completion_(
        EKEntityTypeReminder,
        completion_handler
    )
    
    # Wait for permission response
    timeout = 0
    while not access_granted and timeout < 50:
        time.sleep(0.1)
        timeout += 1
    
    if not access_granted or not access_granted[0]:
        print("Access to Reminders denied")
        return []
    
    # Fetch reminders
    predicate = store.predicateForRemindersInCalendars_(None)
    
    reminders_data = []
    
    def fetch_completion(items):
        if items:
            for reminder in items:
                due_date = None
                if reminder.dueDateComponents():
                    components = reminder.dueDateComponents()
                    # Convert NSDateComponents to dict
                    due_date = {
                        'year': components.year(),
                        'month': components.month(),
                        'day': components.day(),
                        'hour': components.hour(),
                        'minute': components.minute()
                    }
                
                reminders_data.append({
                    'title': reminder.title(),
                    'completed': reminder.isCompleted(),
                    'due_date': due_date,
                    'priority': reminder.priority(),
                    'notes': reminder.notes(),
                    'calendar': reminder.calendar().title()
                })
    
    store.fetchRemindersMatchingPredicate_completion_(
        predicate,
        fetch_completion
    )
    
    # Wait for fetch to complete
    timeout = 0
    while not reminders_data and timeout < 50:
        time.sleep(0.1)
        timeout += 1
    
    return reminders_data

# Usage
if __name__ == '__main__':
    reminders = extract_reminders()
    
    for reminder in reminders:
        status = "✓" if reminder['completed'] else "○"
        print(f"{status} {reminder['title']}")
        if reminder['due_date']:
            print(f"  Due: {reminder['due_date']}")
        if reminder['notes']:
            print(f"  Notes: {reminder['notes']}")
        print()
```

### Extract Reminders (AppleScript)

```applescript
#!/usr/bin/osascript

tell application "Reminders"
    set reminderList to {}
    
    repeat with lst in lists
        set listName to name of lst
        
        repeat with r in reminders of lst
            set reminderData to {listName:listName, reminderName:name of r, completed:completed of r, body:body of r}
            
            -- Add due date if exists
            try
                set dueDate to due date of r
                set reminderData to reminderData & {dueDate:dueDate}
            end try
            
            set end of reminderList to reminderData
        end repeat
    end repeat
    
    return reminderList
end tell
```

**Python wrapper:**
```python
import subprocess
import json

def get_reminders_applescript():
    """Get reminders using AppleScript."""
    script = '''
    tell application "Reminders"
        set output to {}
        repeat with lst in lists
            repeat with r in reminders of lst
                set end of output to {listName:(name of lst), title:(name of r), completed:(completed of r)}
            end repeat
        end repeat
        return output
    end tell
    '''
    
    result = subprocess.run(
        ['osascript', '-e', script],
        capture_output=True,
        text=True
    )
    
    return result.stdout
```

## Calendar Examples

### Extract Calendar Events (Python with PyObjC)

```python
#!/usr/bin/env python3
from EventKit import EKEventStore, EKEntityTypeEvent
from Foundation import NSDate
import time
from datetime import datetime, timedelta
from typing import List, Dict

def extract_calendar_events(days_ahead: int = 30, days_back: int = 30) -> List[Dict]:
    """Extract calendar events using EventKit."""
    store = EKEventStore.alloc().init()
    
    # Request access
    access_granted = []
    
    def completion_handler(granted, error):
        access_granted.append(granted)
    
    store.requestAccessToEntityType_completion_(
        EKEntityTypeEvent,
        completion_handler
    )
    
    # Wait for permission
    timeout = 0
    while not access_granted and timeout < 50:
        time.sleep(0.1)
        timeout += 1
    
    if not access_granted or not access_granted[0]:
        print("Access to Calendar denied")
        return []
    
    # Get calendars
    calendars = store.calendarsForEntityType_(EKEntityTypeEvent)
    
    # Date range
    start_date = NSDate.dateWithTimeIntervalSinceNow_(-days_back * 24 * 3600)
    end_date = NSDate.dateWithTimeIntervalSinceNow_(days_ahead * 24 * 3600)
    
    # Fetch events
    predicate = store.predicateForEventsWithStartDate_endDate_calendars_(
        start_date,
        end_date,
        calendars
    )
    
    events = store.eventsMatchingPredicate_(predicate)
    
    events_data = []
    for event in events:
        events_data.append({
            'title': event.title(),
            'start_date': event.startDate(),
            'end_date': event.endDate(),
            'location': event.location(),
            'notes': event.notes(),
            'all_day': event.isAllDay(),
            'calendar': event.calendar().title(),
            'has_alarms': len(event.alarms()) > 0 if event.alarms() else False
        })
    
    return events_data

# Usage
if __name__ == '__main__':
    events = extract_calendar_events(days_ahead=7, days_back=7)
    
    for event in sorted(events, key=lambda x: x['start_date']):
        print(f"📅 {event['title']}")
        print(f"   {event['start_date']} - {event['end_date']}")
        if event['location']:
            print(f"   📍 {event['location']}")
        print(f"   Calendar: {event['calendar']}")
        print()
```

### Extract Calendar Events (AppleScript)

```applescript
#!/usr/bin/osascript

tell application "Calendar"
    set eventList to {}
    set startDate to (current date) - (7 * days)
    set endDate to (current date) + (7 * days)
    
    repeat with cal in calendars
        set calName to name of cal
        set calEvents to (every event of cal whose start date > startDate and start date < endDate)
        
        repeat with evt in calEvents
            set eventData to {calendar:calName, summary:summary of evt, startDate:start date of evt, endDate:end date of evt, location:location of evt, allDayEvent:allday event of evt}
            set end of eventList to eventData
        end repeat
    end repeat
    
    return eventList
end tell
```

## Utility Functions

### Permission Check (Python)

```python
#!/usr/bin/env python3
import os
import sys

def check_full_disk_access() -> bool:
    """Check if the script has Full Disk Access permission."""
    test_path = os.path.expanduser('~/Library/Safari/History.db')
    
    try:
        with open(test_path, 'rb') as f:
            f.read(1)
        return True
    except (PermissionError, OSError):
        return False

def check_automation_permission() -> bool:
    """Check if automation permission is granted."""
    import subprocess
    
    script = 'tell application "System Events" to return name'
    result = subprocess.run(
        ['osascript', '-e', script],
        capture_output=True,
        text=True
    )
    
    return result.returncode == 0

def request_permissions_guide():
    """Print instructions for granting permissions."""
    print("""
Permission Setup Guide:
    
For Full Disk Access:
1. Open System Preferences → Security & Privacy → Privacy
2. Click on "Full Disk Access" in the left sidebar
3. Click the lock icon and authenticate
4. Click the "+" button and add your Python executable or Terminal app
    
For Automation:
1. Run the script - you'll be prompted automatically
2. Or go to System Preferences → Security & Privacy → Privacy → Automation
3. Enable access for your application
    
For Calendar/Reminders:
1. The app will prompt automatically on first access
2. Or go to System Preferences → Security & Privacy → Privacy → Calendars/Reminders
3. Enable access for your application
    """)

# Usage
if __name__ == '__main__':
    print("Checking permissions...")
    
    if check_full_disk_access():
        print("✓ Full Disk Access: Granted")
    else:
        print("✗ Full Disk Access: Denied")
    
    if check_automation_permission():
        print("✓ Automation: Granted")
    else:
        print("✗ Automation: Denied")
    
    request_permissions_guide()
```

### Apple Epoch Converter

```python
def apple_timestamp_to_datetime(timestamp: float):
    """Convert Apple timestamp (seconds since 2001-01-01) to datetime."""
    from datetime import datetime
    
    apple_epoch = datetime(2001, 1, 1).timestamp()
    return datetime.fromtimestamp(apple_epoch + timestamp)

def datetime_to_apple_timestamp(dt: datetime) -> float:
    """Convert datetime to Apple timestamp."""
    from datetime import datetime
    
    apple_epoch = datetime(2001, 1, 1).timestamp()
    return dt.timestamp() - apple_epoch
```

### Run AppleScript from Python

```python
import subprocess
import json

def run_applescript(script: str) -> str:
    """Execute AppleScript and return output."""
    result = subprocess.run(
        ['osascript', '-e', script],
        capture_output=True,
        text=True
    )
    
    if result.returncode != 0:
        raise Exception(f"AppleScript error: {result.stderr}")
    
    return result.stdout.strip()

def run_jxa(script: str) -> str:
    """Execute JavaScript for Automation and return output."""
    result = subprocess.run(
        ['osascript', '-l', 'JavaScript', '-e', script],
        capture_output=True,
        text=True
    )
    
    if result.returncode != 0:
        raise Exception(f"JXA error: {result.stderr}")
    
    return result.stdout.strip()

# Usage examples
def get_frontmost_app() -> str:
    """Get the name of the frontmost application."""
    script = 'tell application "System Events" to return name of first process whose frontmost is true'
    return run_applescript(script)

def quit_app(app_name: str) -> None:
    """Quit an application."""
    script = f'tell application "{app_name}" to quit'
    run_applescript(script)
```

### Error Handling Template

```python
import os
import sys
from typing import Optional

class DataExtractionError(Exception):
    """Base exception for data extraction errors."""
    pass

class PermissionError(DataExtractionError):
    """Raised when required permissions are not granted."""
    pass

class DataNotFoundError(DataExtractionError):
    """Raised when expected data files are not found."""
    pass

def safe_extract(extract_func, fallback_value=None, error_message: Optional[str] = None):
    """Safely execute an extraction function with error handling."""
    try:
        return extract_func()
    except PermissionError as e:
        print(f"Permission denied: {error_message or str(e)}", file=sys.stderr)
        print("Please grant the required permissions in System Preferences.", file=sys.stderr)
        return fallback_value
    except DataNotFoundError as e:
        print(f"Data not found: {error_message or str(e)}", file=sys.stderr)
        return fallback_value
    except Exception as e:
        print(f"Unexpected error: {error_message or str(e)}", file=sys.stderr)
        return fallback_value

# Usage
def main():
    bookmarks = safe_extract(
        lambda: extract_bookmarks(),
        fallback_value=[],
        error_message="Could not extract bookmarks"
    )
    
    if bookmarks:
        print(f"Found {len(bookmarks)} bookmarks")
    else:
        print("No bookmarks found or access denied")
```

## Installation Requirements

### Python Dependencies

```bash
# For EventKit access (Calendar and Reminders)
pip install pyobjc-framework-EventKit

# For comprehensive macOS framework access
pip install pyobjc-core pyobjc-framework-Cocoa

# For alternative AppleScript execution
pip install py-applescript
```

### Swift/Objective-C

No additional dependencies needed. Use Xcode or Swift compiler:

```bash
# Compile Swift script
swiftc -o extract_calendar extract_calendar.swift

# Run
./extract_calendar
```

### Shell Scripts

No dependencies needed. Just make scripts executable:

```bash
chmod +x script.sh
./script.sh
```

## Best Practices

1. **Always check permissions first** before attempting data access
2. **Handle errors gracefully** - don't crash on permission denied
3. **Close database connections** properly to avoid corruption
4. **Use context managers** (with statements) in Python
5. **Test on multiple macOS versions** - schemas change
6. **Respect user privacy** - only access what you need
7. **Provide clear error messages** about required permissions
8. **Use official APIs** when available (EventKit, AppleScript)

## Common Issues and Solutions

### Issue: "Operation not permitted" error

**Solution:** Grant Full Disk Access permission in System Preferences.

### Issue: AppleScript returns empty results

**Solution:** Make sure the target app is running and automation permission is granted.

### Issue: Database schema errors

**Solution:** Check macOS version - schemas vary. Query `sqlite_master` table first to inspect schema.

### Issue: Apple timestamp conversion is wrong

**Solution:** Remember Apple uses epoch of 2001-01-01, not Unix epoch of 1970-01-01.

### Issue: EventKit async callbacks never fire

**Solution:** Use a proper run loop or sleep-based waiting mechanism as shown in examples.
