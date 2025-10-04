import React from 'react';
import { BoxProps } from '../types';

export const Box: React.FC<BoxProps> = ({ children, ...props }) => {
  return React.createElement('Box', props, children);
};
