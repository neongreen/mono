/**
 * Approach 1: Flexbox-based Constraint System
 * 
 * This approach uses CSS Flexbox with smart defaults and constraints.
 * Features:
 * - Automatic text wrapping and overflow prevention
 * - Intelligent spacing based on content
 * - Constraint-based sizing (min/max/preferred)
 * - No manual pixel calculations needed
 */

import React from 'react';
import { StackProps, CardProps, SpaceProps } from './types';

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
    minWidth: constraints.minWidth,
    maxWidth: constraints.maxWidth,
    minHeight: constraints.minHeight,
    maxHeight: constraints.maxHeight,
    width: constraints.preferredWidth,
    height: constraints.preferredHeight,
    boxSizing: 'border-box',
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

  const style: React.CSSProperties = {
    backgroundColor: colors.bg,
    border: `2px solid ${colors.border}`,
    borderRadius: '8px',
    padding: '16px',
    minWidth: constraints.minWidth || 150,
    maxWidth: constraints.maxWidth || 400,
    minHeight: constraints.minHeight || 60,
    maxHeight: constraints.maxHeight,
    width: constraints.preferredWidth,
    height: constraints.preferredHeight,
    boxSizing: 'border-box',
    wordWrap: 'break-word',
    overflowWrap: 'break-word',
    hyphens: 'auto',
  };

  return (
    <div style={style}>
      {title && (
        <div style={{ fontWeight: 'bold', marginBottom: '8px', fontSize: '16px' }}>
          {title}
        </div>
      )}
      {content && <div style={{ fontSize: '14px' }}>{content}</div>}
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
  const style: React.CSSProperties = {
    flex: `${grow} ${shrink} ${basis === 'auto' ? 'auto' : `${basis}px`}`,
    minWidth: constraints.minWidth,
    maxWidth: constraints.maxWidth,
    minHeight: constraints.minHeight,
    maxHeight: constraints.maxHeight,
    width: constraints.preferredWidth,
    height: constraints.preferredHeight,
    boxSizing: 'border-box',
    display: 'flex',
    flexDirection: 'column',
  };

  return <div style={style}>{children}</div>;
};
