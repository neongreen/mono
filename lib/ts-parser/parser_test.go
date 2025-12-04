package tsparser

import (
	"context"
	"testing"
)

func TestParseTypeScript(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	src := []byte(`/** This is a JSDoc comment for the greet function */
function greet(name: string): string {
    return "Hello, " + name;
}

class Greeter {
    /** JSDoc for the greeting property */
    greeting: string;

    /** JSDoc for the constructor */
    constructor(message: string) {
        this.greeting = message;
    }

    /** JSDoc for the greet method */
    greet(): string {
        return "Hello, " + this.greeting;
    }
}
`)

	tree, err := parser.ParseTS(src)
	if err != nil {
		t.Fatalf("Failed to parse TypeScript: %v", err)
	}
	defer tree.Close()

	root, err := tree.RootNode()
	if err != nil {
		t.Fatalf("Failed to get root node: %v", err)
	}

	typeName, err := root.TypeName()
	if err != nil {
		t.Fatalf("Failed to get type name: %v", err)
	}

	if typeName != "program" {
		t.Errorf("Expected root type 'program', got %q", typeName)
	}

	startByte, err := root.StartByte()
	if err != nil {
		t.Fatalf("Failed to get start byte: %v", err)
	}

	endByte, err := root.EndByte()
	if err != nil {
		t.Fatalf("Failed to get end byte: %v", err)
	}

	if startByte != 0 {
		t.Errorf("Expected start byte 0, got %d", startByte)
	}

	if endByte != uint32(len(src)) {
		t.Errorf("Expected end byte %d, got %d", len(src), endByte)
	}

	// Check children
	childCount, err := root.ChildCount()
	if err != nil {
		t.Fatalf("Failed to get child count: %v", err)
	}

	if childCount < 2 {
		t.Errorf("Expected at least 2 children (comment + function + class), got %d", childCount)
	}

	t.Logf("Root has %d children", childCount)
}

func TestParseTSX(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	src := []byte(`
import React from 'react';

/** JSDoc for the Greeting component */
function Greeting({ name }: { name: string }) {
    return <div>Hello, {name}!</div>;
}

/** JSDoc for the App component */
export function App() {
    return (
        <div>
            <Greeting name="World" />
        </div>
    );
}
`)

	tree, err := parser.ParseTSX(src)
	if err != nil {
		t.Fatalf("Failed to parse TSX: %v", err)
	}
	defer tree.Close()

	root, err := tree.RootNode()
	if err != nil {
		t.Fatalf("Failed to get root node: %v", err)
	}

	typeName, err := root.TypeName()
	if err != nil {
		t.Fatalf("Failed to get type name: %v", err)
	}

	if typeName != "program" {
		t.Errorf("Expected root type 'program', got %q", typeName)
	}

	childCount, err := root.ChildCount()
	if err != nil {
		t.Fatalf("Failed to get child count: %v", err)
	}

	if childCount < 2 {
		t.Errorf("Expected at least 2 children, got %d", childCount)
	}

	t.Logf("TSX root has %d children", childCount)
}

func TestJSDocExtractionTypeScript(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	src := []byte(`/** JSDoc for greet function */
function greet(name: string): string {
    return "Hello, " + name;
}`)

	tree, err := parser.ParseTS(src)
	if err != nil {
		t.Fatalf("Failed to parse TypeScript: %v", err)
	}
	defer tree.Close()

	jsdocs, err := ExtractJSDoc(src, tree)
	if err != nil {
		t.Fatalf("Failed to extract JSDoc: %v", err)
	}

	// Should find exactly one JSDoc
	if len(jsdocs) != 1 {
		t.Fatalf("Expected 1 JSDoc, got %d", len(jsdocs))
	}

	jsdoc := jsdocs[0]

	// Verify comment range
	commentText := string(src[jsdoc.CommentStart:jsdoc.CommentEnd])
	if commentText != "/** JSDoc for greet function */" {
		t.Errorf("Unexpected comment text: %q", commentText)
	}

	// Verify attached declaration
	if jsdoc.AttachedDecl.TypeName != "function_declaration" {
		t.Errorf("Expected attached declaration type 'function_declaration', got %q", jsdoc.AttachedDecl.TypeName)
	}

	if jsdoc.AttachedDecl.Name != "greet" {
		t.Errorf("Expected attached declaration name 'greet', got %q", jsdoc.AttachedDecl.Name)
	}

	// Verify the attached declaration range matches the function
	funcText := string(src[jsdoc.AttachedDecl.StartByte:jsdoc.AttachedDecl.EndByte])
	if len(funcText) == 0 {
		t.Error("Attached declaration range is empty")
	}
	t.Logf("Attached declaration: %s", funcText)

	// Verify container is the source file
	if jsdoc.ContainerDecl.TypeName != "program" {
		t.Errorf("Expected container type 'program', got %q", jsdoc.ContainerDecl.TypeName)
	}

	t.Logf("JSDoc: CommentStart=%d, CommentEnd=%d, AttachedDecl=%+v, ContainerDecl=%+v",
		jsdoc.CommentStart, jsdoc.CommentEnd, jsdoc.AttachedDecl, jsdoc.ContainerDecl)
}

func TestJSDocExtractionTSX(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	src := []byte(`import React from 'react';

/** JSDoc for the Greeting component */
function Greeting({ name }: { name: string }) {
    return <div>Hello, {name}!</div>;
}`)

	tree, err := parser.ParseTSX(src)
	if err != nil {
		t.Fatalf("Failed to parse TSX: %v", err)
	}
	defer tree.Close()

	jsdocs, err := ExtractJSDoc(src, tree)
	if err != nil {
		t.Fatalf("Failed to extract JSDoc: %v", err)
	}

	// Should find exactly one JSDoc
	if len(jsdocs) != 1 {
		t.Fatalf("Expected 1 JSDoc, got %d", len(jsdocs))
	}

	jsdoc := jsdocs[0]

	// Verify comment range
	commentText := string(src[jsdoc.CommentStart:jsdoc.CommentEnd])
	if commentText != "/** JSDoc for the Greeting component */" {
		t.Errorf("Unexpected comment text: %q", commentText)
	}

	// Verify attached declaration
	if jsdoc.AttachedDecl.TypeName != "function_declaration" {
		t.Errorf("Expected attached declaration type 'function_declaration', got %q", jsdoc.AttachedDecl.TypeName)
	}

	if jsdoc.AttachedDecl.Name != "Greeting" {
		t.Errorf("Expected attached declaration name 'Greeting', got %q", jsdoc.AttachedDecl.Name)
	}

	// Verify attached declaration range matches the function
	funcText := string(src[jsdoc.AttachedDecl.StartByte:jsdoc.AttachedDecl.EndByte])
	if len(funcText) == 0 {
		t.Error("Attached declaration range is empty")
	}
	t.Logf("Attached declaration: %s", funcText)

	// Verify container is the source file
	if jsdoc.ContainerDecl.TypeName != "program" {
		t.Errorf("Expected container type 'program', got %q", jsdoc.ContainerDecl.TypeName)
	}

	t.Logf("JSDoc: CommentStart=%d, CommentEnd=%d, AttachedDecl=%+v, ContainerDecl=%+v",
		jsdoc.CommentStart, jsdoc.CommentEnd, jsdoc.AttachedDecl, jsdoc.ContainerDecl)
}

func TestJSDocNestedContainer(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	src := []byte(`class MyClass {
    /** JSDoc for the method */
    myMethod(): void {
        console.log("hello");
    }
}`)

	tree, err := parser.ParseTS(src)
	if err != nil {
		t.Fatalf("Failed to parse TypeScript: %v", err)
	}
	defer tree.Close()

	jsdocs, err := ExtractJSDoc(src, tree)
	if err != nil {
		t.Fatalf("Failed to extract JSDoc: %v", err)
	}

	// Should find exactly one JSDoc
	if len(jsdocs) != 1 {
		t.Fatalf("Expected 1 JSDoc, got %d", len(jsdocs))
	}

	jsdoc := jsdocs[0]

	// Verify attached declaration is the method
	if jsdoc.AttachedDecl.TypeName != "method_definition" {
		t.Errorf("Expected attached declaration type 'method_definition', got %q", jsdoc.AttachedDecl.TypeName)
	}

	if jsdoc.AttachedDecl.Name != "myMethod" {
		t.Errorf("Expected attached declaration name 'myMethod', got %q", jsdoc.AttachedDecl.Name)
	}

	// Verify container is the class body or class
	if jsdoc.ContainerDecl.TypeName != "class_body" && jsdoc.ContainerDecl.TypeName != "class_declaration" {
		t.Errorf("Expected container type 'class_body' or 'class_declaration', got %q", jsdoc.ContainerDecl.TypeName)
	}

	t.Logf("JSDoc container: %+v", jsdoc.ContainerDecl)
}

func TestNodeNavigation(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	src := []byte(`const a = 1;
const b = 2;
const c = 3;`)

	tree, err := parser.ParseTS(src)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer tree.Close()

	root, err := tree.RootNode()
	if err != nil {
		t.Fatalf("Failed to get root: %v", err)
	}

	// Get first child
	first, err := root.Child(0)
	if err != nil {
		t.Fatalf("Failed to get first child: %v", err)
	}

	firstType, _ := first.TypeName()
	t.Logf("First child type: %s", firstType)

	// Get next sibling
	second, err := first.NextSibling()
	if err != nil {
		t.Fatalf("Failed to get next sibling: %v", err)
	}

	isNull, _ := second.IsNull()
	if isNull {
		t.Error("Second sibling should not be null")
	}

	secondType, _ := second.TypeName()
	t.Logf("Second child type: %s", secondType)

	// Test Text() method
	firstText, err := first.Text()
	if err != nil {
		t.Fatalf("Failed to get text: %v", err)
	}
	t.Logf("First child text: %s", firstText)
}

func TestParserReuse(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	sources := [][]byte{
		[]byte(`const x = 1;`),
		[]byte(`function foo() { return 42; }`),
		[]byte(`class Bar { prop: string; }`),
	}

	for i, src := range sources {
		tree, err := parser.ParseTS(src)
		if err != nil {
			t.Fatalf("Parse %d failed: %v", i, err)
		}

		root, err := tree.RootNode()
		if err != nil {
			t.Fatalf("Get root %d failed: %v", i, err)
		}

		typeName, _ := root.TypeName()
		if typeName != "program" {
			t.Errorf("Parse %d: expected 'program', got %q", i, typeName)
		}

		tree.Close()
	}
}
