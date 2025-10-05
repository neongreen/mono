import React from 'react';
import { BoxProps } from '../types';

export const Row: React.FC<BoxProps> = ({ children, ...props }) => {
  return React.createElement('Row', props, children);
};
