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
  borderStyle?: 'solid' | 'dashed' | 'dotted';
  borderDashArray?: string; // e.g., "6 4" for 6px dash, 4px gap
  flexGrow?: number; // Allow boxes to grow in flex containers
  flexShrink?: number;
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
  labelPosition?: 'start' | 'middle' | 'end';
  startLabel?: string;
  endLabel?: string;
  style?: 'solid' | 'dashed' | 'dotted' | 'wave';
  curve?: 'straight' | 'curved' | 'step' | 'arc';
  headType?: 'arrow' | 'none' | 'circle' | 'diamond' | 'square';
  tailType?: 'none' | 'arrow' | 'circle' | 'diamond' | 'square';
  animated?: boolean;
  bidirectional?: boolean;
  thickness?: 'thin' | 'medium' | 'thick' | 'very-thick';
  // Advanced arrow features for complex diagrams
  fromSide?: 'top' | 'bottom' | 'left' | 'right' | 'auto';
  toSide?: 'top' | 'bottom' | 'left' | 'right' | 'auto';
  fromOffset?: number; // Offset from center along the edge (0 = center, positive/negative for offset)
  toOffset?: number;
  shortenStart?: number; // Shorten arrow from start point (in pixels)
  shortenEnd?: number; // Shorten arrow from end point (in pixels)
  // For Y-shaped forks - split into multiple endpoints
  toMultiple?: string[]; // Array of target IDs for forked arrows
}

export interface ImageProps extends LayoutProps {
  src: string;
  alt?: string;
  fit?: 'contain' | 'cover' | 'fill' | 'none';
  borderRadius?: number;
  opacity?: number;
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

// Agent Loop & State Machine Components
export interface StateNodeProps extends BoxProps {
  label: string;
  stateType?: 'initial' | 'active' | 'final' | 'default';
  icon?: string;
}

export interface TransitionProps extends ArrowProps {
  condition?: string;
}

export interface DecisionNodeProps extends BoxProps {
  label: string;
  shape?: 'diamond' | 'hexagon';
}

export interface LoopIndicatorProps {
  target: string;
  iterations?: number | string;
  color?: string;
}

// Timeline Components
export interface TimelineProps extends LayoutProps {
  orientation?: 'horizontal' | 'vertical';
  showAxis?: boolean;
}

export interface TimelineEventProps extends LayoutProps {
  time: string | number;
  label: string;
  description?: string;
  color?: string;
  icon?: string;
}

export interface TimelineRangeProps {
  start: string | number;
  end: string | number;
  label: string;
  color?: string;
  opacity?: number;
}

// Memory & Storage Components
export interface MemoryBlockProps extends BoxProps {
  label: string;
  capacity: number;
  used: number;
  unit?: string;
  showBar?: boolean;
  showPercentage?: boolean;
}

export interface StackVisualizationProps extends LayoutProps {
  items: string[];
  direction?: 'vertical' | 'horizontal';
  highlightTop?: boolean;
  highlightBottom?: boolean;
  label?: string;
}

// Context Engineering Specific
export interface ContextWindowProps extends LayoutProps {
  capacity: number;
  sections: Array<{
    label: string;
    tokens: number;
    color: string;
  }>;
  showLabels?: boolean;
  showPercentages?: boolean;
  orientation?: 'horizontal' | 'vertical';
}

export interface TokenBudgetProps extends LayoutProps {
  total: number;
  allocations: Array<{
    label: string;
    value: number;
    color?: string;
  }>;
  showRemaining?: boolean;
}

// Process & Flow Components
export interface ProcessNodeProps extends BoxProps {
  label: string;
  nodeType?: 'process' | 'data' | 'decision' | 'start' | 'end' | 'subprocess';
  status?: 'pending' | 'active' | 'complete' | 'error';
}

export interface DataTransformProps {
  from: string;
  to: string;
  transformation: string;
  showPreview?: boolean;
  inputData?: any;
  outputData?: any;
}
