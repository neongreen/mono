import React from 'react';
import { BoxProps } from '../types';

export const Column: React.FC<BoxProps> = ({ children, ...props }) => {
  return React.createElement('Column', props, children);
};
