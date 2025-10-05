import React from 'react';
import type { BoxProps } from '../types';

export interface PanelProps extends Partial<BoxProps> {
  children?: React.ReactNode;
  header?: React.ReactNode;
  footer?: React.ReactNode;
  variant?: 'default' | 'primary' | 'secondary' | 'success' | 'warning' | 'danger';
  elevation?: 0 | 1 | 2 | 3;
}

const variantColors = {
  default: { bg: 'white', border: '#e0e0e0', headerBg: '#f5f5f5' },
  primary: { bg: 'white', border: '#1976d2', headerBg: '#e3f2fd' },
  secondary: { bg: 'white', border: '#7b1fa2', headerBg: '#f3e5f5' },
  success: { bg: 'white', border: '#2e7d32', headerBg: '#e8f5e9' },
  warning: { bg: 'white', border: '#ff9800', headerBg: '#fff3e0' },
  danger: { bg: 'white', border: '#c62828', headerBg: '#ffebee' },
};

/**
 * Panel component - a card-like container with optional header and footer
 * Supports elevation (shadow depth) for visual hierarchy
 */
export const Panel: React.FC<PanelProps> = ({
  children,
  header,
  footer,
  variant = 'default',
  elevation = 1,
  padding = 0,
  borderRadius = 8,
  borderWidth = 1,
  ...props
}) => {
  const colors = variantColors[variant];
  
  // Note: SVG doesn't support box-shadow, so elevation is represented through border
  const elevationBorder = elevation > 0 ? borderWidth + elevation : borderWidth;
  
  return React.createElement('Stack', {
    backgroundColor: colors.bg,
    borderColor: colors.border,
    borderWidth: elevationBorder,
    borderRadius,
    padding: 0,
    gap: 0,
    ...props
  },
    header && React.createElement('Box', {
      backgroundColor: colors.headerBg,
      padding: 16,
      borderBottomWidth: 1,
      borderBottomColor: colors.border,
      borderTopLeftRadius: borderRadius,
      borderTopRightRadius: borderRadius
    }, header),
    React.createElement('Box', {
      padding: padding || 20,
      flexGrow: 1
    }, children),
    footer && React.createElement('Box', {
      backgroundColor: colors.headerBg,
      padding: 16,
      borderTopWidth: 1,
      borderTopColor: colors.border,
      borderBottomLeftRadius: borderRadius,
      borderBottomRightRadius: borderRadius
    }, footer)
  );
};
