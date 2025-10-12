/**
 * Approach 2: Grid-based Proportional System
 * 
 * This approach uses CSS Grid with proportional units (fr).
 * Features:
 * - Automatic space distribution using fr units
 * - Grid-aware layout with explicit template areas
 * - Better for complex multi-dimensional layouts
 * - Smart defaults with 12-column grid system
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
  const childCount = React.Children.count(children);
  
  // Calculate grid template based on grow values for horizontal stacks
  let gridTemplateColumns = '1fr';
  if (direction === 'horizontal') {
    const growValues: number[] = [];
    React.Children.forEach(children, (child) => {
      if (React.isValidElement(child) && child.props.grow !== undefined) {
        growValues.push(child.props.grow);
      } else {
        growValues.push(1); // Default grow value
      }
    });
    
    // Create fr units based on actual grow values
    if (growValues.length > 0) {
      gridTemplateColumns = growValues.map(v => `${v}fr`).join(' ');
    } else {
      gridTemplateColumns = `repeat(${childCount}, 1fr)`;
    }
  }
  
  const style: React.CSSProperties = {
    display: 'grid',
    gridTemplateColumns,
    gridTemplateRows: direction === 'vertical' ? `repeat(${childCount}, auto)` : 'auto',
    gap: `${gap}px`,
    alignItems: align === 'start' ? 'start' : align === 'end' ? 'end' : align,
    justifyContent: distribute === 'start' ? 'start' : distribute === 'end' ? 'end' : distribute,
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
    display: 'grid',
    gridTemplateRows: title ? 'auto 1fr' : '1fr',
    gap: '8px',
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
        <div style={{ fontWeight: 'bold', fontSize: '16px', alignSelf: 'start' }}>
          {title}
        </div>
      )}
      {content && <div style={{ fontSize: '14px', alignSelf: 'start' }}>{content}</div>}
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
    // Don't use gridColumn span since the parent grid template
    // is already sized based on grow values
    minWidth: constraints.minWidth,
    maxWidth: constraints.maxWidth,
    minHeight: constraints.minHeight,
    maxHeight: constraints.maxHeight,
    width: constraints.preferredWidth,
    height: constraints.preferredHeight,
    boxSizing: 'border-box',
    display: 'grid',
  };

  return <div style={style}>{children}</div>;
};
