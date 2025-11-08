import * as vscode from 'vscode';
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';

const execFileAsync = promisify(execFile);

// escapeMarkdown escapes special characters in markdown to prevent interpretation
function escapeMarkdown(text: string): string {
  return text
    .replace(/\\/g, '\\\\')
    .replace(/`/g, '\\`')
    .replace(/\*/g, '\\*')
    .replace(/_/g, '\\_')
    .replace(/\{/g, '\\{')
    .replace(/\}/g, '\\}')
    .replace(/\[/g, '\\[')
    .replace(/\]/g, '\\]')
    .replace(/\(/g, '\\(')
    .replace(/\)/g, '\\)')
    .replace(/#/g, '\\#')
    .replace(/\+/g, '\\+')
    .replace(/-/g, '\\-')
    .replace(/\./g, '\\.')
    .replace(/!/g, '\\!')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

interface AxisStatus {
  effective?: string;
}

interface TkNote {
  markdown?: string;
  actor?: string;
  timestamp?: string;
}

interface RelationEdge {
  dst?: string;
  note?: string;
}

interface Relations {
  blocks?: {
    out?: RelationEdge[];
    in?: RelationEdge[];
  };
  subtask?: {
    children?: string[];
    parent?: string;
  };
  related?: {
    out?: RelationEdge[];
  };
  duplicate_of?: {
    out?: RelationEdge[];
  };
  supersedes?: {
    out?: RelationEdge[];
  };
}

interface TkTask {
  uuid?: string;
  display_id?: string;
  project_uuid?: string;
  title?: string;
  axes?: Record<string, AxisStatus | undefined>;
  blocked?: boolean;
  blockers?: Array<{ display_id?: string; title?: string }>;
  notes?: TkNote[];
  relations?: Relations;
}

interface TkGroup {
  group?: string;
  tasks: TkTask[];
}

interface TkProject {
  uid: string;
  name: string;
  local_preferred_alias?: string;
  aliases: string[];
  description?: string;
  type?: string;
}

type TkTreeItem = GroupTreeItem | TaskTreeItem;

class GroupTreeItem extends vscode.TreeItem {
  constructor(
    public readonly groupName: string,
    public readonly children: TaskTreeItem[],
    uniqueId?: string,
  ) {
    super(groupName, vscode.TreeItemCollapsibleState.Collapsed);
    this.iconPath = new vscode.ThemeIcon('folder');
    this.contextValue = 'tkGroup';
    if (uniqueId) {
      this.id = uniqueId;
    }
  }
}

class TaskTreeItem extends vscode.TreeItem {
  public readonly statusColor?: vscode.ThemeColor;
  public children?: TaskTreeItem[];

  constructor(public readonly task: TkTask, children?: TaskTreeItem[]) {
    // Extract just the number from display_id (e.g., "foo-123" -> "#123")
    let label = task.display_id ?? task.title ?? 'unnamed task';
    if (task.display_id) {
      const match = task.display_id.match(/-(\d+)(?:-\w+)?$/);
      if (match) {
        label = `#${match[1]}`;
      }
    }
    
    // Make collapsible if it has children
    const collapsibleState = children && children.length > 0 
      ? vscode.TreeItemCollapsibleState.Collapsed
      : vscode.TreeItemCollapsibleState.None;
    
    super(label, collapsibleState);
    this.children = children;

    const genericAxis = task.axes?.['generic'];
    const state = genericAxis?.effective ?? 'unknown';

    // Show only the title, without status in brackets
    if (task.title && task.title !== label) {
      this.description = task.title;
    }
    const blocked = task.blocked ? 'yes' : 'no';

    const tooltip = new vscode.MarkdownString();
    tooltip.appendMarkdown(`**${escapeMarkdown(label)}**\n\n`);
    if (task.title) {
      tooltip.appendMarkdown(`${escapeMarkdown(task.title)}\n\n`);
    }
    tooltip.appendMarkdown(`Status: ${state}\n`);
    tooltip.appendMarkdown(`Blocked: ${blocked}`);
    if (task.blockers && task.blockers.length > 0) {
      const blockersList = task.blockers
        .map((blocker) => `${blocker.display_id ?? ''} ${blocker.title ?? ''}`.trim())
        .filter((entry) => entry.length > 0)
        .join('\n');
      if (blockersList.length > 0) {
        tooltip.appendMarkdown(`\nBlockers:\n${blockersList}`);
      }
    }
    tooltip.isTrusted = false;

    this.tooltip = tooltip;
    this.contextValue = 'tkTask';

    // Set icon based on status and blocked state with colors
    if (task.blocked) {
      this.statusColor = new vscode.ThemeColor('errorForeground');
      this.iconPath = new vscode.ThemeIcon('circle-slash', this.statusColor);
    } else {
      // Icon based on status with colors
      switch (state) {
        case 'next':
          this.statusColor = new vscode.ThemeColor('charts.blue');
          this.iconPath = new vscode.ThemeIcon('arrow-right', this.statusColor);
          break;
        case 'wip':
          this.statusColor = new vscode.ThemeColor('charts.yellow');
          this.iconPath = new vscode.ThemeIcon('sync', this.statusColor);
          break;
        case 'done':
          this.statusColor = new vscode.ThemeColor('charts.green');
          this.iconPath = new vscode.ThemeIcon('check', this.statusColor);
          break;
        default:
          this.iconPath = new vscode.ThemeIcon('circle-outline');
          break;
      }
    }

    // Create a unique URI for this task to enable file decorations
    if (this.statusColor) {
      this.resourceUri = vscode.Uri.parse(`tk:${task.uuid ?? task.display_id ?? label}`);
    }

    // Set command to show task details when clicked
    this.command = {
      command: 'tk.showTaskDetails',
      title: 'Show Task Details',
      arguments: [task]
    };
  }
}

// Detail view using WebView
class TaskDetailProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'tkDetailView';

  private _view?: vscode.WebviewView;
  private _extensionUri: vscode.Uri;
  private currentTask: TkTask | undefined;
  private allTasks: TkTask[] = [];

  constructor(extensionUri: vscode.Uri) {
    this._extensionUri = extensionUri;
  }

  resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken,
  ): void {
    this._view = webviewView;

    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [
        vscode.Uri.joinPath(this._extensionUri, 'out'),
        vscode.Uri.joinPath(this._extensionUri, 'src', 'webview')
      ],
    };

    // Handle messages from the webview
    webviewView.webview.onDidReceiveMessage(async (message) => {
      if (message.type === 'editTitle' && this.currentTask) {
        await this.handleTitleEdit(message.newTitle);
      } else if (message.type === 'addNote' && this.currentTask) {
        await this.handleAddNote(message.markdown);
      }
    });

    // Set the HTML content
    webviewView.webview.html = this.getHtmlForWebview(webviewView.webview);

    // Send initial task data
    if (this.currentTask) {
      this.updateView();
    }
  }

  private async handleTitleEdit(newTitle: string): Promise<void> {
    if (!this.currentTask) {
      return;
    }

    const taskId = this.currentTask.display_id;
    if (!taskId) {
      void vscode.window.showErrorMessage('Cannot edit title: task has no ID');
      return;
    }

    if (newTitle.trim() === '') {
      void vscode.window.showErrorMessage('Title cannot be empty');
      return;
    }

    try {
      const { binary, cwd } = getTkConfig();
      const args = ['describe', taskId, newTitle];

      await execFileAsync(binary, args, {
        cwd,
        env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
      });

      // Update the current task's title
      this.currentTask.title = newTitle;
      this.updateView();

      // Trigger a refresh of the tree view
      void vscode.commands.executeCommand('tk.refresh');
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      void vscode.window.showErrorMessage(`Failed to update title: ${message}`);
    }
  }

  private async handleAddNote(markdown: string): Promise<void> {
    if (!this.currentTask) {
      return;
    }

    const taskId = this.currentTask.display_id;
    if (!taskId) {
      void vscode.window.showErrorMessage('Cannot add note: task has no ID');
      return;
    }

    if (markdown.trim() === '') {
      void vscode.window.showErrorMessage('Note cannot be empty');
      return;
    }

    try {
      const { binary, cwd } = getTkConfig();
      // Use '--' to force Cobra to treat markdown as positional argument
      // This prevents errors when markdown starts with '-' or '--' (e.g., bullet lists)
      const args = ['note', taskId, '--', markdown];

      await execFileAsync(binary, args, {
        cwd,
        env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
      });

      // Refresh the task data to get the new note
      await this.refreshCurrentTask();

      // Trigger a refresh of the tree view
      void vscode.commands.executeCommand('tk.refresh');
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      void vscode.window.showErrorMessage(`Failed to add note: ${message}`);
    }
  }

  private async refreshCurrentTask(): Promise<void> {
    if (!this.currentTask || !this.currentTask.display_id) {
      return;
    }

    try {
      const { binary, cwd } = getTkConfig();
      const args = ['show', '--json', this.currentTask.display_id];

      const result = await execFileAsync(binary, args, {
        cwd,
        env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
      });

      const updatedTask = JSON.parse(result.stdout) as TkTask;
      this.currentTask = updatedTask;
      this.updateView();
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      void vscode.window.showErrorMessage(`Failed to refresh task: ${message}`);
    }
  }

  showTask(task: TkTask): void {
    this.currentTask = task;
    this.updateView();
  }

  setAllTasks(tasks: TkTask[]): void {
    this.allTasks = tasks;
    this.updateView();
  }

  clear(): void {
    this.currentTask = undefined;
    this.showEmptyState();
  }

  private showEmptyState(): void {
    if (this._view) {
      void this._view.webview.postMessage({ type: 'clear' });
    }
  }

  private updateView(): void {
    if (this._view && this.currentTask) {
      void this._view.webview.postMessage({
        type: 'updateTask',
        task: this.currentTask,
        allTasks: this.allTasks
      });
    }
  }

  private getHtmlForWebview(webview: vscode.Webview): string {
    const scriptUri = webview.asWebviewUri(
      vscode.Uri.joinPath(this._extensionUri, 'out', 'webview.js')
    );
    const styleUri = webview.asWebviewUri(
      vscode.Uri.joinPath(this._extensionUri, 'src', 'webview', 'styles.css')
    );

    const nonce = getNonce();

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${webview.cspSource}; script-src 'nonce-${nonce}';">
    <link href="${styleUri}" rel="stylesheet">
</head>
<body>
    <script nonce="${nonce}" src="${scriptUri}"></script>
</body>
</html>`;
  }
}

function getNonce(): string {
  let text = '';
  const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  for (let i = 0; i < 32; i++) {
    text += possible.charAt(Math.floor(Math.random() * possible.length));
  }
  return text;
}

class TkDecorationProvider implements vscode.FileDecorationProvider {
  private readonly _onDidChangeFileDecorations = new vscode.EventEmitter<vscode.Uri | vscode.Uri[]>();
  readonly onDidChangeFileDecorations = this._onDidChangeFileDecorations.event;

  private taskItems = new Map<string, TaskTreeItem>();

  provideFileDecoration(uri: vscode.Uri): vscode.FileDecoration | undefined {
    if (uri.scheme !== 'tk') {
      return undefined;
    }

    const item = this.taskItems.get(uri.toString());
    if (item?.statusColor) {
      return {
        color: item.statusColor,
      };
    }

    return undefined;
  }

  updateTaskItems(items: TaskTreeItem[]): void {
    this.taskItems.clear();
    const uris: vscode.Uri[] = [];
    for (const item of items) {
      if (item.resourceUri) {
        this.taskItems.set(item.resourceUri.toString(), item);
        uris.push(item.resourceUri);
      }
    }
    // Fire with all URIs to notify VS Code that these decorations have changed
    if (uris.length > 0) {
      this._onDidChangeFileDecorations.fire(uris);
    }
  }
}

class TkDragAndDropController implements vscode.TreeDragAndDropController<TkTreeItem> {
  dropMimeTypes = ['application/vnd.code.tree.tkexplorer'];
  dragMimeTypes = ['application/vnd.code.tree.tkexplorer'];

  constructor(private readonly provider: TkProvider) {}

  public handleDrag(source: readonly TkTreeItem[], dataTransfer: vscode.DataTransfer, _token: vscode.CancellationToken): void {
    // Only allow dragging tasks, not groups
    const tasks = source.filter((item): item is TaskTreeItem => item instanceof TaskTreeItem);
    if (tasks.length === 0) {
      return;
    }
    
    dataTransfer.set(
      'application/vnd.code.tree.tkexplorer',
      new vscode.DataTransferItem(tasks)
    );
  }

  public async handleDrop(target: TkTreeItem | undefined, dataTransfer: vscode.DataTransfer, _token: vscode.CancellationToken): Promise<void> {
    const transferItem = dataTransfer.get('application/vnd.code.tree.tkexplorer');
    if (!transferItem) {
      return;
    }

    const tasks = transferItem.value as TaskTreeItem[];
    if (!tasks || tasks.length === 0) {
      return;
    }

    // Determine the target group
    let targetGroup: GroupTreeItem | undefined;
    if (target instanceof GroupTreeItem) {
      targetGroup = target;
    } else if (target instanceof TaskTreeItem) {
      // If dropped on a task, find the group it belongs to
      targetGroup = this.provider.findGroupForTask(target);
    }

    if (!targetGroup) {
      void vscode.window.showErrorMessage('Cannot drop task: no target group found');
      return;
    }

    // Store target group for use in closures (avoids non-null assertions)
    const targetGroupRef = targetGroup;

    // Move each task to the target group, collecting results
    const results = await Promise.allSettled(
      tasks.map(task => moveTaskToGroup(task, targetGroupRef))
    );

    // Report results to user
    const failures = results.filter((r): r is PromiseRejectedResult => r.status === 'rejected');
    const successCount = tasks.length - failures.length;
    
    if (failures.length === 0) {
      // All succeeded - no notification
    } else if (successCount === 0) {
      // All failed
      const firstError = failures[0].reason;
      const message = firstError instanceof Error ? firstError.message : String(firstError);
      void vscode.window.showErrorMessage(`Failed to move task(s): ${message}`);
    } else {
      // Partial success
      void vscode.window.showWarningMessage(
        `Moved ${successCount} task(s) to ${targetGroupRef.groupName}, but ${failures.length} failed`
      );
    }

    // Refresh the tree view
    await this.provider.refresh();
  }
}

class TkProvider implements vscode.TreeDataProvider<TkTreeItem> {
  private readonly _onDidChangeTreeData = new vscode.EventEmitter<TkTreeItem | undefined>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  private rawGroups: TkGroup[] = [];
  private rawUngrouped: TkTask[] = [];
  private showDone: boolean = false; // Default: hide done tasks
  private groupItems: GroupTreeItem[] = [];
  private treeView?: vscode.TreeView<TkTreeItem>;
  private collapseCounter: number = 0;
  private displayedTaskItems: TaskTreeItem[] = [];
  private detailProvider?: TaskDetailProvider;

  constructor(private readonly decorationProvider: TkDecorationProvider) {
    void this.refresh();
  }

  public setDetailProvider(provider: TaskDetailProvider): void {
    this.detailProvider = provider;
  }

  public setTreeView(treeView: vscode.TreeView<TkTreeItem>): void {
    this.treeView = treeView;
  }

  public toggleDone(): void {
    this.showDone = !this.showDone;
    this._onDidChangeTreeData.fire(undefined);
  }

  public isShowingDone(): boolean {
    return this.showDone;
  }

  public async expandAll(): Promise<void> {
    if (!this.treeView) {
      return;
    }
    // Reveal each group with expand parameter
    for (const group of this.groupItems) {
      await this.treeView.reveal(group, { expand: true, select: false, focus: false });
    }
  }

  public async collapseAllGroups(): Promise<void> {
    // Increment counter to force new IDs for all groups, making VS Code forget expansion state
    this.collapseCounter++;
    this._onDidChangeTreeData.fire(undefined);
  }

  private filterTasksByStatus(tasks: TkTask[]): TkTask[] {
    if (this.showDone) {
      return tasks; // Show all tasks
    }
    return tasks.filter(task => {
      const genericAxis = task.axes?.['generic'];
      const status = genericAxis?.effective ?? '';
      return status !== 'done'; // Hide done tasks
    });
  }

  public findGroupForTask(task: TaskTreeItem): GroupTreeItem | undefined {
    // Find the group containing this task in the raw data
    const taskUuid = task.task.uuid;
    const taskDisplayId = task.task.display_id;
    
    for (const rawGroup of this.rawGroups) {
      const taskFound = rawGroup.tasks.some(t => 
        (taskUuid && t.uuid === taskUuid) || 
        (taskDisplayId && t.display_id === taskDisplayId)
      );
      if (taskFound) {
        // Return a lightweight GroupTreeItem for drag-and-drop operations
        // We only need the group name for the move operation
        return new GroupTreeItem(rawGroup.group ?? 'unnamed', []);
      }
    }
    return undefined;
  }

  getTreeItem(element: TkTreeItem): vscode.TreeItem {
    return element;
  }

  getParent(element: TkTreeItem): vscode.ProviderResult<TkTreeItem> {
    // Tasks have group parents, groups have no parent
    if (element instanceof TaskTreeItem) {
      // Find the group that contains this task
      return this.findGroupForTask(element);
    }
    return undefined;
  }

  getChildren(element?: TkTreeItem): vscode.ProviderResult<TkTreeItem[]> {
    if (!element) {
      // Apply filtering when getting root children
      // Filter groups first, then create TreeItems to avoid unnecessary object creation
      // Store task items for decoration provider
      const allDisplayedTasks: TaskTreeItem[] = [];
      
      // Build task map for subtask resolution
      const taskMap = new Map<string, TkTask>();
      for (const group of this.rawGroups) {
        for (const task of group.tasks) {
          if (task.uuid) {
            taskMap.set(task.uuid, task);
          }
        }
      }
      for (const task of this.rawUngrouped) {
        if (task.uuid) {
          taskMap.set(task.uuid, task);
        }
      }
      
      this.groupItems = this.rawGroups
        .map(group => {
          const filteredTasks = this.filterTasksByStatus(group.tasks);
          const groupName = group.group ?? 'unnamed';
          // Include collapse counter in ID to force VS Code to forget expansion state when collapsed
          const uniqueId = `${groupName}-${this.collapseCounter}`;
          const taskItems = filteredTasks.map((task) => this.buildTaskTree(task, taskMap));
          allDisplayedTasks.push(...taskItems);
          return new GroupTreeItem(
            groupName,
            taskItems,
            uniqueId,
          );
        });

      const filteredUngrouped = this.filterTasksByStatus(this.rawUngrouped);
      const ungrouped = filteredUngrouped.map((task) => this.buildTaskTree(task, taskMap));
      allDisplayedTasks.push(...ungrouped);

      // Update decoration provider with the actual displayed items
      this.displayedTaskItems = allDisplayedTasks;
      this.decorationProvider.updateTaskItems(this.displayedTaskItems);

      return [...this.groupItems, ...ungrouped];
    }

    if (element instanceof GroupTreeItem) {
      return element.children;
    }

    if (element instanceof TaskTreeItem) {
      return element.children ?? [];
    }

    return [];
  }

  private buildTaskTree(task: TkTask, taskMap: Map<string, TkTask>, renderedSubtasks: Set<string> = new Set()): TaskTreeItem {
    // Get subtasks for this task
    const subtaskUUIDs = task.relations?.subtask?.children ?? [];
    const subtaskItems: TaskTreeItem[] = [];
    
    for (const uuid of subtaskUUIDs) {
      // Skip if this subtask was already rendered under another parent (tk-vsc-66)
      if (renderedSubtasks.has(uuid)) {
        continue;
      }
      
      const subtask = taskMap.get(uuid);
      if (subtask) {
        renderedSubtasks.add(uuid);
        // Recursively build subtask tree (renderedSubtasks prevents cycles)
        subtaskItems.push(this.buildTaskTree(subtask, taskMap, renderedSubtasks));
      }
    }
    
    return new TaskTreeItem(task, subtaskItems.length > 0 ? subtaskItems : undefined);
  }

  async refresh(): Promise<void> {
    try {
      const tasks = await fetchTk();

      // Store raw unfiltered data
      this.rawGroups = tasks.groups;
      this.rawUngrouped = tasks.tasks;

      // Update detail provider with all tasks for UUID resolution
      if (this.detailProvider) {
        const allTasks: TkTask[] = [];
        for (const group of this.rawGroups) {
          allTasks.push(...group.tasks);
        }
        allTasks.push(...this.rawUngrouped);
        this.detailProvider.setAllTasks(allTasks);
      }

      // Notify tree view to refresh, which will call getChildren()
      // getChildren() will create the task items and update the decoration provider
      this._onDidChangeTreeData.fire(undefined);
    } catch (error) {
      this.rawGroups = [];
      this.rawUngrouped = [];
      const message = error instanceof Error ? error.message : String(error);
      void vscode.window.showErrorMessage(`tk Tasks: ${message}`);
      this._onDidChangeTreeData.fire(undefined);
    }
  }
}

interface TkConfig {
  binary: string;
  cwd: string;
}

function getTkConfig(): TkConfig {
  const configuration = vscode.workspace.getConfiguration('tk');
  const binary = configuration.get<string>('binaryPath', 'tk') || 'tk';
  const configuredCwd = configuration.get<string>('workingDirectory');

  let cwd: string;

  if (configuredCwd && configuredCwd.trim().length > 0) {
    cwd = configuredCwd;
  } else if (vscode.workspace.workspaceFolders?.[0]?.uri.fsPath) {
    cwd = vscode.workspace.workspaceFolders[0].uri.fsPath;
  } else {
    // Default to home directory if no workspace is open
    cwd = process.env.HOME || process.env.USERPROFILE || '.';
  }

  return { binary, cwd };
}

async function fetchProjects(): Promise<Map<string, TkProject>> {
  const { binary, cwd } = getTkConfig();
  const args = ['project', 'ls', '--json'];

  const result = await execFileAsync(binary, args, {
    cwd,
    env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
  });

  const projects = JSON.parse(result.stdout) as TkProject[];
  const projectMap = new Map<string, TkProject>();
  for (const project of projects) {
    projectMap.set(project.uid, project);
  }
  return projectMap;
}

async function fetchTk(): Promise<{ groups: TkGroup[]; tasks: TkTask[] }> {
  const { binary, cwd } = getTkConfig();

  // NOTE: tk ls --json always returns a flat array of tasks with project_uuid.
  // We fetch project metadata separately and group tasks on the client side.
  // See tk/cmd/display.go:outputTasksJSON for the task JSON format.
  // See tk/cmd/project/ls.go for the project JSON format.
  const args = ['ls', '--json'];

  let stdout: string;
  try {
    const result = await execFileAsync(binary, args, {
      cwd,
      env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
      maxBuffer: 1024 * 1024 * 10,
    });
    stdout = result.stdout;
  } catch (error) {
    if (error instanceof Error && 'stderr' in error) {
      const stderr = (error as { stderr?: string }).stderr;
      if (stderr) {
        throw new Error(stderr.trim());
      }
    }
    throw error;
  }

  // Parse the flat array of tasks from tk ls --json
  let allTasks: TkTask[];
  try {
    allTasks = JSON.parse(stdout) as TkTask[];
  } catch (error) {
    throw new Error(`Failed to parse JSON output from tk: ${error}`);
  }

  // Fetch project metadata
  const projectMap = await fetchProjects();

  // Group tasks by project_uuid
  const grouped = new Map<string, TkTask[]>();
  const ungrouped: TkTask[] = [];

  for (const task of allTasks) {
    if (task.project_uuid) {
      if (!grouped.has(task.project_uuid)) {
        grouped.set(task.project_uuid, []);
      }
      grouped.get(task.project_uuid)!.push(task);
    } else {
      // Tasks without a project_uuid go into ungrouped
      ungrouped.push(task);
    }
  }

  // Convert the Map to an array of groups with display names
  const groups: TkGroup[] = Array.from(grouped.entries()).map(([projectUUID, tasks]) => {
    const project = projectMap.get(projectUUID);
    // Use local_preferred_alias if available, otherwise fall back to name
    const groupName = project?.local_preferred_alias || project?.name || projectUUID;
    return {
      group: groupName,
      tasks: tasks,
    };
  });

  return { groups, tasks: ungrouped };
}

async function rotateStatus(provider: TkProvider, item: TaskTreeItem): Promise<void> {
  const taskId = item.task.display_id;
  if (!taskId) {
    void vscode.window.showErrorMessage('Cannot rotate status: task has no ID');
    return;
  }

  const genericAxis = item.task.axes?.['generic'];
  const currentStatus = genericAxis?.effective ?? '';

  // Determine next status: no status -> next -> wip -> done -> no status
  let nextStatus: string;
  switch (currentStatus) {
    case 'next':
      nextStatus = 'wip';
      break;
    case 'wip':
      nextStatus = 'done';
      break;
    case 'done':
      nextStatus = '';
      break;
    default:
      nextStatus = 'next';
      break;
  }

  try {
    const { binary, cwd } = getTkConfig();

    // Use tk mark command to update the status
    const args = nextStatus === ''
      ? ['mark', '--unset', taskId]
      : ['mark', taskId, nextStatus];

    await execFileAsync(binary, args, {
      cwd,
      env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
    });

    await provider.refresh();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    void vscode.window.showErrorMessage(`Failed to update status: ${message}`);
  }
}

async function markDone(provider: TkProvider, treeView: vscode.TreeView<TkTreeItem>, item?: TaskTreeItem): Promise<void> {
  // If no item is passed (e.g., from keybinding), get the selected item from the tree view
  if (!item) {
    const selection = treeView.selection;
    if (selection.length === 0) {
      void vscode.window.showErrorMessage('No task selected');
      return;
    }
    const selectedItem = selection[0];
    if (!(selectedItem instanceof TaskTreeItem)) {
      void vscode.window.showErrorMessage('Please select a task, not a group');
      return;
    }
    item = selectedItem;
  }

  const taskId = item.task.display_id;
  if (!taskId) {
    void vscode.window.showErrorMessage('Cannot mark task as done: task has no ID');
    return;
  }

  try {
    const { binary, cwd } = getTkConfig();

    // Toggle behavior: if already done, unset status; otherwise mark as done
    const genericAxis = item.task.axes?.['generic'];
    const currentStatus = genericAxis?.effective ?? '';
    
    const args = currentStatus === 'done' 
      ? ['mark', taskId, '--unset']  // Unset if already done (no state arg)
      : ['mark', taskId, 'done'];     // Mark as done otherwise

    await execFileAsync(binary, args, {
      cwd,
      env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
    });

    await provider.refresh();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    void vscode.window.showErrorMessage(`Failed to mark task as done: ${message}`);
  }
}

async function editTitle(provider: TkProvider, item: TaskTreeItem): Promise<void> {
  const taskId = item.task.display_id;
  if (!taskId) {
    void vscode.window.showErrorMessage('Cannot edit title: task has no ID');
    return;
  }

  const currentTitle = item.task.title ?? '';
  const newTitle = await vscode.window.showInputBox({
    prompt: `Edit title for ${taskId}`,
    value: currentTitle,
    placeHolder: 'Enter new title',
  });

  if (newTitle === undefined) {
    return;
  }

  if (newTitle.trim() === '') {
    void vscode.window.showErrorMessage('Title cannot be empty');
    return;
  }

  try {
    const { binary, cwd } = getTkConfig();

    const args = ['describe', taskId, newTitle];

    await execFileAsync(binary, args, {
      cwd,
      env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
    });

    await provider.refresh();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    void vscode.window.showErrorMessage(`Failed to update title: ${message}`);
  }
}

async function createTask(provider: TkProvider, item: GroupTreeItem): Promise<void> {
  const groupName = item.groupName;

  const taskTitle = await vscode.window.showInputBox({
    prompt: `Create new task in group "${groupName}"`,
    placeHolder: 'Enter task title',
  });

  if (taskTitle === undefined) {
    return;
  }

  if (taskTitle.trim() === '') {
    void vscode.window.showErrorMessage('Task title cannot be empty');
    return;
  }

  try {
    const { binary, cwd } = getTkConfig();

    // Create task with project matching the group name
    const args = ['new', '-p', groupName, taskTitle];

    await execFileAsync(binary, args, {
      cwd,
      env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
    });

    await provider.refresh();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    void vscode.window.showErrorMessage(`Failed to create task: ${message}`);
  }
}

async function createProject(provider: TkProvider): Promise<void> {
  const projectName = await vscode.window.showInputBox({
    prompt: 'Create new project',
    placeHolder: 'Enter project name',
  });

  if (projectName === undefined) {
    return;
  }

  if (projectName.trim() === '') {
    void vscode.window.showErrorMessage('Project name cannot be empty');
    return;
  }

  const projectDescription = await vscode.window.showInputBox({
    prompt: `Description for project "${projectName}"`,
    placeHolder: 'Enter project description (optional)',
  });

  try {
    const { binary, cwd } = getTkConfig();

    // Create project using tk project create command
    const args = ['project', 'create', projectName];
    if (projectDescription && projectDescription.trim() !== '') {
      args.push(projectDescription);
    }

    await execFileAsync(binary, args, {
      cwd,
      env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
    });

    await provider.refresh();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    void vscode.window.showErrorMessage(`Failed to create project: ${message}`);
  }
}

async function deleteTask(provider: TkProvider, item: TaskTreeItem): Promise<void> {
  const taskId = item.task.display_id;
  if (!taskId) {
    void vscode.window.showErrorMessage('Cannot delete task: task has no ID');
    return;
  }

  const taskTitle = item.task.title ?? taskId;
  const confirmation = await vscode.window.showWarningMessage(
    `Delete task ${taskId}${taskTitle !== taskId ? ` (${taskTitle})` : ''}?`,
    { modal: true },
    'Delete'
  );

  if (confirmation !== 'Delete') {
    return;
  }

  try {
    const { binary, cwd } = getTkConfig();

    const args = ['rm', taskId];

    await execFileAsync(binary, args, {
      cwd,
      env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
    });

    await provider.refresh();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    void vscode.window.showErrorMessage(`Failed to delete task: ${message}`);
  }
}

async function deleteProject(provider: TkProvider, item: GroupTreeItem): Promise<void> {
  const projectName = item.groupName;

  const confirmation = await vscode.window.showWarningMessage(
    `Delete project ${projectName}? This will also delete all tasks in this project.`,
    { modal: true },
    'Delete'
  );

  if (confirmation !== 'Delete') {
    return;
  }

  try {
    const { binary, cwd } = getTkConfig();

    const args = ['project', 'rm', projectName];

    await execFileAsync(binary, args, {
      cwd,
      env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
    });

    await provider.refresh();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    void vscode.window.showErrorMessage(`Failed to delete project: ${message}`);
  }
}

async function moveTaskToGroup(task: TaskTreeItem, targetGroup: GroupTreeItem): Promise<void> {
  const taskId = task.task.display_id;
  if (!taskId) {
    throw new Error('Cannot move task: task has no ID');
  }

  const targetGroupName = targetGroup.groupName;
  const { binary, cwd } = getTkConfig();

  // Use tk mv command to move the task to the target project/group
  // Use --auto to automatically assign a new number in the target project if needed
  const args = ['mv', '--auto', taskId, targetGroupName];

  await execFileAsync(binary, args, {
    cwd,
    env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
  });
}

async function updateToggleDoneButton(provider: TkProvider): Promise<void> {
  const showing = provider.isShowingDone();
  await vscode.commands.executeCommand('setContext', 'tk.showingDone', showing);
}

export function activate(context: vscode.ExtensionContext): void {
  const decorationProvider = new TkDecorationProvider();
  const provider = new TkProvider(decorationProvider);
  const dragAndDropController = new TkDragAndDropController(provider);
  const detailProvider = new TaskDetailProvider(context.extensionUri);
  
  // Link providers so tasks can update detail view
  provider.setDetailProvider(detailProvider);

  const treeView = vscode.window.createTreeView('tkExplorer', {
    treeDataProvider: provider,
    dragAndDropController: dragAndDropController,
  });

  // Give provider access to tree view for expand/collapse operations
  provider.setTreeView(treeView);

  // Set initial context
  void updateToggleDoneButton(provider);

  context.subscriptions.push(
    vscode.window.registerFileDecorationProvider(decorationProvider),
    treeView,
    vscode.window.registerWebviewViewProvider(TaskDetailProvider.viewType, detailProvider),
    vscode.commands.registerCommand('tk.refresh', () => provider.refresh()),
    vscode.commands.registerCommand('tk.editTitle', (item: TaskTreeItem) => editTitle(provider, item)),
    vscode.commands.registerCommand('tk.rotateStatus', (item: TaskTreeItem) => rotateStatus(provider, item)),
    vscode.commands.registerCommand('tk.markDone', (item?: TaskTreeItem) => markDone(provider, treeView, item)),
    vscode.commands.registerCommand('tk.createTask', (item: GroupTreeItem) => createTask(provider, item)),
    vscode.commands.registerCommand('tk.createProject', () => createProject(provider)),
    vscode.commands.registerCommand('tk.deleteTask', (item: TaskTreeItem) => deleteTask(provider, item)),
    vscode.commands.registerCommand('tk.deleteProject', (item: GroupTreeItem) => deleteProject(provider, item)),
    vscode.commands.registerCommand('tk.toggleDone', () => {
      provider.toggleDone();
      void updateToggleDoneButton(provider);
    }),
    vscode.commands.registerCommand('tk.toggleDoneHide', () => {
      provider.toggleDone();
      void updateToggleDoneButton(provider);
    }),
    vscode.commands.registerCommand('tk.showTaskDetails', (task: TkTask) => {
      detailProvider.showTask(task);
    }),
    vscode.commands.registerCommand('tk.expandAll', async () => {
      await provider.expandAll();
    }),
    vscode.commands.registerCommand('tk.collapseAll', async () => {
      await provider.collapseAllGroups();
    }),
  );
}

export function deactivate(): void {
  // Nothing to clean up.
}
