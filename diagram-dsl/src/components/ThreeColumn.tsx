import React from 'react';
import type { BoxProps } from '../types';

export interface ThreeColumnProps {
  left: React.ReactNode;
  center: React.ReactNode;
  right: React.ReactNode;
  gap?: number;
  width?: number;
  height?: number;
  padding?: number;
  margin?: number;
  marginTop?: number;
  marginBottom?: number;
}

/**
 * ThreeColumn component - creates a three-column layout with equal widths
 */
export const ThreeColumn: React.FC<ThreeColumnProps> = ({
  left,
  center,
  right,
  gap = 24,
  ...props
}) => {
  return React.createElement('Row', { gap, ...props },
    React.createElement('Stack', { flexGrow: 1 }, left),
    React.createElement('Stack', { flexGrow: 1 }, center),
    React.createElement('Stack', { flexGrow: 1 }, right)
  );
};
