import * as vscode from "vscode";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { TkConfig, TkProject, TkTask, TkGroup } from './types';

const execFileAsync = promisify(execFile);
export function getTkConfig(): TkConfig {
  const configuration = vscode.workspace.getConfiguration('tk');
  const binary = configuration.get<string>('binaryPath', 'tk') || 'tk';
  const configuredCwd = configuration.get<string>('workingDirectory');

  let cwd: string;

  if (configuredCwd && configuredCwd.trim().length > 0) {
    cwd = configuredCwd;
  } else if (vscode.workspace.workspaceFolders?.[0]?.uri.fsPath) {
    cwd = vscode.workspace.workspaceFolders[0].uri.fsPath;
  } else {
    // Default to home directory if no workspace is open
    cwd = process.env.HOME || process.env.USERPROFILE || '.';
  }

  return { binary, cwd };
}
export async function fetchProjects(): Promise<Map<string, TkProject>> {
  const { binary, cwd } = getTkConfig();
  const args = ['project', 'ls', '--json'];

  const result = await execFileAsync(binary, args, {
    cwd,
    env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0', TK_SKIP_INVLOG: '1' },
  });

  const projects = JSON.parse(result.stdout) as TkProject[];
  const projectMap = new Map<string, TkProject>();
  for (const project of projects) {
    projectMap.set(project.uid, project);
  }
  return projectMap;
}
export async function fetchTk(): Promise<{ groups: TkGroup[]; tasks: TkTask[] }> {
  const { binary, cwd } = getTkConfig();

  // NOTE: tk ls --json always returns a flat array of tasks with project_uuid.
  // We fetch project metadata separately and group tasks on the client side.
  // See tk/cmd/display.go:outputTasksJSON for the task JSON format.
  // See tk/cmd/project/ls.go for the project JSON format.
  const args = ['ls', '--json'];

  let stdout: string;
  try {
    const result = await execFileAsync(binary, args, {
      cwd,
      env: { ...process.env, FORCE_COLOR: '0', CLICOLOR_FORCE: '0', TK_SKIP_INVLOG: '1' },
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

  // Parse the flat array of tasks from tk ls --json
  let allTasks: TkTask[];
  try {
    allTasks = JSON.parse(stdout) as TkTask[];
  } catch (error) {
    throw new Error(`Failed to parse JSON output from tk: ${error}`);
  }

  // Fetch project metadata
  const projectMap = await fetchProjects();

  // Group tasks by project_uuid
  const grouped = new Map<string, TkTask[]>();
  const ungrouped: TkTask[] = [];

  for (const task of allTasks) {
    if (task.project_uuid) {
      if (!grouped.has(task.project_uuid)) {
        grouped.set(task.project_uuid, []);
      }
      grouped.get(task.project_uuid)!.push(task);
    } else {
      // Tasks without a project_uuid go into ungrouped
      ungrouped.push(task);
    }
  }

  // Convert the Map to an array of groups with display names
  const groups: TkGroup[] = Array.from(grouped.entries()).map(([projectUUID, tasks]) => {
    const project = projectMap.get(projectUUID);
    // Use local_preferred_alias if available, otherwise fall back to name
    const groupName = project?.local_preferred_alias || project?.name || projectUUID;
    return {
      group: groupName,
      tasks: tasks,
    };
  });

  return { groups, tasks: ungrouped };
}
