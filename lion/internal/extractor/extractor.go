package extractor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type DocEntry struct {
	Topic   string
	Content string
	File    string
	Line    int
	Entity  string // Name of the function/type/etc. the comment is attached to
	// Optional overrides
	TopicTitle    string
	HasTopicTitle bool
	SectionTitle  string
	HasSection    bool
}

// Extraction pipeline:
//   - Walks all .go files under the directory, skipping *_test.go.
//   - For TypeScript files (.ts, .tsx), uses the TypeScript compiler API via a Node.js helper.
//   - Parses with comments and pulls lion markers from package doc, func doc, type/const/var
//     doc comments, and inline comments inside functions and other declarations (first name in a
//     const/var block is used as the entity).
//   - Supports single-line markers and block comment markers (marker at top of the doc block).
//   - For TypeScript, uses TSDoc-style @lion tags in JSDoc comments.
//   - Aggregates snippets per topic across files; generator writes one file per topic.
//
//lion:implementation section="Extraction pipeline"
func Extract(dir string) (map[string][]DocEntry, error) {
	docs := make(map[string][]DocEntry)
	fset := token.NewFileSet()

	// Extract from Go files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse the Go file
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// Extract documentation from the AST
		if err := extractFromFile(fset, file, path, docs); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Extract from TypeScript files
	tsDocs, err := ExtractTypeScript(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to extract TypeScript documentation: %w", err)
	}

	// Merge TypeScript docs into the main docs map
	for topic, entries := range tsDocs {
		docs[topic] = append(docs[topic], entries...)
	}

	return docs, nil
}

func extractFromFile(fset *token.FileSet, file *ast.File, filepath string, docs map[string][]DocEntry) error {
	processed := make(map[*ast.CommentGroup]bool)

	// Extract package-level comments
	if file.Doc != nil {
		processed[file.Doc] = true
		if err := extractFromCommentGroup(fset, file.Doc, filepath, "package "+file.Name.Name, docs); err != nil {
			return err
		}
	}

	var funcDecls []*ast.FuncDecl
	var genDecls []*ast.GenDecl

	// Extract doc comments from declarations and record them as processed
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			funcDecls = append(funcDecls, d)
			if d.Doc != nil {
				processed[d.Doc] = true
				if err := extractFromCommentGroup(fset, d.Doc, filepath, d.Name.Name, docs); err != nil {
					return err
				}
			}
		case *ast.GenDecl:
			genDecls = append(genDecls, d)
			if d.Doc != nil {
				processed[d.Doc] = true
				entityName := getEntityName(d)
				if err := extractFromCommentGroup(fset, d.Doc, filepath, entityName, docs); err != nil {
					return err
				}
			}
		}
	}

	// Extract inline lion comment groups that were not attached as doc comments
	for _, cg := range file.Comments {
		if processed[cg] {
			continue
		}

		entityName := determineEntityForComment(cg, file, funcDecls, genDecls)
		if err := extractFromCommentGroup(fset, cg, filepath, entityName, docs); err != nil {
			return err
		}
	}

	return nil
}

// getEntityName extracts the entity name from a general declaration.
func getEntityName(decl *ast.GenDecl) string {
	if len(decl.Specs) > 0 {
		switch spec := decl.Specs[0].(type) {
		case *ast.TypeSpec:
			return spec.Name.Name
		case *ast.ValueSpec:
			if len(spec.Names) > 0 {
				return spec.Names[0].Name
			}
		}
	}
	return ""
}

func determineEntityForComment(cg *ast.CommentGroup, file *ast.File, funcDecls []*ast.FuncDecl, genDecls []*ast.GenDecl) string {
	for _, fn := range funcDecls {
		if fn.Body != nil && cg.Pos() >= fn.Body.Pos() && cg.End() <= fn.Body.End() {
			return fn.Name.Name
		}
	}

	for _, gd := range genDecls {
		if cg.Pos() >= gd.Pos() && cg.End() <= gd.End() {
			if name := getEntityName(gd); name != "" {
				return name
			}
		}
	}

	return "package " + file.Name.Name
}

// Supported syntax formats:
//
//  1. Single-line marker first:
//     //lion:topic-name metadata...
//     // Content lines (optional)
//
//  2. Block comment:
//     /*lion:topic-name metadata...
//     Multi-line content
//     */
//
// Optional metadata on any lion marker:
//   - title="Custom Title" overrides the topic's display title (file H1 + index link).
//   - section="Section Title" overrides the heading for that entry; section="" suppresses it.
//   - Unknown keys stop metadata parsing and are treated as content.
//
// All formats attach documentation to the next declaration (function, type, const, var).
//
//lion:supported-syntax section="Supported syntax"
func extractFromCommentGroup(fset *token.FileSet, cg *ast.CommentGroup, filepath, entityName string, docs map[string][]DocEntry) error {
	topicGroups := make(map[string][]string)
	topicOrder := []string{}
	topicFirstLine := make(map[string]int)
	metaByTopic := make(map[string]metaInfo)
	currentTopic := ""

	for _, comment := range cg.List {
		text := comment.Text

		// Handle block comments /* */
		if strings.HasPrefix(text, "/*") && strings.HasSuffix(text, "*/") {
			text = strings.TrimPrefix(text, "/*")
			text = strings.TrimSuffix(text, "*/")
			text = strings.TrimSpace(text)

			// Check if this is a lion block comment
			if strings.HasPrefix(text, "lion:") {
				topic, content, meta, err := parseLionBlockComment(text)
				if err != nil {
					pos := fset.Position(comment.Pos())
					return fmt.Errorf("invalid lion block comment at %s:%d: %w", filepath, pos.Line, err)
				}
				if topic == "" {
					currentTopic = ""
					continue
				}
				currentTopic = topic
				if _, exists := topicGroups[topic]; !exists {
					topicOrder = append(topicOrder, topic)
				}
				if _, ok := topicFirstLine[topic]; !ok {
					pos := fset.Position(comment.Pos())
					topicFirstLine[topic] = pos.Line
				}
				if content != "" {
					topicGroups[topic] = append(topicGroups[topic], content)
				}
				if existing, ok := metaByTopic[topic]; ok {
					merged, err := mergeMeta(existing, meta)
					if err != nil {
						pos := fset.Position(comment.Pos())
						return fmt.Errorf("conflicting lion metadata at %s:%d: %w", filepath, pos.Line, err)
					}
					metaByTopic[topic] = merged
				} else {
					metaByTopic[topic] = meta
				}
			}
			continue
		}

		// Handle single-line comments //
		if strings.HasPrefix(text, "//") {
			text = strings.TrimPrefix(text, "//")
			text = strings.TrimSpace(text)

			// Check if this is a lion comment
			if strings.HasPrefix(text, "lion:") {
				topic, content, meta, err := parseLionCommentLine(text)
				if err != nil {
					pos := fset.Position(comment.Pos())
					return fmt.Errorf("invalid lion comment at %s:%d: %w", filepath, pos.Line, err)
				}
				if topic == "" {
					currentTopic = ""
					continue
				}
				currentTopic = topic
				if _, exists := topicGroups[topic]; !exists {
					topicOrder = append(topicOrder, topic)
				}
				if _, ok := topicFirstLine[topic]; !ok {
					pos := fset.Position(comment.Pos())
					topicFirstLine[topic] = pos.Line
				}
				if content != "" {
					topicGroups[topic] = append(topicGroups[topic], content)
				}
				if existing, ok := metaByTopic[topic]; ok {
					merged, err := mergeMeta(existing, meta)
					if err != nil {
						pos := fset.Position(comment.Pos())
						return fmt.Errorf("conflicting lion metadata at %s:%d: %w", filepath, pos.Line, err)
					}
					metaByTopic[topic] = merged
				} else {
					metaByTopic[topic] = meta
				}
				continue
			}

			// Non-lion comment: attach to current topic if one is active
			if currentTopic != "" && text != "" {
				topicGroups[currentTopic] = append(topicGroups[currentTopic], text)
			}
		}
	}

	// Create entries for each topic
	for _, topic := range topicOrder {
		contents := topicGroups[topic]
		combinedContent := strings.Join(contents, "\n\n")
		meta := metaByTopic[topic]
		line := topicFirstLine[topic]

		entry := DocEntry{
			Topic:         topic,
			Content:       combinedContent,
			File:          filepath,
			Line:          line,
			Entity:        entityName,
			TopicTitle:    meta.topicTitle,
			HasTopicTitle: meta.hasTopicTitle,
			SectionTitle:  meta.sectionTitle,
			HasSection:    meta.hasSection,
		}

		docs[topic] = append(docs[topic], entry)
	}

	return nil
}

func parseLionCommentLine(text string) (string, string, metaInfo, error) {
	// Remove "lion:" prefix
	text = strings.TrimPrefix(text, "lion:")
	text = strings.TrimSpace(text)

	// Split into topic and content
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 1 {
		return "", "", metaInfo{}, nil
	}

	topic := strings.TrimSpace(parts[0])
	content := ""
	meta := metaInfo{}
	if len(parts) > 1 {
		remainder := strings.TrimSpace(parts[1])
		var parsedContent string
		var err error
		meta, parsedContent, err = parseMetadata(remainder)
		if err != nil {
			return "", "", metaInfo{}, err
		}
		content = parsedContent
	}

	return topic, content, meta, nil
}

/*
lion:errors-and-validation section="Error handling"
Validation and error handling:
- Invalid or empty topics are ignored silently (no explicit validation yet).
- Conflicting topic/section titles within the same comment group fail extraction with file:line.
- Comments that do not attach to package/func/type/const/var doc groups are skipped.
- CLI exits non-zero only when extraction fails (parse/metadata error) or generation fails (write error).
- Bad markers inside otherwise valid files do not stop extraction; they are just skipped.
*/
func parseLionBlockComment(text string) (string, string, metaInfo, error) {
	// Remove "lion:" prefix
	text = strings.TrimPrefix(text, "lion:")
	text = strings.TrimSpace(text)

	// Split by newlines to handle multi-line content
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return "", "", metaInfo{}, nil
	}

	// First line contains topic and optionally the start of content
	firstLine := strings.TrimSpace(lines[0])
	parts := strings.SplitN(firstLine, " ", 2)
	if len(parts) < 1 {
		return "", "", metaInfo{}, nil
	}

	topic := strings.TrimSpace(parts[0])
	if topic == "" {
		return "", "", metaInfo{}, nil
	}

	// Collect content from first line (if any) and remaining lines
	var contentLines []string
	meta := metaInfo{}
	if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
		var parsedContent string
		var err error
		meta, parsedContent, err = parseMetadata(strings.TrimSpace(parts[1]))
		if err != nil {
			return "", "", metaInfo{}, err
		}
		if parsedContent != "" {
			contentLines = append(contentLines, parsedContent)
		}
	}

	// Add remaining lines
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			contentLines = append(contentLines, line)
		}
	}

	content := strings.Join(contentLines, "\n")
	return topic, content, meta, nil
}

type metaInfo struct {
	topicTitle    string
	hasTopicTitle bool
	sectionTitle  string
	hasSection    bool
}

func mergeMeta(existing, incoming metaInfo) (metaInfo, error) {
	if incoming.hasTopicTitle {
		if existing.hasTopicTitle && existing.topicTitle != incoming.topicTitle {
			return existing, fmt.Errorf("conflicting topic titles: %q vs %q", existing.topicTitle, incoming.topicTitle)
		}
		existing.topicTitle = incoming.topicTitle
		existing.hasTopicTitle = true
	}
	if incoming.hasSection {
		if existing.hasSection && existing.sectionTitle != incoming.sectionTitle {
			return existing, fmt.Errorf("conflicting section titles: %q vs %q", existing.sectionTitle, incoming.sectionTitle)
		}
		existing.sectionTitle = incoming.sectionTitle
		existing.hasSection = true
	}
	return existing, nil
}

func parseMetadata(text string) (metaInfo, string, error) {
	meta := metaInfo{}
	rest := strings.TrimSpace(text)
	for {
		rest = strings.TrimLeft(rest, " \t")
		if rest == "" {
			return meta, "", nil
		}

		key, value, remainder, ok, err := parseKeyValue(rest)
		if err != nil {
			return meta, "", err
		}
		if !ok {
			// Not a metadata token; remaining text is content
			return meta, rest, nil
		}

		switch key {
		case "title":
			if meta.hasTopicTitle && meta.topicTitle != value {
				return meta, "", fmt.Errorf("conflicting topic titles: %q vs %q", meta.topicTitle, value)
			}
			meta.topicTitle = value
			meta.hasTopicTitle = true
		case "section":
			if meta.hasSection && meta.sectionTitle != value {
				return meta, "", fmt.Errorf("conflicting section titles: %q vs %q", meta.sectionTitle, value)
			}
			meta.sectionTitle = value
			meta.hasSection = true
		default:
			// Unknown key means treat as content start
			return meta, rest, nil
		}

		rest = remainder
	}
}

func parseKeyValue(text string) (key string, value string, remainder string, ok bool, err error) {
	text = strings.TrimLeft(text, " \t")
	if text == "" {
		return "", "", "", false, nil
	}

	eq := strings.IndexRune(text, '=')
	if eq <= 0 {
		return "", "", "", false, nil
	}

	key = strings.TrimSpace(text[:eq])
	if key == "" {
		return "", "", "", false, nil
	}

	after := strings.TrimLeft(text[eq+1:], " \t")
	if after == "" {
		return "", "", "", false, fmt.Errorf("missing value for key %q", key)
	}

	if after[0] == '"' {
		after = after[1:]
		end := strings.IndexRune(after, '"')
		if end == -1 {
			return "", "", "", false, fmt.Errorf("unterminated quote for key %q", key)
		}
		value = after[:end]
		remainder = after[end+1:]
		return key, value, remainder, true, nil
	}

	// Unquoted value until next space
	fields := strings.Fields(after)
	if len(fields) == 0 {
		return "", "", "", false, fmt.Errorf("missing value for key %q", key)
	}
	value = fields[0]
	remainder = strings.TrimPrefix(after, value)
	return key, value, remainder, true, nil
}
