import { describe, it, expect, beforeEach, vi } from 'vitest';
import { TkProvider } from './treeProvider';
import { TkDecorationProvider } from './decorationProvider';
import { TkGroup, TkTask } from './types';

// Mock the API
vi.mock('./tkApi', () => ({
  fetchTk: vi.fn(),
}));

describe('TkProvider', () => {
  let provider: TkProvider;
  let decorationProvider: TkDecorationProvider;

  beforeEach(() => {
    decorationProvider = new TkDecorationProvider();
    provider = new TkProvider(decorationProvider);
    vi.clearAllMocks();
  });

  describe('Bug fix: subtasks should be included in decoration provider', () => {
    it('should collect all task items including subtasks for decoration provider', () => {
      // Given: A parent task with a subtask
      const mockGroups: TkGroup[] = [
        {
          group: 'test-group',
          tasks: [
            {
              uuid: 'parent-uuid',
              display_id: 'test-1',
              title: 'Parent Task',
              axes: {
                generic: {
                  effective: 'wip',
                },
              },
              relations: {
                subtask: {
                  children: ['child-uuid'],
                },
              },
            },
            {
              uuid: 'child-uuid',
              display_id: 'test-2',
              title: 'Child Task',
              axes: {
                generic: {
                  effective: 'next',
                },
              },
            },
          ],
        },
      ];

      // Set raw data
      provider['rawGroups'] = mockGroups;
      provider['rawUngrouped'] = [];

      // When: getChildren is called (which builds the tree)
      provider.getChildren();

      // Then: displayedTaskItems should include BOTH parent and child
      const displayedItems = provider['displayedTaskItems'];

      expect(displayedItems).toHaveLength(2);

      const parentItem = displayedItems.find(item => item.task.uuid === 'parent-uuid');
      const childItem = displayedItems.find(item => item.task.uuid === 'child-uuid');

      expect(parentItem).toBeDefined();
      expect(childItem).toBeDefined();

      // Both should have resourceUri set (needed for decoration)
      expect(parentItem?.resourceUri).toBeDefined();
      expect(childItem?.resourceUri).toBeDefined();

      // Both should have statusColor set
      expect(parentItem?.statusColor).toBeDefined();
      expect(childItem?.statusColor).toBeDefined();
    });

    it('should pass all items including subtasks to decoration provider', () => {
      // Given: A task hierarchy
      const mockGroups: TkGroup[] = [
        {
          group: 'project',
          tasks: [
            {
              uuid: 'task-1',
              display_id: 'proj-1',
              title: 'Top-level task',
              axes: {
                generic: { effective: 'done' },
              },
              relations: {
                subtask: {
                  children: ['task-2', 'task-3'],
                },
              },
            },
            {
              uuid: 'task-2',
              display_id: 'proj-2',
              title: 'Subtask 1',
              axes: {
                generic: { effective: 'wip' },
              },
            },
            {
              uuid: 'task-3',
              display_id: 'proj-3',
              title: 'Subtask 2',
              axes: {
                generic: { effective: 'next' },
              },
              relations: {
                subtask: {
                  children: ['task-4'],
                },
              },
            },
            {
              uuid: 'task-4',
              display_id: 'proj-4',
              title: 'Nested subtask',
              axes: {
                generic: { effective: 'wip' },
              },
            },
          ],
        },
      ];

      provider['rawGroups'] = mockGroups;
      provider['rawUngrouped'] = [];
      provider['showDone'] = true; // Show all tasks including done ones

      // When: getChildren is called
      provider.getChildren();

      // Then: All 4 tasks should be in displayedTaskItems
      const displayedItems = provider['displayedTaskItems'];
      expect(displayedItems).toHaveLength(4);

      // Verify each task is included
      const uuids = displayedItems.map(item => item.task.uuid);
      expect(uuids).toContain('task-1');
      expect(uuids).toContain('task-2');
      expect(uuids).toContain('task-3');
      expect(uuids).toContain('task-4');
    });

    it('should handle ungrouped tasks with subtasks', () => {
      // Given: Ungrouped tasks with subtask relations
      const mockUngrouped: TkTask[] = [
        {
          uuid: 'ungrouped-1',
          display_id: 'task-1',
          title: 'Ungrouped parent',
          axes: {
            generic: { effective: 'wip' },
          },
          relations: {
            subtask: {
              children: ['ungrouped-2'],
            },
          },
        },
        {
          uuid: 'ungrouped-2',
          display_id: 'task-2',
          title: 'Ungrouped child',
          axes: {
            generic: { effective: 'next' },
          },
        },
      ];

      provider['rawGroups'] = [];
      provider['rawUngrouped'] = mockUngrouped;

      // When: getChildren is called
      provider.getChildren();

      // Then: Both tasks should be in displayedTaskItems
      const displayedItems = provider['displayedTaskItems'];
      expect(displayedItems).toHaveLength(2);

      const uuids = displayedItems.map(item => item.task.uuid);
      expect(uuids).toContain('ungrouped-1');
      expect(uuids).toContain('ungrouped-2');
    });
  });

  describe('collectAllTaskItems helper', () => {
    it('should recursively collect all items in a tree', () => {
      // Create a mock tree structure
      const mockGroups: TkGroup[] = [
        {
          group: 'test',
          tasks: [
            {
              uuid: 'root',
              display_id: 'test-1',
              title: 'Root',
              axes: { generic: { effective: 'wip' } },
              relations: {
                subtask: {
                  children: ['child1', 'child2'],
                },
              },
            },
            {
              uuid: 'child1',
              display_id: 'test-2',
              title: 'Child 1',
              axes: { generic: { effective: 'next' } },
            },
            {
              uuid: 'child2',
              display_id: 'test-3',
              title: 'Child 2',
              axes: { generic: { effective: 'done' } },
            },
          ],
        },
      ];

      provider['rawGroups'] = mockGroups;
      provider['rawUngrouped'] = [];

      provider.getChildren();

      // Verify all items are collected
      const items = provider['displayedTaskItems'];
      expect(items).toHaveLength(3);

      // Verify the tree structure is maintained in the items themselves
      const rootItem = items.find(item => item.task.uuid === 'root');
      expect(rootItem?.children).toHaveLength(2);
    });
  });

  describe('Filter logic', () => {
    it('should filter done tasks when showDone is false', () => {
      const mockGroups: TkGroup[] = [
        {
          group: 'test',
          tasks: [
            {
              uuid: 'task-1',
              display_id: 'test-1',
              title: 'WIP task',
              axes: { generic: { effective: 'wip' } },
            },
            {
              uuid: 'task-2',
              display_id: 'test-2',
              title: 'Done task',
              axes: { generic: { effective: 'done' } },
            },
          ],
        },
      ];

      provider['rawGroups'] = mockGroups;
      provider['rawUngrouped'] = [];
      provider['showDone'] = false;

      const result = provider.getChildren();

      // Should only have the group (with filtered tasks inside)
      expect(result).toHaveLength(1);

      // The displayed items should only include the WIP task
      const displayedItems = provider['displayedTaskItems'];
      expect(displayedItems).toHaveLength(1);
      expect(displayedItems[0].task.uuid).toBe('task-1');
    });

    it('should show all tasks when showDone is true', () => {
      const mockGroups: TkGroup[] = [
        {
          group: 'test',
          tasks: [
            {
              uuid: 'task-1',
              display_id: 'test-1',
              title: 'WIP task',
              axes: { generic: { effective: 'wip' } },
            },
            {
              uuid: 'task-2',
              display_id: 'test-2',
              title: 'Done task',
              axes: { generic: { effective: 'done' } },
            },
          ],
        },
      ];

      provider['rawGroups'] = mockGroups;
      provider['rawUngrouped'] = [];
      provider['showDone'] = true;

      provider.getChildren();

      // Both tasks should be displayed
      const displayedItems = provider['displayedTaskItems'];
      expect(displayedItems).toHaveLength(2);
    });

    it('should filter by search term', () => {
      const mockGroups: TkGroup[] = [
        {
          group: 'test',
          tasks: [
            {
              uuid: 'task-1',
              display_id: 'test-1',
              title: 'Fix bug in authentication',
              axes: { generic: { effective: 'wip' } },
            },
            {
              uuid: 'task-2',
              display_id: 'test-2',
              title: 'Add new feature',
              axes: { generic: { effective: 'next' } },
            },
          ],
        },
      ];

      provider['rawGroups'] = mockGroups;
      provider['rawUngrouped'] = [];
      provider['searchTerm'] = 'bug';

      provider.getChildren();

      // Only the bug task should match
      const displayedItems = provider['displayedTaskItems'];
      expect(displayedItems).toHaveLength(1);
      expect(displayedItems[0].task.uuid).toBe('task-1');
    });
  });
});
