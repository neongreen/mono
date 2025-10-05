import React from 'react';
import type { BoxProps } from '../types';

export interface ContainerSection {
  title?: string;
  content: React.ReactNode;
  flex?: number;
}

export interface ContainerProps extends Partial<BoxProps> {
  sections: ContainerSection[];
  orientation?: 'horizontal' | 'vertical';
  variant?: 'default' | 'primary' | 'secondary' | 'accent';
  showDividers?: boolean;
  dividerColor?: string;
}

const variantColors = {
  default: { border: '#ddd', bg: 'white', header: '#f5f5f5', divider: '#e0e0e0' },
  primary: { border: '#1976d2', bg: '#fafcff', header: '#e3f2fd', divider: '#bbdefb' },
  secondary: { border: '#7b1fa2', bg: '#fdfaff', header: '#f3e5f5', divider: '#e1bee7' },
  accent: { border: '#f57c00', bg: '#fffcfa', header: '#fff3e0', divider: '#ffe0b2' },
};

/**
 * Container component - creates a box with multiple sections separated by dividers
 * Each section can have a title and custom content
 */
export const Container: React.FC<ContainerProps> = ({
  sections,
  orientation = 'vertical',
  variant = 'default',
  showDividers = true,
  dividerColor,
  borderRadius = 8,
  borderWidth = 2,
  padding = 0,
  ...props
}) => {
  const colors = variantColors[variant];
  const Layout = orientation === 'horizontal' ? 'Row' : 'Stack';
  const dividerProps = orientation === 'horizontal'
    ? { width: 2, height: '100%', backgroundColor: dividerColor || colors.divider }
    : { height: 2, width: '100%', backgroundColor: dividerColor || colors.divider };

  const sectionElements: React.ReactNode[] = [];
  
  sections.forEach((section, index) => {
    // Add section
    const sectionContent = section.title
      ? React.createElement('Stack', { flexGrow: section.flex || 1, gap: 0 },
          React.createElement('Box', {
            backgroundColor: colors.header,
            padding: 12,
            borderBottomWidth: 1,
            borderBottomColor: colors.divider
          },
            React.createElement('Text', {
              fontSize: 13,
              fontWeight: 'bold',
              color: '#333'
            }, section.title)
          ),
          React.createElement('Box', { padding: 16, flexGrow: 1 },
            section.content
          )
        )
      : React.createElement('Box', { 
          padding: 16, 
          flexGrow: section.flex || 1 
        }, section.content);
    
    sectionElements.push(
      React.createElement('div', { key: `section-${index}` }, sectionContent)
    );
    
    // Add divider between sections
    if (showDividers && index < sections.length - 1) {
      sectionElements.push(
        React.createElement('Box', { 
          key: `divider-${index}`,
          ...dividerProps 
        })
      );
    }
  });

  return React.createElement('Box', {
    borderColor: colors.border,
    borderWidth,
    borderRadius,
    padding,
    backgroundColor: colors.bg,
    ...props
  },
    React.createElement(Layout, { gap: 0 }, ...sectionElements)
  );
};
