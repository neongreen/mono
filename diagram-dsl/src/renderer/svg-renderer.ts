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
        const style = node.props.style || 'solid';
        const curve = node.props.curve || 'straight';
        const headType = node.props.headType || 'arrow';
        const tailType = node.props.tailType || 'none';

        // Calculate centers of both boxes
        const fromCenterX = fromPos.x + fromPos.width / 2;
        const fromCenterY = fromPos.y + fromPos.height / 2;
        const toCenterX = toPos.x + toPos.width / 2;
        const toCenterY = toPos.y + toPos.height / 2;

        // Calculate attachment points (centers of rectangle sides)
        const fromAttachmentPoints = {
          top: { x: fromCenterX, y: fromPos.y },
          bottom: { x: fromCenterX, y: fromPos.y + fromPos.height },
          left: { x: fromPos.x, y: fromCenterY },
          right: { x: fromPos.x + fromPos.width, y: fromCenterY }
        };

        const toAttachmentPoints = {
          top: { x: toCenterX, y: toPos.y },
          bottom: { x: toCenterX, y: toPos.y + toPos.height },
          left: { x: toPos.x, y: toCenterY },
          right: { x: toPos.x + toPos.width, y: toCenterY }
        };

        // Find closest attachment points
        let minDistance = Infinity;
        let bestFrom = fromAttachmentPoints.bottom;
        let bestTo = toAttachmentPoints.top;

        for (const [fromSide, fromPoint] of Object.entries(fromAttachmentPoints)) {
          for (const [toSide, toPoint] of Object.entries(toAttachmentPoints)) {
            const distance = Math.sqrt(
              Math.pow(toPoint.x - fromPoint.x, 2) + 
              Math.pow(toPoint.y - fromPoint.y, 2)
            );
            if (distance < minDistance) {
              minDistance = distance;
              bestFrom = fromPoint;
              bestTo = toPoint;
            }
          }
        }

        const x1 = bestFrom.x;
        const y1 = bestFrom.y;
        const x2 = bestTo.x;
        const y2 = bestTo.y;

        // Determine stroke dash array based on style
        let strokeDashArray = '';
        if (style === 'dashed') {
          strokeDashArray = `stroke-dasharray="8,4" `;
        } else if (style === 'dotted') {
          strokeDashArray = `stroke-dasharray="2,4" `;
        }

        // Generate path based on curve style
        let pathData = '';
        if (curve === 'straight') {
          pathData = `M ${x1} ${y1} L ${x2} ${y2}`;
        } else if (curve === 'curved') {
          // Bezier curve
          const dx = x2 - x1;
          const dy = y2 - y1;
          const controlX1 = x1 + dx * 0.5;
          const controlY1 = y1;
          const controlX2 = x2 - dx * 0.5;
          const controlY2 = y2;
          pathData = `M ${x1} ${y1} C ${controlX1} ${controlY1}, ${controlX2} ${controlY2}, ${x2} ${y2}`;
        } else if (curve === 'step') {
          // Step/orthogonal path
          const midX = (x1 + x2) / 2;
          pathData = `M ${x1} ${y1} L ${midX} ${y1} L ${midX} ${y2} L ${x2} ${y2}`;
        }

        // Generate marker IDs
        const colorKey = color.replace('#', '');
        const headMarker = headType !== 'none' ? `url(#${headType}-${colorKey})` : '';
        const tailMarker = tailType !== 'none' ? `url(#${tailType}-tail-${colorKey})` : '';

        // Draw path
        svg += `\n${indent}<path d="${pathData}" `;
        svg += `stroke="${color}" stroke-width="${strokeWidth}" fill="none" ${strokeDashArray}`;
        if (headMarker) svg += `marker-end="${headMarker}" `;
        if (tailMarker) svg += `marker-start="${tailMarker}" `;
        svg += `/>`;

        // Add label if present with semi-transparent background
        if (node.props.label) {
          const midX = (x1 + x2) / 2;
          const midY = (y1 + y2) / 2;
          
          // Measure approximate text size for background
          const labelText = this.escapeXml(node.props.label);
          const fontSize = 12;
          const textWidth = labelText.length * fontSize * 0.6; // Approximate width
          const textHeight = fontSize;
          const padding = 4;
          
          // Draw semi-transparent background rectangle
          svg += `\n${indent}<rect x="${midX - textWidth / 2 - padding}" y="${midY - textHeight / 2 - padding}" `;
          svg += `width="${textWidth + padding * 2}" height="${textHeight + padding * 2}" `;
          svg += `fill="white" fill-opacity="0.85" rx="2" />`;
          
          // Draw label text
          svg += `\n${indent}<text x="${midX}" y="${midY + fontSize / 3}" `;
          svg += `font-size="${fontSize}" fill="${color}" text-anchor="middle">${labelText}</text>`;
        }
      }
    }

    // Recursively render arrows in children
    node.children.forEach(child => {
      svg += this.renderArrows(child, indent);
    });

    return svg;
  }

  private escapeXml(text: any): string {
    if (text == null) return '';
    const str = String(text);
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&apos;');
  }

  // Add arrowhead marker definitions to SVG
  renderWithArrowMarkers(tree: LayoutNode, options: RenderOptions = {}): string {
    const svg = this.render(tree, options);
    
    // Extract all unique colors and head types used in arrows
    const arrowConfigs = new Map<string, Set<string>>(); // color -> set of head types
    this.collectArrowConfigs(tree, arrowConfigs);

    // Generate marker definitions
    let markers = '\n  <defs>';
    
    arrowConfigs.forEach((headTypes, color) => {
      const colorId = color.replace('#', '');
      
      headTypes.forEach(headType => {
        if (headType === 'arrow') {
          // Standard arrow marker
          markers += `\n    <marker id="arrow-${colorId}" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">`;
          markers += `\n      <polygon points="0 0, 10 3, 0 6" fill="${color}" />`;
          markers += `\n    </marker>`;
          
          // Tail arrow (reversed)
          markers += `\n    <marker id="arrow-tail-${colorId}" markerWidth="10" markerHeight="10" refX="1" refY="3" orient="auto">`;
          markers += `\n      <polygon points="10 0, 0 3, 10 6" fill="${color}" />`;
          markers += `\n    </marker>`;
        } else if (headType === 'circle') {
          // Circle marker
          markers += `\n    <marker id="circle-${colorId}" markerWidth="8" markerHeight="8" refX="4" refY="4" orient="auto">`;
          markers += `\n      <circle cx="4" cy="4" r="3" fill="${color}" />`;
          markers += `\n    </marker>`;
          
          // Tail circle
          markers += `\n    <marker id="circle-tail-${colorId}" markerWidth="8" markerHeight="8" refX="4" refY="4" orient="auto">`;
          markers += `\n      <circle cx="4" cy="4" r="3" fill="${color}" />`;
          markers += `\n    </marker>`;
        } else if (headType === 'diamond') {
          // Diamond marker
          markers += `\n    <marker id="diamond-${colorId}" markerWidth="12" markerHeight="12" refX="6" refY="6" orient="auto">`;
          markers += `\n      <polygon points="6 0, 12 6, 6 12, 0 6" fill="${color}" />`;
          markers += `\n    </marker>`;
          
          // Tail diamond
          markers += `\n    <marker id="diamond-tail-${colorId}" markerWidth="12" markerHeight="12" refX="6" refY="6" orient="auto">`;
          markers += `\n      <polygon points="6 0, 12 6, 6 12, 0 6" fill="${color}" />`;
          markers += `\n    </marker>`;
        }
      });
    });
    
    markers += '\n  </defs>';

    // Insert markers after the opening svg tag
    return svg.replace('<svg xmlns="http://www.w3.org/2000/svg"', '<svg xmlns="http://www.w3.org/2000/svg"').replace('>', '>' + markers);
  }

  private collectArrowConfigs(node: LayoutNode, configs: Map<string, Set<string>>): void {
    if (node.type === 'Arrow') {
      const color = node.props.color || 'black';
      const headType = node.props.headType || 'arrow';
      const tailType = node.props.tailType || 'none';
      
      if (!configs.has(color)) {
        configs.set(color, new Set());
      }
      const types = configs.get(color)!;
      
      if (headType !== 'none') types.add(headType);
      if (tailType !== 'none') types.add(tailType);
    }
    node.children.forEach(child => this.collectArrowConfigs(child, configs));
  }
}
