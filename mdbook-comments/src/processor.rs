use crate::{CommentsConfig, CssAsset, JsAsset};
use anyhow::Result;
use mdbook::book::Chapter;
use sha2::{Digest, Sha256};
use std::collections::HashMap;

pub struct CommentsProcessor {
    config: CommentsConfig,
}

#[derive(Debug)]
struct MetadataParams<'a> {
    content: &'a str,
    block_index: usize,
    section_index: usize,
    heading_stack: &'a [String],
    path: &'a Option<std::path::PathBuf>,
    prev_content: &'a Option<String>,
    next_content: &'a Option<String>,
}

impl CommentsProcessor {
    pub fn new(config: CommentsConfig) -> Self {
        Self { config }
    }

    pub fn process_chapter(&self, chapter: &mut Chapter) -> Result<()> {
        // Parse markdown and inject comment metadata
        let processed = self.process_markdown(&chapter.content, &chapter.path)?;
        chapter.content = processed;
        Ok(())
    }

    pub fn inject_assets(&self, chapter: &mut Chapter) -> Result<()> {
        let mut asset_html = String::new();

        // Inject CSS
        if let Some(css_file) = CssAsset::get("comments.css") {
            let css_content = std::str::from_utf8(css_file.data.as_ref())?;
            asset_html.push_str(&format!("<style>\n{}\n</style>\n\n", css_content));
        }

        // Inject shared base module (always required)
        if let Some(base_js) = JsAsset::get("comments-base.js") {
            let base_content = std::str::from_utf8(base_js.data.as_ref())?;
            asset_html.push_str(&format!("<script>\n{}\n</script>\n\n", base_content));
        }

        // Inject backend adapter based on backend type
        let adapter_filename = match self.config.backend_type.as_str() {
            "json-server" => "comments-json-server-adapter.js",
            "supabase" => "comments-supabase-adapter.js",
            "google-sheets" => "comments-googlesheets-adapter.js",
            "custom" => "comments-custom-adapter.js",
            _ => {
                eprintln!("Warning: Unknown backend type '{}', defaulting to json-server", self.config.backend_type);
                "comments-json-server-adapter.js"
            }
        };

        if let Some(adapter_js) = JsAsset::get(adapter_filename) {
            let adapter_content = std::str::from_utf8(adapter_js.data.as_ref())?;
            asset_html.push_str(&format!("<script>\n{}\n</script>\n\n", adapter_content));
        }

        // Prepend assets to chapter content
        // Use HTML comment to prevent markdown processing of the injected content
        chapter.content = format!("<!-- mdbook-comments assets -->\n{}\n<!-- end mdbook-comments assets -->\n\n{}", asset_html, chapter.content);

        Ok(())
    }

    fn process_markdown(&self, content: &str, path: &Option<std::path::PathBuf>) -> Result<String> {
        // Split content into lines for processing
        let lines: Vec<&str> = content.lines().collect();
        let mut result = String::new();
        let mut in_code_block = false;
        let mut code_fence_char = ' ';
        let mut block_index = 0;
        let mut heading_stack: Vec<String> = Vec::new();
        let mut current_section_index = 0;

        let mut i = 0;
        while i < lines.len() {
            let line = lines[i];
            let trimmed = line.trim();

            // Track code blocks
            if trimmed.starts_with("```") || trimmed.starts_with("~~~") {
                if !in_code_block {
                    in_code_block = true;
                    code_fence_char = trimmed.chars().next().unwrap();
                } else if trimmed.starts_with(code_fence_char) {
                    in_code_block = false;
                }
            }

            // Track headings for context
            if !in_code_block && trimmed.starts_with('#') {
                let level = trimmed.chars().take_while(|c| *c == '#').count();
                let heading_text = trimmed.trim_start_matches('#').trim();

                // Update heading stack
                heading_stack.truncate(level - 1);
                if level <= heading_stack.len() {
                    heading_stack[level - 1] = heading_text.to_string();
                } else {
                    heading_stack.push(heading_text.to_string());
                }
                current_section_index = 0;
            }

            // Check if this is a commentable block
            if !in_code_block && self.is_commentable_line(line, &lines, i) {
                // Extract the content block
                let (block_content, block_lines) = self.extract_block(&lines, i);

                // Get context (prev and next blocks)
                let prev_content = self.get_prev_block_content(&lines, i);
                let next_content = self.get_next_block_content(&lines, i + block_lines);

                // Generate metadata
                let params = MetadataParams {
                    content: &block_content,
                    block_index,
                    section_index: current_section_index,
                    heading_stack: &heading_stack,
                    path,
                    prev_content: &prev_content,
                    next_content: &next_content,
                };
                let metadata = self.generate_metadata(&params);

                // Add the block with comment link
                result.push_str(&self.add_comment_link(&block_content, &metadata));
                result.push('\n');

                block_index += 1;
                current_section_index += 1;
                i += block_lines;
                continue;
            }

            result.push_str(line);
            result.push('\n');
            i += 1;
        }

        Ok(result)
    }

    fn is_commentable_line(&self, line: &str, _lines: &[&str], _index: usize) -> bool {
        let trimmed = line.trim();

        // Empty lines are not commentable
        if trimmed.is_empty() {
            return false;
        }

        // Check if it's a heading
        if trimmed.starts_with('#') {
            return self.config.elements.headings;
        }

        // Check if it's a list item
        if trimmed.starts_with('-')
            || trimmed.starts_with('*')
            || trimmed.starts_with('+')
            || (trimmed
                .chars()
                .next()
                .map(|c| c.is_numeric())
                .unwrap_or(false)
                && trimmed.contains('.'))
        {
            return self.config.elements.lists;
        }

        // Check if it's a blockquote
        if trimmed.starts_with('>') {
            return self.config.elements.blockquotes;
        }

        // Check if it's a code block
        if trimmed.starts_with("```") || trimmed.starts_with("~~~") {
            return self.config.elements.code_blocks;
        }

        // Check if it's a table
        if trimmed.starts_with('|') {
            return self.config.elements.tables;
        }

        // Otherwise, it's likely a paragraph
        self.config.elements.paragraphs
    }

    fn extract_block(&self, lines: &[&str], start: usize) -> (String, usize) {
        let mut content = String::new();
        let mut count = 0;
        let first_line = lines[start].trim();

        // Handle code blocks specially
        if first_line.starts_with("```") || first_line.starts_with("~~~") {
            let fence_char = first_line.chars().next().unwrap();
            content.push_str(lines[start]);
            content.push('\n');
            count += 1;

            for line in &lines[(start + 1)..] {
                content.push_str(line);
                content.push('\n');
                count += 1;

                if line.trim().starts_with(fence_char) && line.trim().len() >= 3 {
                    break;
                }
            }
            return (content, count);
        }

        // Handle regular blocks (paragraphs, list items, etc.)
        for line in &lines[start..] {
            let trimmed = line.trim();

            // Stop at empty line (end of block)
            if trimmed.is_empty() && count > 0 {
                break;
            }

            // For paragraphs, stop at heading or special blocks
            if count > 0
                && (trimmed.starts_with('#')
                    || trimmed.starts_with("```")
                    || trimmed.starts_with("~~~")
                    || trimmed.starts_with('|'))
            {
                break;
            }

            if !trimmed.is_empty() {
                content.push_str(line);
                content.push('\n');
                count += 1;
            }
        }

        (content.trim_end().to_string(), count)
    }

    fn get_prev_block_content(&self, lines: &[&str], current: usize) -> Option<String> {
        if current == 0 {
            return None;
        }

        // Search backwards for previous non-empty block
        let mut i = current - 1;
        while i > 0 && lines[i].trim().is_empty() {
            i -= 1;
        }

        if lines[i].trim().is_empty() {
            return None;
        }

        // Find start of this block
        let mut start = i;
        while start > 0 && !lines[start - 1].trim().is_empty() {
            start -= 1;
        }

        let content: Vec<&str> = lines[start..=i].to_vec();
        Some(content.join("\n"))
    }

    fn get_next_block_content(&self, lines: &[&str], current: usize) -> Option<String> {
        if current >= lines.len() {
            return None;
        }

        // Skip empty lines
        let mut i = current;
        while i < lines.len() && lines[i].trim().is_empty() {
            i += 1;
        }

        if i >= lines.len() {
            return None;
        }

        // Find end of this block
        let start = i;
        while i < lines.len() && !lines[i].trim().is_empty() {
            i += 1;
        }

        let content: Vec<&str> = lines[start..i].to_vec();
        Some(content.join("\n"))
    }

    fn generate_metadata(&self, params: &MetadataParams) -> HashMap<String, serde_json::Value> {
        use serde_json::json;

        // Generate content hash
        let mut hasher = Sha256::new();
        hasher.update(params.content.as_bytes());
        let hash = hex::encode(hasher.finalize());
        let short_hash = &hash[..8];

        // Generate ID
        let path_str = params.path.as_ref().and_then(|p| p.to_str()).unwrap_or("unknown");
        let id = format!(
            "{}-block-{}-{}",
            path_str.replace(['/', '\\', '.'], "-"),
            params.block_index,
            short_hash
        );

        let mut metadata = HashMap::new();
        metadata.insert("id".to_string(), json!(id));
        metadata.insert(
            "position".to_string(),
            json!({
                "file": path_str,
                "block-index": params.block_index,
                "section-index": params.section_index,
            }),
        );
        metadata.insert("content".to_string(), json!(params.content));

        let mut context = serde_json::Map::new();
        if let Some(prev) = params.prev_content {
            context.insert("prev".to_string(), json!(prev));
        }
        if let Some(next) = params.next_content {
            context.insert("next".to_string(), json!(next));
        }
        context.insert("heading-path".to_string(), json!(params.heading_stack));
        metadata.insert("context".to_string(), json!(context));

        // Get git commit hash (if available)
        if let Ok(output) = std::process::Command::new("git")
            .args(["rev-parse", "--short", "HEAD"])
            .output()
        {
            if output.status.success() {
                let commit = String::from_utf8_lossy(&output.stdout).trim().to_string();
                metadata.insert("commit".to_string(), json!(commit));
            }
        }

        metadata
    }

    fn add_comment_link(
        &self,
        content: &str,
        metadata: &HashMap<String, serde_json::Value>,
    ) -> String {
        let metadata_json = serde_json::to_string(metadata).unwrap_or_default();
        let metadata_escaped = metadata_json
            .replace('&', "&amp;")
            .replace('"', "&quot;")
            .replace('<', "&lt;")
            .replace('>', "&gt;");

        let id = metadata
            .get("id")
            .and_then(|v| v.as_str())
            .unwrap_or("unknown");

        let link_text = &self.config.ui.link_text;

        format!(
            "{} <span class=\"comment-link-wrapper\" data-comment-id=\"{}\" data-comment-meta=\"{}\"><a href=\"#\" class=\"comment-link\" onclick=\"toggleComments('{}'); return false;\">{}</a></span>",
            content,
            id,
            metadata_escaped,
            id,
            link_text
        )
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{CommentsConfig, ElementsConfig, UiConfig};
    use std::path::PathBuf;

    fn create_test_config() -> CommentsConfig {
        CommentsConfig {
            api_url: "http://localhost:3000".to_string(),
            auth_type: "cookie".to_string(),
            backend_type: "json-server".to_string(),
            similarity_threshold: 0.85,
            orphaned_comments_location: "end-of-chapter".to_string(),
            elements: ElementsConfig {
                paragraphs: true,
                lists: true,
                blockquotes: true,
                code_blocks: true,
                tables: true,
                headings: false,
            },
            ui: UiConfig {
                link_text: "comment".to_string(),
                show_comment_count: true,
            },
        }
    }

    #[test]
    fn test_is_commentable_line_paragraph() {
        let processor = CommentsProcessor::new(create_test_config());
        let lines = vec!["This is a paragraph."];
        
        assert!(processor.is_commentable_line("This is a paragraph.", &lines, 0));
        assert!(!processor.is_commentable_line("", &lines, 0));
        assert!(!processor.is_commentable_line("   ", &lines, 0));
    }

    #[test]
    fn test_is_commentable_line_headings() {
        let mut config = create_test_config();
        config.elements.headings = true;
        let processor = CommentsProcessor::new(config);
        let lines = vec!["# Heading"];
        
        assert!(processor.is_commentable_line("# Heading", &lines, 0));
        assert!(processor.is_commentable_line("## Sub heading", &lines, 0));
        assert!(processor.is_commentable_line("### Deep heading", &lines, 0));
        
        // Test with headings disabled
        let mut config = create_test_config();
        config.elements.headings = false;
        let processor = CommentsProcessor::new(config);
        assert!(!processor.is_commentable_line("# Heading", &lines, 0));
    }

    #[test]
    fn test_is_commentable_line_lists() {
        let processor = CommentsProcessor::new(create_test_config());
        let lines = vec!["- List item"];
        
        assert!(processor.is_commentable_line("- List item", &lines, 0));
        assert!(processor.is_commentable_line("* Another list item", &lines, 0));
        assert!(processor.is_commentable_line("+ Plus list item", &lines, 0));
        assert!(processor.is_commentable_line("1. Ordered list item", &lines, 0));
        assert!(processor.is_commentable_line("42. Numbered item", &lines, 0));
        
        // Test with lists disabled
        let mut config = create_test_config();
        config.elements.lists = false;
        let processor = CommentsProcessor::new(config);
        assert!(!processor.is_commentable_line("- List item", &lines, 0));
    }

    #[test]
    fn test_is_commentable_line_blockquotes() {
        let processor = CommentsProcessor::new(create_test_config());
        let lines = vec!["> Quote"];
        
        assert!(processor.is_commentable_line("> Quote", &lines, 0));
        assert!(processor.is_commentable_line("> Another quote", &lines, 0));
        
        // Test with blockquotes disabled
        let mut config = create_test_config();
        config.elements.blockquotes = false;
        let processor = CommentsProcessor::new(config);
        assert!(!processor.is_commentable_line("> Quote", &lines, 0));
    }

    #[test]
    fn test_is_commentable_line_code_blocks() {
        let processor = CommentsProcessor::new(create_test_config());
        let lines = vec!["```rust"];
        
        assert!(processor.is_commentable_line("```rust", &lines, 0));
        assert!(processor.is_commentable_line("~~~python", &lines, 0));
        
        // Test with code blocks disabled
        let mut config = create_test_config();
        config.elements.code_blocks = false;
        let processor = CommentsProcessor::new(config);
        assert!(!processor.is_commentable_line("```rust", &lines, 0));
    }

    #[test]
    fn test_is_commentable_line_tables() {
        let processor = CommentsProcessor::new(create_test_config());
        let lines = vec!["| Header | Header |"];
        
        assert!(processor.is_commentable_line("| Header | Header |", &lines, 0));
        assert!(processor.is_commentable_line("| Data | Data |", &lines, 0));
        
        // Test with tables disabled
        let mut config = create_test_config();
        config.elements.tables = false;
        let processor = CommentsProcessor::new(config);
        assert!(!processor.is_commentable_line("| Header | Header |", &lines, 0));
    }

    #[test]
    fn test_extract_block_paragraph() {
        let processor = CommentsProcessor::new(create_test_config());
        let lines = vec![
            "This is a paragraph.",
            "It spans multiple lines.",
            "",
            "This is the next paragraph."
        ];
        
        let (content, count) = processor.extract_block(&lines, 0);
        assert_eq!(content, "This is a paragraph.\nIt spans multiple lines.");
        assert_eq!(count, 2);
    }

    #[test]
    fn test_extract_block_code_block() {
        let processor = CommentsProcessor::new(create_test_config());
        let lines = vec![
            "```rust",
            "fn main() {",
            "    println!(\"Hello\");",
            "}",
            "```",
            "",
            "Next paragraph."
        ];
        
        let (content, count) = processor.extract_block(&lines, 0);
        assert_eq!(content, "```rust\nfn main() {\n    println!(\"Hello\");\n}\n```\n");
        assert_eq!(count, 5);
    }

    #[test]
    fn test_extract_block_single_line() {
        let processor = CommentsProcessor::new(create_test_config());
        let lines = vec![
            "Single line paragraph.",
            "",
            "Next paragraph."
        ];
        
        let (content, count) = processor.extract_block(&lines, 0);
        assert_eq!(content, "Single line paragraph.");
        assert_eq!(count, 1);
    }

    #[test]
    fn test_get_prev_block_content() {
        let processor = CommentsProcessor::new(create_test_config());
        let lines = vec![
            "Previous paragraph.",
            "",
            "Current paragraph.",
            "",
            "Next paragraph."
        ];
        
        let prev = processor.get_prev_block_content(&lines, 2);
        assert_eq!(prev, Some("Previous paragraph.".to_string()));
        
        let no_prev = processor.get_prev_block_content(&lines, 0);
        assert_eq!(no_prev, None);
    }

    #[test]
    fn test_get_next_block_content() {
        let processor = CommentsProcessor::new(create_test_config());
        let lines = vec![
            "Previous paragraph.",
            "",
            "Current paragraph.",
            "",
            "Next paragraph."
        ];
        
        let next = processor.get_next_block_content(&lines, 3);
        assert_eq!(next, Some("Next paragraph.".to_string()));
        
        let no_next = processor.get_next_block_content(&lines, 5);
        assert_eq!(no_next, None);
    }

    #[test]
    fn test_generate_metadata() {
        let processor = CommentsProcessor::new(create_test_config());
        let path = Some(PathBuf::from("chapter1.md"));
        let heading_stack = vec!["Chapter 1".to_string(), "Section 1".to_string()];
        
        let params = MetadataParams {
            content: "Test content for metadata",
            block_index: 0,
            section_index: 2,
            heading_stack: &heading_stack,
            path: &path,
            prev_content: &Some("Previous content".to_string()),
            next_content: &Some("Next content".to_string()),
        };
        
        let metadata = processor.generate_metadata(&params);
        
        // Test ID generation
        let id = metadata.get("id").unwrap().as_str().unwrap();
        assert!(id.starts_with("chapter1-md-block-0-"));
        assert_eq!(id.len(), "chapter1-md-block-0-".len() + 8); // 8 char hash
        
        // Test position
        let position = metadata.get("position").unwrap().as_object().unwrap();
        assert_eq!(position.get("file").unwrap().as_str().unwrap(), "chapter1.md");
        assert_eq!(position.get("block-index").unwrap().as_u64().unwrap(), 0);
        assert_eq!(position.get("section-index").unwrap().as_u64().unwrap(), 2);
        
        // Test content
        assert_eq!(metadata.get("content").unwrap().as_str().unwrap(), "Test content for metadata");
        
        // Test context
        let context = metadata.get("context").unwrap().as_object().unwrap();
        assert_eq!(context.get("prev").unwrap().as_str().unwrap(), "Previous content");
        assert_eq!(context.get("next").unwrap().as_str().unwrap(), "Next content");
        let heading_path = context.get("heading-path").unwrap().as_array().unwrap();
        assert_eq!(heading_path.len(), 2);
        assert_eq!(heading_path[0].as_str().unwrap(), "Chapter 1");
        assert_eq!(heading_path[1].as_str().unwrap(), "Section 1");
    }

    #[test]
    fn test_add_comment_link() {
        let processor = CommentsProcessor::new(create_test_config());
        let mut metadata = HashMap::new();
        metadata.insert("id".to_string(), serde_json::json!("test-id-12345678"));
        
        let result = processor.add_comment_link("Test paragraph content.", &metadata);
        
        assert!(result.contains("Test paragraph content."));
        assert!(result.contains("data-comment-id=\"test-id-12345678\""));
        assert!(result.contains("onclick=\"toggleComments('test-id-12345678')"));
        assert!(result.contains("comment")); // link text
        assert!(result.contains("comment-link-wrapper"));
        assert!(result.contains("comment-link"));
    }

    #[test]
    fn test_add_comment_link_escapes_metadata() {
        let processor = CommentsProcessor::new(create_test_config());
        let mut metadata = HashMap::new();
        metadata.insert("id".to_string(), serde_json::json!("test-id"));
        metadata.insert("content".to_string(), serde_json::json!("Content with \"quotes\" & <tags>"));
        
        let result = processor.add_comment_link("Test content.", &metadata);
        
        // Check that special characters are escaped in the data attribute
        // The metadata is JSON-encoded then HTML-escaped
        assert!(result.contains("&quot;"));
        assert!(result.contains("&amp;"));
        assert!(result.contains("&lt;"));
    }

    #[test]
    fn test_process_markdown_simple_paragraph() {
        let processor = CommentsProcessor::new(create_test_config());
        let content = "This is a simple paragraph.\n\nThis is another paragraph.";
        let path = Some(PathBuf::from("test.md"));
        
        let result = processor.process_markdown(content, &path).unwrap();
        
        // Should contain both paragraphs with comment links
        assert!(result.contains("This is a simple paragraph."));
        assert!(result.contains("This is another paragraph."));
        assert_eq!(result.matches("comment-link-wrapper").count(), 2);
        assert_eq!(result.matches("data-comment-id=").count(), 2);
    }

    #[test]
    fn test_process_markdown_with_heading() {
        let mut config = create_test_config();
        config.elements.headings = true;
        let processor = CommentsProcessor::new(config);
        
        let content = "# Chapter Title\n\nThis is a paragraph under the heading.";
        let path = Some(PathBuf::from("test.md"));
        
        let result = processor.process_markdown(content, &path).unwrap();
        
        // Should have comment links for both heading and paragraph
        assert_eq!(result.matches("comment-link-wrapper").count(), 2);
        assert!(result.contains("Chapter Title"));
        assert!(result.contains("This is a paragraph"));
    }

    #[test]
    fn test_process_markdown_code_block_handling() {
        let processor = CommentsProcessor::new(create_test_config());
        let content = "Before code.\n\n```rust\nfn main() {\n    // This should not be processed\n}\n```\n\nAfter code.";
        let path = Some(PathBuf::from("test.md"));
        
        let result = processor.process_markdown(content, &path).unwrap();
        
        // Should have comment links for commentable blocks
        // Note: The exact count depends on how the markdown is parsed
        assert!(result.matches("comment-link-wrapper").count() >= 2);
        assert!(result.contains("Before code."));
        assert!(result.contains("After code."));
        assert!(result.contains("```rust"));
    }

    #[test]
    fn test_hash_consistency() {
        let processor = CommentsProcessor::new(create_test_config());
        let path = Some(PathBuf::from("test.md"));
        let heading_stack = vec![];
        
        let params = MetadataParams {
            content: "Same content",
            block_index: 0,
            section_index: 0,
            heading_stack: &heading_stack,
            path: &path,
            prev_content: &None,
            next_content: &None,
        };
        
        let metadata1 = processor.generate_metadata(&params);
        let metadata2 = processor.generate_metadata(&params);
        
        // Same content should generate same ID
        assert_eq!(metadata1.get("id"), metadata2.get("id"));
    }
}
