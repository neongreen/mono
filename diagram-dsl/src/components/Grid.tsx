import React from 'react';
import type { BoxProps } from '../types';

export interface GridProps extends Partial<BoxProps> {
  children?: React.ReactNode;
  columns?: number;
  gap?: number;
}

/**
 * Grid component - arranges children in a grid layout
 * Note: This is a simple row-wrapping implementation
 */
export const Grid: React.FC<GridProps> = ({ 
  children,
  columns = 2,
  gap = 16,
  ...props 
}) => {
  const childArray = React.Children.toArray(children);
  const rows: React.ReactNode[][] = [];
  
  for (let i = 0; i < childArray.length; i += columns) {
    rows.push(childArray.slice(i, i + columns));
  }
  
  return React.createElement('Stack', { gap, ...props },
    ...rows.map((row, rowIndex) =>
      React.createElement('Row', { 
        key: rowIndex, 
        gap,
        justifyContent: 'space-between'
      }, ...row)
    )
  );
};
