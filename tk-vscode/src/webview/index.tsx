import { render } from 'preact';
import { App } from './App';

// @ts-ignore - vscode is injected by the webview
const vscode = acquireVsCodeApi();

// Listen for messages from the extension
window.addEventListener('message', (event) => {
  const message = event.data;
  if (message.type === 'updateTask') {
    render(<App task={message.task} vscode={vscode} />, document.body);
  } else if (message.type === 'clear') {
    render(<App task={null} vscode={vscode} />, document.body);
  }
});

// Initial render
render(<App task={null} vscode={vscode} />, document.body);
