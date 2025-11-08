import * as vscode from "vscode";
import { TaskTreeItem } from './treeItems';

export class TkDecorationProvider implements vscode.FileDecorationProvider {
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
