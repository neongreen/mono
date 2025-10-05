import React from 'react';
import { Box } from './Box';
import { Stack } from './Stack';
import { theme } from '../theme';
import { BoxProps } from '../types';

/**
 * Card component - A styled Box with professional defaults
 * Includes background color, border, border radius, and padding
 */
export interface CardProps extends Omit<BoxProps, 'backgroundColor' | 'borderColor' | 'borderWidth' | 'borderRadius' | 'padding'> {
  children?: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'info' | 'default';
  padding?: number;
  borderRadius?: number;
}

export const Card: React.FC<CardProps> = ({ 
  children, 
  variant = 'default',
  padding = theme.spacing.lg,  // 16px default
  borderRadius = theme.borderRadius.md,  // 8px default
  justifyContent = 'center',
  alignItems = 'center',
  width,
  height,
  ...props 
}) => {
  // Get colors based on variant
  const getColors = () => {
    switch (variant) {
      case 'primary':
        return { bg: theme.colors.primary.light, border: theme.colors.primary.dark };
      case 'secondary':
        return { bg: theme.colors.secondary.light, border: theme.colors.secondary.dark };
      case 'success':
        return { bg: theme.colors.success.light, border: theme.colors.success.dark };
      case 'warning':
        return { bg: theme.colors.warning.light, border: theme.colors.warning.dark };
      case 'error':
        return { bg: theme.colors.error.light, border: theme.colors.error.dark };
      case 'info':
        return { bg: theme.colors.info.light, border: theme.colors.info.dark };
      default:
        return { bg: theme.colors.gray[50], border: theme.colors.gray[300] };
    }
  };

  const colors = getColors();

  return (
    <Box
      width={width}
      height={height}
      backgroundColor={colors.bg}
      borderColor={colors.border}
      borderWidth={theme.border.width.normal}  // 2px
      borderRadius={borderRadius}
      padding={padding}
      justifyContent={justifyContent}
      alignItems={alignItems}
      {...props}
    >
      {children}
    </Box>
  );
};
