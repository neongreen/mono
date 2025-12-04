package extractor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// tsExtractResult represents the JSON output from the TypeScript extractor.
type tsExtractResult struct {
	Entries []tsDocEntry `json:"entries"`
}

// tsDocEntry represents a single documentation entry from TypeScript.
type tsDocEntry struct {
	Topic        string `json:"topic"`
	Content      string `json:"content"`
	File         string `json:"file"`
	Line         int    `json:"line"`
	Entity       string `json:"entity"`
	TopicTitle   string `json:"topicTitle,omitempty"`
	SectionTitle string `json:"sectionTitle,omitempty"`
}

// tsExtractorScript is the embedded TypeScript extractor script.
// It uses the TypeScript compiler API to parse TypeScript files and extract @lion tags.
const tsExtractorScript = `// @ts-nocheck
import * as ts from "typescript";
import * as fs from "fs";
import * as path from "path";

interface DocEntry {
  topic: string;
  content: string;
  file: string;
  line: number;
  entity: string;
  topicTitle?: string;
  sectionTitle?: string;
}

interface MetaInfo {
  topicTitle?: string;
  sectionTitle?: string;
}

function parseMetadata(text: string): { meta: MetaInfo; content: string } {
  const meta: MetaInfo = {};
  let rest = text.trim();
  const keyValuePattern = /^(\w+)="([^"]*)"\s*/;
  while (rest.length > 0) {
    const match = rest.match(keyValuePattern);
    if (!match) break;
    const [fullMatch, key, value] = match;
    if (key === "title") meta.topicTitle = value;
    else if (key === "section") meta.sectionTitle = value;
    else break;
    rest = rest.slice(fullMatch.length);
  }
  return { meta, content: rest.trim() };
}

function parseLionTag(text: string): { topic: string; meta: MetaInfo; content: string } | null {
  const match = text.match(/^@lion\s+(\S+)(?:\s+(.*))?$/);
  if (!match) return null;
  const topic = match[1];
  const remainder = match[2] || "";
  const { meta, content } = parseMetadata(remainder);
  return { topic, meta, content };
}

function extractFromJSDoc(comment: string, filePath: string, line: number, entityName: string): DocEntry[] {
  const entries: DocEntry[] = [];
  const content = comment.replace(/^\/\*\*/, "").replace(/\*\/$/, "");
  const lines = content.split("\n");
  let currentTopic: string | null = null;
  let currentMeta: MetaInfo = {};
  let currentContent: string[] = [];
  let topicLine = line;

  for (const rawLine of lines) {
    const cleanedLine = rawLine.replace(/^\s*\*\s?/, "");
    const lionMatch = parseLionTag(cleanedLine);
    if (lionMatch) {
      if (currentTopic) {
        entries.push({
          topic: currentTopic, content: currentContent.join("\n").trim(),
          file: filePath, line: topicLine, entity: entityName,
          topicTitle: currentMeta.topicTitle, sectionTitle: currentMeta.sectionTitle,
        });
      }
      currentTopic = lionMatch.topic;
      currentMeta = lionMatch.meta;
      currentContent = lionMatch.content ? [lionMatch.content] : [];
      continue;
    }
    if (currentTopic && !cleanedLine.startsWith("@")) {
      currentContent.push(cleanedLine);
    }
  }
  if (currentTopic) {
    entries.push({
      topic: currentTopic, content: currentContent.join("\n").trim(),
      file: filePath, line: topicLine, entity: entityName,
      topicTitle: currentMeta.topicTitle, sectionTitle: currentMeta.sectionTitle,
    });
  }
  return entries;
}

function getEntityName(node: ts.Node): string {
  if (ts.isFunctionDeclaration(node) && node.name) return node.name.text;
  if (ts.isClassDeclaration(node) && node.name) return node.name.text;
  if (ts.isInterfaceDeclaration(node)) return node.name.text;
  if (ts.isTypeAliasDeclaration(node)) return node.name.text;
  if (ts.isEnumDeclaration(node)) return node.name.text;
  if (ts.isVariableStatement(node)) {
    const declarations = node.declarationList.declarations;
    if (declarations.length > 0 && ts.isIdentifier(declarations[0].name)) return declarations[0].name.text;
  }
  if (ts.isMethodDeclaration(node) && node.name && ts.isIdentifier(node.name)) return node.name.text;
  if (ts.isPropertyDeclaration(node) && node.name && ts.isIdentifier(node.name)) return node.name.text;
  if (ts.isModuleDeclaration(node) && ts.isIdentifier(node.name)) return node.name.text;
  return "";
}

function extractFromFile(sourceFile: ts.SourceFile, filePath: string): DocEntry[] {
  const entries: DocEntry[] = [];
  function getLeadingComments(node: ts.Node): string[] {
    const text = sourceFile.getFullText();
    const comments: string[] = [];
    const ranges = ts.getLeadingCommentRanges(text, node.getFullStart());
    if (ranges) {
      for (const range of ranges) {
        const comment = text.slice(range.pos, range.end);
        if (comment.startsWith("/**")) comments.push(comment);
      }
    }
    return comments;
  }

  function visit(node: ts.Node) {
    if (ts.isFunctionDeclaration(node) || ts.isClassDeclaration(node) || ts.isInterfaceDeclaration(node) ||
        ts.isTypeAliasDeclaration(node) || ts.isEnumDeclaration(node) || ts.isVariableStatement(node) ||
        ts.isModuleDeclaration(node)) {
      const comments = getLeadingComments(node);
      const entityName = getEntityName(node);
      const { line } = sourceFile.getLineAndCharacterOfPosition(node.getStart());
      for (const comment of comments) {
        if (comment.includes("@lion")) {
          entries.push(...extractFromJSDoc(comment, filePath, line + 1, entityName));
        }
      }
    }
    if (ts.isClassDeclaration(node)) {
      for (const member of node.members) {
        if (ts.isMethodDeclaration(member) || ts.isPropertyDeclaration(member)) {
          const comments = getLeadingComments(member);
          const entityName = getEntityName(member);
          const { line } = sourceFile.getLineAndCharacterOfPosition(member.getStart());
          for (const comment of comments) {
            if (comment.includes("@lion")) {
              entries.push(...extractFromJSDoc(comment, filePath, line + 1, entityName));
            }
          }
        }
      }
    }
    ts.forEachChild(node, visit);
  }
  visit(sourceFile);
  return entries;
}

function extractFromDirectory(dir: string): { entries: DocEntry[] } {
  const entries: DocEntry[] = [];
  function walkDir(currentPath: string) {
    const items = fs.readdirSync(currentPath);
    for (const item of items) {
      const fullPath = path.join(currentPath, item);
      const stat = fs.statSync(fullPath);
      if (stat.isDirectory()) {
        if (!item.startsWith(".") && item !== "node_modules") walkDir(fullPath);
      } else if (stat.isFile()) {
        if ((item.endsWith(".ts") || item.endsWith(".tsx")) &&
            !item.endsWith(".test.ts") && !item.endsWith(".test.tsx") &&
            !item.endsWith(".spec.ts") && !item.endsWith(".spec.tsx") && !item.endsWith(".d.ts")) {
          const content = fs.readFileSync(fullPath, "utf-8");
          const sourceFile = ts.createSourceFile(fullPath, content, ts.ScriptTarget.Latest, true,
            item.endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS);
          entries.push(...extractFromFile(sourceFile, fullPath));
        }
      }
    }
  }
  walkDir(dir);
  return { entries };
}

const dir = process.argv[2];
if (!dir || !fs.existsSync(dir)) {
  console.error("Usage: extract <directory>");
  process.exit(1);
}
console.log(JSON.stringify(extractFromDirectory(dir)));
`

// ExtractTypeScript extracts lion documentation from TypeScript files in a directory.
// It uses the TypeScript compiler API via an embedded script that is compiled with tsc at runtime.
func ExtractTypeScript(dir string) (map[string][]DocEntry, error) {
	// Check if there are any TypeScript files first
	hasTS := false
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx")) {
			// Skip test and declaration files
			if !strings.HasSuffix(path, ".test.ts") &&
				!strings.HasSuffix(path, ".test.tsx") &&
				!strings.HasSuffix(path, ".spec.ts") &&
				!strings.HasSuffix(path, ".spec.tsx") &&
				!strings.HasSuffix(path, ".d.ts") {
				hasTS = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return nil, err
	}

	if !hasTS {
		return nil, nil
	}

	// Create a temp directory for the extractor
	tmpDir, err := os.MkdirTemp("", "lion-ts-extractor-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write the TypeScript source file
	tsFile := filepath.Join(tmpDir, "extractor.ts")
	if err := os.WriteFile(tsFile, []byte(tsExtractorScript), 0644); err != nil {
		return nil, fmt.Errorf("failed to write extractor script: %w", err)
	}

	// Find the global node_modules directory and symlink typescript
	// This allows tsc to find the typescript module for type checking
	npmRootCmd := exec.Command("npm", "root", "-g")
	npmRootOutput, err := npmRootCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to find global node_modules: %w", err)
	}
	globalNodeModules := strings.TrimSpace(string(npmRootOutput))

	// Create node_modules directory with symlink to global typescript
	nodeModulesDir := filepath.Join(tmpDir, "node_modules")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create node_modules directory: %w", err)
	}
	tsModulePath := filepath.Join(globalNodeModules, "typescript")
	if err := os.Symlink(tsModulePath, filepath.Join(nodeModulesDir, "typescript")); err != nil {
		return nil, fmt.Errorf("failed to symlink typescript module: %w", err)
	}

	// Compile the TypeScript file using tsc directly (no tsconfig needed)
	jsFile := filepath.Join(tmpDir, "extractor.js")
	tscCmd := exec.Command("tsc",
		"--target", "ES2020",
		"--module", "CommonJS",
		"--moduleResolution", "node",
		"--esModuleInterop",
		"--skipLibCheck",
		"--outDir", tmpDir,
		tsFile,
	)
	if output, err := tscCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to compile TypeScript extractor: %s", string(output))
	}

	// Run the compiled JavaScript file
	nodeCmd := exec.Command("node", jsFile, dir)
	output, err := nodeCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("TypeScript extractor failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run TypeScript extractor: %w", err)
	}

	// Parse the JSON output
	var result tsExtractResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse TypeScript extractor output: %w", err)
	}

	// Convert to DocEntry map
	docs := make(map[string][]DocEntry)
	for _, entry := range result.Entries {
		docEntry := DocEntry{
			Topic:         entry.Topic,
			Content:       entry.Content,
			File:          entry.File,
			Line:          entry.Line,
			Entity:        entry.Entity,
			TopicTitle:    entry.TopicTitle,
			HasTopicTitle: entry.TopicTitle != "",
			SectionTitle:  entry.SectionTitle,
			HasSection:    entry.SectionTitle != "",
		}
		docs[entry.Topic] = append(docs[entry.Topic], docEntry)
	}

	return docs, nil
}
