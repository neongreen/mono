import React, { ReactElement } from 'react';
import { LayoutNode } from '../types';
import { YogaLayoutEngine } from '../layout/yoga-engine';
import { SVGRenderer, RenderOptions } from './svg-renderer';

export interface RenderResult {
  svg: string;
  layout: LayoutNode;
}

export async function renderToSVG(element: ReactElement, options: RenderOptions = {}): Promise<string> {
  const result = await renderToSVGWithLayout(element, options);
  return result.svg;
}

export async function renderToSVGWithLayout(element: ReactElement, options: RenderOptions = {}): Promise<RenderResult> {
  // Convert React element tree to layout tree
  const layoutTree = elementToLayoutNode(element);
  
  // Compute layout using Yoga
  const layoutEngine = await YogaLayoutEngine.create();
  const computedTree = layoutEngine.computeLayout(layoutTree, options.width, options.height);
  
  // Render to SVG
  const svgRenderer = new SVGRenderer();
  const svg = svgRenderer.renderWithArrowMarkers(computedTree, options);
  
  return {
    svg,
    layout: computedTree
  };
}

function elementToLayoutNode(element: any): LayoutNode {
  if (typeof element === 'string' || typeof element === 'number') {
    return {
      type: 'Text',
      props: { children: String(element) },
      children: [],
    };
  }

  if (!element || !element.type) {
    return {
      type: 'Box',
      props: {},
      children: [],
    };
  }

  // If the element type is a function, call it to get the actual element
  if (typeof element.type === 'function') {
    const rendered = element.type(element.props || {});
    return elementToLayoutNode(rendered);
  }

  const type = typeof element.type === 'string' 
    ? element.type 
    : element.type.name || 'Box';

  const props = element.props || {};
  const children: LayoutNode[] = [];

  // For Text nodes, keep the children prop
  const finalProps = type === 'Text' 
    ? { ...props }
    : { ...props, children: undefined };

  // Process children (but not for Text or Arrow nodes)
  if (props.children && type !== 'Text' && type !== 'Arrow') {
    const childrenArray = Array.isArray(props.children) 
      ? props.children 
      : [props.children];

    childrenArray.forEach((child: any) => {
      if (child && typeof child === 'object' && child.type) {
        children.push(elementToLayoutNode(child));
      } else if (typeof child === 'string' || typeof child === 'number') {
        // Text children need to be wrapped in a Text node if parent isn't a Text node
        children.push({
          type: 'Text',
          props: { children: String(child) },
          children: [],
        });
      }
    });
  }

  return {
    type,
    props: finalProps,
    children,
  };
}

export { SVGRenderer, YogaLayoutEngine };
