import React from 'react';
import type { BoxProps } from '../types';

export interface CodeBlockProps extends Partial<BoxProps> {
  code: string | string[];
  language?: string;
  fontSize?: number;
  lineNumbers?: boolean;
}

/**
 * CodeBlock component - displays code with monospace font and optional line numbers
 */
export const CodeBlock: React.FC<CodeBlockProps> = ({
  code,
  language,
  fontSize = 11,
  lineNumbers = false,
  padding = 16,
  borderRadius = 6,
  backgroundColor = '#1e1e1e',
  ...props
}) => {
  const lines = Array.isArray(code) ? code : code.split('\n');
  
  return React.createElement('Box', {
    backgroundColor,
    borderColor: '#333',
    borderWidth: 1,
    borderRadius,
    padding,
    ...props
  },
    language && React.createElement('Text', {
      fontSize: fontSize - 1,
      color: '#808080',
      marginBottom: 8
    }, `// ${language}`),
    ...lines.map((line, index) => {
      const content = lineNumbers ? `${(index + 1).toString().padStart(2, ' ')}  ${line}` : line;
      return React.createElement('Text', {
        key: index,
        fontSize,
        color: '#d4d4d4',
        fontFamily: 'Monaco, Consolas, monospace'
      }, content);
    })
  );
};
