import Cocoa
import Foundation

struct Settings: Codable {
    // MARK: - Hotkey Settings
    var hotkeyEnabled: Bool = true
    var hotkeyKeyCode: UInt32 = 0x00 // A key
    var hotkeyModifiers: [String] = ["command", "shift"] // Cmd+Shift+A
    
    // MARK: - Menu Bar Settings
    var showMenuBarIcon: Bool = true
    var menuBarIconName: String = "note.text" // SF Symbol name
    
    // MARK: - Window Settings
    var windowWidth: CGFloat = 500
    var windowHeight: CGFloat = 300
    var windowTitle: String = "Add Annotation"
    var windowLevel: String = "floating" // "normal", "floating", "modalPanel"
    var windowCentered: Bool = true
    var windowOpacity: CGFloat = 1.0
    
    // MARK: - UI Behavior Settings
    var autoFocusAnnotationField: Bool = true
    var closeOnSave: Bool = true
    var closeOnCancel: Bool = true
    var validateEmptyAnnotation: Bool = true
    var restoreClipboard: Bool = true
    var clipboardCaptureDelay: Double = 0.1
    
    // MARK: - Text Display Settings
    var selectedTextLabel: String = "Selected Text:"
    var selectedTextFontSize: CGFloat = 11
    var selectedTextScrollHeight: CGFloat = 80
    var selectedTextEditable: Bool = false
    var selectedTextSelectable: Bool = true
    
    // MARK: - Annotation Field Settings
    var annotationLabel: String = "Your Annotation:"
    var annotationPlaceholder: String = "Enter your annotation here..."
    var annotationFieldHeight: CGFloat = 24
    var annotationFontSize: CGFloat = 13
    var showHintLabel: Bool = true
    var hintText: String = "Press Cmd+Enter to save"
    
    // MARK: - Button Settings
    var showSaveButton: Bool = true
    var saveButtonTitle: String = "Save (⌘↩)"
    var saveButtonWidth: CGFloat = 100
    var showCancelButton: Bool = true
    var cancelButtonTitle: String = "Cancel"
    var cancelButtonWidth: CGFloat = 80
    
    // MARK: - Color Settings
    var backgroundColor: String? = nil // nil = system default
    var textColor: String? = nil
    var labelColor: String? = nil
    var hintColor: String? = nil
    
    // MARK: - Spacing Settings
    var windowPadding: CGFloat = 20
    var sectionSpacing: CGFloat = 20
    var elementSpacing: CGFloat = 8
    var hintSpacing: CGFloat = 4
    var buttonSpacing: CGFloat = 12
    
    // MARK: - Font Settings
    var headerFontSize: CGFloat = 12
    var headerFontWeight: String = "bold" // "regular", "bold", "semibold"
    var hintFontSize: CGFloat = 10
    
    // MARK: - Storage Settings
    var storageFilePath: String = "~/.text-annotations.json"
    var prettyPrintJSON: Bool = true
    var includeTimestamp: Bool = true
    var timestampFormat: String = "iso8601" // "iso8601", "unix", "custom"
    var customTimestampFormat: String? = nil // For "custom" format
    
    // MARK: - Permissions Settings
    var requestAccessibilityOnLaunch: Bool = true
    var showAccessibilityPrompt: Bool = true
    
    // MARK: - Advanced Settings
    var debugMode: Bool = false
    var logToConsole: Bool = true
    
    // Helper to get window level
    func getWindowLevel() -> NSWindow.Level {
        switch windowLevel.lowercased() {
        case "normal":
            return .normal
        case "floating":
            return .floating
        case "modalpanel":
            return .modalPanel
        case "popupmenu":
            return .popUpMenu
        case "statusbar":
            return .statusBar
        default:
            return .floating
        }
    }
    
    // Helper to get font weight
    func getHeaderFontWeight() -> NSFont.Weight {
        switch headerFontWeight.lowercased() {
        case "regular":
            return .regular
        case "bold":
            return .bold
        case "semibold":
            return .semibold
        case "medium":
            return .medium
        case "light":
            return .light
        case "thin":
            return .thin
        case "heavy":
            return .heavy
        case "black":
            return .black
        default:
            return .bold
        }
    }
    
    // Helper to parse color from hex string
    func getColor(from hexString: String?) -> NSColor? {
        guard let hex = hexString else { return nil }
        
        var hexSanitized = hex.trimmingCharacters(in: .whitespacesAndNewlines)
        hexSanitized = hexSanitized.replacingOccurrences(of: "#", with: "")
        
        var rgb: UInt64 = 0
        Scanner(string: hexSanitized).scanHexInt64(&rgb)
        
        let red = CGFloat((rgb & 0xFF0000) >> 16) / 255.0
        let green = CGFloat((rgb & 0x00FF00) >> 8) / 255.0
        let blue = CGFloat(rgb & 0x0000FF) / 255.0
        
        return NSColor(red: red, green: green, blue: blue, alpha: 1.0)
    }
    
    // Helper to get expanded file path
    func getExpandedStoragePath() -> String {
        return NSString(string: storageFilePath).expandingTildeInPath
    }
}

class SettingsManager {
    static let shared = SettingsManager()
    private(set) var settings: Settings
    
    private let configFileName = "config.json"
    private var configFilePath: URL {
        let homeDir = FileManager.default.homeDirectoryForCurrentUser
        return homeDir.appendingPathComponent(".text-annotator").appendingPathComponent(configFileName)
    }
    
    private init() {
        // Try to load settings from file
        if let loadedSettings = SettingsManager.loadSettings() {
            self.settings = loadedSettings
            print("[Settings] Loaded from \(configFilePath.path)")
        } else {
            // Use default settings
            self.settings = Settings()
            print("[Settings] Using default settings")
            
            // Save default settings to file
            self.save()
        }
    }
    
    private static func loadSettings() -> Settings? {
        let manager = FileManager.default
        let homeDir = manager.homeDirectoryForCurrentUser
        let configDir = homeDir.appendingPathComponent(".text-annotator")
        let configFile = configDir.appendingPathComponent("config.json")
        
        guard manager.fileExists(atPath: configFile.path) else {
            return nil
        }
        
        guard let data = try? Data(contentsOf: configFile) else {
            return nil
        }
        
        let decoder = JSONDecoder()
        return try? decoder.decode(Settings.self, from: data)
    }
    
    func save() {
        let manager = FileManager.default
        let homeDir = manager.homeDirectoryForCurrentUser
        let configDir = homeDir.appendingPathComponent(".text-annotator")
        
        // Create directory if it doesn't exist
        if !manager.fileExists(atPath: configDir.path) {
            try? manager.createDirectory(at: configDir, withIntermediateDirectories: true)
        }
        
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
        
        if let data = try? encoder.encode(settings) {
            try? data.write(to: configFilePath)
            print("[Settings] Saved to \(configFilePath.path)")
        }
    }
    
    func reload() {
        if let loadedSettings = SettingsManager.loadSettings() {
            self.settings = loadedSettings
            print("[Settings] Reloaded from \(configFilePath.path)")
        }
    }
}
