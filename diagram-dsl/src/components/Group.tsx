import React from 'react';
import type { BoxProps } from '../types';

export interface GroupProps extends Partial<BoxProps> {
  children?: React.ReactNode;
  label?: string;
  direction?: 'horizontal' | 'vertical';
  spacing?: 'tight' | 'normal' | 'relaxed';
  align?: 'start' | 'center' | 'end' | 'stretch';
}

/**
 * Group component - lightweight grouping without visual borders
 * Provides consistent spacing and alignment for related elements
 */
export const Group: React.FC<GroupProps> = ({
  children,
  label,
  direction = 'vertical',
  spacing = 'normal',
  align = 'start',
  ...props
}) => {
  // Map spacing to gap values
  const gapMap = {
    tight: 8,
    normal: 16,
    relaxed: 24
  };
  
  // Map align to flexbox alignment
  const alignMap = {
    start: 'flex-start',
    center: 'center',
    end: 'flex-end',
    stretch: 'stretch'
  };
  
  const gap = gapMap[spacing];
  const Layout = direction === 'horizontal' ? 'Row' : 'Stack';
  const alignItems = alignMap[align];
  
  return React.createElement('Stack', { gap: label ? 8 : 0, ...props },
    label && React.createElement('Text', {
      fontSize: 12,
      fontWeight: 'bold',
      color: '#666',
      marginBottom: 4
    }, label),
    React.createElement(Layout, { gap, alignItems }, children)
  );
};
