import { LayoutNode } from '../types';
import { measureText } from '../layout/text-measurement';

export interface LayoutLint {
  type: 'warning' | 'info';
  message: string;
  elementId?: string;
  details?: any;
}

interface ArrowInfo {
  fromId: string;
  toId: string;
  tail: { x: number; y: number };
  arrowhead: { x: number; y: number };
  labelRect?: Rect;
}

interface Rect {
  x: number;
  y: number;
  width: number;
  height: number;
}

interface BoxInfo {
  id: string;
  bounds: Rect;
}

const TEXTUAL_TYPES = new Set(['Text', 'Label', 'Subtitle', 'Title']);

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
    const arrowInfos = this.collectArrowInfos(this.layout);
    const boxInfos = this.collectBoxInfos(this.layout);
    this.checkArrowheadSpacing(arrowInfos);
    this.checkArrowCrossings(arrowInfos);
    this.checkArrowLabelOverlaps(arrowInfos, boxInfos);
    this.checkTextOverflow(this.layout);
    this.checkTextSpacing(this.layout);
    
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
        const metrics = this.calculateArrowMetrics(fromNode.computed, toNode.computed);
        const arrowLength = metrics.length;
        
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
  private calculateArrowMetrics(
    fromBox: { x: number; y: number; width: number; height: number },
    toBox: { x: number; y: number; width: number; height: number }
  ): { length: number; fromPoint: { x: number; y: number }; toPoint: { x: number; y: number } } {
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
    let bestFrom = fromAttachmentPoints.bottom;
    let bestTo = toAttachmentPoints.top;

    for (const fromPoint of Object.values(fromAttachmentPoints)) {
      for (const toPoint of Object.values(toAttachmentPoints)) {
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

    return {
      length: minDistance,
      fromPoint: bestFrom,
      toPoint: bestTo
    };
  }

  /**
   * Check if arrowheads are overlapping or too close to each other
   */
  private checkArrowheadSpacing(arrows: ArrowInfo[]): void {
    if (arrows.length < 2) {
      return;
    }

    const minRecommendedDistance = 16; // px, roughly 1.5x arrowhead size

    for (let i = 0; i < arrows.length; i++) {
      for (let j = i + 1; j < arrows.length; j++) {
        const arrowA = arrows[i];
        const arrowB = arrows[j];

        const distance = Math.sqrt(
          Math.pow(arrowA.arrowhead.x - arrowB.arrowhead.x, 2) +
          Math.pow(arrowA.arrowhead.y - arrowB.arrowhead.y, 2)
        );

        if (distance < minRecommendedDistance) {
          this.lints.push({
            type: 'warning',
            message: `Arrowheads for "${arrowA.fromId}" → "${arrowA.toId}" and "${arrowB.fromId}" → "${arrowB.toId}" are very close (${distance.toFixed(1)}px). Increase spacing or adjust positioning so arrowheads have at least ${minRecommendedDistance}px separation.`,
            details: {
              arrows: [
                { from: arrowA.fromId, to: arrowA.toId },
                { from: arrowB.fromId, to: arrowB.toId }
              ],
              distance,
              minRecommendedDistance
            }
          });
        }
      }
    }
  }

  private collectArrowInfos(node: LayoutNode, acc: ArrowInfo[] = []): ArrowInfo[] {
    if (node.type === 'Arrow') {
      const fromId = node.props.from;
      const toId = node.props.to;
      const fromNode = this.findById(this.layout, fromId);
      const toNode = this.findById(this.layout, toId);

      if (fromNode?.computed && toNode?.computed) {
        const metrics = this.calculateArrowMetrics(fromNode.computed, toNode.computed);
        if (!metrics) {
          return acc;
        }
        const { fromPoint, toPoint } = metrics;
        const arrowInfo: ArrowInfo = {
          fromId,
          toId,
          tail: fromPoint,
          arrowhead: toPoint
        };

        if (node.props.label) {
          const labelText = String(node.props.label);
          const fontSize = 12;
          const fontFamily = 'Arial, sans-serif';
          const fontWeight = 'normal';
          const padding = 4;
          const textMetrics = measureText(labelText, fontSize, fontFamily, fontWeight);
          const boxWidth = textMetrics.width + padding * 2;
          const boxHeight = textMetrics.height + padding * 2;
          const midX = (fromPoint.x + toPoint.x) / 2;
          const midY = (fromPoint.y + toPoint.y) / 2;
          arrowInfo.labelRect = {
            x: midX - boxWidth / 2,
            y: midY - boxHeight / 2,
            width: boxWidth,
            height: boxHeight
          };
        }

        acc.push(arrowInfo);
      }
    }

    node.children.forEach(child => this.collectArrowInfos(child, acc));
    return acc;
  }

  private collectBoxInfos(node: LayoutNode, acc: BoxInfo[] = []): BoxInfo[] {
    if (node.type !== 'Arrow' && node.computed && node.props?.id) {
      acc.push({
        id: node.props.id,
        bounds: {
          x: node.computed.x,
          y: node.computed.y,
          width: node.computed.width,
          height: node.computed.height
        }
      });
    }

    node.children.forEach(child => this.collectBoxInfos(child, acc));
    return acc;
  }

  private checkArrowLabelOverlaps(arrows: ArrowInfo[], boxes: BoxInfo[]): void {
    if (boxes.length === 0) return;

    arrows.forEach(arrow => {
      if (!arrow.labelRect) return;

      boxes.forEach(box => {
        if (box.id === arrow.fromId || box.id === arrow.toId) {
          return;
        }

        if (this.rectanglesOverlap(arrow.labelRect!, box.bounds)) {
          this.lints.push({
            type: 'warning',
            message: `Arrow label for "${arrow.fromId}" → "${arrow.toId}" overlaps node "${box.id}". Nudge layout or reroute the arrow so labels remain readable.`,
            details: {
              arrow: { from: arrow.fromId, to: arrow.toId },
              node: box.id
            }
          });
        }
      });
    });
  }

  private rectanglesOverlap(a: Rect, b: Rect): boolean {
    return (
      a.x < b.x + b.width &&
      a.x + a.width > b.x &&
      a.y < b.y + b.height &&
      a.y + a.height > b.y
    );
  }

  /**
   * Detect when arrow segments intersect each other
   */
  private checkArrowCrossings(arrows: ArrowInfo[]): void {
    if (arrows.length < 2) {
      return;
    }

    for (let i = 0; i < arrows.length; i++) {
      for (let j = i + 1; j < arrows.length; j++) {
        const arrowA = arrows[i];
        const arrowB = arrows[j];

        // Ignore arrows that share a common endpoint to avoid duplicating other lint warnings
        const sharesEndpoint =
          arrowA.fromId === arrowB.fromId ||
          arrowA.fromId === arrowB.toId ||
          arrowA.toId === arrowB.fromId ||
          arrowA.toId === arrowB.toId;

        if (sharesEndpoint) {
          continue;
        }

        if (this.segmentsProperlyIntersect(arrowA.tail, arrowA.arrowhead, arrowB.tail, arrowB.arrowhead)) {
          this.lints.push({
            type: 'warning',
            message: `Arrows "${arrowA.fromId}" → "${arrowA.toId}" and "${arrowB.fromId}" → "${arrowB.toId}" cross each other. Adjust layout or reroute arrows to avoid intersecting connectors.`,
            details: {
              arrows: [
                { from: arrowA.fromId, to: arrowA.toId },
                { from: arrowB.fromId, to: arrowB.toId }
              ]
            }
          });
        }
      }
    }
  }

  private segmentsProperlyIntersect(
    a1: { x: number; y: number },
    a2: { x: number; y: number },
    b1: { x: number; y: number },
    b2: { x: number; y: number }
  ): boolean {
    const epsilon = 1e-2;

    const denominator = (b2.y - b1.y) * (a2.x - a1.x) - (b2.x - b1.x) * (a2.y - a1.y);
    if (Math.abs(denominator) < epsilon) {
      // Parallel or coincident
      return false;
    }

    const t = ((b2.x - b1.x) * (a1.y - b1.y) - (b2.y - b1.y) * (a1.x - b1.x)) / denominator;
    const u = ((a2.x - a1.x) * (a1.y - b1.y) - (a2.y - a1.y) * (a1.x - b1.x)) / denominator;

    // We only care about intersections strictly inside both segments (no endpoints)
    if (t <= epsilon || t >= 1 - epsilon || u <= epsilon || u >= 1 - epsilon) {
      return false;
    }

    return true;
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

  private checkTextOverflow(node: LayoutNode, parent?: LayoutNode): void {
    if (TEXTUAL_TYPES.has(node.type) && node.computed && parent?.computed) {
      const availableWidth = Math.max(
        parent.computed.width - this.getPadding(parent.props, 'Left') - this.getPadding(parent.props, 'Right'),
        0
      );

      if (node.computed.width > availableWidth + 0.5) {
        const parentId = parent.props?.id ?? parent.type;
        const contentId = node.props?.id ?? node.type;
        this.lints.push({
          type: 'warning',
          message: `Text "${contentId}" exceeds available width inside "${parentId}". Reduce copy length, adjust padding, or widen the container.`,
          details: {
            textWidth: node.computed.width,
            availableWidth,
            parent: parentId
          }
        });
      }
    }

    node.children.forEach(child => this.checkTextOverflow(child, node));
  }

  private getPadding(props: any, side: 'Left' | 'Right'): number {
    if (!props) return 0;
    const base = typeof props.padding === 'number' ? props.padding : 0;
    const specific = props[`padding${side}`];
    return typeof specific === 'number' ? specific : base;
  }

  private checkTextSpacing(node: LayoutNode): void {
    if (node.children.length >= 2) {
      const textChildren = node.children
        .filter(child => child.computed && TEXTUAL_TYPES.has(child.type))
        .sort((a, b) => a.computed!.y - b.computed!.y);

      const minGap = 6; // px, matches default Stack gap for label/subtitle pairs

      for (let i = 0; i < textChildren.length - 1; i++) {
        const first = textChildren[i];
        const second = textChildren[i + 1];

        const gap = second.computed!.y - (first.computed!.y + first.computed!.height);

        if (gap < minGap) {
          const parentId = node.props?.id ?? node.type;
          const firstId = first.props?.id ?? first.type;
          const secondId = second.props?.id ?? second.type;
          this.lints.push({
            type: 'warning',
            message: `Text nodes "${firstId}" and "${secondId}" inside "${parentId}" are only ${gap.toFixed(1)}px apart. Use a Stack with gap or adjust layout so they have at least ${minGap}px separation.`,
            details: {
              parent: parentId,
              first: firstId,
              second: secondId,
              gap,
              minGap
            }
          });
        }
      }
    }

    node.children.forEach(child => this.checkTextSpacing(child));
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
