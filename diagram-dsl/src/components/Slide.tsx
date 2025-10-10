import React from 'react';
import type { BoxProps } from '../types';

export interface SlideProps extends Partial<BoxProps> {
  children?: React.ReactNode;
  title?: string;
  subtitle?: string;
}

/**
 * Slide component - provides a consistent container for presentation slides
 * Default: 1200x800 with padding
 */
export const Slide: React.FC<SlideProps> = ({ 
  children, 
  title,
  subtitle,
  width = 1200, 
  height = 800,
  padding = 60,
  gap = 32,
  backgroundColor = 'white',
  ...props 
}) => {
  return React.createElement('Stack', {
    gap,
    padding,
    width,
    height,
    backgroundColor,
    ...props
  }, children);
};
