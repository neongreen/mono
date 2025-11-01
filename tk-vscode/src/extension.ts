import * as vscode from 'vscode';
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';

const execFileAsync = promisify(execFile);

interface AxisStatus {
  effective?: string;
}

interface TkNote {
  markdown?: string;
  actor?: string;
  timestamp?: string;
}

interface TkTask {
  task_uuid?: string;
  task_id?: string;
  title?: string;
  axes?: Record<string, AxisStatus | undefined>;
  blocked?: boolean;
  blockers?: Array<{ task_id?: string; title?: string }>;
  notes?: TkNote[];
}

interface TkGroup {
  group?: string;
  tasks: TkTask[];
}

interface TkJsonOutput {
  groups?: TkGroup[];
  tasks?: TkTask[];
}

type TkTreeItem = GroupTreeItem | TaskTreeItem;

class GroupTreeItem extends vscode.TreeItem {
  constructor(
    public readonly groupName: string,
    public readonly children: TaskTreeItem[],
  ) {
    super(groupName, vscode.TreeItemCollapsibleState.Collapsed);
    this.iconPath = new vscode.ThemeIcon('folder');
    this.contextValue = 'tkGroup';
  }
}

class TaskTreeItem extends vscode.TreeItem {
  public readonly statusColor?: vscode.ThemeColor;

  constructor(public readonly task: TkTask) {
    const label = task.task_id ?? task.title ?? 'unnamed task';
    super(label, vscode.TreeItemCollapsibleState.None);

    const genericAxis = task.axes?.['generic'];
    const state = genericAxis?.effective ?? 'unknown';

    // Show only the title, without status in brackets
    if (task.title && task.title !== label) {
      this.description = task.title;
    }
    const blocked = task.blocked ? 'yes' : 'no';

    const tooltip = new vscode.MarkdownString();
    tooltip.appendMarkdown(`**${label}**\n\n`);
    if (task.title) {
      tooltip.appendMarkdown(`${task.title}\n\n`);
    }
    tooltip.appendMarkdown(`Status: ${state}\n`);
    tooltip.appendMarkdown(`Blocked: ${blocked}`);
    if (task.blockers && task.blockers.length > 0) {
      const blockersList = task.blockers
        .map((blocker) => `${blocker.task_id ?? ''} ${blocker.title ?? ''}`.trim())
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
      this.resourceUri = vscode.Uri.parse(`tk:${task.task_uuid ?? task.task_id ?? label}`);
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
  private currentTask: TkTask | undefined;

  resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken,
  ): void {
    this._view = webviewView;

    webviewView.webview.options = {
      enableScripts: false,
    };

    if (this.currentTask) {
      this.updateView();
    } else {
      this.showEmptyState();
    }
  }

  showTask(task: TkTask): void {
    this.currentTask = task;
    this.updateView();
  }

  clear(): void {
    this.currentTask = undefined;
    this.showEmptyState();
  }

  private showEmptyState(): void {
    if (this._view) {
      this._view.webview.html = this.getHtmlForEmptyState();
    }
  }

  private updateView(): void {
    if (this._view && this.currentTask) {
      this._view.webview.html = this.getHtmlForTask(this.currentTask);
    }
  }

  private getHtmlForEmptyState(): string {
    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline';">
    <style>
        body {
            padding: 12px;
            font-family: var(--vscode-font-family);
            font-size: var(--vscode-font-size);
            color: var(--vscode-foreground);
        }
        .empty-state {
            color: var(--vscode-descriptionForeground);
            font-style: italic;
            text-align: center;
            padding: 20px;
        }
    </style>
</head>
<body>
    <div class="empty-state">No task selected</div>
</body>
</html>`;
  }

  private getHtmlForTask(task: TkTask): string {
    const taskId = task.task_id ?? 'unknown';
    const title = task.title ?? 'No title';
    const genericAxis = task.axes?.['generic'];
    const status = genericAxis?.effective ?? 'none';
    const blocked = task.blocked ? 'yes' : 'no';

    let notesHtml = '';
    if (task.notes && task.notes.length > 0) {
      notesHtml = task.notes.map(note => {
        const noteText = this.escapeHtml(note.markdown || '(empty note)');
        const actor = this.escapeHtml(note.actor || 'Unknown');
        const timestamp = note.timestamp ? new Date(note.timestamp).toLocaleString() : '';
        
        return `
          <div class="note">
            <div class="note-content">${noteText}</div>
            <div class="note-meta">
              ${timestamp ? `<span class="note-time">${this.escapeHtml(timestamp)}</span>` : ''}
              <span class="note-actor">by ${actor}</span>
            </div>
          </div>
        `;
      }).join('');
    } else {
      notesHtml = '<div class="empty-section">No notes</div>';
    }

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline';">
    <style>
        body {
            padding: 12px;
            font-family: var(--vscode-font-family);
            font-size: var(--vscode-font-size);
            color: var(--vscode-foreground);
            line-height: 1.5;
        }
        .section {
            margin-bottom: 20px;
        }
        .section-title {
            font-weight: 600;
            color: var(--vscode-foreground);
            margin-bottom: 8px;
            font-size: 11px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            opacity: 0.8;
        }
        .section-content {
            color: var(--vscode-foreground);
            white-space: pre-wrap;
            word-wrap: break-word;
        }
        .task-id {
            font-family: var(--vscode-editor-font-family);
            color: var(--vscode-textLink-foreground);
            font-weight: 500;
        }
        .note {
            background-color: var(--vscode-textBlockQuote-background);
            border-left: 3px solid var(--vscode-textBlockQuote-border);
            padding: 8px 12px;
            margin-bottom: 8px;
            border-radius: 3px;
        }
        .note-content {
            white-space: pre-wrap;
            word-wrap: break-word;
            margin-bottom: 6px;
        }
        .note-meta {
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
            display: flex;
            gap: 8px;
        }
        .note-time::after {
            content: "•";
            margin-left: 8px;
        }
        .empty-section {
            color: var(--vscode-descriptionForeground);
            font-style: italic;
        }
        .metadata-grid {
            display: grid;
            grid-template-columns: auto 1fr;
            gap: 8px 12px;
            margin-bottom: 16px;
        }
        .metadata-label {
            font-weight: 500;
            color: var(--vscode-descriptionForeground);
        }
        .metadata-value {
            color: var(--vscode-foreground);
        }
    </style>
</head>
<body>
    <div class="section">
        <div class="section-title">Task</div>
        <div class="section-content task-id">${this.escapeHtml(taskId)}</div>
    </div>
    
    <div class="section">
        <div class="section-title">Title</div>
        <div class="section-content">${this.escapeHtml(title)}</div>
    </div>

    <div class="section">
        <div class="metadata-grid">
            <div class="metadata-label">Status:</div>
            <div class="metadata-value">${this.escapeHtml(status)}</div>
            <div class="metadata-label">Blocked:</div>
            <div class="metadata-value">${this.escapeHtml(blocked)}</div>
        </div>
    </div>

    <div class="section">
        <div class="section-title">Notes</div>
        ${notesHtml}
    </div>
</body>
</html>`;
  }

  private escapeHtml(text: string): string {
    const map: Record<string, string> = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#039;',
      '`': '&#96;'
    };
    return text.replace(/[&<>"'`]/g, (m) => map[m]);
  }
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
      // All succeeded
      if (tasks.length === 1) {
        void vscode.window.showInformationMessage(`Moved ${tasks[0].task.task_id} to ${targetGroupRef.groupName}`);
      } else {
        void vscode.window.showInformationMessage(`Moved ${successCount} task(s) to ${targetGroupRef.groupName}`);
      }
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

  private groups: GroupTreeItem[] = [];
  private ungrouped: TaskTreeItem[] = [];

  constructor(private readonly decorationProvider: TkDecorationProvider) {
    void this.refresh();
  }

  public findGroupForTask(task: TaskTreeItem): GroupTreeItem | undefined {
    for (const group of this.groups) {
      if (group.children.includes(task)) {
        return group;
      }
    }
    return undefined;
  }

  getTreeItem(element: TkTreeItem): vscode.TreeItem {
    return element;
  }

  getChildren(element?: TkTreeItem): vscode.ProviderResult<TkTreeItem[]> {
    if (!element) {
      return [...this.groups, ...this.ungrouped];
    }

    if (element instanceof GroupTreeItem) {
      return element.children;
    }

    return [];
  }

  async refresh(): Promise<void> {
    try {
      const tasks = await fetchTk();

      this.groups = tasks.groups.map(
        (group) =>
          new GroupTreeItem(
            group.group ?? 'unnamed',
            group.tasks.map((task) => new TaskTreeItem(task)),
          ),
      );

      this.ungrouped = tasks.tasks.map((task) => new TaskTreeItem(task));

      // Collect all task items for decoration provider
      const allTaskItems: TaskTreeItem[] = [];
      for (const group of this.groups) {
        allTaskItems.push(...group.children);
      }
      allTaskItems.push(...this.ungrouped);
      this.decorationProvider.updateTaskItems(allTaskItems);

      this._onDidChangeTreeData.fire(undefined);
    } catch (error) {
      this.groups = [];
      this.ungrouped = [];
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

  let cwd: string | undefined;

  if (configuredCwd && configuredCwd.trim().length > 0) {
    cwd = configuredCwd;
  } else {
    cwd = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
  }

  if (!cwd) {
    throw new Error('No workspace folder is open.');
  }

  return { binary, cwd };
}

async function fetchTk(): Promise<{ groups: TkGroup[]; tasks: TkTask[] }> {
  const configuration = vscode.workspace.getConfiguration('tk');
  const group = configuration.get<string>('group', 'prefix') || 'prefix';
  const { binary, cwd } = getTkConfig();

  const args = ['ls', '--json', '--group', group];

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

  let parsed: TkJsonOutput;
  try {
    parsed = JSON.parse(stdout) as TkJsonOutput;
  } catch (error) {
    throw new Error(`Failed to parse JSON output from tk: ${error}`);
  }

  const groups = Array.isArray(parsed.groups) ? parsed.groups : [];
  const tasks = Array.isArray(parsed.tasks) ? parsed.tasks : [];

  return { groups, tasks };
}

async function rotateStatus(provider: TkProvider, item: TaskTreeItem): Promise<void> {
  const taskId = item.task.task_id;
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

async function editTitle(provider: TkProvider, item: TaskTreeItem): Promise<void> {
  const taskId = item.task.task_id;
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

    void vscode.window.showInformationMessage(`Created project: ${projectName}`);
    await provider.refresh();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    void vscode.window.showErrorMessage(`Failed to create project: ${message}`);
  }
}

async function moveTaskToGroup(task: TaskTreeItem, targetGroup: GroupTreeItem): Promise<void> {
  const taskId = task.task.task_id;
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

export function activate(context: vscode.ExtensionContext): void {
  const decorationProvider = new TkDecorationProvider();
  const provider = new TkProvider(decorationProvider);
  const dragAndDropController = new TkDragAndDropController(provider);
  const detailProvider = new TaskDetailProvider();

  const treeView = vscode.window.createTreeView('tkExplorer', {
    treeDataProvider: provider,
    dragAndDropController: dragAndDropController,
  });

  context.subscriptions.push(
    vscode.window.registerFileDecorationProvider(decorationProvider),
    treeView,
    vscode.window.registerWebviewViewProvider(TaskDetailProvider.viewType, detailProvider),
    vscode.commands.registerCommand('tk.refresh', () => provider.refresh()),
    vscode.commands.registerCommand('tk.editTitle', (item: TaskTreeItem) => editTitle(provider, item)),
    vscode.commands.registerCommand('tk.rotateStatus', (item: TaskTreeItem) => rotateStatus(provider, item)),
    vscode.commands.registerCommand('tk.createTask', (item: GroupTreeItem) => createTask(provider, item)),
    vscode.commands.registerCommand('tk.createProject', () => createProject(provider)),
    vscode.commands.registerCommand('tk.showTaskDetails', (task: TkTask) => {
      detailProvider.showTask(task);
    }),
  );
}

export function deactivate(): void {
  // Nothing to clean up.
}
