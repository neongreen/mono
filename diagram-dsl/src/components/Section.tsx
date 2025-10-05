import React from 'react';
import type { BoxProps } from '../types';

export interface SectionProps extends Partial<BoxProps> {
  title: string;
  children?: React.ReactNode;
  titleSize?: number;
  titleColor?: string;
  variant?: 'default' | 'primary' | 'secondary' | 'accent' | 'success' | 'warning' | 'danger';
}

const variantColors = {
  default: { bg: '#f5f5f5', border: '#999', text: '#333' },
  primary: { bg: '#e3f2fd', border: '#1976d2', text: '#1565c0' },
  secondary: { bg: '#f3e5f5', border: '#7b1fa2', text: '#6a1b9a' },
  accent: { bg: '#fff3e0', border: '#f57c00', text: '#e65100' },
  success: { bg: '#e8f5e9', border: '#2e7d32', text: '#1b5e20' },
  warning: { bg: '#fff3e0', border: '#ff9800', text: '#e65100' },
  danger: { bg: '#ffebee', border: '#c62828', text: '#b71c1c' },
};

/**
 * Section component - a titled box for organizing content
 */
export const Section: React.FC<SectionProps> = ({ 
  title,
  children,
  titleSize = 14,
  titleColor,
  variant = 'default',
  padding = 20,
  borderRadius = 8,
  borderWidth = 2,
  ...props 
}) => {
  const colors = variantColors[variant];
  
  return React.createElement('Box', {
    backgroundColor: colors.bg,
    borderColor: colors.border,
    borderWidth,
    borderRadius,
    padding,
    ...props
  },
    React.createElement('Text', {
      fontSize: titleSize,
      fontWeight: 'bold',
      marginBottom: 12,
      color: titleColor || colors.text
    }, title),
    children
  );
};
