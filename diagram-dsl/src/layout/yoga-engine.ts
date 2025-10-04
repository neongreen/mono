import { LayoutNode, LayoutProps, AlignmentProps, PositionProps } from '../types';

let yogaInstance: any = null;

async function loadYoga(): Promise<any> {
  if (!yogaInstance) {
    const Yoga = await import('yoga-layout');
    yogaInstance = Yoga.default;
  }
  return yogaInstance;
}

export class YogaLayoutEngine {
  private yoga: any;

  private constructor(yoga: any) {
    this.yoga = yoga;
  }

  static async create(): Promise<YogaLayoutEngine> {
    const yoga = await loadYoga();
    return new YogaLayoutEngine(yoga);
  }

  computeLayout(tree: LayoutNode, containerWidth: number = 800, containerHeight: number = 600): LayoutNode {
    const root = this.createYogaNode(tree);
    root.calculateLayout(containerWidth, containerHeight, this.yoga.DIRECTION_LTR);
    this.extractLayout(root, tree, 0, 0);
    this.freeYogaNode(root);
    return tree;
  }

  private createYogaNode(node: LayoutNode): any {
    // Arrow nodes don't participate in layout
    if (node.type === 'Arrow') {
      return null;
    }

    const yogaNode = this.yoga.Node.create();

    const props = node.props;

    // Set dimensions
    if (props.width !== undefined && props.width !== 'auto') {
      yogaNode.setWidth(props.width);
    } else if (node.type === 'Text') {
      // Estimate text width based on font size and character count
      const text = props.children || '';
      const fontSize = props.fontSize || 16;
      const estimatedWidth = text.length * fontSize * 0.6;
      yogaNode.setWidth(estimatedWidth);
    }
    
    if (props.height !== undefined && props.height !== 'auto') {
      yogaNode.setHeight(props.height);
    } else if (node.type === 'Text') {
      // Use font size as height
      const fontSize = props.fontSize || 16;
      yogaNode.setHeight(fontSize * 1.2);
    }
    if (props.minWidth !== undefined) {
      yogaNode.setMinWidth(props.minWidth);
    }
    if (props.minHeight !== undefined) {
      yogaNode.setMinHeight(props.minHeight);
    }
    if (props.maxWidth !== undefined) {
      yogaNode.setMaxWidth(props.maxWidth);
    }
    if (props.maxHeight !== undefined) {
      yogaNode.setMaxHeight(props.maxHeight);
    }

    // Set padding
    if (props.padding !== undefined) {
      yogaNode.setPadding(this.yoga.EDGE_ALL, props.padding);
    }
    if (props.paddingTop !== undefined) {
      yogaNode.setPadding(this.yoga.EDGE_TOP, props.paddingTop);
    }
    if (props.paddingBottom !== undefined) {
      yogaNode.setPadding(this.yoga.EDGE_BOTTOM, props.paddingBottom);
    }
    if (props.paddingLeft !== undefined) {
      yogaNode.setPadding(this.yoga.EDGE_LEFT, props.paddingLeft);
    }
    if (props.paddingRight !== undefined) {
      yogaNode.setPadding(this.yoga.EDGE_RIGHT, props.paddingRight);
    }

    // Set margin
    if (props.margin !== undefined) {
      yogaNode.setMargin(this.yoga.EDGE_ALL, props.margin);
    }
    if (props.marginTop !== undefined) {
      yogaNode.setMargin(this.yoga.EDGE_TOP, props.marginTop);
    }
    if (props.marginBottom !== undefined) {
      yogaNode.setMargin(this.yoga.EDGE_BOTTOM, props.marginBottom);
    }
    if (props.marginLeft !== undefined) {
      yogaNode.setMargin(this.yoga.EDGE_LEFT, props.marginLeft);
    }
    if (props.marginRight !== undefined) {
      yogaNode.setMargin(this.yoga.EDGE_RIGHT, props.marginRight);
    }

    // Set flex direction (for Stack/Row/Column)
    if (props.direction === 'horizontal' || node.type === 'Row') {
      yogaNode.setFlexDirection(this.yoga.FLEX_DIRECTION_ROW);
    } else if (props.direction === 'vertical' || node.type === 'Column' || node.type === 'Stack') {
      yogaNode.setFlexDirection(this.yoga.FLEX_DIRECTION_COLUMN);
    }

    // Set gap (using margin on children)
    if (props.gap !== undefined) {
      yogaNode.setGap(this.yoga.GUTTER_ALL, props.gap);
    }

    // Set alignment
    if (props.alignItems) {
      const alignMap: Record<string, number> = {
        'flex-start': this.yoga.ALIGN_FLEX_START,
        'center': this.yoga.ALIGN_CENTER,
        'flex-end': this.yoga.ALIGN_FLEX_END,
        'stretch': this.yoga.ALIGN_STRETCH,
      };
      yogaNode.setAlignItems(alignMap[props.alignItems] || this.yoga.ALIGN_STRETCH);
    }

    if (props.justifyContent) {
      const justifyMap: Record<string, number> = {
        'flex-start': this.yoga.JUSTIFY_FLEX_START,
        'center': this.yoga.JUSTIFY_CENTER,
        'flex-end': this.yoga.JUSTIFY_FLEX_END,
        'space-between': this.yoga.JUSTIFY_SPACE_BETWEEN,
        'space-around': this.yoga.JUSTIFY_SPACE_AROUND,
        'space-evenly': this.yoga.JUSTIFY_SPACE_EVENLY,
      };
      yogaNode.setJustifyContent(justifyMap[props.justifyContent] || this.yoga.JUSTIFY_FLEX_START);
    }

    // Set position
    if (props.position === 'absolute') {
      yogaNode.setPositionType(this.yoga.POSITION_TYPE_ABSOLUTE);
      if (props.top !== undefined) yogaNode.setPosition(this.yoga.EDGE_TOP, props.top);
      if (props.bottom !== undefined) yogaNode.setPosition(this.yoga.EDGE_BOTTOM, props.bottom);
      if (props.left !== undefined) yogaNode.setPosition(this.yoga.EDGE_LEFT, props.left);
      if (props.right !== undefined) yogaNode.setPosition(this.yoga.EDGE_RIGHT, props.right);
    } else {
      yogaNode.setPositionType(this.yoga.POSITION_TYPE_RELATIVE);
    }

    // Process children
    let childIndex = 0;
    node.children.forEach((child) => {
      const childYogaNode = this.createYogaNode(child);
      if (childYogaNode !== null) {
        yogaNode.insertChild(childYogaNode, childIndex);
        childIndex++;
      }
    });

    return yogaNode;
  }

  private extractLayout(yogaNode: any, node: LayoutNode, parentX: number, parentY: number): void {
    const layout = yogaNode.getComputedLayout();
    
    node.computed = {
      x: parentX + layout.left,
      y: parentY + layout.top,
      width: layout.width,
      height: layout.height,
    };

    let yogaChildIndex = 0;
    node.children.forEach((child) => {
      if (child.type === 'Arrow') {
        // Arrows don't have layout, skip
        return;
      }
      const childYogaNode = yogaNode.getChild(yogaChildIndex);
      if (node.computed && childYogaNode) {
        this.extractLayout(childYogaNode, child, node.computed.x, node.computed.y);
      }
      yogaChildIndex++;
    });
  }

  private freeYogaNode(yogaNode: any): void {
    if (!yogaNode) return;
    
    const childCount = yogaNode.getChildCount();
    for (let i = 0; i < childCount; i++) {
      const child = yogaNode.getChild(i);
      this.freeYogaNode(child);
    }
    yogaNode.free();
  }
}
