import React from 'react';
import type { BoxProps } from '../types';

export interface ListProps extends Partial<BoxProps> {
  items: (string | React.ReactNode)[];
  bullet?: string;
  fontSize?: number;
  color?: string;
  gap?: number;
}

/**
 * List component - renders a list of items with bullets
 */
export const List: React.FC<ListProps> = ({ 
  items, 
  bullet = '•',
  fontSize = 14,
  color,
  gap = 8,
  ...props 
}) => {
  return React.createElement('Stack', { gap, ...props }, 
    items.map((item, index) => 
      React.createElement('Text', { 
        key: index,
        fontSize,
        color
      }, `${bullet} ${typeof item === 'string' ? item : ''}`, 
      typeof item !== 'string' ? item : null)
    )
  );
};
