import React from 'react';
import type { BoxProps } from '../types';

export interface APIEndpointProps extends Partial<BoxProps> {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  path: string;
  description?: string;
  request?: {
    params?: Record<string, string>;
    body?: string;
  };
  response?: {
    status: number;
    body?: string;
  };
}

const methodColors = {
  GET: { bg: '#e3f2fd', border: '#1976d2', text: '#1565c0' },
  POST: { bg: '#e8f5e9', border: '#2e7d32', text: '#1b5e20' },
  PUT: { bg: '#fff3e0', border: '#f57c00', text: '#e65100' },
  DELETE: { bg: '#ffebee', border: '#c62828', text: '#b71c1c' },
  PATCH: { bg: '#f3e5f5', border: '#7b1fa2', text: '#6a1b9a' },
};

/**
 * APIEndpoint component - documents REST API endpoints
 * Perfect for API documentation and technical specifications
 */
export const APIEndpoint: React.FC<APIEndpointProps> = ({
  method,
  path,
  description,
  request,
  response,
  width = 800,
  ...props
}) => {
  const colors = methodColors[method];
  
  return React.createElement('Stack', {
    width,
    borderColor: colors.border,
    borderWidth: 2,
    borderRadius: 8,
    padding: 0,
    gap: 0,
    ...props
  },
    // Header with method and path
    React.createElement('Row', {
      backgroundColor: colors.bg,
      padding: 16,
      gap: 12,
      alignItems: 'center'
    },
      React.createElement('Box', {
        backgroundColor: colors.text,
        padding: 6,
        paddingLeft: 12,
        paddingRight: 12,
        borderRadius: 4
      },
        React.createElement('Text', {
          fontSize: 12,
          fontWeight: 'bold',
          color: 'white',
          children: method
        })
      ),
      React.createElement('Text', {
        fontSize: 14,
        fontFamily: 'Monaco, Courier, monospace',
        fontWeight: 'bold',
        children: path
      })
    ),
    
    // Description
    description && React.createElement('Box', {
      padding: 16,
      borderBottomWidth: 1,
      borderBottomColor: '#e0e0e0'
    },
      React.createElement('Text', {
        fontSize: 13,
        children: description
      })
    ),
    
    // Request section
    (request?.params || request?.body) && React.createElement('Stack', {
      padding: 16,
      gap: 12,
      borderBottomWidth: 1,
      borderBottomColor: '#e0e0e0'
    },
      React.createElement('Text', {
        fontSize: 12,
        fontWeight: 'bold',
        color: '#666',
        children: 'REQUEST'
      }),
      request.params && React.createElement('Stack', { gap: 4 },
        React.createElement('Text', {
          fontSize: 11,
          fontWeight: 'bold',
          children: 'Parameters:'
        }),
        ...Object.entries(request.params).map(([key, value]) =>
          React.createElement('Text', {
            key,
            fontSize: 11,
            fontFamily: 'Monaco, Courier, monospace',
            children: `  ${key}: ${value}`
          })
        )
      ),
      request.body && React.createElement('Stack', { gap: 4 },
        React.createElement('Text', {
          fontSize: 11,
          fontWeight: 'bold',
          children: 'Body:'
        }),
        React.createElement('Box', {
          backgroundColor: '#1e1e1e',
          padding: 12,
          borderRadius: 4
        },
          React.createElement('Text', {
            fontSize: 11,
            fontFamily: 'Monaco, Courier, monospace',
            color: '#d4d4d4',
            children: request.body
          })
        )
      )
    ),
    
    // Response section
    response && React.createElement('Stack', {
      padding: 16,
      gap: 12
    },
      React.createElement('Text', {
        fontSize: 12,
        fontWeight: 'bold',
        color: '#666',
        children: 'RESPONSE'
      }),
      React.createElement('Row', { gap: 8, alignItems: 'center' },
        React.createElement('Box', {
          backgroundColor: response.status < 300 ? '#e8f5e9' : '#ffebee',
          padding: 4,
          paddingLeft: 8,
          paddingRight: 8,
          borderRadius: 4
        },
          React.createElement('Text', {
            fontSize: 11,
            fontWeight: 'bold',
            color: response.status < 300 ? '#2e7d32' : '#c62828',
            children: response.status.toString()
          })
        )
      ),
      response.body && React.createElement('Box', {
        backgroundColor: '#1e1e1e',
        padding: 12,
        borderRadius: 4
      },
        React.createElement('Text', {
          fontSize: 11,
          fontFamily: 'Monaco, Courier, monospace',
          color: '#d4d4d4',
          children: response.body
        })
      )
    )
  );
};
