import { LayoutNode } from '../types';

export class LayoutAssertions {
  constructor(private layout: LayoutNode) {}

  /**
   * Find a node by its ID property
   */
  findById(id: string): LayoutNode | null {
    return this.findNodeById(this.layout, id);
  }

  private findNodeById(node: LayoutNode, id: string): LayoutNode | null {
    if (node.props.id === id) {
      return node;
    }
    for (const child of node.children) {
      const found = this.findNodeById(child, id);
      if (found) return found;
    }
    return null;
  }

  /**
   * Find all nodes of a specific type
   */
  findByType(type: string): LayoutNode[] {
    const results: LayoutNode[] = [];
    this.findNodesByType(this.layout, type, results);
    return results;
  }

  private findNodesByType(node: LayoutNode, type: string, results: LayoutNode[]): void {
    if (node.type === type) {
      results.push(node);
    }
    for (const child of node.children) {
      this.findNodesByType(child, type, results);
    }
  }

  /**
   * Assert that a child element is centered within a container
   */
  assertCentered(childId: string, containerId: string, axis: 'x' | 'y' | 'both' = 'both', tolerance: number = 1): void {
    const child = this.findById(childId);
    const container = this.findById(containerId);

    if (!child || !child.computed) {
      throw new Error(`Child element with id "${childId}" not found or has no computed layout`);
    }
    if (!container || !container.computed) {
      throw new Error(`Container element with id "${containerId}" not found or has no computed layout`);
    }

    if (axis === 'x' || axis === 'both') {
      const childCenterX = child.computed.x + child.computed.width / 2;
      const containerCenterX = container.computed.x + container.computed.width / 2;
      const diffX = Math.abs(childCenterX - containerCenterX);
      
      if (diffX > tolerance) {
        throw new Error(
          `Element "${childId}" is not centered horizontally in "${containerId}". ` +
          `Child center X: ${childCenterX}, Container center X: ${containerCenterX}, Diff: ${diffX}`
        );
      }
    }

    if (axis === 'y' || axis === 'both') {
      const childCenterY = child.computed.y + child.computed.height / 2;
      const containerCenterY = container.computed.y + container.computed.height / 2;
      const diffY = Math.abs(childCenterY - containerCenterY);
      
      if (diffY > tolerance) {
        throw new Error(
          `Element "${childId}" is not centered vertically in "${containerId}". ` +
          `Child center Y: ${childCenterY}, Container center Y: ${containerCenterY}, Diff: ${diffY}`
        );
      }
    }
  }

  /**
   * Assert that a child element fits completely inside a container with optional padding
   */
  assertFitsInside(childId: string, containerId: string, padding: number = 0): void {
    const child = this.findById(childId);
    const container = this.findById(containerId);

    if (!child || !child.computed) {
      throw new Error(`Child element with id "${childId}" not found or has no computed layout`);
    }
    if (!container || !container.computed) {
      throw new Error(`Container element with id "${containerId}" not found or has no computed layout`);
    }

    const minX = container.computed.x + padding;
    const minY = container.computed.y + padding;
    const maxX = container.computed.x + container.computed.width - padding;
    const maxY = container.computed.y + container.computed.height - padding;

    if (child.computed.x < minX) {
      throw new Error(
        `Element "${childId}" overflows left edge of "${containerId}". ` +
        `Child X: ${child.computed.x}, Min X: ${minX}`
      );
    }
    if (child.computed.y < minY) {
      throw new Error(
        `Element "${childId}" overflows top edge of "${containerId}". ` +
        `Child Y: ${child.computed.y}, Min Y: ${minY}`
      );
    }
    if (child.computed.x + child.computed.width > maxX) {
      throw new Error(
        `Element "${childId}" overflows right edge of "${containerId}". ` +
        `Child right edge: ${child.computed.x + child.computed.width}, Max X: ${maxX}`
      );
    }
    if (child.computed.y + child.computed.height > maxY) {
      throw new Error(
        `Element "${childId}" overflows bottom edge of "${containerId}". ` +
        `Child bottom edge: ${child.computed.y + child.computed.height}, Max Y: ${maxY}`
      );
    }
  }

  /**
   * Assert that there is a specific gap between two elements
   */
  assertGap(element1Id: string, element2Id: string, expectedGap: number, tolerance: number = 1): void {
    const elem1 = this.findById(element1Id);
    const elem2 = this.findById(element2Id);

    if (!elem1 || !elem1.computed) {
      throw new Error(`Element with id "${element1Id}" not found or has no computed layout`);
    }
    if (!elem2 || !elem2.computed) {
      throw new Error(`Element with id "${element2Id}" not found or has no computed layout`);
    }

    // Calculate vertical gap (assuming vertical stacking)
    const verticalGap = elem2.computed.y - (elem1.computed.y + elem1.computed.height);
    const horizontalGap = elem2.computed.x - (elem1.computed.x + elem1.computed.width);

    // Use the gap that makes sense based on layout direction
    const actualGap = Math.abs(verticalGap) < Math.abs(horizontalGap) ? verticalGap : horizontalGap;
    const diff = Math.abs(actualGap - expectedGap);

    if (diff > tolerance) {
      throw new Error(
        `Gap between "${element1Id}" and "${element2Id}" is incorrect. ` +
        `Expected: ${expectedGap}, Actual: ${actualGap}, Diff: ${diff}`
      );
    }
  }

  /**
   * Assert that elements are aligned on a specific edge
   */
  assertAligned(elementIds: string[], alignment: 'left' | 'center' | 'right' | 'top' | 'bottom', tolerance: number = 1): void {
    if (elementIds.length < 2) {
      throw new Error('Need at least 2 elements to check alignment');
    }

    const elements = elementIds.map(id => {
      const elem = this.findById(id);
      if (!elem || !elem.computed) {
        throw new Error(`Element with id "${id}" not found or has no computed layout`);
      }
      return elem;
    });

    let referenceValue: number;
    let getValueFn: (node: LayoutNode) => number;

    switch (alignment) {
      case 'left':
        getValueFn = (node) => node.computed!.x;
        break;
      case 'right':
        getValueFn = (node) => node.computed!.x + node.computed!.width;
        break;
      case 'center':
        getValueFn = (node) => node.computed!.x + node.computed!.width / 2;
        break;
      case 'top':
        getValueFn = (node) => node.computed!.y;
        break;
      case 'bottom':
        getValueFn = (node) => node.computed!.y + node.computed!.height;
        break;
    }

    referenceValue = getValueFn(elements[0]);

    for (let i = 1; i < elements.length; i++) {
      const value = getValueFn(elements[i]);
      const diff = Math.abs(value - referenceValue);
      
      if (diff > tolerance) {
        throw new Error(
          `Elements are not aligned on "${alignment}". ` +
          `Element "${elementIds[i]}" value: ${value}, Reference value: ${referenceValue}, Diff: ${diff}`
        );
      }
    }
  }

  /**
   * Assert that two elements don't overlap
   */
  assertNoOverlap(element1Id: string, element2Id: string): void {
    const elem1 = this.findById(element1Id);
    const elem2 = this.findById(element2Id);

    if (!elem1 || !elem1.computed) {
      throw new Error(`Element with id "${element1Id}" not found or has no computed layout`);
    }
    if (!elem2 || !elem2.computed) {
      throw new Error(`Element with id "${element2Id}" not found or has no computed layout`);
    }

    const overlap = !(
      elem1.computed.x + elem1.computed.width <= elem2.computed.x ||
      elem2.computed.x + elem2.computed.width <= elem1.computed.x ||
      elem1.computed.y + elem1.computed.height <= elem2.computed.y ||
      elem2.computed.y + elem2.computed.height <= elem1.computed.y
    );

    if (overlap) {
      throw new Error(
        `Elements "${element1Id}" and "${element2Id}" overlap. ` +
        `Elem1: (${elem1.computed.x}, ${elem1.computed.y}, ${elem1.computed.width}, ${elem1.computed.height}), ` +
        `Elem2: (${elem2.computed.x}, ${elem2.computed.y}, ${elem2.computed.width}, ${elem2.computed.height})`
      );
    }
  }

  /**
   * Get computed layout information for debugging
   */
  getLayoutInfo(elementId: string): string {
    const elem = this.findById(elementId);
    if (!elem || !elem.computed) {
      return `Element "${elementId}" not found or has no computed layout`;
    }

    return `Element "${elementId}" (${elem.type}):
  Position: (${elem.computed.x}, ${elem.computed.y})
  Size: ${elem.computed.width} x ${elem.computed.height}
  Center: (${elem.computed.x + elem.computed.width / 2}, ${elem.computed.y + elem.computed.height / 2})`;
  }
}
