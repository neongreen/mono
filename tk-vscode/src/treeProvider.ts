import * as vscode from "vscode";
import { TkGroup, TkTask } from './types';
import { GroupTreeItem, TaskTreeItem, TkTreeItem } from './treeItems';
import { TaskDetailProvider } from './detailProvider';
import { TkDecorationProvider } from './decorationProvider';
import { fetchTk } from './tkApi';

export class TkProvider implements vscode.TreeDataProvider<TkTreeItem> {
  private readonly _onDidChangeTreeData = new vscode.EventEmitter<TkTreeItem | undefined>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  private rawGroups: TkGroup[] = [];
  private rawUngrouped: TkTask[] = [];
  private showDone: boolean = false; // Default: hide done tasks
  private searchTerm: string = '';
  private groupItems: GroupTreeItem[] = [];
  private treeView?: vscode.TreeView<TkTreeItem>;
  private collapseCounter: number = 0;
  private displayedTaskItems: TaskTreeItem[] = [];
  private detailProvider?: TaskDetailProvider;
  private context?: vscode.ExtensionContext;

  constructor(private readonly decorationProvider: TkDecorationProvider, context?: vscode.ExtensionContext) {
    this.context = context;
    // Load persisted showDone state (tk-vsc-96)
    if (context) {
      this.showDone = context.globalState.get<boolean>('tk.showDone', false);
    }
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
    // Persist the state (tk-vsc-96)
    if (this.context) {
      void this.context.globalState.update('tk.showDone', this.showDone);
    }
    this._onDidChangeTreeData.fire(undefined);
  }

  public isShowingDone(): boolean {
    return this.showDone;
  }

  public async search(): Promise<void> {
    const term = await vscode.window.showInputBox({
      prompt: 'Search tasks by title or ID',
      value: this.searchTerm,
      placeHolder: 'Enter search term...',
    });

    if (term === undefined) {
      return; // User cancelled
    }

    this.searchTerm = term.trim();
    await this.updateSearchContext();
    this._onDidChangeTreeData.fire(undefined);

    // Auto-expand all groups when searching to show matches
    if (this.searchTerm !== '' && this.treeView) {
      await this.expandAll();
    }
  }

  public clearSearch(): void {
    this.searchTerm = '';
    void this.updateSearchContext();
    this._onDidChangeTreeData.fire(undefined);
  }

  public getSearchTerm(): string {
    return this.searchTerm;
  }

  private async updateSearchContext(): Promise<void> {
    await vscode.commands.executeCommand('setContext', 'tk.searching', this.searchTerm !== '');
    
    // Update tree view message to show search term
    if (this.treeView) {
      if (this.searchTerm !== '') {
        this.treeView.message = `Searching: "${this.searchTerm}"`;
      } else {
        this.treeView.message = undefined;
      }
    }
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

  private matchesSearch(task: TkTask): boolean {
    if (this.searchTerm === '') {
      return true; // No search term, show all
    }

    const searchLower = this.searchTerm.toLowerCase();
    
    // Match against display_id
    if (task.display_id?.toLowerCase().includes(searchLower)) {
      return true;
    }

    // Match against title
    if (task.title?.toLowerCase().includes(searchLower)) {
      return true;
    }

    return false;
  }

  private filterTasks(tasks: TkTask[]): TkTask[] {
    // First filter by status (done/not done)
    let filtered = this.filterTasksByStatus(tasks);
    
    // Then filter by search term
    if (this.searchTerm !== '') {
      filtered = filtered.filter(task => this.matchesSearch(task));
    }

    return filtered;
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

      // Build a set of all task UUIDs that are subtasks of any parent (tk-vsc-67)
      const allSubtaskUUIDs = new Set<string>();
      for (const task of taskMap.values()) {
        const subtaskUUIDs = task.relations?.subtask?.children ?? [];
        for (const uuid of subtaskUUIDs) {
          allSubtaskUUIDs.add(uuid);
        }
      }

      this.groupItems = this.rawGroups
        .map(group => {
          const filteredTasks = this.filterTasks(group.tasks);
          // Exclude tasks that are subtasks of any parent (tk-vsc-67)
          const topLevelTasks = filteredTasks.filter(task => !task.uuid || !allSubtaskUUIDs.has(task.uuid));
          const groupName = group.group ?? 'unnamed';
          // Include collapse counter in ID to force VS Code to forget expansion state when collapsed
          const uniqueId = `${groupName}-${this.collapseCounter}`;
          const taskItems = topLevelTasks.map((task) => this.buildTaskTree(task, taskMap));
          // Collect all tasks including subtasks recursively
          for (const item of taskItems) {
            this.collectAllTaskItems(item, allDisplayedTasks);
          }
          return new GroupTreeItem(
            groupName,
            taskItems,
            uniqueId,
          );
        });

      const filteredUngrouped = this.filterTasks(this.rawUngrouped);
      // Exclude tasks that are subtasks of any parent (tk-vsc-67)
      const topLevelUngrouped = filteredUngrouped.filter(task => !task.uuid || !allSubtaskUUIDs.has(task.uuid));
      const ungrouped = topLevelUngrouped.map((task) => this.buildTaskTree(task, taskMap));
      // Collect all tasks including subtasks recursively
      for (const item of ungrouped) {
        this.collectAllTaskItems(item, allDisplayedTasks);
      }

      // Update decoration provider with the actual displayed items (now including subtasks)
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

  private collectAllTaskItems(item: TaskTreeItem, collection: TaskTreeItem[]): void {
    // Add the current item to the collection
    collection.push(item);

    // Recursively add all children (subtasks)
    if (item.children) {
      for (const child of item.children) {
        this.collectAllTaskItems(child, collection);
      }
    }
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
