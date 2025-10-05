import React, { ReactNode } from 'react';
import { TimelineProps } from '../types';

export const Timeline: React.FC<TimelineProps & { children?: ReactNode }> = (props) => {
  return React.createElement('Timeline', props, props.children);
};
