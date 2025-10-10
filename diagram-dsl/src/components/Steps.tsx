import React from 'react';
import type { BoxProps } from '../types';

export interface Step {
  number: number;
  title: string;
  description?: string;
  status?: 'pending' | 'active' | 'complete';
}

export interface StepsProps extends Partial<BoxProps> {
  steps: Step[];
  orientation?: 'vertical' | 'horizontal';
  showConnectors?: boolean;
}

const statusColors = {
  pending: { bg: '#f5f5f5', border: '#e0e0e0', text: '#999', number: '#666' },
  active: { bg: '#e3f2fd', border: '#1976d2', text: '#1565c0', number: '#1976d2' },
  complete: { bg: '#e8f5e9', border: '#2e7d32', text: '#1b5e20', number: '#2e7d32' },
};

/**
 * Steps component - displays a sequence of steps with status indicators
 * Perfect for tutorials, wizards, and process documentation
 */
export const Steps: React.FC<StepsProps> = ({
  steps,
  orientation = 'vertical',
  showConnectors = true,
  gap = 16,
  ...props
}) => {
  const Layout = orientation === 'vertical' ? 'Stack' : 'Row';
  
  const stepElements: React.ReactNode[] = [];
  
  steps.forEach((step, index) => {
    const status = step.status || 'pending';
    const colors = statusColors[status];
    
    // Step element
    const stepContent = React.createElement('Stack', {
      key: `step-${index}`,
      gap: 12
    },
      React.createElement('Row', { gap: 12, alignItems: 'center' },
        // Number circle
        React.createElement('Box', {
          width: 32,
          height: 32,
          backgroundColor: colors.bg,
          borderColor: colors.border,
          borderWidth: 2,
          borderRadius: 16,
          justifyContent: 'center',
          alignItems: 'center'
        },
          React.createElement('Text', {
            fontSize: 14,
            fontWeight: 'bold',
            color: colors.number,
            children: status === 'complete' ? '✓' : step.number.toString()
          })
        ),
        // Title
        React.createElement('Text', {
          fontSize: 14,
          fontWeight: 'bold',
          color: colors.text,
          children: step.title
        })
      ),
      // Description
      step.description && React.createElement('Text', {
        fontSize: 12,
        color: '#666',
        marginLeft: 44,
        children: step.description
      })
    );
    
    stepElements.push(stepContent);
    
    // Connector line (except for last step)
    if (showConnectors && index < steps.length - 1) {
      const connectorProps = orientation === 'vertical'
        ? { height: gap, width: 2, marginLeft: 15 }
        : { width: gap, height: 2, marginTop: 15 };
      
      stepElements.push(
        React.createElement('Box', {
          key: `connector-${index}`,
          backgroundColor: '#e0e0e0',
          ...connectorProps
        })
      );
    }
  });
  
  return React.createElement(Layout, { gap: 0, ...props }, ...stepElements);
};
