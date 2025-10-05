import React from 'react';
import type { BoxProps } from '../types';

export interface ClusterProps extends Partial<BoxProps> {
  title?: string;
  children?: React.ReactNode;
  variant?: 'default' | 'primary' | 'secondary' | 'accent' | 'success' | 'warning' | 'danger';
  dashed?: boolean;
}

const variantColors = {
  default: { border: '#999', bg: '#fafafa', titleBg: '#f5f5f5', text: '#333' },
  primary: { border: '#1976d2', bg: '#f0f7ff', titleBg: '#e3f2fd', text: '#1565c0' },
  secondary: { border: '#7b1fa2', bg: '#f9f3fb', titleBg: '#f3e5f5', text: '#6a1b9a' },
  accent: { border: '#f57c00', bg: '#fff8f0', titleBg: '#fff3e0', text: '#e65100' },
  success: { border: '#2e7d32', bg: '#f1f8f1', titleBg: '#e8f5e9', text: '#1b5e20' },
  warning: { border: '#ff9800', bg: '#fff8f0', titleBg: '#fff3e0', text: '#e65100' },
  danger: { border: '#c62828', bg: '#fff0f0', titleBg: '#ffebee', text: '#b71c1c' },
};

/**
 * Cluster component - groups related elements with optional title
 * Perfect for showing logical groupings in diagrams
 */
export const Cluster: React.FC<ClusterProps> = ({
  title,
  children,
  variant = 'default',
  dashed = false,
  padding = 24,
  borderRadius = 8,
  borderWidth = 2,
  ...props
}) => {
  const colors = variantColors[variant];
  
  return React.createElement('Stack', {
    backgroundColor: colors.bg,
    borderColor: colors.border,
    borderWidth,
    borderRadius,
    padding: title ? 0 : padding,
    ...props
  },
    title && React.createElement('Box', {
      backgroundColor: colors.titleBg,
      padding: 12,
      borderTopLeftRadius: borderRadius,
      borderTopRightRadius: borderRadius,
      borderBottomWidth: 1,
      borderBottomColor: colors.border
    },
      React.createElement('Text', {
        fontSize: 14,
        fontWeight: 'bold',
        color: colors.text
      }, title)
    ),
    title ? React.createElement('Box', { padding }, children) : children
  );
};
