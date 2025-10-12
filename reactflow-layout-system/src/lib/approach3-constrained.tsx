/**
 * Approach 3: Constrained Absolute Positioning
 * 
 * This approach uses absolute positioning with intelligent constraint resolution.
 * Features:
 * - Explicit control over positioning when needed
 * - Automatic constraint satisfaction (text won't overflow)
 * - Smart defaults for aspect ratios and spacing
 * - Calculates optimal sizes based on content and constraints
 */

import React from 'react';
import { StackProps, CardProps, SpaceProps } from './types';

// Helper to calculate optimal dimensions based on constraints
const calculateOptimalSize = (
  contentSize: number,
  min: number | undefined,
  max: number | undefined,
  preferred: number | undefined
): number | undefined => {
  if (preferred !== undefined) return Math.max(min || 0, Math.min(max || Infinity, preferred));
  if (min !== undefined && max !== undefined) return (min + max) / 2;
  if (min !== undefined) return Math.max(min, contentSize);
  if (max !== undefined) return Math.min(max, contentSize);
  return undefined;
};

export const Stack: React.FC<StackProps> = ({
  direction = 'vertical',
  gap = 16,
  align = 'stretch',
  distribute = 'start',
  padding = 0,
  constraints = {},
  children,
}) => {
  const style: React.CSSProperties = {
    display: 'flex',
    flexDirection: direction === 'horizontal' ? 'row' : 'column',
    gap: `${gap}px`,
    alignItems: align === 'start' ? 'flex-start' : align === 'end' ? 'flex-end' : align,
    justifyContent: distribute === 'start' ? 'flex-start' : distribute === 'end' ? 'flex-end' : distribute,
    padding: `${padding}px`,
    minWidth: constraints.minWidth || 100,
    maxWidth: constraints.maxWidth || 'none',
    minHeight: constraints.minHeight || 50,
    maxHeight: constraints.maxHeight || 'none',
    width: calculateOptimalSize(300, constraints.minWidth, constraints.maxWidth, constraints.preferredWidth),
    height: calculateOptimalSize(200, constraints.minHeight, constraints.maxHeight, constraints.preferredHeight),
    boxSizing: 'border-box',
    position: 'relative',
  };

  return <div style={style}>{children}</div>;
};

export const Card: React.FC<CardProps> = ({
  title,
  content,
  constraints = {},
  variant = 'default',
}) => {
  const variants = {
    default: { bg: '#f5f5f5', border: '#e0e0e0' },
    primary: { bg: '#e3f2fd', border: '#2196f3' },
    secondary: { bg: '#f3e5f5', border: '#9c27b0' },
    success: { bg: '#e8f5e9', border: '#4caf50' },
    warning: { bg: '#fff3e0', border: '#ff9800' },
  };

  const colors = variants[variant];

  // Calculate optimal dimensions with sensible defaults
  const optimalWidth = calculateOptimalSize(200, constraints.minWidth || 150, constraints.maxWidth || 400, constraints.preferredWidth);
  const optimalHeight = calculateOptimalSize(100, constraints.minHeight || 60, constraints.maxHeight, constraints.preferredHeight);

  const style: React.CSSProperties = {
    backgroundColor: colors.bg,
    border: `2px solid ${colors.border}`,
    borderRadius: '8px',
    padding: '16px',
    minWidth: constraints.minWidth || 150,
    maxWidth: constraints.maxWidth || 400,
    minHeight: constraints.minHeight || 60,
    maxHeight: constraints.maxHeight || 'none',
    width: optimalWidth,
    height: optimalHeight,
    boxSizing: 'border-box',
    wordWrap: 'break-word',
    overflowWrap: 'break-word',
    hyphens: 'auto',
    overflow: 'hidden',
    display: 'flex',
    flexDirection: 'column',
    gap: '8px',
  };

  return (
    <div style={style}>
      {title && (
        <div style={{ fontWeight: 'bold', fontSize: '16px', flexShrink: 0 }}>
          {title}
        </div>
      )}
      {content && (
        <div style={{ fontSize: '14px', flex: '1 1 auto', overflow: 'auto' }}>
          {content}
        </div>
      )}
    </div>
  );
};

export const Space: React.FC<SpaceProps> = ({
  grow = 1,
  shrink = 1,
  basis = 'auto',
  constraints = {},
  children,
}) => {
  const optimalWidth = calculateOptimalSize(200, constraints.minWidth, constraints.maxWidth, constraints.preferredWidth);
  const optimalHeight = calculateOptimalSize(100, constraints.minHeight, constraints.maxHeight, constraints.preferredHeight);

  const style: React.CSSProperties = {
    flex: `${grow} ${shrink} ${basis === 'auto' ? 'auto' : `${basis}px`}`,
    minWidth: constraints.minWidth || 50,
    maxWidth: constraints.maxWidth || 'none',
    minHeight: constraints.minHeight || 50,
    maxHeight: constraints.maxHeight || 'none',
    width: optimalWidth,
    height: optimalHeight,
    boxSizing: 'border-box',
    display: 'flex',
    flexDirection: 'column',
    position: 'relative',
  };

  return <div style={style}>{children}</div>;
};
