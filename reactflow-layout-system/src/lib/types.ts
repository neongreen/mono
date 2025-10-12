import { ReactNode } from 'react';

export type Direction = 'horizontal' | 'vertical';
export type Alignment = 'start' | 'center' | 'end' | 'stretch';
export type Distribution = 'start' | 'center' | 'end' | 'space-between' | 'space-around' | 'space-evenly';

export interface LayoutConstraints {
  minWidth?: number;
  maxWidth?: number;
  minHeight?: number;
  maxHeight?: number;
  preferredWidth?: number;
  preferredHeight?: number;
}

export interface StackProps {
  direction?: Direction;
  gap?: number;
  align?: Alignment;
  distribute?: Distribution;
  padding?: number;
  constraints?: LayoutConstraints;
  children: ReactNode;
}

export interface CardProps {
  title?: string;
  content?: ReactNode;
  constraints?: LayoutConstraints;
  variant?: 'default' | 'primary' | 'secondary' | 'success' | 'warning';
}

export interface SpaceProps {
  grow?: number;
  shrink?: number;
  basis?: number | 'auto';
  constraints?: LayoutConstraints;
  children: ReactNode;
}
