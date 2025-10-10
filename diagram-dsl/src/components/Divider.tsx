import React from 'react';
import type { BoxProps } from '../types';

export interface DividerProps extends Partial<BoxProps> {
  variant?: 'solid' | 'dashed';
  thickness?: number;
  color?: string;
  orientation?: 'horizontal' | 'vertical';
}

/**
 * Divider component - creates a visual separator line
 */
export const Divider: React.FC<DividerProps> = ({
  variant = 'solid',
  thickness = 1,
  color = '#ddd',
  orientation = 'horizontal',
  width,
  height,
  margin = 16,
  ...props
}) => {
  if (orientation === 'horizontal') {
    return React.createElement('Box', {
      width: width || '100%',
      height: thickness,
      backgroundColor: variant === 'solid' ? color : 'transparent',
      borderColor: variant === 'dashed' ? color : 'transparent',
      borderWidth: variant === 'dashed' ? thickness : 0,
      borderStyle: variant === 'dashed' ? 'dashed' : 'solid',
      marginTop: margin,
      marginBottom: margin,
      ...props
    });
  } else {
    return React.createElement('Box', {
      width: thickness,
      height: height || '100%',
      backgroundColor: variant === 'solid' ? color : 'transparent',
      borderColor: variant === 'dashed' ? color : 'transparent',
      borderWidth: variant === 'dashed' ? thickness : 0,
      borderStyle: variant === 'dashed' ? 'dashed' : 'solid',
      marginLeft: margin,
      marginRight: margin,
      ...props
    });
  }
};
