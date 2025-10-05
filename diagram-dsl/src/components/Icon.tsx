import React from 'react';
import type { BoxProps } from '../types';

export interface IconProps extends Partial<BoxProps> {
  symbol: string;
  size?: 'small' | 'medium' | 'large' | 'xlarge';
  color?: string;
  backgroundColor?: string;
  circular?: boolean;
}

const sizeMap = {
  small: { fontSize: 16, padding: 4, circle: 24 },
  medium: { fontSize: 24, padding: 6, circle: 36 },
  large: { fontSize: 32, padding: 8, circle: 48 },
  xlarge: { fontSize: 48, padding: 12, circle: 72 },
};

/**
 * Icon component - displays emoji or unicode symbols with optional circular background
 * Perfect for visual markers, status indicators, and decorative elements
 */
export const Icon: React.FC<IconProps> = ({
  symbol,
  size = 'medium',
  color,
  backgroundColor,
  circular = false,
  ...props
}) => {
  const sizing = sizeMap[size];
  
  if (circular) {
    return React.createElement('Box', {
      width: sizing.circle,
      height: sizing.circle,
      backgroundColor: backgroundColor || '#f5f5f5',
      borderRadius: sizing.circle / 2,
      justifyContent: 'center',
      alignItems: 'center',
      ...props
    },
      React.createElement('Text', {
        fontSize: sizing.fontSize,
        color: color || '#333',
        children: symbol
      })
    );
  }
  
  return React.createElement('Text', {
    fontSize: sizing.fontSize,
    color: color || '#333',
    ...props,
    children: symbol
  });
};
