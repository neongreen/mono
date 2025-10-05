import React from 'react';
import { Box } from './Box';
import { Text } from './Text';
import { LayoutProps } from '../types';

export interface ImageProps extends LayoutProps {
  src: string;
  alt?: string;
  width?: number;
  height?: number;
  fit?: 'contain' | 'cover' | 'fill' | 'none';
  borderRadius?: number;
  opacity?: number;
}

/**
 * Image component for embedding images in presentations
 * Currently renders a placeholder with the image path
 * TODO: Full image rendering support in SVG renderer
 */
export function Image({
  src,
  alt = '',
  width = 400,
  height = 300,
  fit = 'contain',
  borderRadius = 0,
  opacity = 1,
  ...layoutProps
}: ImageProps) {
  // For now, create a placeholder box with image info
  // This will be enhanced when we add full SVG image support to the renderer
  return (
    <Box
      width={width}
      height={height}
      backgroundColor="#f0f0f0"
      borderColor="#ccc"
      borderWidth={1}
      borderRadius={borderRadius}
      justifyContent="center"
      alignItems="center"
      {...layoutProps}
    >
      <Text fontSize={12} color="#666" textAlign="center">
        {`📷 Image: ${src.substring(0, 50)}${src.length > 50 ? '...' : ''}`}
      </Text>
      {alt && (
        <Text fontSize={10} color="#999" textAlign="center" marginTop={4}>
          {alt}
        </Text>
      )}
    </Box>
  );
}
