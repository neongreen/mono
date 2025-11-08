import * as vscode from "vscode";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { TkTask } from './types';
import { getTkConfig } from './tkApi';
import { getNonce } from './utils';

const execFileAsync = promisify(execFile);
export class TaskDetailProvider implements vscode.WebviewViewProvider {
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
