import * as vscode from 'vscode';
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';

const execFileAsync = promisify(execFile);

interface AxisStatus {
  effective?: string;
}

interface TkTask {
  task_uuid?: string;
  task_id?: string;
  title?: string;
  axes?: Record<string, AxisStatus | undefined>;
  blocked?: boolean;
  blockers?: Array<{ task_id?: string; title?: string }>;
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
  }
}

class TaskTreeItem extends vscode.TreeItem {
  constructor(public readonly task: TkTask) {
    const label = task.task_id ?? task.title ?? 'unnamed task';
    super(label, vscode.TreeItemCollapsibleState.None);

    const genericAxis = task.axes?.['generic'];
    const state = genericAxis?.effective ?? 'unknown';

    if (task.title && task.title !== label) {
      this.description = `${task.title} [${state}]`;
    } else {
      this.description = `[${state}]`;
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

    // Set icon based on status and blocked state
    if (task.blocked) {
      this.iconPath = new vscode.ThemeIcon('circle-slash');
    } else {
      // Icon based on status
      switch (state) {
        case 'next':
          this.iconPath = new vscode.ThemeIcon('arrow-right');
          break;
        case 'wip':
          this.iconPath = new vscode.ThemeIcon('sync');
          break;
        case 'done':
          this.iconPath = new vscode.ThemeIcon('check');
          break;
        default:
          this.iconPath = new vscode.ThemeIcon('circle-outline');
          break;
      }
    }
  }
}

class TkProvider implements vscode.TreeDataProvider<TkTreeItem> {
  private readonly _onDidChangeTreeData = new vscode.EventEmitter<TkTreeItem | undefined>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  private groups: GroupTreeItem[] = [];
  private ungrouped: TaskTreeItem[] = [];

  constructor() {
    void this.refresh();
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

async function fetchTk(): Promise<{ groups: TkGroup[]; tasks: TkTask[] }> {
  const configuration = vscode.workspace.getConfiguration('tk');
  const binary = configuration.get<string>('binaryPath', 'tk') || 'tk';
  const group = configuration.get<string>('group', 'prefix') || 'prefix';
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
      void vscode.window.showErrorMessage('No workspace folder is open.');
      return;
    }

    // Use tk mark command to update the status
    const args = nextStatus === ''
      ? ['mark', taskId, '--unset']
      : ['mark', taskId, nextStatus];

    await execFileAsync(binary, args, {
      cwd,
      env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
    });

    const statusText = nextStatus === '' ? 'no status' : nextStatus;
    void vscode.window.showInformationMessage(`Updated status for ${taskId} to ${statusText}`);
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
      void vscode.window.showErrorMessage('No workspace folder is open.');
      return;
    }

    const args = ['describe', taskId, newTitle];

    await execFileAsync(binary, args, {
      cwd,
      env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0' },
    });

    void vscode.window.showInformationMessage(`Updated title for ${taskId}`);
    await provider.refresh();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    void vscode.window.showErrorMessage(`Failed to update title: ${message}`);
  }
}

export function activate(context: vscode.ExtensionContext): void {
  const provider = new TkProvider();

  context.subscriptions.push(
    vscode.window.registerTreeDataProvider('tkExplorer', provider),
    vscode.commands.registerCommand('tk.refresh', () => provider.refresh()),
    vscode.commands.registerCommand('tk.editTitle', (item: TaskTreeItem) => editTitle(provider, item)),
    vscode.commands.registerCommand('tk.rotateStatus', (item: TaskTreeItem) => rotateStatus(provider, item)),
  );
}

export function deactivate(): void {
  // Nothing to clean up.
}
