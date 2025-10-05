import React from 'react';
import type { BoxProps } from '../types';

export interface Actor {
  id: string;
  name: string;
  type?: 'user' | 'service' | 'database' | 'system';
}

export interface Message {
  from: string;
  to: string;
  message: string;
  type?: 'sync' | 'async' | 'return';
  style?: 'solid' | 'dashed';
}

export interface SequenceDiagramProps extends Partial<BoxProps> {
  actors: Actor[];
  messages: Message[];
  title?: string;
}

const actorTypeColors = {
  user: { bg: '#e3f2fd', border: '#1976d2', icon: '👤' },
  service: { bg: '#f3e5f5', border: '#7b1fa2', icon: '⚙️' },
  database: { bg: '#fff3e0', border: '#f57c00', icon: '🗄️' },
  system: { bg: '#e8f5e9', border: '#2e7d32', icon: '💻' },
};

/**
 * SequenceDiagram component - shows message flow between actors over time
 * Perfect for API interactions, microservices communication, user flows
 */
export const SequenceDiagram: React.FC<SequenceDiagramProps> = ({
  actors,
  messages,
  title,
  width = 900,
  ...props
}) => {
  const diagramWidth = typeof width === 'number' ? width : 900;
  const actorWidth = 120;
  const actorSpacing = (diagramWidth - actors.length * actorWidth) / (actors.length + 1);
  const messageHeight = 60;
  const headerHeight = title ? 40 : 0;
  const actorHeight = 60;
  const totalHeight = headerHeight + actorHeight + messages.length * messageHeight + 60;

  const actorElements = actors.map((actor, index) => {
    const x = actorSpacing + index * (actorWidth + actorSpacing);
    const colors = actorTypeColors[actor.type || 'service'];
    
    return React.createElement('g', { key: `actor-${actor.id}` },
      // Actor box
      React.createElement('Box', {
        position: 'absolute',
        left: x,
        top: headerHeight,
        width: actorWidth,
        height: actorHeight,
        backgroundColor: colors.bg,
        borderColor: colors.border,
        borderWidth: 2,
        borderRadius: 8,
        justifyContent: 'center',
        alignItems: 'center',
        id: `actor-${actor.id}`
      },
        React.createElement('Text', {
          fontSize: 12,
          fontWeight: 'bold',
          children: `${colors.icon} ${actor.name}`
        })
      ),
      // Lifeline
      React.createElement('Box', {
        position: 'absolute',
        left: x + actorWidth / 2 - 1,
        top: headerHeight + actorHeight,
        width: 2,
        height: messages.length * messageHeight + 20,
        backgroundColor: '#ddd'
      })
    );
  });

  // Create message arrows
  const messageElements = messages.map((msg, index) => {
    const fromIndex = actors.findIndex(a => a.id === msg.from);
    const toIndex = actors.findIndex(a => a.id === msg.to);
    
    if (fromIndex === -1 || toIndex === -1) return null;
    
    const fromX = actorSpacing + fromIndex * (actorWidth + actorSpacing) + actorWidth / 2;
    const toX = actorSpacing + toIndex * (actorWidth + actorSpacing) + actorWidth / 2;
    const y = headerHeight + actorHeight + 30 + index * messageHeight;
    
    const arrowStyle = msg.style || (msg.type === 'return' ? 'dashed' : 'solid');
    const arrowHead = msg.type === 'return' ? 'none' : 'arrow';
    
    return React.createElement('g', { key: `message-${index}` },
      // Arrow
      React.createElement('Arrow', {
        from: `msg-start-${index}`,
        to: `msg-end-${index}`,
        style: arrowStyle,
        curve: 'straight',
        headType: arrowHead,
        label: msg.message,
        color: msg.type === 'async' ? '#7b1fa2' : '#1976d2'
      }),
      // Invisible anchor points for arrow
      React.createElement('Box', {
        id: `msg-start-${index}`,
        position: 'absolute',
        left: fromX,
        top: y,
        width: 1,
        height: 1
      }),
      React.createElement('Box', {
        id: `msg-end-${index}`,
        position: 'absolute',
        left: toX,
        top: y,
        width: 1,
        height: 1
      })
    );
  });

  return React.createElement('Box', {
    width: diagramWidth,
    height: totalHeight,
    padding: 20,
    ...props
  },
    title && React.createElement('Text', {
      fontSize: 18,
      fontWeight: 'bold',
      marginBottom: 20,
      children: title
    }),
    ...actorElements,
    ...messageElements
  );
};
