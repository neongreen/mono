import React from 'react';
import type { BoxProps } from '../types';

export interface HighlightProps extends Partial<BoxProps> {
  children?: React.ReactNode;
  color?: string;
  variant?: 'info' | 'success' | 'warning' | 'danger';
}

const variantColors = {
  info: { bg: '#e3f2fd', border: '#1976d2' },
  success: { bg: '#e8f5e9', border: '#2e7d32' },
  warning: { bg: '#fff3e0', border: '#ff9800' },
  danger: { bg: '#ffebee', border: '#c62828' },
};

/**
 * Highlight component - highlights important content
 */
export const Highlight: React.FC<HighlightProps> = ({ 
  children,
  color,
  variant = 'info',
  padding = 12,
  borderRadius = 6,
  backgroundColor,
  borderColor,
  ...props 
}) => {
  const colors = variantColors[variant];
  
  return React.createElement('Box', {
    backgroundColor: backgroundColor || colors.bg,
    borderColor: borderColor || colors.border,
    borderWidth: 1,
    borderRadius,
    padding,
    ...props
  }, children);
};
