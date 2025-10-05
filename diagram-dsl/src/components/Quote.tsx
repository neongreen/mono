import React from 'react';
import type { BoxProps } from '../types';

export interface QuoteProps extends Partial<BoxProps> {
  text: string;
  author?: string;
  fontSize?: number;
  variant?: 'default' | 'primary' | 'accent';
}

const variantColors = {
  default: { border: '#999', text: '#333', author: '#666' },
  primary: { border: '#1976d2', text: '#1565c0', author: '#1976d2' },
  accent: { border: '#f57c00', text: '#e65100', author: '#f57c00' },
};

/**
 * Quote component - displays a blockquote with optional attribution
 */
export const Quote: React.FC<QuoteProps> = ({
  text,
  author,
  fontSize = 16,
  variant = 'default',
  padding = 20,
  borderRadius = 4,
  ...props
}) => {
  const colors = variantColors[variant];
  
  return React.createElement('Box', {
    backgroundColor: '#f9f9f9',
    borderColor: colors.border,
    borderWidth: 0,
    borderLeftWidth: 4,
    borderRadius,
    padding,
    ...props
  },
    React.createElement('Text', {
      fontSize,
      color: colors.text,
      fontStyle: 'italic',
      marginBottom: author ? 12 : 0
    }, `"${text}"`),
    author && React.createElement('Text', {
      fontSize: fontSize - 2,
      color: colors.author,
      textAlign: 'right'
    }, `— ${author}`)
  );
};
