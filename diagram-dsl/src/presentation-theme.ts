/**
 * Presentation Theme System
 * Allows customizing presentation-wide colors, fonts, and spacing
 */

export interface PresentationTheme {
  // Colors
  primary: string;
  secondary: string;
  accent: string;
  success: string;
  warning: string;
  danger: string;
  text: string;
  textSecondary: string;
  background: string;
  
  // Typography
  fontFamily: string;
  fontFamilyMono: string;
  titleSize: number;
  subtitleSize: number;
  bodySize: number;
  smallSize: number;
  
  // Spacing
  slideWidth: number;
  slideHeight: number;
  slidePadding: number;
  gap: number;
  gapSmall: number;
  gapLarge: number;
  
  // Borders
  borderRadius: number;
  borderWidth: number;
}

export const defaultTheme: PresentationTheme = {
  // Colors
  primary: '#1976d2',
  secondary: '#7b1fa2',
  accent: '#f57c00',
  success: '#2e7d32',
  warning: '#ff9800',
  danger: '#c62828',
  text: '#333',
  textSecondary: '#666',
  background: 'white',
  
  // Typography
  fontFamily: 'Arial, sans-serif',
  fontFamilyMono: 'Monaco, Consolas, monospace',
  titleSize: 32,
  subtitleSize: 20,
  bodySize: 14,
  smallSize: 12,
  
  // Spacing
  slideWidth: 1200,
  slideHeight: 800,
  slidePadding: 60,
  gap: 16,
  gapSmall: 8,
  gapLarge: 32,
  
  // Borders
  borderRadius: 8,
  borderWidth: 2,
};

// Professional theme - more muted colors
export const professionalTheme: PresentationTheme = {
  ...defaultTheme,
  primary: '#2c3e50',
  secondary: '#34495e',
  accent: '#e67e22',
  success: '#27ae60',
  warning: '#f39c12',
  danger: '#c0392b',
  text: '#2c3e50',
  textSecondary: '#7f8c8d',
};

// Dark theme
export const darkTheme: PresentationTheme = {
  ...defaultTheme,
  primary: '#64b5f6',
  secondary: '#ba68c8',
  accent: '#ffb74d',
  success: '#81c784',
  warning: '#ffb74d',
  danger: '#e57373',
  text: '#e0e0e0',
  textSecondary: '#bdbdbd',
  background: '#1e1e1e',
};

// Vibrant theme
export const vibrantTheme: PresentationTheme = {
  ...defaultTheme,
  primary: '#3f51b5',
  secondary: '#9c27b0',
  accent: '#ff5722',
  success: '#4caf50',
  warning: '#ff9800',
  danger: '#f44336',
};

// Minimal theme
export const minimalTheme: PresentationTheme = {
  ...defaultTheme,
  primary: '#000000',
  secondary: '#424242',
  accent: '#757575',
  success: '#000000',
  warning: '#424242',
  danger: '#000000',
  text: '#000000',
  textSecondary: '#757575',
  borderRadius: 0,
};

// Presentation theme context (for potential future React Context use)
let currentTheme: PresentationTheme = defaultTheme;

export function setCurrentTheme(theme: PresentationTheme) {
  currentTheme = theme;
}

export function getCurrentTheme(): PresentationTheme {
  return currentTheme;
}

export function createCustomTheme(overrides: Partial<PresentationTheme>): PresentationTheme {
  return {
    ...defaultTheme,
    ...overrides,
  };
}
