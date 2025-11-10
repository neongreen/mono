import * as vscode from "vscode";
import { TaskTreeItem } from './treeItems';

export class TkDecorationProvider implements vscode.FileDecorationProvider {
  private readonly _onDidChangeFileDecorations = new vscode.EventEmitter<vscode.Uri | vscode.Uri[]>();
  readonly onDidChangeFileDecorations = this._onDidChangeFileDecorations.event;

  private taskItems = new Map<string, TaskTreeItem>();
  private previousUris = new Set<string>();

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
    // Collect all URIs that need to be refreshed (tk-vsc-51)
    // Include both current URIs and previous URIs to handle status changes
    const allUris = new Set<string>();

    // Add previous URIs (for tasks that lost their status/decoration)
    for (const uri of this.previousUris) {
      allUris.add(uri);
    }

    this.taskItems.clear();
    this.previousUris.clear();

    for (const item of items) {
      if (item.resourceUri) {
        const uriString = item.resourceUri.toString();
        this.taskItems.set(uriString, item);
        this.previousUris.add(uriString);
        allUris.add(uriString);
      }
    }

    // Fire with all URIs (both current and previous) to ensure decorations update properly
    if (allUris.size > 0) {
      const uris = Array.from(allUris).map(uriString => vscode.Uri.parse(uriString));
      this._onDidChangeFileDecorations.fire(uris);
    }
  }
}
