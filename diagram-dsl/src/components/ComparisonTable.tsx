import React from 'react';
import type { BoxProps } from '../types';

export interface TableColumn {
  header: string;
  key: string;
  width?: number;
  align?: 'left' | 'center' | 'right';
}

export interface TableRow {
  [key: string]: string | boolean;
}

export interface ComparisonTableProps extends Partial<BoxProps> {
  columns: TableColumn[];
  rows: TableRow[];
  highlightColumn?: string;
  striped?: boolean;
}

/**
 * ComparisonTable component - displays data in tabular format
 * Perfect for feature comparisons, benchmarks, specifications
 */
export const ComparisonTable: React.FC<ComparisonTableProps> = ({
  columns,
  rows,
  highlightColumn,
  striped = true,
  width = 800,
  ...props
}) => {
  const tableWidth = typeof width === 'number' ? width : 800;
  const defaultColumnWidth = tableWidth / columns.length;
  
  return React.createElement('Stack', {
    width: tableWidth,
    borderColor: '#e0e0e0',
    borderWidth: 1,
    borderRadius: 8,
    gap: 0,
    ...props
  },
    // Header row
    React.createElement('Row', {
      backgroundColor: '#f5f5f5',
      gap: 0,
      borderTopLeftRadius: 8,
      borderTopRightRadius: 8
    },
      ...columns.map((col, index) =>
        React.createElement('Box', {
          key: col.key,
          width: col.width || defaultColumnWidth,
          padding: 12,
          borderRightWidth: index < columns.length - 1 ? 1 : 0,
          borderRightColor: '#e0e0e0',
          backgroundColor: col.key === highlightColumn ? '#e3f2fd' : 'transparent'
        },
          React.createElement('Text', {
            fontSize: 13,
            fontWeight: 'bold',
            textAlign: col.align || 'left',
            children: col.header
          })
        )
      )
    ),
    
    // Data rows
    ...rows.map((row, rowIndex) =>
      React.createElement('Row', {
        key: rowIndex,
        backgroundColor: striped && rowIndex % 2 === 0 ? '#fafafa' : 'white',
        gap: 0,
        borderTopWidth: 1,
        borderTopColor: '#e0e0e0'
      },
        ...columns.map((col, colIndex) => {
          const value = row[col.key];
          const displayValue = typeof value === 'boolean' 
            ? (value ? '✓' : '✗')
            : String(value);
          const textColor = typeof value === 'boolean'
            ? (value ? '#2e7d32' : '#c62828')
            : '#333';
          
          return React.createElement('Box', {
            key: `${rowIndex}-${col.key}`,
            width: col.width || defaultColumnWidth,
            padding: 12,
            borderRightWidth: colIndex < columns.length - 1 ? 1 : 0,
            borderRightColor: '#e0e0e0',
            backgroundColor: col.key === highlightColumn ? '#f0f7ff' : 'transparent'
          },
            React.createElement('Text', {
              fontSize: 12,
              textAlign: col.align || 'left',
              color: textColor,
              fontWeight: typeof value === 'boolean' ? 'bold' : 'normal',
              children: displayValue
            })
          );
        })
      )
    )
  );
};
