import * as vscode from "vscode";
import { TkTask } from './types';
import { escapeMarkdown } from './utils';

export class GroupTreeItem extends vscode.TreeItem {
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
export class TaskTreeItem extends vscode.TreeItem {
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

export type TkTreeItem = GroupTreeItem | TaskTreeItem;
