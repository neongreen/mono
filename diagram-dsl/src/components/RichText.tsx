import React from 'react';
import type { BoxProps } from '../types';

export interface TextSegment {
  text: string;
  bold?: boolean;
  color?: string;
  fontSize?: number;
}

export interface RichTextProps extends Partial<BoxProps> {
  segments: (string | TextSegment)[];
  fontSize?: number;
  color?: string;
  gap?: number;
}

/**
 * RichText component - renders text with mixed formatting in a single line
 * Note: This is a workaround until we support proper inline text formatting
 */
export const RichText: React.FC<RichTextProps> = ({ 
  segments,
  fontSize = 14,
  color = 'black',
  gap = 4,
  ...props 
}) => {
  const textElements = segments.map((segment, index) => {
    if (typeof segment === 'string') {
      return React.createElement('Text', {
        key: index,
        fontSize,
        color
      }, segment);
    } else {
      return React.createElement('Text', {
        key: index,
        fontSize: segment.fontSize || fontSize,
        color: segment.color || color,
        fontWeight: segment.bold ? 'bold' : 'normal'
      }, segment.text);
    }
  });

  return React.createElement('Row', { gap, alignItems: 'baseline', ...props }, ...textElements);
};
