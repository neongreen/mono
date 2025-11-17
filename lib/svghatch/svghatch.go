package svghatch

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// PatternType represents different hatching patterns available
type PatternType string

const (
	PatternHorizontal    PatternType = "horizontal"
	PatternVertical      PatternType = "vertical"
	PatternDiagonalLeft  PatternType = "diagonal-left"  // from top-left to bottom-right
	PatternDiagonalRight PatternType = "diagonal-right" // from top-right to bottom-left
	PatternCrosshatch    PatternType = "crosshatch"
	PatternDots          PatternType = "dots"
	PatternGrid          PatternType = "grid"
)

// PatternConfig configures how a pattern is generated
type PatternConfig struct {
	Type    PatternType // Pattern type to use
	Spacing float64     // Spacing between lines/dots (default: 5)
	Width   float64     // Line width (default: 1)
	Angle   float64     // Angle for diagonal patterns (default: 45)
}

// DefaultPatternConfig returns a default pattern configuration
func DefaultPatternConfig(patternType PatternType) PatternConfig {
	return PatternConfig{
		Type:    patternType,
		Spacing: 5.0,
		Width:   1.0,
		Angle:   45.0,
	}
}

// ColorMapping maps a source color to a pattern configuration
type ColorMapping struct {
	SourceColor string        // Color to replace (e.g., "#FF0000", "red")
	Pattern     PatternConfig // Pattern to use as replacement
}

// Replacer performs SVG color-to-pattern replacements
type Replacer struct {
	mappings []ColorMapping
}

// NewReplacer creates a new SVG replacer with the given color mappings
func NewReplacer(mappings []ColorMapping) *Replacer {
	return &Replacer{
		mappings: mappings,
	}
}

// Replace reads an SVG, replaces colors with patterns, and writes the result
func (r *Replacer) Replace(input io.Reader, output io.Writer) error {
	// Read the entire SVG
	data, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read SVG: %w", err)
	}

	// Parse the SVG
	svg, err := parseSVG(data)
	if err != nil {
		return fmt.Errorf("failed to parse SVG: %w", err)
	}

	// Apply replacements
	if err := r.applyReplacements(svg); err != nil {
		return fmt.Errorf("failed to apply replacements: %w", err)
	}

	// Write the modified SVG
	if err := writeSVG(svg, output); err != nil {
		return fmt.Errorf("failed to write SVG: %w", err)
	}

	return nil
}

// SVGNode represents a generic SVG element
type SVGNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Content  []byte     `xml:",innerxml"`
	Children []*SVGNode
}

func parseSVG(data []byte) (*SVGNode, error) {
	var node SVGNode
	decoder := xml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&node); err != nil {
		return nil, fmt.Errorf("XML decode failed: %w", err)
	}
	// Parse children from innerxml content
	node.Children = parseChildren(node.Content)
	return &node, nil
}

func parseChildren(content []byte) []*SVGNode {
	if len(content) == 0 {
		return nil
	}

	var children []*SVGNode
	decoder := xml.NewDecoder(bytes.NewReader(content))
	for {
		var child SVGNode
		err := decoder.Decode(&child)
		if err == io.EOF {
			break
		}
		if err != nil {
			// Not parseable XML, skip
			break
		}
		child.Children = parseChildren(child.Content)
		children = append(children, &child)
	}
	return children
}

func writeSVG(node *SVGNode, w io.Writer) error {
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")

	// Reconstruct the node with children as innerxml
	outputNode := &SVGNode{
		XMLName: node.XMLName,
		Attrs:   node.Attrs,
	}

	if len(node.Children) > 0 {
		var buf bytes.Buffer
		childEncoder := xml.NewEncoder(&buf)
		for _, child := range node.Children {
			writeNode(child, childEncoder)
		}
		childEncoder.Flush()
		outputNode.Content = buf.Bytes()
	}

	if err := encoder.Encode(outputNode); err != nil {
		return fmt.Errorf("XML encode failed: %w", err)
	}
	if err := encoder.Flush(); err != nil {
		return fmt.Errorf("encoder flush failed: %w", err)
	}
	return nil
}

func writeNode(node *SVGNode, encoder *xml.Encoder) error {
	outputNode := &SVGNode{
		XMLName: node.XMLName,
		Attrs:   node.Attrs,
	}

	if len(node.Children) > 0 {
		var buf bytes.Buffer
		childEncoder := xml.NewEncoder(&buf)
		for _, child := range node.Children {
			writeNode(child, childEncoder)
		}
		childEncoder.Flush()
		outputNode.Content = buf.Bytes()
	}

	return encoder.Encode(outputNode)
}

// applyReplacements walks through the SVG and applies pattern replacements
func (r *Replacer) applyReplacements(svg *SVGNode) error {
	// First, add pattern definitions to the SVG
	if err := r.addPatternDefs(svg); err != nil {
		return fmt.Errorf("failed to add pattern definitions: %w", err)
	}

	// Walk through all nodes and replace colors
	r.replaceColorsInNode(svg)

	return nil
}

// addPatternDefs adds <defs> with all pattern definitions to the SVG
func (r *Replacer) addPatternDefs(svg *SVGNode) error {
	// Find or create <defs> element
	var defs *SVGNode
	for _, child := range svg.Children {
		if child.XMLName.Local == "defs" {
			defs = child
			break
		}
	}

	if defs == nil {
		defs = &SVGNode{
			XMLName:  xml.Name{Local: "defs"},
			Children: []*SVGNode{},
		}
		// Add defs as first child
		svg.Children = append([]*SVGNode{defs}, svg.Children...)
	}

	// Add pattern definitions
	for i, mapping := range r.mappings {
		patternNode := r.createPatternNode(i, mapping.Pattern)
		defs.Children = append(defs.Children, patternNode)
	}

	return nil
}

// createPatternNode creates an SVG <pattern> node for the given configuration
func (r *Replacer) createPatternNode(id int, config PatternConfig) *SVGNode {
	patternID := fmt.Sprintf("pattern-%d", id)
	spacing := config.Spacing
	width := config.Width

	pattern := &SVGNode{
		XMLName: xml.Name{Local: "pattern"},
		Attrs: []xml.Attr{
			{Name: xml.Name{Local: "id"}, Value: patternID},
			{Name: xml.Name{Local: "patternUnits"}, Value: "userSpaceOnUse"},
			{Name: xml.Name{Local: "width"}, Value: fmt.Sprintf("%.1f", spacing)},
			{Name: xml.Name{Local: "height"}, Value: fmt.Sprintf("%.1f", spacing)},
		},
		Children: []*SVGNode{},
	}

	// Create pattern elements based on type
	switch config.Type {
	case PatternHorizontal:
		line := &SVGNode{
			XMLName: xml.Name{Local: "line"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "x1"}, Value: "0"},
				{Name: xml.Name{Local: "y1"}, Value: fmt.Sprintf("%.1f", spacing/2)},
				{Name: xml.Name{Local: "x2"}, Value: fmt.Sprintf("%.1f", spacing)},
				{Name: xml.Name{Local: "y2"}, Value: fmt.Sprintf("%.1f", spacing/2)},
				{Name: xml.Name{Local: "stroke"}, Value: "black"},
				{Name: xml.Name{Local: "stroke-width"}, Value: fmt.Sprintf("%.1f", width)},
			},
		}
		pattern.Children = append(pattern.Children, line)

	case PatternVertical:
		line := &SVGNode{
			XMLName: xml.Name{Local: "line"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "x1"}, Value: fmt.Sprintf("%.1f", spacing/2)},
				{Name: xml.Name{Local: "y1"}, Value: "0"},
				{Name: xml.Name{Local: "x2"}, Value: fmt.Sprintf("%.1f", spacing/2)},
				{Name: xml.Name{Local: "y2"}, Value: fmt.Sprintf("%.1f", spacing)},
				{Name: xml.Name{Local: "stroke"}, Value: "black"},
				{Name: xml.Name{Local: "stroke-width"}, Value: fmt.Sprintf("%.1f", width)},
			},
		}
		pattern.Children = append(pattern.Children, line)

	case PatternDiagonalLeft:
		line := &SVGNode{
			XMLName: xml.Name{Local: "line"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "x1"}, Value: "0"},
				{Name: xml.Name{Local: "y1"}, Value: "0"},
				{Name: xml.Name{Local: "x2"}, Value: fmt.Sprintf("%.1f", spacing)},
				{Name: xml.Name{Local: "y2"}, Value: fmt.Sprintf("%.1f", spacing)},
				{Name: xml.Name{Local: "stroke"}, Value: "black"},
				{Name: xml.Name{Local: "stroke-width"}, Value: fmt.Sprintf("%.1f", width)},
			},
		}
		pattern.Children = append(pattern.Children, line)

	case PatternDiagonalRight:
		line := &SVGNode{
			XMLName: xml.Name{Local: "line"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "x1"}, Value: "0"},
				{Name: xml.Name{Local: "y1"}, Value: fmt.Sprintf("%.1f", spacing)},
				{Name: xml.Name{Local: "x2"}, Value: fmt.Sprintf("%.1f", spacing)},
				{Name: xml.Name{Local: "y2"}, Value: "0"},
				{Name: xml.Name{Local: "stroke"}, Value: "black"},
				{Name: xml.Name{Local: "stroke-width"}, Value: fmt.Sprintf("%.1f", width)},
			},
		}
		pattern.Children = append(pattern.Children, line)

	case PatternCrosshatch:
		line1 := &SVGNode{
			XMLName: xml.Name{Local: "line"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "x1"}, Value: "0"},
				{Name: xml.Name{Local: "y1"}, Value: "0"},
				{Name: xml.Name{Local: "x2"}, Value: fmt.Sprintf("%.1f", spacing)},
				{Name: xml.Name{Local: "y2"}, Value: fmt.Sprintf("%.1f", spacing)},
				{Name: xml.Name{Local: "stroke"}, Value: "black"},
				{Name: xml.Name{Local: "stroke-width"}, Value: fmt.Sprintf("%.1f", width)},
			},
		}
		line2 := &SVGNode{
			XMLName: xml.Name{Local: "line"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "x1"}, Value: "0"},
				{Name: xml.Name{Local: "y1"}, Value: fmt.Sprintf("%.1f", spacing)},
				{Name: xml.Name{Local: "x2"}, Value: fmt.Sprintf("%.1f", spacing)},
				{Name: xml.Name{Local: "y2"}, Value: "0"},
				{Name: xml.Name{Local: "stroke"}, Value: "black"},
				{Name: xml.Name{Local: "stroke-width"}, Value: fmt.Sprintf("%.1f", width)},
			},
		}
		pattern.Children = append(pattern.Children, line1, line2)

	case PatternDots:
		circle := &SVGNode{
			XMLName: xml.Name{Local: "circle"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "cx"}, Value: fmt.Sprintf("%.1f", spacing/2)},
				{Name: xml.Name{Local: "cy"}, Value: fmt.Sprintf("%.1f", spacing/2)},
				{Name: xml.Name{Local: "r"}, Value: fmt.Sprintf("%.1f", width)},
				{Name: xml.Name{Local: "fill"}, Value: "black"},
			},
		}
		pattern.Children = append(pattern.Children, circle)

	case PatternGrid:
		line1 := &SVGNode{
			XMLName: xml.Name{Local: "line"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "x1"}, Value: "0"},
				{Name: xml.Name{Local: "y1"}, Value: "0"},
				{Name: xml.Name{Local: "x2"}, Value: fmt.Sprintf("%.1f", spacing)},
				{Name: xml.Name{Local: "y2"}, Value: "0"},
				{Name: xml.Name{Local: "stroke"}, Value: "black"},
				{Name: xml.Name{Local: "stroke-width"}, Value: fmt.Sprintf("%.1f", width)},
			},
		}
		line2 := &SVGNode{
			XMLName: xml.Name{Local: "line"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "x1"}, Value: "0"},
				{Name: xml.Name{Local: "y1"}, Value: "0"},
				{Name: xml.Name{Local: "x2"}, Value: "0"},
				{Name: xml.Name{Local: "y2"}, Value: fmt.Sprintf("%.1f", spacing)},
				{Name: xml.Name{Local: "stroke"}, Value: "black"},
				{Name: xml.Name{Local: "stroke-width"}, Value: fmt.Sprintf("%.1f", width)},
			},
		}
		pattern.Children = append(pattern.Children, line1, line2)
	}

	return pattern
}

// replaceColorsInNode recursively replaces colors in a node and its children
func (r *Replacer) replaceColorsInNode(node *SVGNode) {
	// Replace colors in this node's attributes
	for i := range node.Attrs {
		attr := &node.Attrs[i]
		if attr.Name.Local == "fill" || attr.Name.Local == "style" {
			r.replaceColorInAttr(attr, i)
		}
	}

	// Recursively process child nodes
	for _, child := range node.Children {
		r.replaceColorsInNode(child)
	}
}

// replaceColorInAttr replaces a color in an attribute value with a pattern reference
func (r *Replacer) replaceColorInAttr(attr *xml.Attr, patternIdx int) {
	if attr.Name.Local == "fill" {
		// Direct fill attribute
		for i, mapping := range r.mappings {
			if normalizeColor(attr.Value) == normalizeColor(mapping.SourceColor) {
				attr.Value = fmt.Sprintf("url(#pattern-%d)", i)
				return
			}
		}
	} else if attr.Name.Local == "style" {
		// Style attribute containing fill
		style := attr.Value
		for i, mapping := range r.mappings {
			normalized := normalizeColor(mapping.SourceColor)
			// Replace fill:color patterns in style
			style = strings.ReplaceAll(style, "fill:"+mapping.SourceColor, fmt.Sprintf("fill:url(#pattern-%d)", i))
			style = strings.ReplaceAll(style, "fill: "+mapping.SourceColor, fmt.Sprintf("fill: url(#pattern-%d)", i))
			style = strings.ReplaceAll(style, "fill:"+normalized, fmt.Sprintf("fill:url(#pattern-%d)", i))
			style = strings.ReplaceAll(style, "fill: "+normalized, fmt.Sprintf("fill: url(#pattern-%d)", i))
		}
		attr.Value = style
	}
}

// normalizeColor normalizes color values for comparison
func normalizeColor(color string) string {
	color = strings.TrimSpace(strings.ToLower(color))

	// Convert common color names to hex
	colorMap := map[string]string{
		"black": "#000000",
		"white": "#ffffff",
		"red":   "#ff0000",
		"green": "#008000",
		"blue":  "#0000ff",
		"gray":  "#808080",
		"grey":  "#808080",
	}

	if hex, ok := colorMap[color]; ok {
		return hex
	}

	// Expand 3-digit hex to 6-digit
	if len(color) == 4 && color[0] == '#' {
		r, g, b := color[1:2], color[2:3], color[3:4]
		return "#" + r + r + g + g + b + b
	}

	// Convert rgb() to hex
	if strings.HasPrefix(color, "rgb(") {
		parts := strings.TrimSuffix(strings.TrimPrefix(color, "rgb("), ")")
		components := strings.Split(parts, ",")
		if len(components) == 3 {
			var values [3]int
			for i, comp := range components {
				val, err := strconv.Atoi(strings.TrimSpace(comp))
				if err == nil && val >= 0 && val <= 255 {
					values[i] = val
				}
			}
			return fmt.Sprintf("#%02x%02x%02x", values[0], values[1], values[2])
		}
	}

	return color
}
