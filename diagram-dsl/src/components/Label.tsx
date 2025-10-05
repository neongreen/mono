import React from 'react';
import { Text } from './Text';
import { theme } from '../theme';
import { TextProps } from '../types';

/**
 * Label component - Regular text with sensible defaults
 * For body text and standard labels
 */
export interface LabelProps extends Omit<TextProps, 'children' | 'fontSize'> {
  children: string;
  size?: 'sm' | 'base' | 'lg';  // Size variant
  bold?: boolean;
}

export const Label: React.FC<LabelProps> = ({ 
  children, 
  size = 'base', 
  bold = false,
  color,
  textAlign = 'center',
  ...props 
}) => {
  const fontSize = size === 'sm' 
    ? theme.typography.fontSize.sm    // 12
    : size === 'lg'
    ? theme.typography.fontSize.lg    // 16
    : theme.typography.fontSize.base; // 14

  return (
    <Text
      fontSize={fontSize}
      fontWeight={bold ? 'bold' : 'normal'}
      color={color || theme.colors.text.primary}
      textAlign={textAlign}
      {...props}
    >
      {children}
    </Text>
  );
};
