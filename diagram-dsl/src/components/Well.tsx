import React from 'react';
import type { BoxProps } from '../types';

export interface WellProps extends Partial<BoxProps> {
  children?: React.ReactNode;
  variant?: 'default' | 'info' | 'success' | 'warning' | 'danger';
  inset?: boolean;
}

const variantColors = {
  default: { bg: '#f5f5f5', border: '#e0e0e0' },
  info: { bg: '#e3f2fd', border: '#90caf9' },
  success: { bg: '#e8f5e9', border: '#a5d6a7' },
  warning: { bg: '#fff3e0', border: '#ffcc80' },
  danger: { bg: '#ffebee', border: '#ef9a9a' },
};

/**
 * Well component - an inset container for de-emphasized content
 * Perfect for examples, notes, or secondary information
 */
export const Well: React.FC<WellProps> = ({
  children,
  variant = 'default',
  inset = true,
  padding = 16,
  borderRadius = 6,
  ...props
}) => {
  const colors = variantColors[variant];
  
  return React.createElement('Box', {
    backgroundColor: colors.bg,
    borderColor: inset ? colors.border : 'transparent',
    borderWidth: inset ? 1 : 0,
    borderRadius,
    padding,
    ...props
  }, children);
};
