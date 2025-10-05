import Cocoa

class AnnotationWindow: NSWindowController, NSWindowDelegate, NSTextFieldDelegate {
    let selectedText: String
    let settings: Settings
    var textLabel: NSTextField!
    var annotationField: NSTextField!
    var saveButton: NSButton!
    var cancelButton: NSButton!
    
    init(selectedText: String, settings: Settings) {
        self.selectedText = selectedText
        self.settings = settings
        
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: settings.windowWidth, height: settings.windowHeight),
            styleMask: [.titled, .closable, .miniaturizable],
            backing: .buffered,
            defer: false
        )
        
        super.init(window: window)
        
        window.delegate = self
        window.title = settings.windowTitle
        if settings.windowCentered {
            window.center()
        }
        window.level = settings.getWindowLevel()
        window.alphaValue = settings.windowOpacity
        
        // Apply background color if specified
        if let bgColor = settings.getColor(from: settings.backgroundColor) {
            window.backgroundColor = bgColor
        }
        
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    func setupUI() {
        guard let contentView = window?.contentView else { return }
        
        // Selected text label header
        let headerLabel = NSTextField(labelWithString: settings.selectedTextLabel)
        headerLabel.font = NSFont.systemFont(ofSize: settings.headerFontSize, weight: settings.getHeaderFontWeight())
        headerLabel.translatesAutoresizingMaskIntoConstraints = false
        if let labelColor = settings.getColor(from: settings.labelColor) {
            headerLabel.textColor = labelColor
        }
        contentView.addSubview(headerLabel)
        
        // Selected text display (read-only)
        let scrollView = NSScrollView(frame: .zero)
        scrollView.translatesAutoresizingMaskIntoConstraints = false
        scrollView.hasVerticalScroller = true
        scrollView.autohidesScrollers = true
        scrollView.borderType = .bezelBorder
        
        textLabel = NSTextField(labelWithString: selectedText)
        textLabel.isEditable = settings.selectedTextEditable
        textLabel.isSelectable = settings.selectedTextSelectable
        textLabel.lineBreakMode = .byWordWrapping
        textLabel.maximumNumberOfLines = 0
        textLabel.font = NSFont.systemFont(ofSize: settings.selectedTextFontSize)
        textLabel.translatesAutoresizingMaskIntoConstraints = false
        if let textColor = settings.getColor(from: settings.textColor) {
            textLabel.textColor = textColor
        }
        
        scrollView.documentView = textLabel
        contentView.addSubview(scrollView)
        
        // Annotation label
        let annotationLabel = NSTextField(labelWithString: settings.annotationLabel)
        annotationLabel.font = NSFont.systemFont(ofSize: settings.headerFontSize, weight: settings.getHeaderFontWeight())
        annotationLabel.translatesAutoresizingMaskIntoConstraints = false
        if let labelColor = settings.getColor(from: settings.labelColor) {
            annotationLabel.textColor = labelColor
        }
        contentView.addSubview(annotationLabel)
        
        // Annotation input field
        annotationField = NSTextField(frame: .zero)
        annotationField.placeholderString = settings.annotationPlaceholder
        annotationField.font = NSFont.systemFont(ofSize: settings.annotationFontSize)
        annotationField.translatesAutoresizingMaskIntoConstraints = false
        annotationField.delegate = self
        contentView.addSubview(annotationField)
        
        // Hint label
        var hintLabel: NSTextField? = nil
        if settings.showHintLabel {
            hintLabel = NSTextField(labelWithString: settings.hintText)
            hintLabel!.font = NSFont.systemFont(ofSize: settings.hintFontSize)
            hintLabel!.textColor = settings.getColor(from: settings.hintColor) ?? .secondaryLabelColor
            hintLabel!.translatesAutoresizingMaskIntoConstraints = false
            contentView.addSubview(hintLabel!)
        }
        
        // Buttons
        if settings.showSaveButton {
            saveButton = NSButton(title: settings.saveButtonTitle, target: self, action: #selector(saveAnnotation))
            saveButton.keyEquivalent = "\r"
            saveButton.keyEquivalentModifierMask = .command
            saveButton.translatesAutoresizingMaskIntoConstraints = false
            contentView.addSubview(saveButton)
        }
        
        if settings.showCancelButton {
            cancelButton = NSButton(title: settings.cancelButtonTitle, target: self, action: #selector(cancel))
            cancelButton.keyEquivalent = "\u{1b}" // Escape key
            cancelButton.translatesAutoresizingMaskIntoConstraints = false
            contentView.addSubview(cancelButton)
        }
        
        // Layout constraints
        var constraints: [NSLayoutConstraint] = [
            // Header label
            headerLabel.topAnchor.constraint(equalTo: contentView.topAnchor, constant: settings.windowPadding),
            headerLabel.leadingAnchor.constraint(equalTo: contentView.leadingAnchor, constant: settings.windowPadding),
            headerLabel.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -settings.windowPadding),
            
            // Selected text scroll view
            scrollView.topAnchor.constraint(equalTo: headerLabel.bottomAnchor, constant: settings.elementSpacing),
            scrollView.leadingAnchor.constraint(equalTo: contentView.leadingAnchor, constant: settings.windowPadding),
            scrollView.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -settings.windowPadding),
            scrollView.heightAnchor.constraint(equalToConstant: settings.selectedTextScrollHeight),
            
            // Text label inside scroll view
            textLabel.topAnchor.constraint(equalTo: scrollView.topAnchor),
            textLabel.leadingAnchor.constraint(equalTo: scrollView.leadingAnchor),
            textLabel.trailingAnchor.constraint(equalTo: scrollView.trailingAnchor),
            
            // Annotation label
            annotationLabel.topAnchor.constraint(equalTo: scrollView.bottomAnchor, constant: settings.sectionSpacing),
            annotationLabel.leadingAnchor.constraint(equalTo: contentView.leadingAnchor, constant: settings.windowPadding),
            annotationLabel.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -settings.windowPadding),
            
            // Annotation field
            annotationField.topAnchor.constraint(equalTo: annotationLabel.bottomAnchor, constant: settings.elementSpacing),
            annotationField.leadingAnchor.constraint(equalTo: contentView.leadingAnchor, constant: settings.windowPadding),
            annotationField.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -settings.windowPadding),
            annotationField.heightAnchor.constraint(equalToConstant: settings.annotationFieldHeight),
        ]
        
        // Hint label constraints (if enabled)
        if let hintLabel = hintLabel {
            constraints.append(contentsOf: [
                hintLabel.topAnchor.constraint(equalTo: annotationField.bottomAnchor, constant: settings.hintSpacing),
                hintLabel.leadingAnchor.constraint(equalTo: contentView.leadingAnchor, constant: settings.windowPadding),
            ])
        }
        
        // Button constraints
        if settings.showCancelButton && settings.showSaveButton {
            constraints.append(contentsOf: [
                cancelButton.bottomAnchor.constraint(equalTo: contentView.bottomAnchor, constant: -settings.windowPadding),
                cancelButton.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -settings.windowPadding),
                cancelButton.widthAnchor.constraint(equalToConstant: settings.cancelButtonWidth),
                
                saveButton.bottomAnchor.constraint(equalTo: contentView.bottomAnchor, constant: -settings.windowPadding),
                saveButton.trailingAnchor.constraint(equalTo: cancelButton.leadingAnchor, constant: -settings.buttonSpacing),
                saveButton.widthAnchor.constraint(equalToConstant: settings.saveButtonWidth),
            ])
        } else if settings.showSaveButton {
            constraints.append(contentsOf: [
                saveButton.bottomAnchor.constraint(equalTo: contentView.bottomAnchor, constant: -settings.windowPadding),
                saveButton.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -settings.windowPadding),
                saveButton.widthAnchor.constraint(equalToConstant: settings.saveButtonWidth),
            ])
        } else if settings.showCancelButton {
            constraints.append(contentsOf: [
                cancelButton.bottomAnchor.constraint(equalTo: contentView.bottomAnchor, constant: -settings.windowPadding),
                cancelButton.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -settings.windowPadding),
                cancelButton.widthAnchor.constraint(equalToConstant: settings.cancelButtonWidth),
            ])
        }
        
        NSLayoutConstraint.activate(constraints)
        
        // Focus on annotation field
        if settings.autoFocusAnnotationField {
            window?.makeFirstResponder(annotationField)
        }
    }
    
    @objc func saveAnnotation() {
        let annotation = annotationField.stringValue.trimmingCharacters(in: .whitespacesAndNewlines)
        
        if settings.validateEmptyAnnotation && annotation.isEmpty {
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
        
        // Close window if configured
        if settings.closeOnSave {
            close()
        }
    }
    
    @objc func cancel() {
        if settings.closeOnCancel {
            close()
        }
    }
    
    func saveToJSON(selectedText: String, annotation: String) {
        let jsonFilePath = URL(fileURLWithPath: settings.getExpandedStoragePath())
        
        // Create directory if needed
        let directory = jsonFilePath.deletingLastPathComponent()
        try? FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        
        var annotations: [[String: Any]] = []
        
        // Load existing annotations if file exists
        if FileManager.default.fileExists(atPath: jsonFilePath.path) {
            if let data = try? Data(contentsOf: jsonFilePath),
               let json = try? JSONSerialization.jsonObject(with: data) as? [[String: Any]] {
                annotations = json
            }
        }
        
        // Add new annotation with optional timestamp
        var newAnnotation: [String: Any] = [
            "selectedText": selectedText,
            "annotation": annotation
        ]
        
        if settings.includeTimestamp {
            let timestamp: Any
            switch settings.timestampFormat.lowercased() {
            case "unix":
                timestamp = Date().timeIntervalSince1970
            case "custom":
                if let format = settings.customTimestampFormat {
                    let formatter = DateFormatter()
                    formatter.dateFormat = format
                    timestamp = formatter.string(from: Date())
                } else {
                    timestamp = ISO8601DateFormatter().string(from: Date())
                }
            default: // iso8601
                timestamp = ISO8601DateFormatter().string(from: Date())
            }
            newAnnotation["timestamp"] = timestamp
        }
        
        annotations.append(newAnnotation)
        
        // Save to file
        let options: JSONSerialization.WritingOptions = settings.prettyPrintJSON ? .prettyPrinted : []
        if let jsonData = try? JSONSerialization.data(withJSONObject: annotations, options: options) {
            try? jsonData.write(to: jsonFilePath)
            if settings.logToConsole {
                print("Annotation saved to \(jsonFilePath.path)")
            }
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
