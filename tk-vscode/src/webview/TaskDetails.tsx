import { h } from 'preact';
import { useState, useEffect } from 'preact/hooks';
import { encode as encodeHtml } from 'he';
import { marked } from 'marked';
import DOMPurify from 'isomorphic-dompurify';
import type { TkTask, VSCodeAPI } from './types';

interface TaskDetailsProps {
  task: TkTask;
  vscode: VSCodeAPI;
}

export function TaskDetails({ task, vscode }: TaskDetailsProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [titleValue, setTitleValue] = useState(task.title ?? 'No title');
  const [isAddingNote, setIsAddingNote] = useState(false);
  const [newNoteValue, setNewNoteValue] = useState('');

  // Update title when task changes
  useEffect(() => {
    setTitleValue(task.title ?? 'No title');
    setIsEditing(false);
    setIsAddingNote(false);
  }, [task.task_id, task.task_uuid, task.title]);

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

  const handleAddNote = () => {
    setIsAddingNote(true);
    setNewNoteValue('');
  };

  const handleCancelNewNote = () => {
    setIsAddingNote(false);
    setNewNoteValue('');
  };

  const handleSaveNewNote = () => {
    if (newNoteValue.trim() === '') {
      return;
    }
    vscode.postMessage({
      type: 'addNote',
      markdown: newNoteValue,
    });
    setIsAddingNote(false);
    setNewNoteValue('');
  };

  const handleNoteKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      handleSaveNewNote();
    }
    if (e.key === 'Escape') {
      e.preventDefault();
      handleCancelNewNote();
    }
  };

  const renderMarkdown = (markdown: string) => {
    try {
      // Parse markdown to HTML
      const rawHtml = marked.parse(markdown, { 
        async: false,
        breaks: true,
        gfm: true
      }) as string;
      
      // Sanitize the HTML to prevent XSS attacks
      // DOMPurify removes any potentially dangerous content while preserving safe HTML
      const sanitizedHtml = DOMPurify.sanitize(rawHtml);
      return sanitizedHtml;
    } catch (e) {
      return encodeHtml(markdown);
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
        <div class="section-header">
          <div class="section-title">Notes</div>
          {!isAddingNote && (
            <button class="btn btn-small" onClick={handleAddNote}>
              Add Note
            </button>
          )}
        </div>
        
        {isAddingNote && (
          <div class="note-editor">
            <textarea
              class="note-textarea"
              value={newNoteValue}
              onInput={(e) => setNewNoteValue((e.target as HTMLTextAreaElement).value)}
              onKeyDown={handleNoteKeyDown}
              placeholder="Write your note in markdown..."
              autoFocus
            />
            <div class="note-buttons">
              <button class="btn btn-primary" onClick={handleSaveNewNote}>
                Save Note
              </button>
              <button class="btn btn-secondary" onClick={handleCancelNewNote}>
                Cancel
              </button>
            </div>
          </div>
        )}

        {task.notes && task.notes.length > 0 ? (
          task.notes.map((note, i) => (
            <div class="note" key={i}>
              <div 
                class="note-content markdown-content" 
                dangerouslySetInnerHTML={{ __html: renderMarkdown(note.markdown || '(empty note)') }}
              />
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
          !isAddingNote && <div class="empty-section">No notes</div>
        )}
      </div>
    </div>
  );
}
