import { LayoutNode } from '../types';

export interface RenderOptions {
  width?: number;
  height?: number;
  backgroundColor?: string;
}

export class SVGRenderer {
  private idToPosition: Map<string, { x: number; y: number; width: number; height: number }> = new Map();

  render(tree: LayoutNode, options: RenderOptions = {}): string {
    this.idToPosition.clear();
    
    const width = options.width || 800;
    const height = options.height || 600;
    const bgColor = options.backgroundColor || 'white';

    let svgContent = `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">`;
    svgContent += `\n  <rect width="${width}" height="${height}" fill="${bgColor}"/>`;
    
    // First pass: collect all nodes with IDs
    this.collectIds(tree);
    
    // Second pass: render nodes
    svgContent += this.renderNode(tree);
    
    // Third pass: render arrows
    svgContent += this.renderArrows(tree);
    
    svgContent += '\n</svg>';
    
    return svgContent;
  }

  private collectIds(node: LayoutNode): void {
    if (node.props.id && node.computed) {
      this.idToPosition.set(node.props.id, {
        x: node.computed.x,
        y: node.computed.y,
        width: node.computed.width,
        height: node.computed.height,
      });
    }
    node.children.forEach(child => this.collectIds(child));
  }

  private renderNode(node: LayoutNode, indent: string = '  '): string {
    if (!node.computed) return '';

    let svg = '';
    const { x, y, width, height } = node.computed;

    switch (node.type) {
      case 'Box':
      case 'Stack':
      case 'Row':
      case 'Column':
        const bgColor = node.props.backgroundColor || 'transparent';
        const borderColor = node.props.borderColor || 'black';
        const borderWidth = node.props.borderWidth || 0;
        const borderRadius = node.props.borderRadius || 0;

        if (bgColor !== 'transparent' || borderWidth > 0) {
          svg += `\n${indent}<rect x="${x}" y="${y}" width="${width}" height="${height}" `;
          svg += `fill="${bgColor}" `;
          if (borderWidth > 0) {
            svg += `stroke="${borderColor}" stroke-width="${borderWidth}" `;
          }
          if (borderRadius > 0) {
            svg += `rx="${borderRadius}" `;
          }
          svg += `/>`;
        }
        break;

      case 'Text':
        const text = node.props.children || '';
        const fontSize = node.props.fontSize || 16;
        const fontFamily = node.props.fontFamily || 'Arial, sans-serif';
        const color = node.props.color || 'black';
        const fontWeight = node.props.fontWeight || 'normal';
        const textAlign = node.props.textAlign || 'left';

        let textX = x;
        let textAnchor = 'start';
        
        if (textAlign === 'center') {
          textX = x + width / 2;
          textAnchor = 'middle';
        } else if (textAlign === 'right') {
          textX = x + width;
          textAnchor = 'end';
        }

        const textY = y + height / 2 + fontSize / 3; // Vertically center text

        svg += `\n${indent}<text x="${textX}" y="${textY}" `;
        svg += `font-size="${fontSize}" font-family="${fontFamily}" `;
        svg += `fill="${color}" font-weight="${fontWeight}" `;
        svg += `text-anchor="${textAnchor}">${this.escapeXml(text)}</text>`;
        break;
    }

    // Render children
    node.children.forEach(child => {
      svg += this.renderNode(child, indent + '  ');
    });

    return svg;
  }

  private renderArrows(node: LayoutNode, indent: string = '  '): string {
    let svg = '';

    if (node.type === 'Arrow') {
      const fromPos = this.idToPosition.get(node.props.from);
      const toPos = this.idToPosition.get(node.props.to);

      if (fromPos && toPos) {
        const color = node.props.color || 'black';
        const strokeWidth = node.props.strokeWidth || 2;

        // Calculate arrow start and end points (center of boxes)
        const x1 = fromPos.x + fromPos.width / 2;
        const y1 = fromPos.y + fromPos.height / 2;
        const x2 = toPos.x + toPos.width / 2;
        const y2 = toPos.y + toPos.height / 2;

        // Calculate arrow direction
        const angle = Math.atan2(y2 - y1, x2 - x1);
        
        // Adjust end point to be at the edge of the target box
        const adjustedX2 = toPos.x + toPos.width / 2 - Math.cos(angle) * (toPos.width / 2);
        const adjustedY2 = toPos.y + toPos.height / 2 - Math.sin(angle) * (toPos.height / 2);

        // Draw line
        svg += `\n${indent}<line x1="${x1}" y1="${y1}" x2="${adjustedX2}" y2="${adjustedY2}" `;
        svg += `stroke="${color}" stroke-width="${strokeWidth}" marker-end="url(#arrowhead-${color.replace('#', '')})" />`;

        // Add label if present
        if (node.props.label) {
          const midX = (x1 + adjustedX2) / 2;
          const midY = (y1 + adjustedY2) / 2;
          svg += `\n${indent}<text x="${midX}" y="${midY - 5}" `;
          svg += `font-size="12" fill="${color}" text-anchor="middle">${this.escapeXml(node.props.label)}</text>`;
        }
      }
    }

    // Recursively render arrows in children
    node.children.forEach(child => {
      svg += this.renderArrows(child, indent);
    });

    return svg;
  }

  private escapeXml(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&apos;');
  }

  // Add arrowhead marker definitions to SVG
  renderWithArrowMarkers(tree: LayoutNode, options: RenderOptions = {}): string {
    const svg = this.render(tree, options);
    
    // Extract all unique colors used in arrows
    const colors = new Set<string>();
    this.collectArrowColors(tree, colors);

    // Generate marker definitions
    let markers = '\n  <defs>';
    colors.forEach(color => {
      const colorId = color.replace('#', '');
      markers += `\n    <marker id="arrowhead-${colorId}" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">`;
      markers += `\n      <polygon points="0 0, 10 3, 0 6" fill="${color}" />`;
      markers += `\n    </marker>`;
    });
    markers += '\n  </defs>';

    // Insert markers after the opening svg tag
    return svg.replace('<svg xmlns="http://www.w3.org/2000/svg"', '<svg xmlns="http://www.w3.org/2000/svg"').replace('>', '>' + markers);
  }

  private collectArrowColors(node: LayoutNode, colors: Set<string>): void {
    if (node.type === 'Arrow') {
      colors.add(node.props.color || 'black');
    }
    node.children.forEach(child => this.collectArrowColors(child, colors));
  }
}
