import Cocoa

class AnnotationWindow: NSWindowController, NSWindowDelegate, NSTextFieldDelegate {
    let selectedText: String
    var textLabel: NSTextField!
    var annotationField: NSTextField!
    var saveButton: NSButton!
    var cancelButton: NSButton!
    
    init(selectedText: String) {
        self.selectedText = selectedText
        
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 500, height: 300),
            styleMask: [.titled, .closable, .miniaturizable],
            backing: .buffered,
            defer: false
        )
        
        super.init(window: window)
        
        window.delegate = self
        window.title = "Add Annotation"
        window.center()
        window.level = .floating
        
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    func setupUI() {
        guard let contentView = window?.contentView else { return }
        
        // Selected text label header
        let headerLabel = NSTextField(labelWithString: "Selected Text:")
        headerLabel.font = NSFont.boldSystemFont(ofSize: 12)
        headerLabel.translatesAutoresizingMaskIntoConstraints = false
        contentView.addSubview(headerLabel)
        
        // Selected text display (read-only)
        let scrollView = NSScrollView(frame: .zero)
        scrollView.translatesAutoresizingMaskIntoConstraints = false
        scrollView.hasVerticalScroller = true
        scrollView.autohidesScrollers = true
        scrollView.borderType = .bezelBorder
        
        textLabel = NSTextField(labelWithString: selectedText)
        textLabel.isEditable = false
        textLabel.isSelectable = true
        textLabel.lineBreakMode = .byWordWrapping
        textLabel.maximumNumberOfLines = 0
        textLabel.font = NSFont.systemFont(ofSize: 11)
        textLabel.translatesAutoresizingMaskIntoConstraints = false
        
        scrollView.documentView = textLabel
        contentView.addSubview(scrollView)
        
        // Annotation label
        let annotationLabel = NSTextField(labelWithString: "Your Annotation:")
        annotationLabel.font = NSFont.boldSystemFont(ofSize: 12)
        annotationLabel.translatesAutoresizingMaskIntoConstraints = false
        contentView.addSubview(annotationLabel)
        
        // Annotation input field
        annotationField = NSTextField(frame: .zero)
        annotationField.placeholderString = "Enter your annotation here..."
        annotationField.translatesAutoresizingMaskIntoConstraints = false
        annotationField.delegate = self
        contentView.addSubview(annotationField)
        
        // Hint label
        let hintLabel = NSTextField(labelWithString: "Press Cmd+Enter to save")
        hintLabel.font = NSFont.systemFont(ofSize: 10)
        hintLabel.textColor = .secondaryLabelColor
        hintLabel.translatesAutoresizingMaskIntoConstraints = false
        contentView.addSubview(hintLabel)
        
        // Buttons
        saveButton = NSButton(title: "Save (⌘↩)", target: self, action: #selector(saveAnnotation))
        saveButton.keyEquivalent = "\r"
        saveButton.keyEquivalentModifierMask = .command
        saveButton.translatesAutoresizingMaskIntoConstraints = false
        contentView.addSubview(saveButton)
        
        cancelButton = NSButton(title: "Cancel", target: self, action: #selector(cancel))
        cancelButton.keyEquivalent = "\u{1b}" // Escape key
        cancelButton.translatesAutoresizingMaskIntoConstraints = false
        contentView.addSubview(cancelButton)
        
        // Layout constraints
        NSLayoutConstraint.activate([
            // Header label
            headerLabel.topAnchor.constraint(equalTo: contentView.topAnchor, constant: 20),
            headerLabel.leadingAnchor.constraint(equalTo: contentView.leadingAnchor, constant: 20),
            headerLabel.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -20),
            
            // Selected text scroll view
            scrollView.topAnchor.constraint(equalTo: headerLabel.bottomAnchor, constant: 8),
            scrollView.leadingAnchor.constraint(equalTo: contentView.leadingAnchor, constant: 20),
            scrollView.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -20),
            scrollView.heightAnchor.constraint(equalToConstant: 80),
            
            // Text label inside scroll view
            textLabel.topAnchor.constraint(equalTo: scrollView.topAnchor),
            textLabel.leadingAnchor.constraint(equalTo: scrollView.leadingAnchor),
            textLabel.trailingAnchor.constraint(equalTo: scrollView.trailingAnchor),
            
            // Annotation label
            annotationLabel.topAnchor.constraint(equalTo: scrollView.bottomAnchor, constant: 20),
            annotationLabel.leadingAnchor.constraint(equalTo: contentView.leadingAnchor, constant: 20),
            annotationLabel.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -20),
            
            // Annotation field
            annotationField.topAnchor.constraint(equalTo: annotationLabel.bottomAnchor, constant: 8),
            annotationField.leadingAnchor.constraint(equalTo: contentView.leadingAnchor, constant: 20),
            annotationField.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -20),
            annotationField.heightAnchor.constraint(equalToConstant: 24),
            
            // Hint label
            hintLabel.topAnchor.constraint(equalTo: annotationField.bottomAnchor, constant: 4),
            hintLabel.leadingAnchor.constraint(equalTo: contentView.leadingAnchor, constant: 20),
            
            // Buttons
            cancelButton.bottomAnchor.constraint(equalTo: contentView.bottomAnchor, constant: -20),
            cancelButton.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -20),
            cancelButton.widthAnchor.constraint(equalToConstant: 80),
            
            saveButton.bottomAnchor.constraint(equalTo: contentView.bottomAnchor, constant: -20),
            saveButton.trailingAnchor.constraint(equalTo: cancelButton.leadingAnchor, constant: -12),
            saveButton.widthAnchor.constraint(equalToConstant: 100),
        ])
        
        // Focus on annotation field
        window?.makeFirstResponder(annotationField)
    }
    
    @objc func saveAnnotation() {
        let annotation = annotationField.stringValue.trimmingCharacters(in: .whitespacesAndNewlines)
        
        if annotation.isEmpty {
            let alert = NSAlert()
            alert.messageText = "Empty Annotation"
            alert.informativeText = "Please enter an annotation before saving."
            alert.alertStyle = .warning
            alert.addButton(withTitle: "OK")
            alert.runModal()
            return
        }
        
        // Save to JSON file
        saveToJSON(selectedText: selectedText, annotation: annotation)
        
        // Close window
        close()
    }
    
    @objc func cancel() {
        close()
    }
    
    func saveToJSON(selectedText: String, annotation: String) {
        let homeDir = FileManager.default.homeDirectoryForCurrentUser
        let jsonFilePath = homeDir.appendingPathComponent(".text-annotations.json")
        
        var annotations: [[String: Any]] = []
        
        // Load existing annotations if file exists
        if FileManager.default.fileExists(atPath: jsonFilePath.path) {
            if let data = try? Data(contentsOf: jsonFilePath),
               let json = try? JSONSerialization.jsonObject(with: data) as? [[String: Any]] {
                annotations = json
            }
        }
        
        // Add new annotation
        let newAnnotation: [String: Any] = [
            "timestamp": ISO8601DateFormatter().string(from: Date()),
            "selectedText": selectedText,
            "annotation": annotation
        ]
        annotations.append(newAnnotation)
        
        // Save to file
        if let jsonData = try? JSONSerialization.data(withJSONObject: annotations, options: .prettyPrinted) {
            try? jsonData.write(to: jsonFilePath)
            print("Annotation saved to \(jsonFilePath.path)")
        }
    }
    
    func control(_ control: NSControl, textView: NSTextView, doCommandBy commandSelector: Selector) -> Bool {
        // Handle Cmd+Enter in text field
        if commandSelector == #selector(NSResponder.insertNewline(_:)) {
            if NSEvent.modifierFlags.contains(.command) {
                saveAnnotation()
                return true
            }
        }
        return false
    }
}
