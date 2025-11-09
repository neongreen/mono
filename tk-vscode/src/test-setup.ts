// Mock VS Code API for testing
import { vi } from 'vitest';

// Create EventEmitter class
class EventEmitter<T = any> {
  fire = vi.fn<[T], void>();
  event = vi.fn();
}

// Create ThemeColor class
class ThemeColor {
  constructor(public id: string) {}
}

// Create ThemeIcon class
class ThemeIcon {
  constructor(public id: string, public color?: ThemeColor) {}
}

// Uri mock
const Uri = {
  parse: (str: string) => ({
    scheme: str.split(':')[0] || 'tk',
    toString: () => str,
    path: str,
    fsPath: str,
  }),
};

// TreeItemCollapsibleState enum
const TreeItemCollapsibleState = {
  None: 0,
  Collapsed: 1,
  Expanded: 2,
} as const;

// TreeItem class
class TreeItem {
  resourceUri?: any;
  iconPath?: any;
  description?: string;
  tooltip?: any;
  contextValue?: string;
  command?: any;
  id?: string;

  constructor(public label: string, public collapsibleState?: number) {}
}

// MarkdownString class
class MarkdownString {
  value = '';
  isTrusted?: boolean;
  appendMarkdown(value: string) {
    this.value += value;
    return this;
  }
}

// FileDecoration class
class FileDecoration {
  constructor(
    public badge?: string,
    public tooltip?: string,
    public color?: ThemeColor,
  ) {}
}

// window namespace
const window = {
  showErrorMessage: vi.fn(),
  showInputBox: vi.fn(),
};

// commands namespace
const commands = {
  executeCommand: vi.fn(),
};

// workspace namespace
const workspace = {
  getConfiguration: vi.fn(() => ({
    get: vi.fn((key: string, defaultValue: any) => defaultValue),
  })),
};

// Export all as named exports (matching VS Code API)
export {
  EventEmitter,
  ThemeColor,
  ThemeIcon,
  Uri,
  TreeItemCollapsibleState,
  TreeItem,
  MarkdownString,
  FileDecoration,
  window,
  commands,
  workspace,
};
