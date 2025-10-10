import React from 'react';
import type { BoxProps } from '../types';

export interface TwoColumnProps {
  left: React.ReactNode;
  right: React.ReactNode;
  leftWidth?: number | string;
  rightWidth?: number | string;
  gap?: number;
  width?: number;
  height?: number;
  padding?: number;
  margin?: number;
  marginTop?: number;
  marginBottom?: number;
  marginLeft?: number;
  marginRight?: number;
}

/**
 * TwoColumn component - creates a two-column layout with flexible widths
 */
export const TwoColumn: React.FC<TwoColumnProps> = ({
  left,
  right,
  leftWidth,
  rightWidth,
  gap = 24,
  ...props
}) => {
  return React.createElement('Row', { gap, ...props },
    React.createElement('Stack', {
      width: leftWidth,
      flexGrow: leftWidth === undefined ? 1 : undefined
    }, left),
    React.createElement('Stack', {
      width: rightWidth,
      flexGrow: rightWidth === undefined ? 1 : undefined
    }, right)
  );
};
