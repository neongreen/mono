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
        const borderStyle = node.props.borderStyle || 'solid';
        const borderDashArray = node.props.borderDashArray;

        if (bgColor !== 'transparent' || borderWidth > 0) {
          svg += `\n${indent}<rect x="${x}" y="${y}" width="${width}" height="${height}" `;
          svg += `fill="${bgColor}" `;
          if (borderWidth > 0) {
            svg += `stroke="${borderColor}" stroke-width="${borderWidth}" `;
            
            // Handle dashed borders
            if (borderStyle === 'dashed' || borderDashArray) {
              const dashArray = borderDashArray || '6 4'; // Default: 6px dash, 4px gap
              svg += `stroke-dasharray="${dashArray}" `;
            } else if (borderStyle === 'dotted') {
              svg += `stroke-dasharray="2 2" `;
            }
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

      case 'StateNode':
        svg += this.renderStateNode(node, x, y, width, height, indent);
        break;

      case 'DecisionNode':
        svg += this.renderDecisionNode(node, x, y, width, height, indent);
        break;

      case 'ProcessNode':
        svg += this.renderProcessNode(node, x, y, width, height, indent);
        break;

      case 'MemoryBlock':
        svg += this.renderMemoryBlock(node, x, y, width, height, indent);
        break;

      case 'ContextWindow':
        svg += this.renderContextWindow(node, x, y, width, height, indent);
        break;

      case 'Timeline':
        svg += this.renderTimeline(node, x, y, width, height, indent);
        break;

      case 'TimelineEvent':
        svg += this.renderTimelineEvent(node, x, y, width, height, indent);
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

        // Helper to calculate attachment point with optional offset
        const getAttachmentPoint = (pos: any, side: string, offset: number = 0) => {
          const centerX = pos.x + pos.width / 2;
          const centerY = pos.y + pos.height / 2;
          
          switch (side) {
            case 'top':
              return { x: centerX + offset, y: pos.y };
            case 'bottom':
              return { x: centerX + offset, y: pos.y + pos.height };
            case 'left':
              return { x: pos.x, y: centerY + offset };
            case 'right':
              return { x: pos.x + pos.width, y: centerY + offset };
            default:
              return { x: centerX, y: centerY };
          }
        };

        // Calculate attachment points (centers of rectangle sides)
        const fromAttachmentPoints = {
          top: getAttachmentPoint(fromPos, 'top'),
          bottom: getAttachmentPoint(fromPos, 'bottom'),
          left: getAttachmentPoint(fromPos, 'left'),
          right: getAttachmentPoint(fromPos, 'right')
        };

        const toAttachmentPoints = {
          top: getAttachmentPoint(toPos, 'top'),
          bottom: getAttachmentPoint(toPos, 'bottom'),
          left: getAttachmentPoint(toPos, 'left'),
          right: getAttachmentPoint(toPos, 'right')
        };

        // Use specified sides or find closest attachment points
        let bestFrom: { x: number; y: number };
        let bestTo: { x: number; y: number };

        if (node.props.fromSide && node.props.fromSide !== 'auto') {
          bestFrom = getAttachmentPoint(fromPos, node.props.fromSide, node.props.fromOffset || 0);
        } else {
          // Auto-detect best from side
          let minDistance = Infinity;
          bestFrom = fromAttachmentPoints.bottom;
          
          for (const [fromSide, fromPoint] of Object.entries(fromAttachmentPoints)) {
            for (const [toSide, toPoint] of Object.entries(toAttachmentPoints)) {
              const distance = Math.sqrt(
                Math.pow(toPoint.x - fromPoint.x, 2) + 
                Math.pow(toPoint.y - fromPoint.y, 2)
              );
              if (distance < minDistance) {
                minDistance = distance;
                bestFrom = fromPoint;
              }
            }
          }
        }

        if (node.props.toSide && node.props.toSide !== 'auto') {
          bestTo = getAttachmentPoint(toPos, node.props.toSide, node.props.toOffset || 0);
        } else {
          // Auto-detect best to side
          let minDistance = Infinity;
          bestTo = toAttachmentPoints.top;
          
          for (const [toSide, toPoint] of Object.entries(toAttachmentPoints)) {
            const distance = Math.sqrt(
              Math.pow(toPoint.x - bestFrom.x, 2) + 
              Math.pow(toPoint.y - bestFrom.y, 2)
            );
            if (distance < minDistance) {
              minDistance = distance;
              bestTo = toPoint;
            }
          }
        }

        let x1 = bestFrom.x;
        let y1 = bestFrom.y;
        let x2 = bestTo.x;
        let y2 = bestTo.y;

        // Apply shortening if specified
        if (node.props.shortenStart || node.props.shortenEnd) {
          const dx = x2 - x1;
          const dy = y2 - y1;
          const length = Math.sqrt(dx * dx + dy * dy);
          const dirX = dx / length;
          const dirY = dy / length;

          if (node.props.shortenStart) {
            x1 += dirX * node.props.shortenStart;
            y1 += dirY * node.props.shortenStart;
          }

          if (node.props.shortenEnd) {
            x2 -= dirX * node.props.shortenEnd;
            y2 -= dirY * node.props.shortenEnd;
          }
        }

        // Handle thickness option
        let finalStrokeWidth = strokeWidth;
        if (node.props.thickness) {
          const thicknessMap: Record<string, number> = { thin: 1, medium: 2, thick: 3, 'very-thick': 5 };
          finalStrokeWidth = thicknessMap[node.props.thickness as string] || strokeWidth;
        }

        // Determine stroke dash array based on style
        let strokeDashArray = '';
        if (style === 'dashed') {
          strokeDashArray = `stroke-dasharray="8,4" `;
        } else if (style === 'dotted') {
          strokeDashArray = `stroke-dasharray="2,4" `;
        } else if (style === 'wave') {
          // Wave style uses a wavy path instead of dash array
          strokeDashArray = '';
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
        } else if (curve === 'arc') {
          // Arc curve
          const dx = x2 - x1;
          const dy = y2 - y1;
          const distance = Math.sqrt(dx * dx + dy * dy);
          const radius = distance * 0.5;
          const sweep = dy > 0 ? 1 : 0;
          pathData = `M ${x1} ${y1} A ${radius} ${radius} 0 0 ${sweep} ${x2} ${y2}`;
        }

        // Handle bidirectional arrows
        let finalHeadType = headType;
        let finalTailType = tailType;
        if (node.props.bidirectional) {
          finalTailType = finalHeadType; // Make tail same as head for bidirectional
        }

        // Generate marker IDs
        const colorKey = color.replace('#', '');
        const headMarker = finalHeadType !== 'none' ? `url(#${finalHeadType}-${colorKey})` : '';
        const tailMarker = finalTailType !== 'none' ? `url(#${finalTailType}-tail-${colorKey})` : '';

        // Add animation if requested
        const animationAttr = node.props.animated 
          ? `><animate attributeName="stroke-dashoffset" from="0" to="-20" dur="1s" repeatCount="indefinite" /></path` 
          : '';

        // Draw path
        svg += `\n${indent}<path d="${pathData}" `;
        svg += `stroke="${color}" stroke-width="${finalStrokeWidth}" fill="none" ${strokeDashArray}`;
        if (headMarker) svg += `marker-end="${headMarker}" `;
        if (tailMarker) svg += `marker-start="${tailMarker}" `;
        svg += animationAttr || '/>';

        // Helper function to render a label
        const renderLabel = (text: string, posX: number, posY: number) => {
          const labelText = this.escapeXml(text);
          const fontSize = 12;
          const textWidth = labelText.length * fontSize * 0.6;
          const textHeight = fontSize;
          const padding = 4;
          
          svg += `\n${indent}<rect x="${posX - textWidth / 2 - padding}" y="${posY - textHeight / 2 - padding}" `;
          svg += `width="${textWidth + padding * 2}" height="${textHeight + padding * 2}" `;
          svg += `fill="white" fill-opacity="0.85" rx="2" />`;
          
          svg += `\n${indent}<text x="${posX}" y="${posY + fontSize / 3}" `;
          svg += `font-size="${fontSize}" fill="${color}" text-anchor="middle">${labelText}</text>`;
        };

        // Add labels
        if (node.props.startLabel) {
          const startX = x1 + (x2 - x1) * 0.15;
          const startY = y1 + (y2 - y1) * 0.15;
          renderLabel(node.props.startLabel, startX, startY);
        }

        if (node.props.label) {
          const midX = (x1 + x2) / 2;
          const midY = (y1 + y2) / 2;
          renderLabel(node.props.label, midX, midY);
        }

        if (node.props.endLabel) {
          const endX = x1 + (x2 - x1) * 0.85;
          const endY = y1 + (y2 - y1) * 0.85;
          renderLabel(node.props.endLabel, endX, endY);
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
        } else if (headType === 'square') {
          // Square marker
          markers += `\n    <marker id="square-${colorId}" markerWidth="8" markerHeight="8" refX="4" refY="4" orient="auto">`;
          markers += `\n      <rect x="1" y="1" width="6" height="6" fill="${color}" />`;
          markers += `\n    </marker>`;
          
          // Tail square
          markers += `\n    <marker id="square-tail-${colorId}" markerWidth="8" markerHeight="8" refX="4" refY="4" orient="auto">`;
          markers += `\n      <rect x="1" y="1" width="6" height="6" fill="${color}" />`;
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
      let headType = node.props.headType || 'arrow';
      let tailType = node.props.tailType || 'none';
      
      // Handle bidirectional arrows
      if (node.props.bidirectional) {
        tailType = headType;
      }
      
      if (!configs.has(color)) {
        configs.set(color, new Set());
      }
      const types = configs.get(color)!;
      
      if (headType !== 'none') types.add(headType);
      if (tailType !== 'none') types.add(tailType);
    }
    node.children.forEach(child => this.collectArrowConfigs(child, configs));
  }

  private renderStateNode(node: LayoutNode, x: number, y: number, width: number, height: number, indent: string): string {
    let svg = '';
    const { label, stateType = 'default', icon } = node.props;
    const bgColor = node.props.backgroundColor || (stateType === 'initial' ? '#e3f2fd' : 
                                                    stateType === 'active' ? '#fff3e0' :
                                                    stateType === 'final' ? '#e8f5e9' : '#f5f5f5');
    const borderColor = node.props.borderColor || (stateType === 'initial' ? '#2196f3' :
                                                     stateType === 'active' ? '#ff9800' :
                                                     stateType === 'final' ? '#4caf50' : '#9e9e9e');
    const borderWidth = node.props.borderWidth || 2;
    const borderRadius = node.props.borderRadius || 8;

    // Draw rounded rectangle
    svg += `\n${indent}<rect x="${x}" y="${y}" width="${width}" height="${height}" `;
    svg += `fill="${bgColor}" stroke="${borderColor}" stroke-width="${borderWidth}" rx="${borderRadius}" />`;

    // Draw icon if present
    let textY = y + height / 2;
    if (icon) {
      const iconSize = 20;
      const iconX = x + 10;
      const iconY = y + height / 2 - iconSize / 2;
      svg += `\n${indent}<text x="${iconX}" y="${iconY + iconSize}" font-size="${iconSize}" text-anchor="start">${icon}</text>`;
      
      // Draw label next to icon
      const labelX = iconX + iconSize + 8;
      svg += `\n${indent}<text x="${labelX}" y="${y + height / 2 + 5}" font-size="14" text-anchor="start" font-weight="500">${this.escapeXml(label)}</text>`;
    } else {
      // Draw centered label
      svg += `\n${indent}<text x="${x + width / 2}" y="${y + height / 2 + 5}" font-size="14" text-anchor="middle" font-weight="500">${this.escapeXml(label)}</text>`;
    }

    return svg;
  }

  private renderDecisionNode(node: LayoutNode, x: number, y: number, width: number, height: number, indent: string): string {
    let svg = '';
    const { label } = node.props;
    const bgColor = node.props.backgroundColor || '#fff9c4';
    const borderColor = node.props.borderColor || '#f57f17';
    const borderWidth = node.props.borderWidth || 2;

    // Draw diamond shape
    const centerX = x + width / 2;
    const centerY = y + height / 2;
    const points = `${centerX},${y} ${x + width},${centerY} ${centerX},${y + height} ${x},${centerY}`;
    
    svg += `\n${indent}<polygon points="${points}" fill="${bgColor}" stroke="${borderColor}" stroke-width="${borderWidth}" />`;
    
    // Draw label
    svg += `\n${indent}<text x="${centerX}" y="${centerY + 5}" font-size="12" text-anchor="middle" font-weight="500">${this.escapeXml(label)}</text>`;

    return svg;
  }

  private renderProcessNode(node: LayoutNode, x: number, y: number, width: number, height: number, indent: string): string {
    let svg = '';
    const { label, nodeType = 'process', status = 'default' } = node.props;
    
    let bgColor = node.props.backgroundColor;
    let borderColor = node.props.borderColor;
    let shape = 'rectangle';

    // Determine colors based on status
    if (!bgColor) {
      bgColor = status === 'active' ? '#e3f2fd' :
                status === 'complete' ? '#e8f5e9' :
                status === 'error' ? '#ffebee' :
                '#f5f5f5';
    }
    
    if (!borderColor) {
      borderColor = status === 'active' ? '#2196f3' :
                    status === 'complete' ? '#4caf50' :
                    status === 'error' ? '#f44336' :
                    '#9e9e9e';
    }

    const borderWidth = node.props.borderWidth || 2;

    // Draw based on node type
    if (nodeType === 'start' || nodeType === 'end') {
      // Rounded/oval shape
      const rx = width / 2;
      const ry = height / 2;
      svg += `\n${indent}<ellipse cx="${x + width / 2}" cy="${y + height / 2}" rx="${rx}" ry="${ry}" `;
      svg += `fill="${bgColor}" stroke="${borderColor}" stroke-width="${borderWidth}" />`;
    } else {
      // Rectangle with rounded corners
      const borderRadius = node.props.borderRadius || (nodeType === 'subprocess' ? 12 : 6);
      svg += `\n${indent}<rect x="${x}" y="${y}" width="${width}" height="${height}" `;
      svg += `fill="${bgColor}" stroke="${borderColor}" stroke-width="${borderWidth}" rx="${borderRadius}" />`;
      
      // Add double border for subprocess
      if (nodeType === 'subprocess') {
        const innerPadding = 4;
        svg += `\n${indent}<rect x="${x + innerPadding}" y="${y + innerPadding}" width="${width - innerPadding * 2}" height="${height - innerPadding * 2}" `;
        svg += `fill="none" stroke="${borderColor}" stroke-width="1" rx="${borderRadius - 2}" />`;
      }
    }

    // Draw label
    svg += `\n${indent}<text x="${x + width / 2}" y="${y + height / 2 + 5}" font-size="13" text-anchor="middle" font-weight="500">${this.escapeXml(label)}</text>`;

    // Add status indicator
    if (status === 'active') {
      svg += `\n${indent}<circle cx="${x + width - 8}" cy="${y + 8}" r="4" fill="#ff9800" />`;
    }

    return svg;
  }

  private renderMemoryBlock(node: LayoutNode, x: number, y: number, width: number, height: number, indent: string): string {
    let svg = '';
    const { label, capacity, used, unit = 'tokens', showBar = true, showPercentage = true } = node.props;
    const bgColor = node.props.backgroundColor || '#f5f5f5';
    const borderColor = node.props.borderColor || '#757575';
    const borderWidth = node.props.borderWidth || 2;
    const borderRadius = node.props.borderRadius || 6;

    // Draw container
    svg += `\n${indent}<rect x="${x}" y="${y}" width="${width}" height="${height}" `;
    svg += `fill="${bgColor}" stroke="${borderColor}" stroke-width="${borderWidth}" rx="${borderRadius}" />`;

    // Draw label
    const labelY = y + 20;
    svg += `\n${indent}<text x="${x + width / 2}" y="${labelY}" font-size="14" text-anchor="middle" font-weight="bold">${this.escapeXml(label)}</text>`;

    if (showBar) {
      // Draw capacity bar
      const barY = y + 35;
      const barHeight = 20;
      const barPadding = 10;
      const barWidth = width - barPadding * 2;
      const fillWidth = (used / capacity) * barWidth;
      
      // Background bar
      svg += `\n${indent}<rect x="${x + barPadding}" y="${barY}" width="${barWidth}" height="${barHeight}" `;
      svg += `fill="#e0e0e0" stroke="#bdbdbd" stroke-width="1" rx="4" />`;
      
      // Filled portion
      const fillColor = used / capacity > 0.9 ? '#f44336' : used / capacity > 0.7 ? '#ff9800' : '#4caf50';
      svg += `\n${indent}<rect x="${x + barPadding}" y="${barY}" width="${fillWidth}" height="${barHeight}" `;
      svg += `fill="${fillColor}" rx="4" />`;
    }

    // Draw stats
    const statsY = y + (showBar ? 70 : 40);
    const percentage = ((used / capacity) * 100).toFixed(1);
    const statsText = showPercentage 
      ? `${used.toLocaleString()} / ${capacity.toLocaleString()} ${unit} (${percentage}%)`
      : `${used.toLocaleString()} / ${capacity.toLocaleString()} ${unit}`;
    
    svg += `\n${indent}<text x="${x + width / 2}" y="${statsY}" font-size="12" text-anchor="middle" fill="#616161">${this.escapeXml(statsText)}</text>`;

    return svg;
  }

  private renderContextWindow(node: LayoutNode, x: number, y: number, width: number, height: number, indent: string): string {
    let svg = '';
    const { capacity, sections, showLabels = true, showPercentages = true, orientation = 'horizontal' } = node.props;
    const borderColor = node.props.borderColor || '#757575';
    const borderWidth = node.props.borderWidth || 2;
    const borderRadius = node.props.borderRadius || 6;

    // Draw outer container
    svg += `\n${indent}<rect x="${x}" y="${y}" width="${width}" height="${height}" `;
    svg += `fill="white" stroke="${borderColor}" stroke-width="${borderWidth}" rx="${borderRadius}" />`;

    // Calculate section dimensions
    let currentPos = orientation === 'horizontal' ? x : y;
    const totalTokens = sections.reduce((sum: number, s: any) => sum + s.tokens, 0);

    sections.forEach((section: any, index: number) => {
      const ratio = section.tokens / capacity;
      const sectionSize = orientation === 'horizontal' 
        ? ratio * (width - 4) 
        : ratio * (height - 4);

      if (orientation === 'horizontal') {
        // Draw horizontal section
        svg += `\n${indent}<rect x="${currentPos + 2}" y="${y + 2}" width="${sectionSize}" height="${height - 4}" `;
        svg += `fill="${section.color}" fill-opacity="0.7" />`;

        if (showLabels && sectionSize > 40) {
          const labelX = currentPos + sectionSize / 2;
          const labelY = y + height / 2;
          const percentage = ((section.tokens / capacity) * 100).toFixed(1);
          
          svg += `\n${indent}<text x="${labelX}" y="${labelY - 5}" font-size="11" text-anchor="middle" font-weight="500">${this.escapeXml(section.label)}</text>`;
          
          if (showPercentages) {
            svg += `\n${indent}<text x="${labelX}" y="${labelY + 10}" font-size="10" text-anchor="middle" fill="#424242">${section.tokens} (${percentage}%)</text>`;
          }
        }

        currentPos += sectionSize;
      } else {
        // Draw vertical section
        svg += `\n${indent}<rect x="${x + 2}" y="${currentPos + 2}" width="${width - 4}" height="${sectionSize}" `;
        svg += `fill="${section.color}" fill-opacity="0.7" />`;

        if (showLabels && sectionSize > 30) {
          const labelX = x + width / 2;
          const labelY = currentPos + sectionSize / 2;
          const percentage = ((section.tokens / capacity) * 100).toFixed(1);
          
          svg += `\n${indent}<text x="${labelX}" y="${labelY}" font-size="11" text-anchor="middle" font-weight="500">${this.escapeXml(section.label)}</text>`;
          
          if (showPercentages && sectionSize > 50) {
            svg += `\n${indent}<text x="${labelX}" y="${labelY + 15}" font-size="10" text-anchor="middle" fill="#424242">${section.tokens} (${percentage}%)</text>`;
          }
        }

        currentPos += sectionSize;
      }
    });

    return svg;
  }

  private renderTimeline(node: LayoutNode, x: number, y: number, width: number, height: number, indent: string): string {
    let svg = '';
    const { orientation = 'horizontal', showAxis = true } = node.props;
    const lineColor = '#757575';
    
    if (showAxis) {
      if (orientation === 'horizontal') {
        // Draw horizontal axis line
        const lineY = y + height / 2;
        svg += `\n${indent}<line x1="${x}" y1="${lineY}" x2="${x + width}" y2="${lineY}" stroke="${lineColor}" stroke-width="2" />`;
      } else {
        // Draw vertical axis line
        const lineX = x + width / 2;
        svg += `\n${indent}<line x1="${lineX}" y1="${y}" x2="${lineX}" y2="${y + height}" stroke="${lineColor}" stroke-width="2" />`;
      }
    }

    return svg;
  }

  private renderTimelineEvent(node: LayoutNode, x: number, y: number, width: number, height: number, indent: string): string {
    let svg = '';
    const { label, description, color = '#2196f3', icon } = node.props;
    const circleRadius = 8;

    // Draw event marker circle
    const centerX = x + width / 2;
    const centerY = y + circleRadius + 5;
    
    svg += `\n${indent}<circle cx="${centerX}" cy="${centerY}" r="${circleRadius}" fill="${color}" stroke="white" stroke-width="2" />`;

    // Draw icon in circle if present
    if (icon) {
      svg += `\n${indent}<text x="${centerX}" y="${centerY + 4}" font-size="10" text-anchor="middle" fill="white">${icon}</text>`;
    }

    // Draw label below
    const labelY = centerY + circleRadius + 15;
    svg += `\n${indent}<text x="${centerX}" y="${labelY}" font-size="12" text-anchor="middle" font-weight="500">${this.escapeXml(label)}</text>`;

    // Draw description if present
    if (description) {
      const descY = labelY + 15;
      svg += `\n${indent}<text x="${centerX}" y="${descY}" font-size="10" text-anchor="middle" fill="#757575">${this.escapeXml(description)}</text>`;
    }

    return svg;
  }
}
