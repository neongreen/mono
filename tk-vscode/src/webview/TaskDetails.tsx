import { h } from 'preact';
import { useState } from 'preact/hooks';
import { encode as encodeHtml } from 'he';
import type { TkTask, VSCodeAPI } from './types';

interface TaskDetailsProps {
  task: TkTask;
  vscode: VSCodeAPI;
}

export function TaskDetails({ task, vscode }: TaskDetailsProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [titleValue, setTitleValue] = useState(task.title ?? 'No title');

  const taskId = task.task_id ?? 'unknown';
  const genericAxis = task.axes?.['generic'];
  const status = genericAxis?.effective ?? 'none';
  const blocked = task.blocked ? 'yes' : 'no';

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
    <div class="task-details">
      <div class="task-header">
        <span class="task-id">{taskId}</span>
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

      <div class="section">
        <div class="metadata-grid">
          <div class="metadata-label">Status:</div>
          <div class="metadata-value">{status}</div>
          <div class="metadata-label">Blocked:</div>
          <div class="metadata-value">{blocked}</div>
        </div>
      </div>

      <div class="section">
        <div class="section-title">Notes</div>
        {task.notes && task.notes.length > 0 ? (
          task.notes.map((note, i) => (
            <div class="note" key={i}>
              <div class="note-content">{note.markdown || '(empty note)'}</div>
              <div class="note-meta">
                {note.timestamp && (
                  <span class="note-time">
                    {new Date(note.timestamp).toLocaleString()}
                  </span>
                )}
                <span class="note-actor">by {note.actor || 'Unknown'}</span>
              </div>
            </div>
          ))
        ) : (
          <div class="empty-section">No notes</div>
        )}
      </div>
    </div>
  );
}
