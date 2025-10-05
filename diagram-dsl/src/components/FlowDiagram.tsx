import React from 'react';
import type { BoxProps } from '../types';

export interface FlowStep {
  id: string;
  label: string;
  subtitle?: string;
  variant?: 'primary' | 'secondary' | 'accent' | 'success' | 'warning' | 'danger' | 'default';
}

export interface FlowDiagramProps extends Partial<BoxProps> {
  steps: FlowStep[];
  orientation?: 'horizontal' | 'vertical';
  stepWidth?: number;
  stepHeight?: number;
  showArrows?: boolean;
  arrowColor?: string;
}

/**
 * FlowDiagram component - creates a flow diagram with connected steps
 */
export const FlowDiagram: React.FC<FlowDiagramProps> = ({
  steps,
  orientation = 'horizontal',
  stepWidth = 180,
  stepHeight,
  showArrows = true,
  arrowColor = '#1976d2',
  gap = 20,
  ...props
}) => {
  const Container = orientation === 'horizontal' ? 'Row' : 'Stack';
  
  const stepElements: React.ReactNode[] = [];
  
  steps.forEach((step, index) => {
    // Add the step card
    stepElements.push(
      React.createElement('Card', {
        key: `step-${index}`,
        id: step.id,
        variant: step.variant || 'primary',
        width: stepWidth,
        height: stepHeight
      },
        React.createElement('Label', null, step.label),
        step.subtitle && React.createElement('Text', { fontSize: 10 }, step.subtitle)
      )
    );

    // Add arrow to next step (except for last step)
    if (showArrows && index < steps.length - 1) {
      const currentId = step.id;
      const nextId = steps[index + 1].id;
      
      stepElements.push(
        React.createElement('Stack', {
          key: `connector-${index}`,
          justifyContent: 'center'
        },
          React.createElement('Arrow', {
            from: currentId,
            to: nextId,
            color: arrowColor,
            strokeWidth: 2
          })
        )
      );
    }
  });

  return React.createElement(Container, { gap, ...props }, ...stepElements);
};
