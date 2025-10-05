/**
 * Presentation Theme System
 * Allows customizing presentation-wide colors, fonts, and spacing
 */

export interface PresentationTheme {
  name: string;
  
  // Colors
  primary: string;
  secondary: string;
  accent: string;
  success: string;
  warning: string;
  danger: string;
  info: string;
  text: string;
  textSecondary: string;
  textMuted: string;
  background: string;
  backgroundSecondary: string;
  border: string;
  
  // Typography
  fontFamily: string;
  fontFamilyMono: string;
  titleSize: number;
  subtitleSize: number;
  bodySize: number;
  smallSize: number;
  lineHeight: number;
  
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
  
  // Shadows
  shadowLight: string;
  shadowMedium: string;
  shadowHeavy: string;
}

export const defaultTheme: PresentationTheme = {
  name: 'Default',
  
  // Colors
  primary: '#1976d2',
  secondary: '#7b1fa2',
  accent: '#f57c00',
  success: '#2e7d32',
  warning: '#ff9800',
  danger: '#c62828',
  info: '#0288d1',
  text: '#333',
  textSecondary: '#666',
  textMuted: '#999',
  background: 'white',
  backgroundSecondary: '#f5f5f5',
  border: '#ddd',
  
  // Typography
  fontFamily: 'Arial, sans-serif',
  fontFamilyMono: 'Monaco, Consolas, monospace',
  titleSize: 32,
  subtitleSize: 20,
  bodySize: 14,
  smallSize: 12,
  lineHeight: 1.6,
  
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
  
  // Shadows
  shadowLight: 'rgba(0, 0, 0, 0.05)',
  shadowMedium: 'rgba(0, 0, 0, 0.1)',
  shadowHeavy: 'rgba(0, 0, 0, 0.2)',
};

// Professional theme - more muted colors
export const professionalTheme: PresentationTheme = {
  ...defaultTheme,
  name: 'Professional',
  primary: '#2c3e50',
  secondary: '#34495e',
  accent: '#e67e22',
  success: '#27ae60',
  warning: '#f39c12',
  danger: '#c0392b',
  info: '#3498db',
  text: '#2c3e50',
  textSecondary: '#7f8c8d',
  textMuted: '#95a5a6',
  backgroundSecondary: '#ecf0f1',
  border: '#bdc3c7',
};

// Dark theme
export const darkTheme: PresentationTheme = {
  ...defaultTheme,
  name: 'Dark',
  primary: '#64b5f6',
  secondary: '#ba68c8',
  accent: '#ffb74d',
  success: '#81c784',
  warning: '#ffb74d',
  danger: '#e57373',
  info: '#4fc3f7',
  text: '#e0e0e0',
  textSecondary: '#bdbdbd',
  textMuted: '#757575',
  background: '#1e1e1e',
  backgroundSecondary: '#2d2d2d',
  border: '#404040',
  shadowLight: 'rgba(0, 0, 0, 0.3)',
  shadowMedium: 'rgba(0, 0, 0, 0.5)',
  shadowHeavy: 'rgba(0, 0, 0, 0.7)',
};

// Vibrant theme
export const vibrantTheme: PresentationTheme = {
  ...defaultTheme,
  name: 'Vibrant',
  primary: '#3f51b5',
  secondary: '#9c27b0',
  accent: '#ff5722',
  success: '#4caf50',
  warning: '#ff9800',
  danger: '#f44336',
  info: '#2196f3',
  backgroundSecondary: '#f3f4ff',
};

// Minimal theme
export const minimalTheme: PresentationTheme = {
  ...defaultTheme,
  name: 'Minimal',
  primary: '#000000',
  secondary: '#424242',
  accent: '#757575',
  success: '#000000',
  warning: '#424242',
  danger: '#000000',
  info: '#616161',
  text: '#000000',
  textSecondary: '#757575',
  textMuted: '#9e9e9e',
  backgroundSecondary: '#fafafa',
  border: '#e0e0e0',
  borderRadius: 0,
  shadowLight: 'rgba(0, 0, 0, 0.03)',
  shadowMedium: 'rgba(0, 0, 0, 0.06)',
  shadowHeavy: 'rgba(0, 0, 0, 0.12)',
};

// Solarized Light theme
export const solarizedLightTheme: PresentationTheme = {
  ...defaultTheme,
  name: 'Solarized Light',
  primary: '#268bd2',
  secondary: '#6c71c4',
  accent: '#cb4b16',
  success: '#859900',
  warning: '#b58900',
  danger: '#dc322f',
  info: '#2aa198',
  text: '#657b83',
  textSecondary: '#839496',
  textMuted: '#93a1a1',
  background: '#fdf6e3',
  backgroundSecondary: '#eee8d5',
  border: '#93a1a1',
};

// Solarized Dark theme
export const solarizedDarkTheme: PresentationTheme = {
  ...defaultTheme,
  name: 'Solarized Dark',
  primary: '#268bd2',
  secondary: '#6c71c4',
  accent: '#cb4b16',
  success: '#859900',
  warning: '#b58900',
  danger: '#dc322f',
  info: '#2aa198',
  text: '#839496',
  textSecondary: '#93a1a1',
  textMuted: '#586e75',
  background: '#002b36',
  backgroundSecondary: '#073642',
  border: '#586e75',
  shadowLight: 'rgba(0, 0, 0, 0.4)',
  shadowMedium: 'rgba(0, 0, 0, 0.6)',
  shadowHeavy: 'rgba(0, 0, 0, 0.8)',
};

// Nord theme
export const nordTheme: PresentationTheme = {
  ...defaultTheme,
  name: 'Nord',
  primary: '#5e81ac',
  secondary: '#81a1c1',
  accent: '#d08770',
  success: '#a3be8c',
  warning: '#ebcb8b',
  danger: '#bf616a',
  info: '#88c0d0',
  text: '#2e3440',
  textSecondary: '#4c566a',
  textMuted: '#d8dee9',
  background: '#eceff4',
  backgroundSecondary: '#e5e9f0',
  border: '#d8dee9',
};

// Dracula theme
export const draculaTheme: PresentationTheme = {
  ...defaultTheme,
  name: 'Dracula',
  primary: '#bd93f9',
  secondary: '#ff79c6',
  accent: '#ffb86c',
  success: '#50fa7b',
  warning: '#f1fa8c',
  danger: '#ff5555',
  info: '#8be9fd',
  text: '#f8f8f2',
  textSecondary: '#e6e6e6',
  textMuted: '#6272a4',
  background: '#282a36',
  backgroundSecondary: '#44475a',
  border: '#6272a4',
  shadowLight: 'rgba(0, 0, 0, 0.5)',
  shadowMedium: 'rgba(0, 0, 0, 0.7)',
  shadowHeavy: 'rgba(0, 0, 0, 0.9)',
};

// GitHub theme
export const githubTheme: PresentationTheme = {
  ...defaultTheme,
  name: 'GitHub',
  primary: '#0969da',
  secondary: '#8250df',
  accent: '#bf8700',
  success: '#1a7f37',
  warning: '#9a6700',
  danger: '#cf222e',
  info: '#0969da',
  text: '#24292f',
  textSecondary: '#57606a',
  textMuted: '#8c959f',
  background: '#ffffff',
  backgroundSecondary: '#f6f8fa',
  border: '#d0d7de',
  borderRadius: 6,
};

// High Contrast theme
export const highContrastTheme: PresentationTheme = {
  ...defaultTheme,
  name: 'High Contrast',
  primary: '#0000ff',
  secondary: '#8b00ff',
  accent: '#ff8c00',
  success: '#008000',
  warning: '#ff8c00',
  danger: '#ff0000',
  info: '#0000ff',
  text: '#000000',
  textSecondary: '#000000',
  textMuted: '#666666',
  background: '#ffffff',
  backgroundSecondary: '#f0f0f0',
  border: '#000000',
  borderWidth: 3,
  shadowLight: 'rgba(0, 0, 0, 0.2)',
  shadowMedium: 'rgba(0, 0, 0, 0.4)',
  shadowHeavy: 'rgba(0, 0, 0, 0.6)',
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
