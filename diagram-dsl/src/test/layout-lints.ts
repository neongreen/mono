import { LayoutNode } from '../types';

export interface LayoutLint {
  type: 'warning' | 'info';
  message: string;
  elementId?: string;
  details?: any;
}

export class LayoutLinter {
  private lints: LayoutLint[] = [];

  constructor(private layout: LayoutNode) {}

  /**
   * Run all linting checks and return the results
   */
  runAllLints(): LayoutLint[] {
    this.lints = [];
    
    // Run various lint checks
    this.checkShortArrows(this.layout);
    this.checkInternalVsExternalSpacing(this.layout);
    this.checkOverlappingElements(this.layout);
    this.checkMinimumFontSize(this.layout);
    this.checkInconsistentSpacing(this.layout);
    this.checkArrowCrossings(this.layout);
    
    return this.lints;
  }

  /**
   * Get all lints
   */
  getLints(): LayoutLint[] {
    return this.lints;
  }

  /**
   * Check for arrows that are too short (almost as short as the marker)
   * Minimum recommended arrow length is ~20px (marker is ~10px)
   */
  private checkShortArrows(node: LayoutNode): void {
    if (node.type === 'Arrow') {
      const fromId = node.props.from;
      const toId = node.props.to;
      
      const fromNode = this.findById(this.layout, fromId);
      const toNode = this.findById(this.layout, toId);
      
      if (fromNode?.computed && toNode?.computed) {
        const arrowLength = this.calculateArrowLength(fromNode.computed, toNode.computed);
        
        // Marker is about 10px, so arrows shorter than 20px are problematic
        const minRecommendedLength = 20;
        
        if (arrowLength < minRecommendedLength) {
          this.lints.push({
            type: 'warning',
            message: `Arrow from "${fromId}" to "${toId}" is very short (${arrowLength.toFixed(1)}px). Consider increasing spacing between elements (minimum recommended: ${minRecommendedLength}px).`,
            details: {
              from: fromId,
              to: toId,
              length: arrowLength,
              minRecommended: minRecommendedLength
            }
          });
        }
      }
    }
    
    // Recursively check children
    node.children.forEach(child => this.checkShortArrows(child));
  }

  /**
   * Calculate the actual arrow length between two boxes
   */
  private calculateArrowLength(
    fromBox: { x: number; y: number; width: number; height: number },
    toBox: { x: number; y: number; width: number; height: number }
  ): number {
    // Calculate centers
    const fromCenterX = fromBox.x + fromBox.width / 2;
    const fromCenterY = fromBox.y + fromBox.height / 2;
    const toCenterX = toBox.x + toBox.width / 2;
    const toCenterY = toBox.y + toBox.height / 2;

    // Calculate attachment points (same logic as SVGRenderer)
    const fromAttachmentPoints = {
      top: { x: fromCenterX, y: fromBox.y },
      bottom: { x: fromCenterX, y: fromBox.y + fromBox.height },
      left: { x: fromBox.x, y: fromCenterY },
      right: { x: fromBox.x + fromBox.width, y: fromCenterY }
    };

    const toAttachmentPoints = {
      top: { x: toCenterX, y: toBox.y },
      bottom: { x: toCenterX, y: toBox.y + toBox.height },
      left: { x: toBox.x, y: toCenterY },
      right: { x: toBox.x + toBox.width, y: toCenterY }
    };

    // Find closest attachment points
    let minDistance = Infinity;

    for (const fromPoint of Object.values(fromAttachmentPoints)) {
      for (const toPoint of Object.values(toAttachmentPoints)) {
        const distance = Math.sqrt(
          Math.pow(toPoint.x - fromPoint.x, 2) + 
          Math.pow(toPoint.y - fromPoint.y, 2)
        );
        if (distance < minDistance) {
          minDistance = distance;
        }
      }
    }

    return minDistance;
  }

  /**
   * Check that internal spacing (border to content) is smaller than external spacing (border to other elements)
   * This applies to vertical stacks where we check gap between boxes vs internal padding
   */
  private checkInternalVsExternalSpacing(node: LayoutNode): void {
    // Only check Stack nodes (vertical stacks)
    if (node.type === 'Stack' && node.children.length >= 2) {
      // Check if this is a vertical stack by examining the layout direction
      const firstChild = node.children[0];
      const secondChild = node.children[1];
      
      if (firstChild?.computed && secondChild?.computed) {
        // If children are stacked vertically (second child's Y is greater than first)
        const isVerticalStack = secondChild.computed.y > firstChild.computed.y + firstChild.computed.height - 1;
        
        if (isVerticalStack) {
          // Check gaps between consecutive children
          for (let i = 0; i < node.children.length - 1; i++) {
            const child1 = node.children[i];
            const child2 = node.children[i + 1];
            
            if (child1.computed && child2.computed) {
              // Calculate external gap (distance between two boxes)
              const externalGap = child2.computed.y - (child1.computed.y + child1.computed.height);
              
              // Check internal spacing for both children
              this.checkChildInternalSpacing(child1, externalGap, i);
              this.checkChildInternalSpacing(child2, externalGap, i + 1);
            }
          }
        }
      }
    }
    
    // Recursively check children
    node.children.forEach(child => this.checkInternalVsExternalSpacing(child));
  }

  /**
   * Check if a child's internal spacing (padding to content) is larger than external gap
   */
  private checkChildInternalSpacing(child: LayoutNode, externalGap: number, index: number): void {
    if (!child.computed) return;
    
    // Only check boxes that have visual borders
    const hasVisualBorder = child.props.borderWidth && child.props.borderWidth > 0;
    if (!hasVisualBorder && child.type !== 'Box') return;
    
    // Calculate padding (use explicit padding or default to 0)
    const padding = child.props.padding || 0;
    const paddingTop = child.props.paddingTop ?? padding;
    const paddingBottom = child.props.paddingBottom ?? padding;
    
    // Get the minimum internal spacing (smallest distance from border to content)
    const minInternalSpacing = Math.min(paddingTop, paddingBottom);
    
    // Check if we have any content children to make this check meaningful
    const hasContentChildren = this.hasTextualContent(child);
    
    if (hasContentChildren && minInternalSpacing > externalGap && externalGap > 0) {
      const childId = child.props.id || `child-${index}`;
      
      this.lints.push({
        type: 'warning',
        message: `Box "${childId}" has internal spacing (${minInternalSpacing}px) > external gap (${externalGap.toFixed(1)}px). Internal distances should be smaller than external distances for better visual grouping.`,
        elementId: childId,
        details: {
          internalSpacing: minInternalSpacing,
          externalGap: externalGap,
          recommendation: `Consider increasing the gap between boxes (current: ${externalGap.toFixed(1)}px, recommend: >${minInternalSpacing}px) or reducing internal padding.`
        }
      });
    }
  }

  /**
   * Check if a node or its descendants have textual content
   */
  private hasTextualContent(node: LayoutNode): boolean {
    if (node.type === 'Text' || node.type === 'Label' || node.type === 'Title' || node.type === 'Subtitle') {
      return true;
    }
    return node.children.some(child => this.hasTextualContent(child));
  }

  /**
   * Find a node by its ID property
   */
  private findById(node: LayoutNode, id: string): LayoutNode | null {
    if (node.props.id === id) {
      return node;
    }
    for (const child of node.children) {
      const found = this.findById(child, id);
      if (found) return found;
    }
    return null;
  }

  /**
   * Check for overlapping elements that might obscure each other
   */
  private checkOverlappingElements(node: LayoutNode): void {
    const elements: Array<{ id: string; computed: NonNullable<LayoutNode['computed']>; type: string }> = [];
    
    // Collect all elements with computed positions (except arrows and text)
    this.collectPositionedElements(node, elements);
    
    // Check each pair for overlap
    for (let i = 0; i < elements.length; i++) {
      for (let j = i + 1; j < elements.length; j++) {
        const elem1 = elements[i];
        const elem2 = elements[j];
        
        if (this.isOverlapping(elem1.computed, elem2.computed)) {
          this.lints.push({
            type: 'warning',
            message: `Elements "${elem1.id}" and "${elem2.id}" are overlapping. This may obscure content and create visual confusion.`,
            elementId: elem1.id,
            details: {
              element1: elem1.id,
              element2: elem2.id,
              element1Bounds: elem1.computed,
              element2Bounds: elem2.computed
            }
          });
        }
      }
    }
  }

  /**
   * Collect all elements with IDs and computed positions
   */
  private collectPositionedElements(
    node: LayoutNode, 
    elements: Array<{ id: string; computed: NonNullable<LayoutNode['computed']>; type: string }>
  ): void {
    // Only collect boxes/cards with IDs and computed positions, skip text and arrows
    if (node.props.id && node.computed && 
        (node.type === 'Box' || node.type === 'Card' || node.type === 'Stack' || node.type === 'Row')) {
      elements.push({
        id: node.props.id,
        computed: node.computed,
        type: node.type
      });
    }
    
    node.children.forEach(child => this.collectPositionedElements(child, elements));
  }

  /**
   * Check if two rectangular bounds overlap
   */
  private isOverlapping(
    box1: { x: number; y: number; width: number; height: number },
    box2: { x: number; y: number; width: number; height: number }
  ): boolean {
    // Add a small tolerance (1px) to avoid false positives from floating point errors
    const tolerance = 1;
    
    return !(
      box1.x + box1.width - tolerance < box2.x ||
      box2.x + box2.width - tolerance < box1.x ||
      box1.y + box1.height - tolerance < box2.y ||
      box2.y + box2.height - tolerance < box1.y
    );
  }

  /**
   * Check for text that's too small to read comfortably
   */
  private checkMinimumFontSize(node: LayoutNode): void {
    const minRecommendedSize = 10; // 10px is generally the minimum for readability
    
    if (node.type === 'Text' && node.props.fontSize) {
      if (node.props.fontSize < minRecommendedSize) {
        this.lints.push({
          type: 'info',
          message: `Text has very small font size (${node.props.fontSize}px). Minimum recommended: ${minRecommendedSize}px for readability.`,
          details: {
            fontSize: node.props.fontSize,
            minRecommended: minRecommendedSize,
            text: typeof node.props.children === 'string' ? node.props.children.substring(0, 50) : 'N/A'
          }
        });
      }
    }
    
    node.children.forEach(child => this.checkMinimumFontSize(child));
  }

  /**
   * Check for inconsistent spacing patterns in containers
   */
  private checkInconsistentSpacing(node: LayoutNode): void {
    // Only check Stack and Row containers with multiple children
    if ((node.type === 'Stack' || node.type === 'Row') && node.children.length >= 3) {
      const gaps: number[] = [];
      
      // Calculate gaps between consecutive children
      for (let i = 0; i < node.children.length - 1; i++) {
        const child1 = node.children[i];
        const child2 = node.children[i + 1];
        
        if (child1.computed && child2.computed) {
          let gap: number;
          
          // For vertical stacks, measure Y distance
          if (node.type === 'Stack') {
            gap = child2.computed.y - (child1.computed.y + child1.computed.height);
          } else {
            // For horizontal rows, measure X distance
            gap = child2.computed.x - (child1.computed.x + child1.computed.width);
          }
          
          if (gap >= 0) {
            gaps.push(gap);
          }
        }
      }
      
      // Check for variance in gaps (more than 50% difference is notable)
      if (gaps.length >= 2) {
        const minGap = Math.min(...gaps);
        const maxGap = Math.max(...gaps);
        
        // Skip if gaps are very small (might be intentional tight spacing)
        if (minGap > 5 && maxGap > minGap * 1.5) {
          const containerId = node.props.id || `${node.type}-container`;
          
          this.lints.push({
            type: 'info',
            message: `Container "${containerId}" has inconsistent spacing between children (${minGap.toFixed(1)}px to ${maxGap.toFixed(1)}px). Consider using uniform gaps for visual consistency.`,
            elementId: containerId,
            details: {
              minGap: minGap,
              maxGap: maxGap,
              gaps: gaps.map(g => Math.round(g * 10) / 10),
              recommendation: 'Use consistent gap values for better visual rhythm'
            }
          });
        }
      }
    }
    
    node.children.forEach(child => this.checkInconsistentSpacing(child));
  }

  /**
   * Check for arrows that cross each other, which can create visual confusion
   */
  private checkArrowCrossings(node: LayoutNode): void {
    const arrows: Array<{
      from: string;
      to: string;
      fromPos: { x: number; y: number };
      toPos: { x: number; y: number };
    }> = [];
    
    // Collect all arrows with their positions
    this.collectArrows(node, arrows);
    
    // Check each pair of arrows for crossings
    for (let i = 0; i < arrows.length; i++) {
      for (let j = i + 1; j < arrows.length; j++) {
        const arrow1 = arrows[i];
        const arrow2 = arrows[j];
        
        // Skip if arrows share endpoints (they're supposed to connect)
        if (arrow1.from === arrow2.from || arrow1.from === arrow2.to ||
            arrow1.to === arrow2.from || arrow1.to === arrow2.to) {
          continue;
        }
        
        if (this.doLineSegmentsIntersect(
          arrow1.fromPos, arrow1.toPos,
          arrow2.fromPos, arrow2.toPos
        )) {
          this.lints.push({
            type: 'info',
            message: `Arrows "${arrow1.from}→${arrow1.to}" and "${arrow2.from}→${arrow2.to}" are crossing. Consider rearranging elements to avoid crossed connections.`,
            details: {
              arrow1: `${arrow1.from} → ${arrow1.to}`,
              arrow2: `${arrow2.from} → ${arrow2.to}`,
              recommendation: 'Rearrange elements or use different layout to avoid arrow crossings'
            }
          });
        }
      }
    }
  }

  /**
   * Collect all arrows with their computed positions
   */
  private collectArrows(
    node: LayoutNode,
    arrows: Array<{
      from: string;
      to: string;
      fromPos: { x: number; y: number };
      toPos: { x: number; y: number };
    }>
  ): void {
    if (node.type === 'Arrow') {
      const fromNode = this.findById(this.layout, node.props.from);
      const toNode = this.findById(this.layout, node.props.to);
      
      if (fromNode?.computed && toNode?.computed) {
        // Use center points for crossing detection
        arrows.push({
          from: node.props.from,
          to: node.props.to,
          fromPos: {
            x: fromNode.computed.x + fromNode.computed.width / 2,
            y: fromNode.computed.y + fromNode.computed.height / 2
          },
          toPos: {
            x: toNode.computed.x + toNode.computed.width / 2,
            y: toNode.computed.y + toNode.computed.height / 2
          }
        });
      }
    }
    
    node.children.forEach(child => this.collectArrows(child, arrows));
  }

  /**
   * Check if two line segments intersect
   * Uses the line segment intersection algorithm
   */
  private doLineSegmentsIntersect(
    p1: { x: number; y: number },
    p2: { x: number; y: number },
    p3: { x: number; y: number },
    p4: { x: number; y: number }
  ): boolean {
    // Calculate direction of the cross products
    const d1 = this.direction(p3, p4, p1);
    const d2 = this.direction(p3, p4, p2);
    const d3 = this.direction(p1, p2, p3);
    const d4 = this.direction(p1, p2, p4);
    
    if (((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
        ((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0))) {
      return true;
    }
    
    return false;
  }

  /**
   * Calculate the direction/orientation of point c relative to line segment ab
   */
  private direction(
    a: { x: number; y: number },
    b: { x: number; y: number },
    c: { x: number; y: number }
  ): number {
    return (b.x - a.x) * (c.y - a.y) - (b.y - a.y) * (c.x - a.x);
  }

  /**
   * Format all lints as a readable string
   */
  static formatLints(lints: LayoutLint[]): string {
    if (lints.length === 0) {
      return 'No layout issues found.';
    }

    let output = `\nLayout Lints (${lints.length}):\n`;
    output += '='.repeat(50) + '\n';
    
    lints.forEach((lint, index) => {
      const icon = lint.type === 'warning' ? '⚠' : 'ℹ';
      output += `\n${index + 1}. ${icon} ${lint.message}\n`;
      
      if (lint.details) {
        output += `   Details: ${JSON.stringify(lint.details, null, 2)}\n`;
      }
    });
    
    output += '='.repeat(50) + '\n';
    
    return output;
  }
}
