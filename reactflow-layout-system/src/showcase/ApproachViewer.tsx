import React from 'react';
import './ApproachViewer.css';

interface ApproachViewerProps {
  name: string;
  description: string;
  component: React.ReactNode;
  sourceCode: string;
}

export const ApproachViewer: React.FC<ApproachViewerProps> = ({
  name,
  description,
  component,
  sourceCode,
}) => {
  return (
    <div className="approach-viewer">
      <div className="approach-header">
        <h3>{name}</h3>
        <p>{description}</p>
      </div>
      <div className="approach-content">
        <div className="preview-section">
          <h4>Preview</h4>
          <div className="preview-container">
            {component}
          </div>
        </div>
        <div className="code-section">
          <h4>Source Code</h4>
          <pre className="code-container">
            <code>{sourceCode}</code>
          </pre>
        </div>
      </div>
    </div>
  );
};
