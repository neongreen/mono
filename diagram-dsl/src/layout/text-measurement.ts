import { createCanvas } from 'canvas';

export interface TextMetrics {
  width: number;
  height: number;
}

/**
 * Measures the actual dimensions of text using canvas API
 * @param text The text to measure
 * @param fontSize Font size in pixels
 * @param fontFamily Font family (default: Arial, sans-serif)
 * @param fontWeight Font weight (default: normal)
 * @returns Accurate width and height of the text
 */
export function measureText(
  text: string,
  fontSize: number = 16,
  fontFamily: string = 'Arial, sans-serif',
  fontWeight: string = 'normal'
): TextMetrics {
  // Create a minimal canvas for text measurement
  const canvas = createCanvas(1, 1);
  const ctx = canvas.getContext('2d');
  
  // Set font properties
  ctx.font = `${fontWeight} ${fontSize}px ${fontFamily}`;
  
  // Measure the text
  const metrics = ctx.measureText(text);
  
  // Calculate actual width and height
  const width = metrics.width;
  
  // Height is calculated from ascent and descent
  // For better accuracy, we use actualBoundingBox metrics when available
  const height = metrics.actualBoundingBoxAscent + metrics.actualBoundingBoxDescent;
  
  return {
    width: Math.ceil(width),
    height: Math.ceil(height || fontSize * 1.2) // Fallback to fontSize * 1.2 if metrics not available
  };
}
