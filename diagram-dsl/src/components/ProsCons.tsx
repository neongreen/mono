import React from 'react';
import type { BoxProps } from '../types';

export interface ProsConsProps extends Partial<BoxProps> {
  pros: string[];
  cons: string[];
  prosTitle?: string;
  consTitle?: string;
  fontSize?: number;
}

/**
 * ProsCons component - displays pros and cons side by side
 */
export const ProsCons: React.FC<ProsConsProps> = ({ 
  pros,
  cons,
  prosTitle = '✓ Pros',
  consTitle = '✗ Cons',
  fontSize = 11,
  gap = 32,
  ...props 
}) => {
  return React.createElement('Row', { gap, ...props },
    // Pros column
    React.createElement('Stack', { gap: 8, flexGrow: 1 },
      React.createElement('Text', { 
        fontSize: fontSize + 2, 
        fontWeight: 'bold', 
        color: '#2e7d32' 
      }, prosTitle),
      ...pros.map((pro, i) => 
        React.createElement('Text', { key: `pro-${i}`, fontSize }, `• ${pro}`)
      )
    ),
    // Cons column
    React.createElement('Stack', { gap: 8, flexGrow: 1 },
      React.createElement('Text', { 
        fontSize: fontSize + 2, 
        fontWeight: 'bold', 
        color: '#c62828' 
      }, consTitle),
      ...cons.map((con, i) => 
        React.createElement('Text', { key: `con-${i}`, fontSize }, `• ${con}`)
      )
    )
  );
};
