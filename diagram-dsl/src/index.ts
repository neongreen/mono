export { Box } from './components/Box';
export { Stack } from './components/Stack';
export { Row } from './components/Row';
export { Column } from './components/Column';
export { Text } from './components/Text';
export { Arrow } from './components/Arrow';

// Semantic/styled components
export { Card } from './components/Card';
export { Title } from './components/Title';
export { Subtitle } from './components/Subtitle';
export { Label } from './components/Label';

// Presentation components
export { Slide } from './components/Slide';
export { List } from './components/List';
export { ProsCons } from './components/ProsCons';
export { Section } from './components/Section';
export { Highlight } from './components/Highlight';
export { RichText } from './components/RichText';
export { Spacer } from './components/Spacer';
export { Grid } from './components/Grid';
export { Callout } from './components/Callout';

export { renderToSVG, renderToSVGWithLayout } from './renderer';
export type { RenderResult } from './renderer';

export { LayoutAssertions } from './test/layout-assertions';
export { LayoutLinter } from './test/layout-lints';
export type { LayoutLint } from './test/layout-lints';
export { measureText } from './layout/text-measurement';
export type { TextMetrics } from './layout/text-measurement';

// Export theme for advanced customization
export { theme } from './theme';
export type { Theme } from './theme';

export type {
  LayoutProps,
  AlignmentProps,
  PositionProps,
  BoxProps,
  StackProps,
  TextProps,
  ArrowProps,
  LayoutNode,
} from './types';
