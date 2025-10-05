import React from 'react';
import { StateNodeProps } from '../types';

export const StateNode: React.FC<StateNodeProps> = (props) => {
  return React.createElement('StateNode', props, null);
};
