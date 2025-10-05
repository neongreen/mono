import Cocoa
import Carbon

@main
class AppDelegate: NSObject, NSApplicationDelegate {
    var statusItem: NSStatusItem?
    var annotationWindow: AnnotationWindow?
    var hotKeyRef: EventHotKeyRef?
    
    func applicationDidFinishLaunching(_ aNotification: Notification) {
        // Set up status bar item
        statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
        if let button = statusItem?.button {
            button.image = NSImage(systemSymbolName: "note.text", accessibilityDescription: "Text Annotator")
            button.action = #selector(statusBarButtonClicked)
        }
        
        // Register global hotkey (Cmd+Shift+A)
        registerGlobalHotkey()
        
        // Request accessibility permissions
        requestAccessibilityPermissions()
    }
    
    @objc func statusBarButtonClicked() {
        showAnnotationWindow()
    }
    
    func requestAccessibilityPermissions() {
        let options: NSDictionary = [kAXTrustedCheckOptionPrompt.takeRetainedValue() as String: true]
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
        
        // Register Cmd+Shift+A
        let status = RegisterEventHotKey(
            UInt32(kVK_ANSI_A),
            UInt32(cmdKey | shiftKey),
            hotKeyID,
            GetApplicationEventTarget(),
            0,
            &hotKeyRef
        )
        
        if status != noErr {
            print("Failed to register hotkey: \(status)")
        }
    }
    
    func showAnnotationWindow() {
        // Get selected text from system
        let selectedText = getSelectedText()
        
        // Create and show annotation window
        annotationWindow = AnnotationWindow(selectedText: selectedText)
        annotationWindow?.showWindow(nil)
        annotationWindow?.window?.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
    }
    
    func getSelectedText() -> String {
        // Save current clipboard
        let pasteboard = NSPasteboard.general
        let oldContents = pasteboard.string(forType: .string)
        
        // Simulate Cmd+C to copy selected text
        let source = CGEventSource(stateID: .combinedSessionState)
        
        let keyDownEvent = CGEvent(keyboardEventSource: source, virtualKey: 0x08, keyDown: true) // C key
        keyDownEvent?.flags = .maskCommand
        
        let keyUpEvent = CGEvent(keyboardEventSource: source, virtualKey: 0x08, keyDown: false)
        keyUpEvent?.flags = .maskCommand
        
        keyDownEvent?.post(tap: .cghidEventTap)
        keyUpEvent?.post(tap: .cghidEventTap)
        
        // Wait a bit for clipboard to update
        Thread.sleep(forTimeInterval: 0.1)
        
        // Get the new clipboard contents
        let selectedText = pasteboard.string(forType: .string) ?? ""
        
        // Restore old clipboard if it was different
        if let oldContents = oldContents, oldContents != selectedText {
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
