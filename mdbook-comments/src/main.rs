use anyhow::Result;
use mdbook::book::{Book, BookItem};
use mdbook::preprocess::{CmdPreprocessor, Preprocessor, PreprocessorContext};
use serde::{Deserialize, Serialize};
use std::io;
use std::process;

mod processor;

use processor::CommentsProcessor;

/// Configuration for the comments preprocessor
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default, rename_all = "kebab-case")]
pub struct CommentsConfig {
    /// API endpoint for comment storage
    pub api_url: String,
    /// Authentication type (cookie, bearer-token, oauth)
    pub auth_type: String,
    /// Similarity threshold for fuzzy matching (0.0 - 1.0)
    pub similarity_threshold: f64,
    /// Where to show orphaned comments (end-of-chapter, end-of-page)
    pub orphaned_comments_location: String,
    /// Configuration for commentable elements
    pub elements: ElementsConfig,
    /// UI customization
    pub ui: UiConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default, rename_all = "kebab-case")]
pub struct ElementsConfig {
    pub paragraphs: bool,
    pub lists: bool,
    pub blockquotes: bool,
    pub code_blocks: bool,
    pub tables: bool,
    pub headings: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default, rename_all = "kebab-case")]
pub struct UiConfig {
    /// Text for comment links
    pub link_text: String,
    /// Show comment count after link
    pub show_comment_count: bool,
}

impl Default for CommentsConfig {
    fn default() -> Self {
        Self {
            api_url: String::from("http://localhost:3000/api"),
            auth_type: String::from("cookie"),
            similarity_threshold: 0.85,
            orphaned_comments_location: String::from("end-of-chapter"),
            elements: ElementsConfig::default(),
            ui: UiConfig::default(),
        }
    }
}

impl Default for ElementsConfig {
    fn default() -> Self {
        Self {
            paragraphs: true,
            lists: true,
            blockquotes: true,
            code_blocks: true,
            tables: true,
            headings: false,
        }
    }
}

impl Default for UiConfig {
    fn default() -> Self {
        Self {
            link_text: String::from("comment"),
            show_comment_count: true,
        }
    }
}

pub struct CommentsPreprocessor;

impl CommentsPreprocessor {
    pub fn new() -> Self {
        CommentsPreprocessor
    }
}

impl Preprocessor for CommentsPreprocessor {
    fn name(&self) -> &str {
        "comments"
    }

    fn run(&self, ctx: &PreprocessorContext, mut book: Book) -> Result<Book> {
        let config: CommentsConfig = ctx
            .config
            .get_preprocessor(self.name())
            .and_then(|table| serde_json::from_value(serde_json::to_value(table).ok()?).ok())
            .unwrap_or_default();

        let processor = CommentsProcessor::new(config);

        book.for_each_mut(|item| {
            if let BookItem::Chapter(chapter) = item {
                if let Err(e) = processor.process_chapter(chapter) {
                    eprintln!("Error processing chapter {}: {}", chapter.name, e);
                }
            }
        });

        Ok(book)
    }

    fn supports_renderer(&self, renderer: &str) -> bool {
        renderer == "html"
    }
}

fn main() -> Result<()> {
    let mut args = std::env::args().skip(1);

    // Handle preprocessor commands
    match args.next().as_deref() {
        Some("supports") => {
            // Check if we support the renderer
            let renderer = args.next().unwrap_or_default();
            let preprocessor = CommentsPreprocessor::new();
            if preprocessor.supports_renderer(&renderer) {
                process::exit(0);
            } else {
                process::exit(1);
            }
        }
        _ => {
            // Run as preprocessor
            let preprocessor = CommentsPreprocessor::new();
            let (ctx, book) = CmdPreprocessor::parse_input(io::stdin())?;
            let processed_book = preprocessor.run(&ctx, book)?;
            serde_json::to_writer(io::stdout(), &processed_book)?;
        }
    }

    Ok(())
}
