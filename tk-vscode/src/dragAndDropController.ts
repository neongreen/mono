import * as vscode from "vscode";
import { GroupTreeItem, TaskTreeItem, TkTreeItem } from './treeItems';
import { TkProvider } from './treeProvider';
import { moveTaskToGroup } from './commands';

export class TkDragAndDropController implements vscode.TreeDragAndDropController<TkTreeItem> {
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
