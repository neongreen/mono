import { h } from 'preact';
import { useState, useEffect } from 'preact/hooks';
import { encode as encodeHtml } from 'he';
import { marked } from 'marked';
import DOMPurify from 'isomorphic-dompurify';
import type { TkTask, VSCodeAPI } from './types';

interface NotesSectionProps {
  task: TkTask;
  vscode: VSCodeAPI;
}

// Helper to render markdown safely
function renderMarkdown(markdown: string): string {
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
}

export function NotesSection({ task, vscode }: NotesSectionProps) {
  const [isAddingNote, setIsAddingNote] = useState(false);
  const [newNoteValue, setNewNoteValue] = useState('');

  // Reset note editor when task changes
  useEffect(() => {
    setIsAddingNote(false);
    setNewNoteValue('');
  }, [task.display_id, task.uuid]);

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

  return (
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
  );
}
