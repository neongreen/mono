import React from 'react';
import type { BoxProps } from '../types';

export interface SpacerProps extends Partial<BoxProps> {
  size?: number;
  flexible?: boolean;
}

/**
 * Spacer component - adds space between elements
 */
export const Spacer: React.FC<SpacerProps> = ({ 
  size = 20,
  flexible = false,
  ...props 
}) => {
  if (flexible) {
    return React.createElement('Box', {
      flexGrow: 1,
      ...props
    });
  }
  
  return React.createElement('Box', {
    width: size,
    height: size,
    ...props
  });
};
