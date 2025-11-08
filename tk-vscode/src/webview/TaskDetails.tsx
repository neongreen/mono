import { h } from 'preact';
import type { TkTask, VSCodeAPI } from './types';
import { TaskHeader } from './TaskHeader';
import { RelationsSection } from './RelationsSection';
import { NotesSection } from './NotesSection';

interface TaskDetailsProps {
  task: TkTask;
  vscode: VSCodeAPI;
  allTasks?: TkTask[]; // For resolving UUIDs to display IDs
}

export function TaskDetails({ task, vscode, allTasks }: TaskDetailsProps) {
  return (
    <div class="task-details">
      <TaskHeader task={task} vscode={vscode} />
      <RelationsSection task={task} allTasks={allTasks} />
      <NotesSection task={task} vscode={vscode} />
    </div>
  );
}
