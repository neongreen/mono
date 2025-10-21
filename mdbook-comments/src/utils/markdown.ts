/**
 * Markdown processing utilities for comment formatting
 */

import { marked } from 'marked';
import DOMPurify from 'dompurify';

/**
 * Configure marked for basic, safe markdown parsing
 */
function configureMarked() {
  // Configure marked with basic options for internal documentation
  marked.setOptions({
    // Disable dangerous features
    sanitize: false, // We use DOMPurify instead
    silent: true, // Don't throw on malformed markdown
    breaks: true, // Convert line breaks to <br>
    gfm: true, // GitHub Flavored Markdown (for better code formatting)
  });

  // Custom renderer for security and simplicity
  const renderer = new marked.Renderer();

  // Only allow basic formatting - no headers, tables, etc.
  renderer.heading = (text: string) => text; // Strip headers, just return text
  renderer.table = () => ''; // Remove tables
  renderer.tablecell = () => '';
  renderer.tablerow = () => '';

  // Safe link handling - add target="_blank" and rel attributes
  renderer.link = (href: string, title: string | null, text: string) => {
    const titleAttr = title ? ` title="${escapeHtml(title)}"` : '';
    return `<a href="${escapeHtml(href)}" target="_blank" rel="noopener noreferrer"${titleAttr}>${text}</a>`;
  };

  marked.use({ renderer });
}

/**
 * Escape HTML characters to prevent XSS
 */
function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

/**
 * Parse markdown text and return safe HTML
 * 
 * Supports:
 * - **bold** and *italic* text
 * - `inline code`
 * - [links](url)
 * - Line breaks and paragraphs
 * - Basic lists
 * 
 * Security: All output is sanitized with DOMPurify
 */
export function parseMarkdown(text: string): string {
  if (!text || typeof text !== 'string') {
    return '';
  }

  // Configure marked on first use
  configureMarked();

  try {
    // Parse markdown
    const html = marked.parse(text) as string;

    // Sanitize HTML with DOMPurify - only allow safe tags and attributes
    const cleanHtml = DOMPurify.sanitize(html, {
      ALLOWED_TAGS: [
        'p', 'br', 'strong', 'b', 'em', 'i', 'code', 'pre', 
        'a', 'ul', 'ol', 'li', 'blockquote'
      ],
      ALLOWED_ATTR: ['href', 'title', 'target', 'rel'],
      // Ensure links open in new tabs safely
      ADD_ATTR: ['target', 'rel'],
    });

    return cleanHtml;
  } catch (error) {
    console.error('Failed to parse markdown:', error);
    // Fallback to escaped plain text
    return escapeHtml(text);
  }
}

/**
 * Check if text contains markdown formatting
 */
export function hasMarkdownFormatting(text: string): boolean {
  if (!text) return false;
  
  // Check for common markdown patterns
  const markdownPatterns = [
    /\*\*[^*]+\*\*/, // **bold**
    /\*[^*]+\*/, // *italic*
    /`[^`]+`/, // `code`
    /\[[^\]]+\]\([^)]+\)/, // [link](url)
    /^[-*+]\s/, // list items
    /^\d+\.\s/, // numbered lists
    /^>\s/, // blockquotes
  ];

  return markdownPatterns.some(pattern => pattern.test(text));
}