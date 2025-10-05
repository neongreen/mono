import React from 'react';
import type { BoxProps } from '../types';

export interface CalloutProps extends Partial<BoxProps> {
  title?: string;
  children?: React.ReactNode;
  variant?: 'info' | 'success' | 'warning' | 'danger';
  icon?: string;
}

const variantConfig = {
  info: { 
    bg: '#e3f2fd', 
    border: '#1976d2', 
    titleColor: '#1565c0',
    icon: 'ℹ️'
  },
  success: { 
    bg: '#e8f5e9', 
    border: '#2e7d32', 
    titleColor: '#1b5e20',
    icon: '✓'
  },
  warning: { 
    bg: '#fff3e0', 
    border: '#ff9800', 
    titleColor: '#e65100',
    icon: '⚠️'
  },
  danger: { 
    bg: '#ffebee', 
    border: '#c62828', 
    titleColor: '#b71c1c',
    icon: '✗'
  },
};

/**
 * Callout component - highlights important information with title and icon
 */
export const Callout: React.FC<CalloutProps> = ({ 
  title,
  children,
  variant = 'info',
  icon,
  padding = 24,
  borderRadius = 8,
  borderWidth = 3,
  ...props 
}) => {
  const config = variantConfig[variant];
  const displayIcon = icon !== undefined ? icon : config.icon;
  
  return React.createElement('Box', {
    backgroundColor: config.bg,
    borderColor: config.border,
    borderWidth,
    borderRadius,
    padding,
    ...props
  },
    title && React.createElement('Row', { gap: 8, marginBottom: 12, alignItems: 'center' },
      displayIcon && React.createElement('Text', { fontSize: 18 }, displayIcon),
      React.createElement('Text', {
        fontSize: 16,
        fontWeight: 'bold',
        color: config.titleColor
      }, title)
    ),
    children
  );
};
