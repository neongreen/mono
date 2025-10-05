import { ReactNode } from 'react';

export interface LayoutProps {
  width?: number | 'auto';
  height?: number | 'auto';
  minWidth?: number;
  minHeight?: number;
  maxWidth?: number;
  maxHeight?: number;
  padding?: number;
  paddingTop?: number;
  paddingBottom?: number;
  paddingLeft?: number;
  paddingRight?: number;
  margin?: number;
  marginTop?: number;
  marginBottom?: number;
  marginLeft?: number;
  marginRight?: number;
  gap?: number;
}

export interface AlignmentProps {
  alignItems?: 'flex-start' | 'center' | 'flex-end' | 'stretch';
  justifyContent?: 'flex-start' | 'center' | 'flex-end' | 'space-between' | 'space-around' | 'space-evenly';
}

export interface PositionProps {
  position?: 'relative' | 'absolute';
  top?: number;
  bottom?: number;
  left?: number;
  right?: number;
}

export interface BoxProps extends LayoutProps, AlignmentProps, PositionProps {
  children?: ReactNode;
  backgroundColor?: string;
  borderColor?: string;
  borderWidth?: number;
  borderRadius?: number;
  id?: string;
}

export interface StackProps extends BoxProps {
  direction?: 'vertical' | 'horizontal';
}

export interface TextProps extends LayoutProps {
  children: string;
  fontSize?: number;
  fontFamily?: string;
  color?: string;
  fontWeight?: 'normal' | 'bold';
  textAlign?: 'left' | 'center' | 'right';
  id?: string;
}

export interface ArrowProps {
  from: string;
  to: string;
  color?: string;
  strokeWidth?: number;
  label?: string;
}

export interface LayoutNode {
  type: string;
  props: any;
  children: LayoutNode[];
  computed?: {
    x: number;
    y: number;
    width: number;
    height: number;
  };
}
