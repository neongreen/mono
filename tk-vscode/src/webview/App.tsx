import { h } from 'preact';
import { useState } from 'preact/hooks';
import { TaskDetails } from './TaskDetails';
import { EmptyState } from './EmptyState';
import type { TkTask, VSCodeAPI } from './types';

interface AppProps {
  task: TkTask | null;
  allTasks?: TkTask[];
  vscode: VSCodeAPI;
}

export function App({ task, allTasks, vscode }: AppProps) {
  return (
    <div class="app">
      {task ? <TaskDetails task={task} allTasks={allTasks} vscode={vscode} /> : <EmptyState />}
    </div>
  );
}
