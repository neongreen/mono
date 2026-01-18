# macOS Application Data Extraction

Research and documentation on programmatic extraction of data from native macOS applications.

## Overview

This folder contains comprehensive research on how to programmatically extract data from macOS system applications including Safari, Notes, Voice Memos, Reminders, Notifications, and Calendar.

## Contents

- **[RESEARCH_REPORT.md](RESEARCH_REPORT.md)** - Complete research report covering all extraction methods, tools, and approaches
- **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - Fast lookup guide with one-liners and decision tree
- **[CODE_EXAMPLES.md](CODE_EXAMPLES.md)** - Ready-to-use code examples in Python, AppleScript, and Swift

## What's Covered

### Applications
- Safari (bookmarks, tabs, reading list, history)
- Notes app
- Voice Memos
- Reminders
- Notifications
- Calendar

### Extraction Methods
1. **Direct Database Access** - SQLite and plist file parsing
2. **AppleScript/JXA** - Using Apple's official scripting interfaces
3. **UI Automation** - Accessibility-based interaction
4. **System APIs** - EventKit and other macOS frameworks

### Additional Resources
- Open source projects for macOS data extraction
- Security and privacy considerations
- Code examples in Python, Swift, and AppleScript
- Comparison of different approaches
- Best practices and recommendations

## Quick Links

For specific use cases:
- **Safari data**: See [Safari Data Extraction](RESEARCH_REPORT.md#safari-data-extraction) section
- **Notes**: See [Notes App Data Extraction](RESEARCH_REPORT.md#notes-app-data-extraction) section
- **Voice Memos**: See [Voice Memos Data Extraction](RESEARCH_REPORT.md#voice-memos-data-extraction) section
- **Reminders**: See [Reminders Data Extraction](RESEARCH_REPORT.md#reminders-data-extraction) section
- **Notifications**: See [Notifications Data Extraction](RESEARCH_REPORT.md#notifications-data-extraction) section
- **Calendar**: See [Calendar Data Extraction](RESEARCH_REPORT.md#calendar-data-extraction) section

## Method Comparison Table

| Method | Speed | Reliability | Permissions | Complexity |
|--------|-------|-------------|-------------|------------|
| SQLite/Plist | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | Full Disk Access | ⭐⭐⭐⭐ |
| AppleScript/JXA | ⭐⭐⭐ | ⭐⭐⭐⭐ | Automation | ⭐⭐ |
| UI Automation | ⭐ | ⭐ | Accessibility | ⭐⭐⭐⭐⭐ |
| System APIs | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Specific | ⭐⭐⭐⭐ |

## Recommended Approaches

- **Safari**: Direct plist/SQLite access for bookmarks and history; AppleScript for current tabs
- **Notes**: AppleScript/JXA (most reliable)
- **Voice Memos**: Direct file system + SQLite (only option)
- **Reminders**: EventKit framework or AppleScript
- **Notifications**: Real-time monitoring only (history not accessible)
- **Calendar**: EventKit framework or AppleScript

## Getting Started

1. Read the [full research report](RESEARCH_REPORT.md)
2. Choose your programming language (Python, Swift, or Shell/AppleScript)
3. Select the appropriate method for your app and requirements
4. Review permission requirements for your chosen approach
5. Check out the open source projects section for existing tools

## Security Note

All data extraction methods require appropriate macOS permissions:
- **Full Disk Access** - for direct database/file access
- **Automation** - for AppleScript/JXA
- **Accessibility** - for UI automation
- **App-specific** - Calendar, Reminders, etc.

Always respect user privacy and request only the permissions you need.
