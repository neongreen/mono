import { h } from 'preact';
import type { TkTask, RelationEdge } from './types';

interface RelationsSectionProps {
  task: TkTask;
  allTasks?: TkTask[];
}

// Helper to resolve UUID to display ID + title
function resolveTask(uuid: string, allTasks?: TkTask[]): { display_id: string; title?: string } {
  if (allTasks) {
    const found = allTasks.find(t => t.uuid === uuid);
    if (found) {
      return { display_id: found.display_id ?? uuid, title: found.title };
    }
  }
  return { display_id: uuid };
}

// Reusable component for rendering a single relation item
function RelationItem({ uuid, edge, allTasks }: { uuid?: string; edge?: RelationEdge; allTasks?: TkTask[] }) {
  const targetUuid = uuid ?? edge?.dst;
  if (!targetUuid) return null;

  const resolved = resolveTask(targetUuid, allTasks);

  return (
    <li>
      <span class="relation-task-id">{resolved.display_id}</span>
      {resolved.title && <span class="relation-title"> - {resolved.title}</span>}
      {edge?.note && <span class="relation-note"> ({edge.note})</span>}
    </li>
  );
}

export function RelationsSection({ task, allTasks }: RelationsSectionProps) {
  if (!task.relations) {
    return null;
  }

  const { relations } = task;
  const hasAnyRelations =
    (relations.related?.out && relations.related.out.length > 0) ||
    (relations.subtask?.children && relations.subtask.children.length > 0) ||
    relations.subtask?.parent ||
    (relations.blocks?.out && relations.blocks.out.length > 0) ||
    (relations.blocks?.in && relations.blocks.in.length > 0) ||
    (relations.duplicate_of?.out && relations.duplicate_of.out.length > 0);

  if (!hasAnyRelations) {
    return null;
  }

  return (
    <div class="section">
      <div class="section-title">Relations</div>

      {relations.related?.out && relations.related.out.length > 0 && (
        <div class="relation-group">
          <div class="relation-label">Related:</div>
          <ul class="relation-list">
            {relations.related.out.map((edge, i) => (
              <RelationItem key={i} edge={edge} allTasks={allTasks} />
            ))}
          </ul>
        </div>
      )}

      {relations.subtask?.children && relations.subtask.children.length > 0 && (
        <div class="relation-group">
          <div class="relation-label">Subtasks:</div>
          <ul class="relation-list">
            {relations.subtask.children.map((uuid, i) => (
              <RelationItem key={i} uuid={uuid} allTasks={allTasks} />
            ))}
          </ul>
        </div>
      )}

      {relations.subtask?.parent && (
        <div class="relation-group">
          <div class="relation-label">Parent:</div>
          <div class="relation-value">
            {(() => {
              const resolved = resolveTask(relations.subtask!.parent, allTasks);
              return (
                <>
                  <span class="relation-task-id">{resolved.display_id}</span>
                  {resolved.title && <span class="relation-title"> - {resolved.title}</span>}
                </>
              );
            })()}
          </div>
        </div>
      )}

      {relations.blocks?.out && relations.blocks.out.length > 0 && (
        <div class="relation-group">
          <div class="relation-label">Blocks:</div>
          <ul class="relation-list">
            {relations.blocks.out.map((edge, i) => (
              <RelationItem key={i} edge={edge} allTasks={allTasks} />
            ))}
          </ul>
        </div>
      )}

      {relations.blocks?.in && relations.blocks.in.length > 0 && (
        <div class="relation-group">
          <div class="relation-label">Blocked by:</div>
          <ul class="relation-list">
            {relations.blocks.in.map((edge, i) => (
              <RelationItem key={i} edge={edge} allTasks={allTasks} />
            ))}
          </ul>
        </div>
      )}

      {relations.duplicate_of?.out && relations.duplicate_of.out.length > 0 && (
        <div class="relation-group">
          <div class="relation-label">Duplicate of:</div>
          <ul class="relation-list">
            {relations.duplicate_of.out.map((edge, i) => (
              <RelationItem key={i} edge={edge} allTasks={allTasks} />
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
