import React from 'react';
import type { BoxProps } from '../types';

export interface TerminalProps extends Partial<BoxProps> {
  commands: string[];
  prompt?: string;
  title?: string;
  theme?: 'dark' | 'light';
}

/**
 * Terminal component - displays command-line interface output
 * Perfect for showing CLI commands, scripts, and terminal sessions
 */
export const Terminal: React.FC<TerminalProps> = ({
  commands,
  prompt = '$',
  title = 'Terminal',
  theme = 'dark',
  width = 800,
  ...props
}) => {
  const colors = theme === 'dark'
    ? { bg: '#1e1e1e', text: '#d4d4d4', header: '#2d2d2d', prompt: '#4ec9b0', border: '#3e3e3e' }
    : { bg: '#ffffff', text: '#333333', header: '#f5f5f5', prompt: '#2e7d32', border: '#e0e0e0' };
  
  return React.createElement('Stack', {
    width,
    borderColor: colors.border,
    borderWidth: 1,
    borderRadius: 8,
    padding: 0,
    gap: 0,
    ...props
  },
    // Terminal header
    React.createElement('Row', {
      backgroundColor: colors.header,
      padding: 12,
      gap: 8,
      alignItems: 'center',
      borderTopLeftRadius: 8,
      borderTopRightRadius: 8
    },
      // Window controls
      React.createElement('Row', { gap: 6 },
        React.createElement('Box', {
          width: 12,
          height: 12,
          borderRadius: 6,
          backgroundColor: '#ff5f56'
        }),
        React.createElement('Box', {
          width: 12,
          height: 12,
          borderRadius: 6,
          backgroundColor: '#ffbd2e'
        }),
        React.createElement('Box', {
          width: 12,
          height: 12,
          borderRadius: 6,
          backgroundColor: '#27c93f'
        })
      ),
      React.createElement('Text', {
        fontSize: 12,
        color: colors.text,
        marginLeft: 8,
        children: title
      })
    ),
    
    // Terminal content
    React.createElement('Stack', {
      backgroundColor: colors.bg,
      padding: 16,
      gap: 4,
      borderBottomLeftRadius: 8,
      borderBottomRightRadius: 8
    },
      ...commands.map((command, index) =>
        React.createElement('Row', {
          key: index,
          gap: 8,
          alignItems: 'flex-start'
        },
          React.createElement('Text', {
            fontSize: 13,
            fontFamily: 'Monaco, Courier, monospace',
            color: colors.prompt,
            children: prompt
          }),
          React.createElement('Text', {
            fontSize: 13,
            fontFamily: 'Monaco, Courier, monospace',
            color: colors.text,
            children: command
          })
        )
      )
    )
  );
};
