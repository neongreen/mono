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
      // Parse markdown to HTML using synchronous marked
      // marked() is the synchronous version, marked.parse() is async
      const rawHtml = marked(markdown, { 
        breaks: true,
        gfm: true
      });
      
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

      {/* Relations section */}
      {task.relations && (
        <div class="section">
          <div class="section-title">Relations</div>
          
          {task.relations.related?.out && task.relations.related.out.length > 0 && (
            <div class="relation-group">
              <div class="relation-label">Related:</div>
              <ul class="relation-list">
                {task.relations.related.out.map((edge, i) => (
                  <li key={i}>
                    <span class="relation-task-id">{edge.dst}</span>
                    {edge.note && <span class="relation-note"> - {edge.note}</span>}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {task.relations.subtask?.children && task.relations.subtask.children.length > 0 && (
            <div class="relation-group">
              <div class="relation-label">Subtasks:</div>
              <ul class="relation-list">
                {task.relations.subtask.children.map((uuid, i) => (
                  <li key={i}>
                    <span class="relation-task-id">{uuid}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {task.relations.subtask?.parent && (
            <div class="relation-group">
              <div class="relation-label">Parent:</div>
              <div class="relation-value">
                <span class="relation-task-id">{task.relations.subtask.parent}</span>
              </div>
            </div>
          )}

          {task.relations.blocks?.out && task.relations.blocks.out.length > 0 && (
            <div class="relation-group">
              <div class="relation-label">Blocks:</div>
              <ul class="relation-list">
                {task.relations.blocks.out.map((edge, i) => (
                  <li key={i}>
                    <span class="relation-task-id">{edge.dst}</span>
                    {edge.note && <span class="relation-note"> - {edge.note}</span>}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {task.relations.blocks?.in && task.relations.blocks.in.length > 0 && (
            <div class="relation-group">
              <div class="relation-label">Blocked by:</div>
              <ul class="relation-list">
                {task.relations.blocks.in.map((edge, i) => (
                  <li key={i}>
                    <span class="relation-task-id">{edge.dst}</span>
                    {edge.note && <span class="relation-note"> - {edge.note}</span>}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {task.relations.duplicate_of?.out && task.relations.duplicate_of.out.length > 0 && (
            <div class="relation-group">
              <div class="relation-label">Duplicate of:</div>
              <ul class="relation-list">
                {task.relations.duplicate_of.out.map((edge, i) => (
                  <li key={i}>
                    <span class="relation-task-id">{edge.dst}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

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
