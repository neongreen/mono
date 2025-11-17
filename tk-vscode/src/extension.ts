import * as vscode from 'vscode';
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';
import { AxisStatus, TkTask } from "./types";
import { escapeMarkdown } from "./utils";
import { GroupTreeItem, TaskTreeItem } from "./treeItems";
import { TaskDetailProvider } from "./detailProvider";
import { TkDecorationProvider } from "./decorationProvider";
import { TkDragAndDropController } from "./dragAndDropController";
import { TkProvider } from "./treeProvider";
import { getTkConfig } from "./tkApi";
import { rotateStatus, markDone, editTitle, createTask, createProject, deleteTask, deleteProject, quickCreateTask, updateToggleDoneButton } from "./commands";

const execFileAsync = promisify(execFile);

// Update UI context variables based on configuration
async function updateUIContext(): Promise<void> {
  const configuration = vscode.workspace.getConfiguration('tk');
  const showDeleteButton = configuration.get<boolean>('ui.showDeleteButton', true);
  const showEditTitleButton = configuration.get<boolean>('ui.showEditTitleButton', true);

  await vscode.commands.executeCommand('setContext', 'tk.ui.showDeleteButton', showDeleteButton);
  await vscode.commands.executeCommand('setContext', 'tk.ui.showEditTitleButton', showEditTitleButton);
}

export function activate(context: vscode.ExtensionContext): void {
  const decorationProvider = new TkDecorationProvider();
  const provider = new TkProvider(decorationProvider, context);
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
  void updateUIContext();

  // Set up auto-refresh if enabled (tk-vsc-35)
  let autoRefreshTimer: NodeJS.Timeout | undefined;

  function setupAutoRefresh() {
    // Clear existing timer
    if (autoRefreshTimer) {
      clearInterval(autoRefreshTimer);
      autoRefreshTimer = undefined;
    }

    const configuration = vscode.workspace.getConfiguration('tk');
    const enabled = configuration.get<boolean>('autoRefresh.enabled', false);
    const intervalSeconds = configuration.get<number>('autoRefresh.intervalSeconds', 5);

    if (enabled && intervalSeconds > 0) {
      autoRefreshTimer = setInterval(() => {
        void provider.refresh();
      }, intervalSeconds * 1000);
    }
  }

  setupAutoRefresh();

  // Listen for configuration changes
  context.subscriptions.push(
    vscode.workspace.onDidChangeConfiguration(e => {
      if (e.affectsConfiguration('tk.ui')) {
        void updateUIContext();
      }
      if (e.affectsConfiguration('tk.autoRefresh')) {
        setupAutoRefresh();
      }
    })
  );

  // Clean up timer on deactivation
  context.subscriptions.push(
    new vscode.Disposable(() => {
      if (autoRefreshTimer) {
        clearInterval(autoRefreshTimer);
      }
    })
  );

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
    vscode.commands.registerCommand('tk.search', async () => {
      await provider.search();
    }),
    vscode.commands.registerCommand('tk.clearSearch', () => {
      provider.clearSearch();
    }),
    vscode.commands.registerCommand('tk.quickCreateTask', async () => {
      await quickCreateTask(provider, treeView);
    }),
  );
}

export function deactivate(): void {
  // Auto-refresh timer is cleaned up automatically via context.subscriptions
}
