use crate::{CommentsConfig, CssAsset, JsAsset};
use anyhow::{anyhow, Result};
use mdbook::book::Chapter;
use pulldown_cmark::{CowStr, Event, HeadingLevel, Options, Parser, Tag};
use pulldown_cmark_to_cmark::cmark;
use sha2::{Digest, Sha256};
use std::collections::HashMap;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum SpanPosition {
    BeforeEnd,
    AfterEnd,
}

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
                eprintln!(
                    "Warning: Unknown backend type '{}', defaulting to json-server",
                    self.config.backend_type
                );
                "comments-json-server-adapter.js"
            }
        };

        if let Some(adapter_js) = JsAsset::get(adapter_filename) {
            let adapter_content = std::str::from_utf8(adapter_js.data.as_ref())?;
            asset_html.push_str(&format!("<script>\n{}\n</script>\n\n", adapter_content));
        }

        // Prepend assets to chapter content
        // Use HTML comment to prevent markdown processing of the injected content
        chapter.content = format!(
            "<!-- mdbook-comments assets -->\n{}\n<!-- end mdbook-comments assets -->\n\n{}",
            asset_html, chapter.content
        );

        Ok(())
    }

    fn process_markdown(&self, content: &str, path: &Option<std::path::PathBuf>) -> Result<String> {
        #[derive(Debug)]
        struct BlockRecord {
            end_index: usize,
            content: String,
            heading_path: Vec<String>,
            block_index: usize,
            section_index: usize,
            prev_content: Option<String>,
            next_content: Option<String>,
            span_position: SpanPosition,
        }

        #[derive(Debug)]
        struct StackEntry<'a> {
            tag: Tag<'a>,
            start_index: usize,
            is_commentable: bool,
            span_position: SpanPosition,
            heading_level: Option<HeadingLevel>,
        }

        let mut events: Vec<Event<'_>> = Vec::new();
        let mut stack: Vec<StackEntry<'_>> = Vec::new();
        let mut blocks: Vec<BlockRecord> = Vec::new();
        let mut block_index = 0usize;
        let mut section_index = 0usize;
        let mut heading_stack: Vec<String> = Vec::new();
        let mut active_commentable_depth = 0usize;

        let parser = Parser::new_ext(content, Options::all());
        for event in parser {
            if let Event::Start(tag) = &event {
                let is_candidate = self.is_commentable_tag(tag);
                let is_commentable = is_candidate && active_commentable_depth == 0usize;
                if is_commentable {
                    active_commentable_depth += 1;
                }

                stack.push(StackEntry {
                    tag: tag.clone(),
                    start_index: events.len(),
                    is_commentable,
                    span_position: self.span_position_for_tag(tag),
                    heading_level: match tag {
                        Tag::Heading(level, ..) => Some(*level),
                        _ => None,
                    },
                });
            }

            events.push(event.clone());

            if let Event::End(tag) = &event {
                let entry = stack
                    .pop()
                    .ok_or_else(|| anyhow!("unbalanced markdown tags"))?;

                debug_assert!(
                    entry.tag == *tag,
                    "mismatched tags: start={:?}, end={:?}",
                    entry.tag,
                    tag
                );

                if let Some(level) = entry.heading_level {
                    let slice = &events[entry.start_index..events.len()];
                    let heading_text = self.extract_plain_text(slice);
                    self.update_heading_stack(&mut heading_stack, level, heading_text);
                    section_index = 0;
                }

                if entry.is_commentable {
                    active_commentable_depth = active_commentable_depth.saturating_sub(1);

                    let slice = &events[entry.start_index..events.len()];
                    let content = self
                        .render_events_to_markdown(slice)?
                        .trim_end_matches('\n')
                        .to_string();

                    blocks.push(BlockRecord {
                        end_index: events.len() - 1,
                        content,
                        heading_path: heading_stack.clone(),
                        block_index,
                        section_index,
                        prev_content: None,
                        next_content: None,
                        span_position: entry.span_position,
                    });

                    block_index += 1;
                    section_index += 1;
                }
            }
        }

        for idx in 0..blocks.len() {
            if idx > 0 {
                blocks[idx].prev_content = Some(blocks[idx - 1].content.clone());
            }
            if idx + 1 < blocks.len() {
                blocks[idx].next_content = Some(blocks[idx + 1].content.clone());
            }
        }

        let mut block_metadata = Vec::with_capacity(blocks.len());
        for block in &blocks {
            let params = MetadataParams {
                content: &block.content,
                block_index: block.block_index,
                section_index: block.section_index,
                heading_stack: &block.heading_path,
                path,
                prev_content: &block.prev_content,
                next_content: &block.next_content,
            };
            block_metadata.push(self.generate_metadata(&params));
        }

        let mut output_events: Vec<Event<'_>> = Vec::new();
        let mut end_to_block: HashMap<usize, usize> = HashMap::new();
        for (idx, block) in blocks.iter().enumerate() {
            end_to_block.insert(block.end_index, idx);
        }

        for (idx, event) in events.into_iter().enumerate() {
            if let Some(&block_idx) = end_to_block.get(&idx) {
                if blocks[block_idx].span_position == SpanPosition::BeforeEnd {
                    let html = self.add_comment_link(&block_metadata[block_idx], " ");
                    output_events.push(Event::Html(CowStr::from(html)));
                }
            }

            output_events.push(event);

            if let Some(&block_idx) = end_to_block.get(&idx) {
                if blocks[block_idx].span_position == SpanPosition::AfterEnd {
                    let html = self.add_comment_link(&block_metadata[block_idx], " ");
                    output_events.push(Event::Html(CowStr::from(html)));
                }
            }
        }

        let mut rendered = String::new();
        cmark(output_events.into_iter(), &mut rendered)
            .map_err(|_| anyhow!("failed to render markdown from events"))?;
        Ok(rendered)
    }

    fn is_commentable_tag(&self, tag: &Tag<'_>) -> bool {
        match tag {
            Tag::Paragraph => self.config.elements.paragraphs,
            Tag::Heading(..) => self.config.elements.headings,
            Tag::BlockQuote => self.config.elements.blockquotes,
            Tag::Item => self.config.elements.lists,
            Tag::CodeBlock(_) => self.config.elements.code_blocks,
            Tag::Table(_) => self.config.elements.tables,
            _ => false,
        }
    }

    fn span_position_for_tag(&self, tag: &Tag<'_>) -> SpanPosition {
        match tag {
            Tag::CodeBlock(_) | Tag::Table(_) => SpanPosition::AfterEnd,
            _ => SpanPosition::BeforeEnd,
        }
    }

    fn render_events_to_markdown(&self, events: &[Event<'_>]) -> Result<String> {
        let mut buf = String::new();
        cmark(events.iter().cloned(), &mut buf)
            .map_err(|_| anyhow!("failed to render markdown slice"))?;
        Ok(buf)
    }

    fn extract_plain_text(&self, events: &[Event<'_>]) -> String {
        let mut text = String::new();
        for event in events {
            match event {
                Event::Text(t) | Event::Code(t) | Event::Html(t) => {
                    text.push_str(t.as_ref());
                }
                Event::SoftBreak | Event::HardBreak => {
                    text.push(' ');
                }
                _ => {}
            }
        }
        text.trim().to_string()
    }

    fn update_heading_stack(
        &self,
        heading_stack: &mut Vec<String>,
        level: HeadingLevel,
        text: String,
    ) {
        let level = level as usize;
        if level == 0 {
            return;
        }
        if level > 1 {
            heading_stack.truncate(level - 1);
        } else {
            heading_stack.clear();
        }
        heading_stack.push(text);
    }

    fn generate_metadata(&self, params: &MetadataParams) -> HashMap<String, serde_json::Value> {
        use serde_json::json;

        // Generate content hash
        let mut hasher = Sha256::new();
        hasher.update(params.content.as_bytes());
        let hash = hex::encode(hasher.finalize());
        let short_hash = &hash[..8];

        // Generate ID
        let path_str = params
            .path
            .as_ref()
            .and_then(|p| p.to_str())
            .unwrap_or("unknown");
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
        metadata: &HashMap<String, serde_json::Value>,
        prefix: &str,
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
            "{}<span class=\"comment-link-wrapper\" data-comment-id=\"{}\" data-comment-meta=\"{}\"><a href=\"#\" class=\"comment-link\" onclick=\"toggleComments('{}'); return false;\">{}</a></span>",
            prefix,
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
    use serde_json::Value;
    use std::collections::HashMap;
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

    fn extract_metadata(markdown: &str) -> Vec<Value> {
        let mut values = Vec::new();
        let pattern = "data-comment-meta=\"";
        let mut search_start = 0;

        while let Some(idx) = markdown[search_start..].find(pattern) {
            let start = search_start + idx + pattern.len();
            if let Some(end_offset) = markdown[start..].find('\"') {
                let end = start + end_offset;
                let raw = &markdown[start..end];
                let decoded = raw
                    .replace("&quot;", "\"")
                    .replace("&lt;", "<")
                    .replace("&gt;", ">")
                    .replace("&amp;", "&");
                if let Ok(value) = serde_json::from_str::<Value>(&decoded) {
                    values.push(value);
                }
                search_start = end + 1;
            } else {
                break;
            }
        }

        values
    }

    #[test]
    fn test_process_markdown_paragraphs_have_comment_links() {
        let processor = CommentsProcessor::new(create_test_config());
        let path = Some(PathBuf::from("sample.md"));
        let markdown = "First paragraph.\n\nSecond paragraph.";
        let result = processor.process_markdown(markdown, &path).unwrap();

        assert_eq!(result.match_indices("comment-link-wrapper").count(), 2);
        assert!(result.contains("First paragraph. <span class=\"comment-link-wrapper\""));
        assert!(result.contains("Second paragraph. <span class=\"comment-link-wrapper\""));
    }

    #[test]
    fn test_process_markdown_code_block_and_following_paragraphs() {
        let processor = CommentsProcessor::new(create_test_config());
        let path = Some(PathBuf::from("code.md"));
        let markdown = "```bash\nls\n```\n\nAfter code.";
        let result = processor.process_markdown(markdown, &path).unwrap();

        assert_eq!(result.match_indices("comment-link-wrapper").count(), 2);
        let metadata = extract_metadata(&result);
        assert_eq!(metadata.len(), 2);
        let first_content = metadata[0]["content"].as_str().unwrap();
        assert!(first_content.contains("```"));
        assert!(first_content.contains("ls"));
        let second_content = metadata[1]["content"].as_str().unwrap();
        assert!(second_content.contains("After code."));
    }

    #[test]
    fn test_process_markdown_headings_reset_section_index() {
        let mut config = create_test_config();
        config.elements.headings = true;
        let processor = CommentsProcessor::new(config);
        let path = Some(PathBuf::from("chapter.md"));
        let markdown = "# Heading\n\nIntro paragraph.\n\n## Subheading\n\nBody paragraph.";
        let result = processor.process_markdown(markdown, &path).unwrap();

        let metadata = extract_metadata(&result);
        assert_eq!(metadata.len(), 4);
        assert_eq!(metadata[0]["position"]["section-index"].as_u64(), Some(0));
        assert_eq!(metadata[1]["position"]["section-index"].as_u64(), Some(1));
        assert_eq!(metadata[2]["position"]["section-index"].as_u64(), Some(0));
        assert_eq!(metadata[3]["position"]["section-index"].as_u64(), Some(1));
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

        let id = metadata.get("id").unwrap().as_str().unwrap();
        assert!(id.starts_with("chapter1-md-block-0-"));
        assert_eq!(id.len(), "chapter1-md-block-0-".len() + 8);

        let position = metadata.get("position").unwrap().as_object().unwrap();
        assert_eq!(
            position.get("file").unwrap().as_str().unwrap(),
            "chapter1.md"
        );
        assert_eq!(position.get("block-index").unwrap().as_u64().unwrap(), 0);
        assert_eq!(position.get("section-index").unwrap().as_u64().unwrap(), 2);

        assert_eq!(
            metadata.get("content").unwrap().as_str().unwrap(),
            "Test content for metadata"
        );

        let context = metadata.get("context").unwrap().as_object().unwrap();
        assert_eq!(
            context.get("prev").unwrap().as_str().unwrap(),
            "Previous content"
        );
        assert_eq!(
            context.get("next").unwrap().as_str().unwrap(),
            "Next content"
        );
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

        let result = processor.add_comment_link(&metadata, " ");

        assert!(result.starts_with(" <span"));
        assert!(result.contains("data-comment-id=\"test-id-12345678\""));
        assert!(result.contains("onclick=\"toggleComments('test-id-12345678'); return false;\""));
        assert!(result.contains("comment-link-wrapper"));
    }

    #[test]
    fn test_add_comment_link_escapes_metadata() {
        let processor = CommentsProcessor::new(create_test_config());
        let mut metadata = HashMap::new();
        metadata.insert("id".to_string(), serde_json::json!("test-id"));
        metadata.insert(
            "content".to_string(),
            serde_json::json!("Content with \"quotes\" & <tags>"),
        );

        let result = processor.add_comment_link(&metadata, " ");

        assert!(result.contains("&quot;"));
        assert!(result.contains("&amp;"));
        assert!(result.contains("&lt;"));
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

        assert_eq!(metadata1.get("id"), metadata2.get("id"));
    }
}
