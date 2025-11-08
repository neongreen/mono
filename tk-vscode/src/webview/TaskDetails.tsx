import { h } from 'preact';
import type { TkTask, VSCodeAPI } from './types';
import { TaskHeader } from './TaskHeader';
import { RelationsSection } from './RelationsSection';
import { NotesSection } from './NotesSection';

interface TaskDetailsProps {
  task: TkTask;
  vscode: VSCodeAPI;
  allTasks?: TkTask[]; // For resolving UUIDs to display IDs
  showDeleteButton?: boolean;
}

export function TaskDetails({ task, vscode, allTasks, showDeleteButton }: TaskDetailsProps) {
  const handleDelete = () => {
    const taskId = task.display_id ?? 'unknown';
    const taskTitle = task.title ?? taskId;
    const confirmMessage = `Delete task ${taskId}${taskTitle !== taskId ? ` (${taskTitle})` : ''}?`;

    if (confirm(confirmMessage)) {
      vscode.postMessage({
        type: 'deleteTask',
        taskId: task.display_id
      });
    }
  };

  return (
    <div class="task-details">
      <TaskHeader task={task} vscode={vscode} />
      {showDeleteButton && (
        <div class="section">
          <button class="btn btn-danger" onClick={handleDelete}>
            Delete Task
          </button>
        </div>
      )}
      <RelationsSection task={task} allTasks={allTasks} />
      <NotesSection task={task} vscode={vscode} />
    </div>
  );
}
