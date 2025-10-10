use crate::CommentsConfig;
use anyhow::Result;
use mdbook::book::Chapter;
use sha2::{Digest, Sha256};
use std::collections::HashMap;

pub struct CommentsProcessor {
    config: CommentsConfig,
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
                let metadata = self.generate_metadata(
                    &block_content,
                    block_index,
                    current_section_index,
                    &heading_stack,
                    path,
                    &prev_content,
                    &next_content,
                );
                
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
        if trimmed.starts_with('-') || trimmed.starts_with('*') || trimmed.starts_with('+') 
            || (trimmed.chars().next().map(|c| c.is_numeric()).unwrap_or(false) && trimmed.contains('.')) {
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
            
            for i in (start + 1)..lines.len() {
                content.push_str(lines[i]);
                content.push('\n');
                count += 1;
                
                if lines[i].trim().starts_with(fence_char) && lines[i].trim().len() >= 3 {
                    break;
                }
            }
            return (content, count);
        }
        
        // Handle regular blocks (paragraphs, list items, etc.)
        for i in start..lines.len() {
            let line = lines[i];
            let trimmed = line.trim();
            
            // Stop at empty line (end of block)
            if trimmed.is_empty() && count > 0 {
                break;
            }
            
            // For paragraphs, stop at heading or special blocks
            if count > 0 && (trimmed.starts_with('#') || trimmed.starts_with("```") 
                || trimmed.starts_with("~~~") || trimmed.starts_with('|')) {
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
        
        let content: Vec<&str> = lines[start..=i].iter().copied().collect();
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
        
        let content: Vec<&str> = lines[start..i].iter().copied().collect();
        Some(content.join("\n"))
    }
    
    fn generate_metadata(
        &self,
        content: &str,
        block_index: usize,
        section_index: usize,
        heading_stack: &[String],
        path: &Option<std::path::PathBuf>,
        prev_content: &Option<String>,
        next_content: &Option<String>,
    ) -> HashMap<String, serde_json::Value> {
        use serde_json::json;
        
        // Generate content hash
        let mut hasher = Sha256::new();
        hasher.update(content.as_bytes());
        let hash = hex::encode(hasher.finalize());
        let short_hash = &hash[..8];
        
        // Generate ID
        let path_str = path.as_ref()
            .and_then(|p| p.to_str())
            .unwrap_or("unknown");
        let id = format!("{}-block-{}-{}", 
            path_str.replace(['/', '\\', '.'], "-"),
            block_index,
            short_hash
        );
        
        let mut metadata = HashMap::new();
        metadata.insert("id".to_string(), json!(id));
        metadata.insert("position".to_string(), json!({
            "file": path_str,
            "block-index": block_index,
            "section-index": section_index,
        }));
        metadata.insert("content".to_string(), json!(content));
        
        let mut context = serde_json::Map::new();
        if let Some(prev) = prev_content {
            context.insert("prev".to_string(), json!(prev));
        }
        if let Some(next) = next_content {
            context.insert("next".to_string(), json!(next));
        }
        context.insert("heading-path".to_string(), json!(heading_stack));
        metadata.insert("context".to_string(), json!(context));
        
        // Get git commit hash (if available)
        if let Ok(output) = std::process::Command::new("git")
            .args(["rev-parse", "--short", "HEAD"])
            .output() {
            if output.status.success() {
                let commit = String::from_utf8_lossy(&output.stdout).trim().to_string();
                metadata.insert("commit".to_string(), json!(commit));
            }
        }
        
        metadata
    }
    
    fn add_comment_link(&self, content: &str, metadata: &HashMap<String, serde_json::Value>) -> String {
        let metadata_json = serde_json::to_string(metadata).unwrap_or_default();
        let metadata_escaped = metadata_json
            .replace('&', "&amp;")
            .replace('"', "&quot;")
            .replace('<', "&lt;")
            .replace('>', "&gt;");
        
        let id = metadata.get("id")
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
