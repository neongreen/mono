import React from 'react';
import type { BoxProps } from '../types';

export interface BadgeProps extends Partial<BoxProps> {
  text: string;
  variant?: 'primary' | 'secondary' | 'accent' | 'success' | 'warning' | 'danger' | 'info' | 'default';
  size?: 'small' | 'medium' | 'large';
}

const variantColors = {
  primary: { bg: '#1976d2', text: '#fff' },
  secondary: { bg: '#7b1fa2', text: '#fff' },
  accent: { bg: '#f57c00', text: '#fff' },
  success: { bg: '#2e7d32', text: '#fff' },
  warning: { bg: '#ff9800', text: '#fff' },
  danger: { bg: '#c62828', text: '#fff' },
  info: { bg: '#0288d1', text: '#fff' },
  default: { bg: '#666', text: '#fff' },
};

const sizeConfig = {
  small: { fontSize: 10, padding: 4 },
  medium: { fontSize: 12, padding: 6 },
  large: { fontSize: 14, padding: 8 },
};

/**
 * Badge component - displays a small label or tag
 */
export const Badge: React.FC<BadgeProps> = ({
  text,
  variant = 'default',
  size = 'medium',
  borderRadius = 4,
  ...props
}) => {
  const colors = variantColors[variant];
  const sizing = sizeConfig[size];
  
  return React.createElement('Box', {
    backgroundColor: colors.bg,
    borderRadius,
    padding: sizing.padding,
    paddingLeft: sizing.padding * 2,
    paddingRight: sizing.padding * 2,
    justifyContent: 'center',
    alignItems: 'center',
    ...props
  },
    React.createElement('Text', {
      fontSize: sizing.fontSize,
      color: colors.text,
      fontWeight: 'bold'
    }, text)
  );
};
