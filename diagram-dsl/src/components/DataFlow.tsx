import React from 'react';
import type { BoxProps } from '../types';

export interface DataNode {
  id: string;
  label: string;
  type: 'input' | 'process' | 'output' | 'storage';
  description?: string;
}

export interface DataConnection {
  from: string;
  to: string;
  label?: string;
  data?: string;
}

export interface DataFlowProps extends Partial<BoxProps> {
  nodes: DataNode[];
  connections: DataConnection[];
  orientation?: 'horizontal' | 'vertical';
}

const nodeTypeStyles = {
  input: { bg: '#e3f2fd', border: '#1976d2', shape: 'rounded', icon: '📥' },
  process: { bg: '#f3e5f5', border: '#7b1fa2', shape: 'rectangle', icon: '⚙️' },
  output: { bg: '#e8f5e9', border: '#2e7d32', shape: 'rounded', icon: '📤' },
  storage: { bg: '#fff3e0', border: '#f57c00', shape: 'cylinder', icon: '💾' },
};

/**
 * DataFlow component - visualizes data processing pipelines
 * Perfect for ETL processes, data transformations, system integrations
 */
export const DataFlow: React.FC<DataFlowProps> = ({
  nodes,
  connections,
  orientation = 'horizontal',
  ...props
}) => {
  const nodeWidth = 160;
  const nodeHeight = 80;
  const gap = 80;
  
  const nodeElements = nodes.map((node, index) => {
    const style = nodeTypeStyles[node.type];
    const x = orientation === 'horizontal' ? index * (nodeWidth + gap) : 0;
    const y = orientation === 'vertical' ? index * (nodeHeight + gap) : 0;
    const borderRadius = style.shape === 'rounded' ? 40 : 8;
    
    return React.createElement('Box', {
      key: node.id,
      id: node.id,
      position: 'absolute',
      left: x,
      top: y,
      width: nodeWidth,
      height: nodeHeight,
      backgroundColor: style.bg,
      borderColor: style.border,
      borderWidth: 2,
      borderRadius,
      padding: 12,
      justifyContent: 'center',
      alignItems: 'center'
    },
      React.createElement('Text', {
        fontSize: 20,
        marginBottom: 4,
        children: style.icon
      }),
      React.createElement('Text', {
        fontSize: 13,
        fontWeight: 'bold',
        textAlign: 'center',
        children: node.label
      }),
      node.description && React.createElement('Text', {
        fontSize: 10,
        textAlign: 'center',
        color: '#666',
        marginTop: 4,
        children: node.description
      })
    );
  });
  
  const connectionElements = connections.map((conn, index) => {
    return React.createElement('Arrow', {
      key: `conn-${index}`,
      from: conn.from,
      to: conn.to,
      label: conn.data || conn.label,
      color: '#1976d2',
      strokeWidth: 2,
      curve: 'straight'
    });
  });
  
  const totalWidth = orientation === 'horizontal' 
    ? nodes.length * (nodeWidth + gap) - gap 
    : nodeWidth;
  const totalHeight = orientation === 'vertical'
    ? nodes.length * (nodeHeight + gap) - gap
    : nodeHeight;
  
  return React.createElement('Box', {
    width: totalWidth + 40,
    height: totalHeight + 40,
    padding: 20,
    position: 'relative',
    ...props
  },
    ...nodeElements,
    ...connectionElements
  );
};
