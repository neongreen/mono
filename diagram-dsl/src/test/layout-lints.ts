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
