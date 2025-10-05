import React from 'react';
import { Text } from './Text';
import { theme } from '../theme';
import { TextProps } from '../types';

/**
 * Subtitle component - Smaller, gray text for secondary information
 * Perfect for subtitles, descriptions, or secondary labels
 */
export interface SubtitleProps extends Omit<TextProps, 'children' | 'fontSize' | 'color'> {
  children: string;
  size?: 'sm' | 'base';  // Size variant
}

export const Subtitle: React.FC<SubtitleProps> = ({ children, size = 'sm', textAlign = 'center', ...props }) => {
  const fontSize = size === 'sm' 
    ? theme.typography.fontSize.sm   // 12
    : theme.typography.fontSize.base; // 14

  return (
    <Text
      fontSize={fontSize}
      fontWeight="normal"
      color={theme.colors.text.secondary}
      textAlign={textAlign}
      {...props}
    >
      {children}
    </Text>
  );
};
