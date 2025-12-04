/**
 * lion TypeScript extractor
 *
 * Uses the TypeScript compiler API to extract documentation from TSDoc-style comments.
 *
 * Syntax:
 *   @lion topic-name
 *   @lion topic-name Content on same line
 *   @lion topic-name title="Custom Title" section="Section Name"
 *
 * Example:
 *   /**
 *    * @lion api
 *    * The User interface represents a user in the system.
 *    * It contains the basic user information.
 *    *\/
 *   interface User {
 *     name: string;
 *   }
 */

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

interface ExtractResult {
  entries: DocEntry[];
}

/**
 * Parse metadata from the remainder of a @lion line.
 * Format: title="value" section="value" remaining content
 */
function parseMetadata(text: string): { meta: MetaInfo; content: string } {
  const meta: MetaInfo = {};
  let rest = text.trim();

  const keyValuePattern = /^(\w+)="([^"]*)"\s*/;

  while (rest.length > 0) {
    const match = rest.match(keyValuePattern);
    if (!match) break;

    const [fullMatch, key, value] = match;
    if (key === "title") {
      meta.topicTitle = value;
    } else if (key === "section") {
      meta.sectionTitle = value;
    } else {
      // Unknown key, treat rest as content
      break;
    }
    rest = rest.slice(fullMatch.length);
  }

  return { meta, content: rest.trim() };
}

/**
 * Parse a @lion tag line and extract topic, metadata, and content.
 */
function parseLionTag(
  text: string
): { topic: string; meta: MetaInfo; content: string } | null {
  // Match @lion followed by topic name
  const match = text.match(/^@lion\s+(\S+)(?:\s+(.*))?$/);
  if (!match) return null;

  const topic = match[1];
  const remainder = match[2] || "";

  const { meta, content } = parseMetadata(remainder);

  return { topic, meta, content };
}

/**
 * Extract lion documentation from a JSDoc comment.
 */
function extractFromJSDoc(
  comment: string,
  filePath: string,
  line: number,
  entityName: string
): DocEntry[] {
  const entries: DocEntry[] = [];

  // Remove /** and */ and split into lines
  const content = comment.replace(/^\/\*\*/, "").replace(/\*\/$/, "");
  const lines = content.split("\n");

  let currentTopic: string | null = null;
  let currentMeta: MetaInfo = {};
  let currentContent: string[] = [];
  let topicLine = line;

  for (const rawLine of lines) {
    // Remove leading * and whitespace
    const cleanedLine = rawLine.replace(/^\s*\*\s?/, "");

    // Check for @lion tag
    const lionMatch = parseLionTag(cleanedLine);
    if (lionMatch) {
      // Save previous topic if exists
      if (currentTopic) {
        entries.push({
          topic: currentTopic,
          content: currentContent.join("\n").trim(),
          file: filePath,
          line: topicLine,
          entity: entityName,
          topicTitle: currentMeta.topicTitle,
          sectionTitle: currentMeta.sectionTitle,
        });
      }

      // Start new topic
      currentTopic = lionMatch.topic;
      currentMeta = lionMatch.meta;
      currentContent = lionMatch.content ? [lionMatch.content] : [];
      // Keep the original line number from the comment
      continue;
    }

    // If we're in a topic, add content (skip other JSDoc tags)
    if (currentTopic && !cleanedLine.startsWith("@")) {
      currentContent.push(cleanedLine);
    }
  }

  // Save last topic
  if (currentTopic) {
    entries.push({
      topic: currentTopic,
      content: currentContent.join("\n").trim(),
      file: filePath,
      line: topicLine,
      entity: entityName,
      topicTitle: currentMeta.topicTitle,
      sectionTitle: currentMeta.sectionTitle,
    });
  }

  return entries;
}

/**
 * Get the entity name from a TypeScript node.
 */
function getEntityName(node: ts.Node): string {
  if (ts.isFunctionDeclaration(node) && node.name) {
    return node.name.text;
  }
  if (ts.isClassDeclaration(node) && node.name) {
    return node.name.text;
  }
  if (ts.isInterfaceDeclaration(node)) {
    return node.name.text;
  }
  if (ts.isTypeAliasDeclaration(node)) {
    return node.name.text;
  }
  if (ts.isEnumDeclaration(node)) {
    return node.name.text;
  }
  if (ts.isVariableStatement(node)) {
    const declarations = node.declarationList.declarations;
    if (declarations.length > 0 && ts.isIdentifier(declarations[0].name)) {
      return declarations[0].name.text;
    }
  }
  if (ts.isMethodDeclaration(node) && node.name) {
    if (ts.isIdentifier(node.name)) {
      return node.name.text;
    }
  }
  if (ts.isPropertyDeclaration(node) && node.name) {
    if (ts.isIdentifier(node.name)) {
      return node.name.text;
    }
  }
  if (ts.isModuleDeclaration(node)) {
    if (ts.isIdentifier(node.name)) {
      return node.name.text;
    }
  }
  return "";
}

/**
 * Extract documentation from a TypeScript source file.
 */
function extractFromFile(
  sourceFile: ts.SourceFile,
  filePath: string
): DocEntry[] {
  const entries: DocEntry[] = [];

  // Get all leading comments for a node
  function getLeadingComments(node: ts.Node): string[] {
    const text = sourceFile.getFullText();
    const comments: string[] = [];

    const ranges = ts.getLeadingCommentRanges(text, node.getFullStart());
    if (ranges) {
      for (const range of ranges) {
        const comment = text.slice(range.pos, range.end);
        // Only process JSDoc-style comments
        if (comment.startsWith("/**")) {
          comments.push(comment);
        }
      }
    }

    return comments;
  }

  function visit(node: ts.Node) {
    // Check for JSDoc comments on declarations
    if (
      ts.isFunctionDeclaration(node) ||
      ts.isClassDeclaration(node) ||
      ts.isInterfaceDeclaration(node) ||
      ts.isTypeAliasDeclaration(node) ||
      ts.isEnumDeclaration(node) ||
      ts.isVariableStatement(node) ||
      ts.isModuleDeclaration(node)
    ) {
      const comments = getLeadingComments(node);
      const entityName = getEntityName(node);
      const { line } = sourceFile.getLineAndCharacterOfPosition(node.getStart());

      for (const comment of comments) {
        if (comment.includes("@lion")) {
          const docEntries = extractFromJSDoc(
            comment,
            filePath,
            line + 1,
            entityName
          );
          entries.push(...docEntries);
        }
      }
    }

    // Also check class members
    if (ts.isClassDeclaration(node)) {
      for (const member of node.members) {
        if (ts.isMethodDeclaration(member) || ts.isPropertyDeclaration(member)) {
          const comments = getLeadingComments(member);
          const entityName = getEntityName(member);
          const { line } = sourceFile.getLineAndCharacterOfPosition(
            member.getStart()
          );

          for (const comment of comments) {
            if (comment.includes("@lion")) {
              const docEntries = extractFromJSDoc(
                comment,
                filePath,
                line + 1,
                entityName
              );
              entries.push(...docEntries);
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

/**
 * Extract documentation from all TypeScript files in a directory.
 */
function extractFromDirectory(dir: string): ExtractResult {
  const entries: DocEntry[] = [];

  function walkDir(currentPath: string) {
    const items = fs.readdirSync(currentPath);

    for (const item of items) {
      const fullPath = path.join(currentPath, item);
      const stat = fs.statSync(fullPath);

      if (stat.isDirectory()) {
        // Skip node_modules and hidden directories
        if (!item.startsWith(".") && item !== "node_modules") {
          walkDir(fullPath);
        }
      } else if (stat.isFile()) {
        // Process .ts and .tsx files, skip test files
        if (
          (item.endsWith(".ts") || item.endsWith(".tsx")) &&
          !item.endsWith(".test.ts") &&
          !item.endsWith(".test.tsx") &&
          !item.endsWith(".spec.ts") &&
          !item.endsWith(".spec.tsx") &&
          !item.endsWith(".d.ts")
        ) {
          const content = fs.readFileSync(fullPath, "utf-8");
          const sourceFile = ts.createSourceFile(
            fullPath,
            content,
            ts.ScriptTarget.Latest,
            true,
            item.endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS
          );

          const fileEntries = extractFromFile(sourceFile, fullPath);
          entries.push(...fileEntries);
        }
      }
    }
  }

  walkDir(dir);
  return { entries };
}

// Main entry point
function main() {
  const args = process.argv.slice(2);
  if (args.length < 1) {
    console.error("Usage: lion-ts-extract <directory>");
    process.exit(1);
  }

  const dir = args[0];
  if (!fs.existsSync(dir)) {
    console.error(`Directory not found: ${dir}`);
    process.exit(1);
  }

  const result = extractFromDirectory(dir);
  console.log(JSON.stringify(result, null, 2));
}

main();
