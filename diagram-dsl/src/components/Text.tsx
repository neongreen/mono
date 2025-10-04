import React from 'react';
import { TextProps } from '../types';

export const Text: React.FC<TextProps> = ({ children, ...props }) => {
  return React.createElement('Text', props, children);
};
