import React from 'react';
import { DecisionNodeProps } from '../types';

export const DecisionNode: React.FC<DecisionNodeProps> = (props) => {
  return React.createElement('DecisionNode', props, null);
};
