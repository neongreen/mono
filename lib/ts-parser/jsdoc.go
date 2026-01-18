package tsparser

import (
	"strings"
)

// JSDoc comment markers
const (
	jsdocStart = "/**"
	jsdocEnd   = "*/"
)

// JSDoc represents a JSDoc comment with information about what it's attached to.
type JSDoc struct {
	// CommentStart is the byte offset where the JSDoc comment starts.
	CommentStart int
	// CommentEnd is the byte offset where the JSDoc comment ends.
	CommentEnd int
	// AttachedDecl is a reference to the declaration this JSDoc is attached to.
	AttachedDecl NodeRef
	// ContainerDecl is a reference to the nearest enclosing declaration.
	ContainerDecl NodeRef
}

// NodeRef is a stable reference to a node, containing byte range and display name.
type NodeRef struct {
	// StartByte is the byte offset where the node starts.
	StartByte int
	// EndByte is the byte offset where the node ends.
	EndByte int
	// Name is the display name of the node (may be empty if unavailable).
	Name string
	// TypeName is the type of the node in the grammar.
	TypeName string
}

// ExtractJSDoc extracts all JSDoc comments from a parsed tree and determines
// what declarations they are attached to and what containers they are inside.
func ExtractJSDoc(src []byte, tree *Tree) ([]JSDoc, error) {
	root, err := tree.RootNode()
	if err != nil {
		return nil, err
	}

	var jsdocs []JSDoc
	var containers []NodeRef // Stack of container declarations

	// Add the source file as the outermost container
	endByte, _ := root.EndByte()
	containers = append(containers, NodeRef{
		StartByte: 0,
		EndByte:   int(endByte),
		Name:      "<source>",
		TypeName:  "program",
	})

	// Recursively walk the tree
	err = walkTree(root, src, &jsdocs, &containers)
	if err != nil {
		return nil, err
	}

	return jsdocs, nil
}

// walkTree recursively walks the tree, finding JSDoc comments and their attachments.
func walkTree(node *Node, src []byte, jsdocs *[]JSDoc, containers *[]NodeRef) error {
	typeName, err := node.TypeName()
	if err != nil {
		return err
	}

	startByte, err := node.StartByte()
	if err != nil {
		return err
	}
	endByte, err := node.EndByte()
	if err != nil {
		return err
	}

	// Update container stack - pop containers that we've exited
	for len(*containers) > 1 {
		top := (*containers)[len(*containers)-1]
		if int(startByte) >= top.EndByte {
			*containers = (*containers)[:len(*containers)-1]
		} else {
			break
		}
	}

	// Check if this node is a container (has a body/block scope)
	if isContainer(typeName) {
		name := extractNodeName(node, src)
		*containers = append(*containers, NodeRef{
			StartByte: int(startByte),
			EndByte:   int(endByte),
			Name:      name,
			TypeName:  typeName,
		})
	}

	// Check if this is a comment node
	if typeName == "comment" {
		text, _ := node.Text()
		if isJSDocComment(text) {
			// Find the attached declaration
			attachedDecl, err := findAttachedDeclaration(node, src)
			if err != nil {
				return err
			}

			// Get current container
			containerDecl := (*containers)[len(*containers)-1]

			*jsdocs = append(*jsdocs, JSDoc{
				CommentStart:  int(startByte),
				CommentEnd:    int(endByte),
				AttachedDecl:  attachedDecl,
				ContainerDecl: containerDecl,
			})
		}
	}

	// Recurse into children
	childCount, err := node.ChildCount()
	if err != nil {
		return err
	}

	for i := uint32(0); i < childCount; i++ {
		child, err := node.Child(i)
		if err != nil {
			return err
		}

		err = walkTree(child, src, jsdocs, containers)
		if err != nil {
			return err
		}
	}

	return nil
}

// isJSDocComment returns true if the comment text is a JSDoc comment (/** ... */).
func isJSDocComment(text string) bool {
	return strings.HasPrefix(text, jsdocStart) && strings.HasSuffix(text, jsdocEnd) && len(text) > len(jsdocStart)+len(jsdocEnd)-1
}

// isContainer returns true if the node type represents a container declaration.
func isContainer(typeName string) bool {
	switch typeName {
	case "program",
		"function_declaration",
		"function",
		"arrow_function",
		"generator_function_declaration",
		"generator_function",
		"method_definition",
		"class_declaration",
		"class",
		"class_body",
		"interface_declaration",
		"enum_declaration",
		"module",
		"namespace_declaration",
		"ambient_declaration":
		return true
	}
	return false
}

// isDeclaration returns true if the node type represents a declaration.
func isDeclaration(typeName string) bool {
	switch typeName {
	case "function_declaration",
		"function",
		"arrow_function",
		"generator_function_declaration",
		"generator_function",
		"method_definition",
		"class_declaration",
		"class",
		"interface_declaration",
		"type_alias_declaration",
		"enum_declaration",
		"variable_declaration",
		"lexical_declaration",
		"export_statement",
		"public_field_definition",
		"property_signature",
		"method_signature",
		"call_signature",
		"construct_signature",
		"index_signature",
		"abstract_method_signature":
		return true
	}
	return false
}

// findAttachedDeclaration finds the declaration that a JSDoc comment is attached to.
func findAttachedDeclaration(commentNode *Node, src []byte) (NodeRef, error) {
	commentEnd, err := commentNode.EndByte()
	if err != nil {
		return NodeRef{}, err
	}

	// Look for the next sibling that is a declaration
	sibling := commentNode
	for {
		nextSibling, err := sibling.NextSibling()
		if err != nil {
			return NodeRef{}, err
		}

		isNull, err := nextSibling.IsNull()
		if err != nil {
			return NodeRef{}, err
		}
		if isNull {
			break
		}

		siblingType, err := nextSibling.TypeName()
		if err != nil {
			return NodeRef{}, err
		}

		siblingStart, err := nextSibling.StartByte()
		if err != nil {
			return NodeRef{}, err
		}

		// Skip other comments, whitespace nodes, and decorators/modifiers
		if isSkippableBeforeDeclaration(siblingType) {
			sibling = nextSibling
			continue
		}

		// Check if this is a declaration
		if isDeclaration(siblingType) {
			siblingEnd, err := nextSibling.EndByte()
			if err != nil {
				return NodeRef{}, err
			}

			// Get the declaration name
			name := extractNodeName(nextSibling, src)

			return NodeRef{
				StartByte: int(siblingStart),
				EndByte:   int(siblingEnd),
				Name:      name,
				TypeName:  siblingType,
			}, nil
		}

		// Not a declaration - stop looking
		break
	}

	// Try to find a declaration in the parent's children after this comment
	parent, err := commentNode.Parent()
	if err != nil {
		return NodeRef{}, err
	}

	isParentNull, err := parent.IsNull()
	if err != nil {
		return NodeRef{}, err
	}
	if isParentNull {
		return NodeRef{}, nil
	}

	// Find the declaration that starts after the comment in the parent
	childCount, err := parent.ChildCount()
	if err != nil {
		return NodeRef{}, err
	}

	foundComment := false
	for i := uint32(0); i < childCount; i++ {
		child, err := parent.Child(i)
		if err != nil {
			return NodeRef{}, err
		}

		childStart, err := child.StartByte()
		if err != nil {
			return NodeRef{}, err
		}
		childEnd, err := child.EndByte()
		if err != nil {
			return NodeRef{}, err
		}
		childType, err := child.TypeName()
		if err != nil {
			return NodeRef{}, err
		}

		// Track when we've passed the comment
		if childStart == uint32(commentEnd)-uint32(len(jsdocEnd)) {
			foundComment = true
			continue
		}

		if foundComment && childStart >= commentEnd {
			// Skip skippable nodes
			if isSkippableBeforeDeclaration(childType) {
				continue
			}

			// Check if this is a declaration
			if isDeclaration(childType) {
				name := extractNodeName(child, src)
				return NodeRef{
					StartByte: int(childStart),
					EndByte:   int(childEnd),
					Name:      name,
					TypeName:  childType,
				}, nil
			}

			// Not a declaration - stop looking
			break
		}
	}

	return NodeRef{}, nil
}

// isSkippableBeforeDeclaration returns true for nodes that can appear between
// a JSDoc and its attached declaration.
func isSkippableBeforeDeclaration(typeName string) bool {
	switch typeName {
	case "comment",
		"decorator",
		"accessibility_modifier",
		"static",
		"readonly",
		"abstract",
		"async",
		"override":
		return true
	}
	return false
}

// extractNodeName extracts a human-readable name from a declaration node.
func extractNodeName(node *Node, src []byte) string {
	typeName, err := node.TypeName()
	if err != nil {
		return ""
	}

	// For different declaration types, look for the identifier child
	switch typeName {
	case "function_declaration",
		"generator_function_declaration",
		"class_declaration",
		"interface_declaration",
		"type_alias_declaration",
		"enum_declaration",
		"namespace_declaration":
		// These have a direct 'name' or 'identifier' child
		return findIdentifierInChildren(node, src)

	case "variable_declaration", "lexical_declaration":
		// These have variable_declarator children
		childCount, _ := node.NamedChildCount()
		for i := uint32(0); i < childCount; i++ {
			child, _ := node.NamedChild(i)
			childType, _ := child.TypeName()
			if childType == "variable_declarator" {
				return findIdentifierInChildren(child, src)
			}
		}

	case "method_definition", "public_field_definition",
		"property_signature", "method_signature":
		// These have a 'name' property child
		return findIdentifierInChildren(node, src)

	case "export_statement":
		// For exports, look at the declaration inside
		childCount, _ := node.NamedChildCount()
		for i := uint32(0); i < childCount; i++ {
			child, _ := node.NamedChild(i)
			childType, _ := child.TypeName()
			if isDeclaration(childType) {
				return extractNodeName(child, src)
			}
		}

	case "arrow_function", "function":
		// Anonymous functions - might have parent variable declarator
		parent, _ := node.Parent()
		if parent != nil {
			parentType, _ := parent.TypeName()
			if parentType == "variable_declarator" {
				return findIdentifierInChildren(parent, src)
			}
		}
		return "<anonymous>"
	}

	return ""
}

// findIdentifierInChildren looks for an identifier or property_identifier child.
func findIdentifierInChildren(node *Node, src []byte) string {
	childCount, err := node.NamedChildCount()
	if err != nil {
		return ""
	}

	for i := uint32(0); i < childCount; i++ {
		child, err := node.NamedChild(i)
		if err != nil {
			continue
		}

		childType, err := child.TypeName()
		if err != nil {
			continue
		}

		if childType == "identifier" || childType == "property_identifier" || childType == "type_identifier" {
			text, _ := child.Text()
			return text
		}
	}

	return ""
}
