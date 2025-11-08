import { h } from 'preact';
import { useState, useEffect } from 'preact/hooks';
import type { TkTask, VSCodeAPI } from './types';

interface TaskHeaderProps {
  task: TkTask;
  vscode: VSCodeAPI;
}

export function TaskHeader({ task, vscode }: TaskHeaderProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [titleValue, setTitleValue] = useState(task.title ?? 'No title');

  // Update title when task changes
  useEffect(() => {
    setTitleValue(task.title ?? 'No title');
    setIsEditing(false);
  }, [task.display_id, task.uuid, task.title]);

  const taskId = task.display_id ?? 'unknown';
  const genericAxis = task.axes?.['generic'];
  const status = genericAxis?.effective ?? 'none';
  const blocked = task.blocked ?? false;

  const handleStartEdit = () => {
    setIsEditing(true);
  };

  const handleCancel = () => {
    setTitleValue(task.title ?? 'No title');
    setIsEditing(false);
  };

  const handleSave = () => {
    if (titleValue.trim() === '') {
      return;
    }
    vscode.postMessage({
      type: 'editTitle',
      newTitle: titleValue,
    });
    setIsEditing(false);
  };

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      handleSave();
    }
    if (e.key === 'Escape') {
      e.preventDefault();
      handleCancel();
    }
  };

  return (
    <div class="task-header">
      <div class="task-header-line">
        <span class="task-id">{taskId}</span>
        <span class="task-meta">
          <span class={`status-badge status-${status}`}>{status}</span>
          {blocked && <span class="blocked-badge">⛔ blocked</span>}
        </span>
      </div>
      <div class="title-container">
        {!isEditing ? (
          <div class="title-display" onClick={handleStartEdit}>
            {titleValue}
          </div>
        ) : (
          <div class="title-edit">
            <textarea
              class="title-textarea"
              value={titleValue}
              onInput={(e) => setTitleValue((e.target as HTMLTextAreaElement).value)}
              onKeyDown={handleKeyDown}
              autoFocus
            />
            <div class="title-buttons">
              <button class="btn btn-primary" onClick={handleSave}>
                Save
              </button>
              <button class="btn btn-secondary" onClick={handleCancel}>
                Cancel
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
