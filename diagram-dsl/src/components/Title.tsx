import React from 'react';
import { Text } from './Text';
import { theme } from '../theme';
import { TextProps } from '../types';

/**
 * Title component - Large, bold text for diagram titles
 * Uses semantic styling with good defaults
 */
export interface TitleProps extends Omit<TextProps, 'children' | 'fontSize' | 'fontWeight'> {
  children: string;
  level?: 1 | 2 | 3;  // Title hierarchy: 1 = largest, 3 = smallest
}

export const Title: React.FC<TitleProps> = ({ children, level = 1, color, textAlign = 'center', ...props }) => {
  const fontSize = level === 1 
    ? theme.typography.fontSize['4xl']  // 36
    : level === 2 
    ? theme.typography.fontSize['2xl']  // 24
    : theme.typography.fontSize.xl;     // 20

  return (
    <Text
      fontSize={fontSize}
      fontWeight="bold"
      color={color || theme.colors.text.primary}
      textAlign={textAlign}
      {...props}
    >
      {children}
    </Text>
  );
};
