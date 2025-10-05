import Cocoa
import Carbon

@main
class AppDelegate: NSObject, NSApplicationDelegate {
    var statusItem: NSStatusItem?
    var annotationWindow: AnnotationWindow?
    var hotKeyRef: EventHotKeyRef?
    let settings = SettingsManager.shared.settings
    
    func applicationDidFinishLaunching(_ aNotification: Notification) {
        if settings.debugMode {
            print("[Debug] Application launched with settings:")
            print("[Debug] Config file: \(FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent(".text-annotator/config.json").path)")
        }
        
        // Set up status bar item
        if settings.showMenuBarIcon {
            statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
            if let button = statusItem?.button {
                button.image = NSImage(systemSymbolName: settings.menuBarIconName, accessibilityDescription: "Text Annotator")
                button.action = #selector(statusBarButtonClicked)
            }
        }
        
        // Register global hotkey
        if settings.hotkeyEnabled {
            registerGlobalHotkey()
        }
        
        // Request accessibility permissions
        if settings.requestAccessibilityOnLaunch {
            requestAccessibilityPermissions()
        }
    }
    
    @objc func statusBarButtonClicked() {
        showAnnotationWindow()
    }
    
    func requestAccessibilityPermissions() {
        let promptValue = settings.showAccessibilityPrompt
        let options: NSDictionary = [kAXTrustedCheckOptionPrompt.takeRetainedValue() as String: promptValue]
        AXIsProcessTrustedWithOptions(options)
    }
    
    func registerGlobalHotkey() {
        var hotKeyID = EventHotKeyID()
        hotKeyID.signature = OSType("TXAN".fourCharCodeValue)
        hotKeyID.id = 1
        
        var eventHandler: EventHandlerRef?
        
        let eventSpec = [
            EventTypeSpec(eventClass: OSType(kEventClassKeyboard), eventKind: UInt32(kEventHotKeyPressed))
        ]
        
        InstallEventHandler(GetApplicationEventTarget(), { (nextHandler, theEvent, userData) -> OSStatus in
            guard let appDelegate = userData?.load(as: AppDelegate.self) else {
                return OSStatus(eventNotHandledErr)
            }
            appDelegate.showAnnotationWindow()
            return noErr
        }, 1, eventSpec, Unmanaged.passUnretained(self).toOpaque(), &eventHandler)
        
        // Parse modifiers from settings
        var modifiers: UInt32 = 0
        for modifier in settings.hotkeyModifiers {
            switch modifier.lowercased() {
            case "command", "cmd":
                modifiers |= UInt32(cmdKey)
            case "shift":
                modifiers |= UInt32(shiftKey)
            case "option", "alt":
                modifiers |= UInt32(optionKey)
            case "control", "ctrl":
                modifiers |= UInt32(controlKey)
            default:
                break
            }
        }
        
        // Register hotkey with custom key code and modifiers
        let status = RegisterEventHotKey(
            settings.hotkeyKeyCode,
            modifiers,
            hotKeyID,
            GetApplicationEventTarget(),
            0,
            &hotKeyRef
        )
        
        if status != noErr {
            print("Failed to register hotkey: \(status)")
        } else if settings.debugMode {
            print("[Debug] Hotkey registered: keyCode=\(settings.hotkeyKeyCode), modifiers=\(settings.hotkeyModifiers)")
        }
    }
    
    func showAnnotationWindow() {
        // Get selected text from system
        let selectedText = getSelectedText()
        
        if settings.debugMode {
            print("[Debug] Showing annotation window with text: \(selectedText.prefix(50))...")
        }
        
        // Create and show annotation window
        annotationWindow = AnnotationWindow(selectedText: selectedText, settings: settings)
        annotationWindow?.showWindow(nil)
        annotationWindow?.window?.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
    }
    
    func getSelectedText() -> String {
        // Save current clipboard
        let pasteboard = NSPasteboard.general
        let oldContents = settings.restoreClipboard ? pasteboard.string(forType: .string) : nil
        
        // Simulate Cmd+C to copy selected text
        let source = CGEventSource(stateID: .combinedSessionState)
        
        let keyDownEvent = CGEvent(keyboardEventSource: source, virtualKey: 0x08, keyDown: true) // C key
        keyDownEvent?.flags = .maskCommand
        
        let keyUpEvent = CGEvent(keyboardEventSource: source, virtualKey: 0x08, keyDown: false)
        keyUpEvent?.flags = .maskCommand
        
        keyDownEvent?.post(tap: .cghidEventTap)
        keyUpEvent?.post(tap: .cghidEventTap)
        
        // Wait for clipboard to update (configurable delay)
        Thread.sleep(forTimeInterval: settings.clipboardCaptureDelay)
        
        // Get the new clipboard contents
        let selectedText = pasteboard.string(forType: .string) ?? ""
        
        // Restore old clipboard if it was different and restore is enabled
        if settings.restoreClipboard, let oldContents = oldContents, oldContents != selectedText {
            pasteboard.clearContents()
            pasteboard.setString(oldContents, forType: .string)
        }
        
        return selectedText
    }
    
    func applicationWillTerminate(_ aNotification: Notification) {
        if let hotKeyRef = hotKeyRef {
            UnregisterEventHotKey(hotKeyRef)
        }
    }
}

extension String {
    var fourCharCodeValue: Int {
        var result: Int = 0
        for char in self.utf8 {
            result = result << 8 + Int(char)
        }
        return result
    }
}
