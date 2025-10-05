import React from 'react';
import { StackProps } from '../types';

export const Stack: React.FC<StackProps> = ({ children, direction = 'vertical', ...props }) => {
  return React.createElement('Stack', { ...props, direction }, children);
};
