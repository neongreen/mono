import { render } from 'preact';
import { App } from './App';

// @ts-ignore - vscode is injected by the webview
const vscode = acquireVsCodeApi();

// Listen for messages from the extension
window.addEventListener('message', (event) => {
  const message = event.data;
  if (message.type === 'updateTask') {
    // Apply font settings to document root (tk-vsc-94)
    if (message.fontFamily) {
      document.documentElement.style.setProperty('--font-family', message.fontFamily);
    }
    if (message.fontSize) {
      document.documentElement.style.setProperty('--font-size', `${message.fontSize}px`);
    }

    render(
      <App
        task={message.task}
        allTasks={message.allTasks}
        vscode={vscode}
        showDeleteButton={message.showDeleteButton}
      />,
      document.body
    );
  } else if (message.type === 'clear') {
    render(<App task={null} vscode={vscode} />, document.body);
  }
});

// Initial render
render(<App task={null} vscode={vscode} />, document.body);
