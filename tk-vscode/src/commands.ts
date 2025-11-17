import * as vscode from "vscode";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { GroupTreeItem, TaskTreeItem, TkTreeItem } from './treeItems';
import { TkProvider } from './treeProvider';
import { getTkConfig } from './tkApi';

const execFileAsync = promisify(execFile);
export async function rotateStatus(provider: TkProvider, item: TaskTreeItem): Promise<void> {
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
export async function markDone(provider: TkProvider, treeView: vscode.TreeView<TkTreeItem>, item?: TaskTreeItem): Promise<void> {
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
export async function editTitle(provider: TkProvider, item: TaskTreeItem): Promise<void> {
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
export async function createTask(provider: TkProvider, item: GroupTreeItem): Promise<void> {
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
export async function quickCreateTask(provider: TkProvider, treeView: vscode.TreeView<TkTreeItem>): Promise<void> {
  // Prompt for task title with inline project support
  const input = await vscode.window.showInputBox({
    prompt: 'Create new task',
    placeHolder: 'Task title (or "project: title" to specify project)',
  });

  if (input === undefined) {
    return; // User cancelled
  }

  if (input.trim() === '') {
    void vscode.window.showErrorMessage('Task title cannot be empty');
    return;
  }

  // Parse input for "project: title" format
  const colonIndex = input.indexOf(':');
  let projectName: string | undefined;
  let taskTitle: string;

  if (colonIndex > 0 && colonIndex < input.length - 1) {
    // Has format "project: title"
    const potentialProject = input.substring(0, colonIndex).trim();
    const potentialTitle = input.substring(colonIndex + 1).trim();
    
    // Only treat as project prefix if there's actual content after the colon
    if (potentialTitle.length > 0) {
      projectName = potentialProject;
      taskTitle = potentialTitle;
    } else {
      // Just a colon with nothing after, treat whole thing as title
      taskTitle = input.trim();
    }
  } else {
    // No colon or colon at start/end, use whole input as title
    taskTitle = input.trim();
  }

  try {
    const { binary, cwd } = getTkConfig();

    // Create task with or without project
    const args = projectName 
      ? ['new', '-p', projectName, taskTitle]
      : ['new', taskTitle];

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
export async function createProject(provider: TkProvider): Promise<void> {
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
export async function deleteTask(provider: TkProvider, item: TaskTreeItem): Promise<void> {
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
export async function deleteProject(provider: TkProvider, item: GroupTreeItem): Promise<void> {
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
export async function moveTaskToGroup(task: TaskTreeItem, targetGroup: GroupTreeItem): Promise<void> {
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
export async function updateToggleDoneButton(provider: TkProvider): Promise<void> {
  const showing = provider.isShowingDone();
  await vscode.commands.executeCommand('setContext', 'tk.showingDone', showing);
}
