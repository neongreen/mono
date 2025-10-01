import Cocoa
import SQLite

class AppDelegate: NSObject, NSApplicationDelegate, NSTableViewDataSource, NSTableViewDelegate, NSWindowDelegate {
    var window: NSWindow!
    var tableView: NSTableView!
    var textField: NSTextField!
    var items: [String] = []

    var db: Connection!
    let todosTable = Table("todos")
    let id = Expression<Int64>("id")
    let text = Expression<String>("text")

    func applicationDidFinishLaunching(_ notification: Notification) {
        setupMenu()
        setupDatabase()
        loadItems()

        window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 480, height: 400),
            styleMask: [.titled, .closable, .miniaturizable, .resizable],
            backing: .buffered,
            defer: false
        )
        window.center()
        window.title = "Tiger"
        window.delegate = self

        setupUI()
        window.makeKeyAndOrderFront(nil)
    }

    func setupMenu() {
        let mainMenu = NSMenu()

        // App menu
        let appMenuItem = NSMenuItem()
        mainMenu.addItem(appMenuItem)

        let appMenu = NSMenu()
        appMenuItem.submenu = appMenu

        appMenu.addItem(NSMenuItem(title: "About Tiger", action: #selector(NSApplication.orderFrontStandardAboutPanel(_:)), keyEquivalent: ""))
        appMenu.addItem(NSMenuItem.separator())
        appMenu.addItem(NSMenuItem(title: "Hide Tiger", action: #selector(NSApplication.hide(_:)), keyEquivalent: "h"))

        let hideOthersItem = NSMenuItem(title: "Hide Others", action: #selector(NSApplication.hideOtherApplications(_:)), keyEquivalent: "h")
        hideOthersItem.keyEquivalentModifierMask = [.command, .option]
        appMenu.addItem(hideOthersItem)

        appMenu.addItem(NSMenuItem(title: "Show All", action: #selector(NSApplication.unhideAllApplications(_:)), keyEquivalent: ""))
        appMenu.addItem(NSMenuItem.separator())
        appMenu.addItem(NSMenuItem(title: "Quit Tiger", action: #selector(NSApplication.terminate(_:)), keyEquivalent: "q"))

        // Edit menu
        let editMenuItem = NSMenuItem()
        mainMenu.addItem(editMenuItem)

        let editMenu = NSMenu(title: "Edit")
        editMenuItem.submenu = editMenu

        editMenu.addItem(NSMenuItem(title: "Cut", action: #selector(NSText.cut(_:)), keyEquivalent: "x"))
        editMenu.addItem(NSMenuItem(title: "Copy", action: #selector(NSText.copy(_:)), keyEquivalent: "c"))
        editMenu.addItem(NSMenuItem(title: "Paste", action: #selector(NSText.paste(_:)), keyEquivalent: "v"))
        editMenu.addItem(NSMenuItem(title: "Select All", action: #selector(NSText.selectAll(_:)), keyEquivalent: "a"))

        NSApp.mainMenu = mainMenu
    }

    func setupDatabase() {
        do {
            let path = NSSearchPathForDirectoriesInDomains(
                .applicationSupportDirectory, .userDomainMask, true
            ).first! + "/Tiger"

            try FileManager.default.createDirectory(
                atPath: path,
                withIntermediateDirectories: true,
                attributes: nil
            )

            db = try Connection("\(path)/todos.sqlite3")

            try db.run(todosTable.create(ifNotExists: true) { t in
                t.column(id, primaryKey: .autoincrement)
                t.column(text)
            })
        } catch {
            print("Database setup error: \(error)")
        }
    }

    func loadItems() {
        do {
            items = try db.prepare(todosTable).map { row in
                row[text]
            }
        } catch {
            print("Failed to load items: \(error)")
        }
    }

    func setupUI() {
        let contentView = NSView(frame: window.contentView!.bounds)
        contentView.autoresizingMask = [.width, .height]

        // Create table view
        tableView = NSTableView(frame: .zero)
        tableView.autoresizingMask = [.width, .height]

        let column = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("TodoColumn"))
        column.title = "To-Do Items"
        column.width = 400
        tableView.addTableColumn(column)

        tableView.dataSource = self
        tableView.delegate = self
        tableView.headerView = nil

        // Create scroll view for table
        let scrollView = NSScrollView(frame: NSRect(x: 20, y: 60, width: 440, height: 300))
        scrollView.documentView = tableView
        scrollView.hasVerticalScroller = true
        scrollView.autoresizingMask = [.width, .height]

        // Create text field
        textField = NSTextField(frame: NSRect(x: 20, y: 20, width: 340, height: 24))
        textField.placeholderString = "Enter new to-do item..."
        textField.autoresizingMask = [.width]

        // Create add button
        let addButton = NSButton(frame: NSRect(x: 370, y: 20, width: 90, height: 24))
        addButton.title = "Add"
        addButton.bezelStyle = .rounded
        addButton.target = self
        addButton.action = #selector(addItem)
        addButton.autoresizingMask = [.minXMargin]
        addButton.keyEquivalent = "\r" // Enter key

        contentView.addSubview(scrollView)
        contentView.addSubview(textField)
        contentView.addSubview(addButton)

        window.contentView = contentView
    }

    @objc func addItem() {
        let itemText = textField.stringValue.trimmingCharacters(in: .whitespaces)
        guard !itemText.isEmpty else { return }

        do {
            try db.run(todosTable.insert(text <- itemText))
            items.append(itemText)
            tableView.reloadData()
            textField.stringValue = ""
        } catch {
            print("Failed to insert item: \(error)")
        }
    }

    func addItemFromText(_ itemText: String) {
        guard !itemText.isEmpty else { return }

        do {
            try db.run(todosTable.insert(text <- itemText))
            items.append(itemText)
            tableView.reloadData()
        } catch {
            print("Failed to insert item: \(error)")
        }
    }

    // MARK: - URL Handling

    func application(_ application: NSApplication, open urls: [URL]) {
        for url in urls {
            guard url.scheme == "tiger",
                  url.host == "add",
                  let components = URLComponents(url: url, resolvingAgainstBaseURL: false),
                  let queryItems = components.queryItems,
                  let urlParam = queryItems.first(where: { $0.name == "url" })?.value else {
                continue
            }

            addItemFromText(urlParam)

            // Bring window to front if it's hidden
            window.makeKeyAndOrderFront(nil)
            NSApp.activate(ignoringOtherApps: true)
        }
    }

    // MARK: - App Lifecycle

    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool {
        return true
    }

    func applicationShouldTerminate(_ sender: NSApplication) -> NSApplication.TerminateReply {
        return .terminateNow
    }

    // MARK: - NSTableViewDataSource

    func numberOfRows(in tableView: NSTableView) -> Int {
        return items.count
    }

    func tableView(_ tableView: NSTableView, objectValueFor tableColumn: NSTableColumn?, row: Int) -> Any? {
        return items[row]
    }
}

let app = NSApplication.shared
let delegate = AppDelegate()
app.delegate = delegate
app.run()
