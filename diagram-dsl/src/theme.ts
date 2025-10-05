/**
 * Default theme for diagram-dsl
 * Provides professional colors, typography, and spacing
 */

export const theme = {
  // Color palette - professional and accessible
  colors: {
    // Primary colors
    primary: {
      main: '#2196f3',
      light: '#e3f2fd',
      dark: '#1976d2',
    },
    secondary: {
      main: '#9c27b0',
      light: '#f3e5f5',
      dark: '#7b1fa2',
    },
    success: {
      main: '#4caf50',
      light: '#e8f5e9',
      dark: '#388e3c',
    },
    warning: {
      main: '#ff9800',
      light: '#fff3e0',
      dark: '#f57c00',
    },
    error: {
      main: '#f44336',
      light: '#ffebee',
      dark: '#d32f2f',
    },
    info: {
      main: '#00bcd4',
      light: '#e0f7fa',
      dark: '#0097a7',
    },
    // Neutral colors
    gray: {
      50: '#fafafa',
      100: '#f5f5f5',
      200: '#eeeeee',
      300: '#e0e0e0',
      400: '#bdbdbd',
      500: '#9e9e9e',
      600: '#757575',
      700: '#616161',
      800: '#424242',
      900: '#212121',
    },
    text: {
      primary: '#212121',
      secondary: '#757575',
      disabled: '#bdbdbd',
    },
    background: {
      default: '#ffffff',
      paper: '#fafafa',
    },
  },

  // Typography scale
  typography: {
    fontFamily: 'Arial, sans-serif',
    fontSize: {
      xs: 10,
      sm: 12,
      base: 14,
      lg: 16,
      xl: 20,
      '2xl': 24,
      '3xl': 30,
      '4xl': 36,
      '5xl': 48,
    },
    fontWeight: {
      normal: 'normal' as const,
      bold: 'bold' as const,
    },
    lineHeight: {
      tight: 1.2,
      normal: 1.5,
      relaxed: 1.75,
    },
  },

  // Spacing scale (based on 4px grid)
  spacing: {
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 20,
    '2xl': 24,
    '3xl': 32,
    '4xl': 40,
    '5xl': 48,
  },

  // Border radius
  borderRadius: {
    none: 0,
    sm: 4,
    md: 8,
    lg: 12,
    xl: 16,
    full: 9999,
  },

  // Borders
  border: {
    width: {
      thin: 1,
      normal: 2,
      thick: 3,
    },
  },

  // Shadows (for potential future use)
  shadows: {
    sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
    md: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
    lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1)',
  },
};

export type Theme = typeof theme;
