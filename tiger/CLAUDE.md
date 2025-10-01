# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A native macOS to-do list application built using Swift Package Manager and AppKit (Cocoa). The app avoids Xcode and uses a custom Python build script to package the Swift executable into a `.app` bundle.

## Build System

This project uses Swift Package Manager but **does not use Xcode**. Instead, it has a custom build pipeline:

1. `swift build -c release` - Compiles the executable
2. `build_app.py` - Packages the executable into a macOS `.app` bundle structure

**Build and run:**
```bash
./build_app.py
open MacApp.app
```

**Quick development iteration:**
```bash
swift run  # Run without packaging
```

The build script:
- Creates `MacApp.app/` bundle structure
- Copies the compiled executable to `Contents/MacOS/`
- Copies `Info.plist` to `Contents/`

## Architecture

**Single-file application:** All code lives in `Sources/MacApp/main.swift`

The app uses imperative AppKit (no SwiftUI, no Storyboards):
- `AppDelegate` handles application lifecycle and owns all UI state
- Direct instantiation of `NSWindow`, `NSTableView`, `NSTextField`, etc.
- Manual layout using frames and autoresizing masks (no Auto Layout)
- `AppDelegate` conforms to `NSTableViewDataSource` and `NSTableViewDelegate`

**Current state:** The app stores to-do items in memory only (`items: [String]` array). No persistence yet.

## Key Files

- `Package.swift` - Swift Package Manager configuration (executable target, macOS 13+ platform)
- `Sources/MacApp/main.swift` - Entire application code
- `Info.plist` - App bundle metadata (bundle ID: `com.neongreen.MacApp`)
- `build_app.py` - Python script that packages the executable into `.app` bundle
