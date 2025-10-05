import React from 'react';
import { TimelineEventProps } from '../types';

export const TimelineEvent: React.FC<TimelineEventProps> = (props) => {
  return React.createElement('TimelineEvent', props, null);
};
