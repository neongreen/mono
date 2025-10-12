import React from 'react';
import { ApproachViewer } from './ApproachViewer';
import * as Approach1 from '../lib/approach1-flexbox';
import * as Approach2 from '../lib/approach2-grid';
import * as Approach3 from '../lib/approach3-constrained';

interface ExamplePageProps {
  createExample: (components: any) => React.ReactNode;
  exampleInfo: {
    name: string;
    description: string;
  };
  sourceCode: string;
}

export const ExamplePage: React.FC<ExamplePageProps> = ({
  createExample,
  exampleInfo,
  sourceCode,
}) => {
  return (
    <div style={{ padding: '40px', maxWidth: '1600px', margin: '0 auto' }}>
      <div style={{ marginBottom: '40px' }}>
        <h1 style={{ fontSize: '32px', marginBottom: '16px', color: '#222' }}>
          {exampleInfo.name}
        </h1>
        <p style={{ fontSize: '18px', color: '#666', marginBottom: '32px' }}>
          {exampleInfo.description}
        </p>
        <p style={{ fontSize: '14px', color: '#999', fontStyle: 'italic' }}>
          Compare how the same layout looks with different implementation approaches
        </p>
      </div>

      <ApproachViewer
        name="Approach 1: Flexbox-based Constraint System"
        description="Uses CSS Flexbox with smart defaults and constraints. Automatic text wrapping and overflow prevention."
        component={createExample(Approach1)}
        sourceCode={sourceCode}
      />

      <ApproachViewer
        name="Approach 2: Grid-based Proportional System"
        description="Uses CSS Grid with proportional units (fr). Better for complex multi-dimensional layouts."
        component={createExample(Approach2)}
        sourceCode={sourceCode}
      />

      <ApproachViewer
        name="Approach 3: Constrained Absolute Positioning"
        description="Uses absolute positioning with intelligent constraint resolution. Explicit control when needed."
        component={createExample(Approach3)}
        sourceCode={sourceCode}
      />
    </div>
  );
};
