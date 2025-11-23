package generator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/neongreen/mono/lion/internal/extractor"
)

var warningWriter io.Writer = os.Stderr

/*lion:snippet-order section="Snippet ordering"
Snippets appear in the generated file in traversal order; within a file they follow source order.
Multiple snippets with the same topic are combined into a single markdown file named after the topic slug.
*/
/*lion:output-behavior section="Stale files"
Generate overwrites files for topics that still exist but does not delete files for topics that disappeared; stale topic files remain even though the index links only current topics, so clean them up yourself if needed.
*/
/*lion:file-title section="Titles and headings"
The generated file title comes from the topic display title:
- Default: Title Case of the topic slug (e.g., "getting-started" -> "Getting Started").
- Override: set title="Custom Title" on any lion marker for that topic; conflicts fail generation.
Entry headings:
- Default: the attached entity name (package <name>, function name, first const/var in the block).
- Override per entry: section="Custom Section"; section="" suppresses the heading entirely (will emit a warning).
- Conflicting section titles within the same comment group fail extraction.
*/
func Generate(docs map[string][]extractor.DocEntry, outputDir string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Sort topics for consistent ordering
	topics := make([]string, 0, len(docs))
	for topic := range docs {
		topics = append(topics, topic)
	}
	sort.Strings(topics)

	// Resolve titles per topic and validate conflicts
	titles := make(map[string]string, len(topics))
	for _, topic := range topics {
		entries := docs[topic]
		displayTitle := formatTopic(topic)
		var customTitle string
		var customFile string
		var customLine int
		for _, entry := range entries {
			if entry.HasTopicTitle {
				if customTitle == "" {
					customTitle = entry.TopicTitle
					customFile = entry.File
					customLine = entry.Line
				} else if customTitle != entry.TopicTitle {
					return fmt.Errorf("conflicting titles for topic %s: %q at %s:%d vs %q", topic, customTitle, customFile, customLine, entry.TopicTitle)
				}
			}
		}
		if customTitle != "" {
			displayTitle = customTitle
		}
		titles[topic] = displayTitle
	}

	// Generate a markdown file for each topic
	for _, topic := range topics {
		entries := docs[topic]
		if err := generateTopicFile(topic, titles[topic], entries, outputDir); err != nil {
			return fmt.Errorf("failed to generate file for topic %s: %w", topic, err)
		}
	}

	// Generate index file
	if err := generateIndex(topics, titles, outputDir); err != nil {
		return fmt.Errorf("failed to generate index: %w", err)
	}

	return nil
}

func generateTopicFile(topic string, displayTitle string, entries []extractor.DocEntry, outputDir string) error {
	// Sort entries by file and line for consistent ordering
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].File != entries[j].File {
			return entries[i].File < entries[j].File
		}
		return entries[i].Line < entries[j].Line
	})

	var content strings.Builder

	// Write title
	content.WriteString(fmt.Sprintf("# %s\n\n", displayTitle))

	// Write each entry
	for _, entry := range entries {
		heading := entry.Entity
		if entry.HasSection {
			heading = entry.SectionTitle
		}
		if heading != "" {
			content.WriteString(fmt.Sprintf("## %s\n\n", heading))
		} else {
			fmt.Fprintf(warningWriter, "Warning: topic %s entry from %s:%d has no heading\n", topic, entry.File, entry.Line)
		}

		if entry.Content != "" {
			content.WriteString(entry.Content)
			content.WriteString("\n\n")
		}

		// Add source reference
		relPath := entry.File
		content.WriteString(fmt.Sprintf("*Source: `%s:%d`*\n\n", relPath, entry.Line))
	}

	// Write to file
	filename := filepath.Join(outputDir, topic+".md")
	if err := os.WriteFile(filename, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func generateIndex(topics []string, titles map[string]string, outputDir string) error {
	var content strings.Builder

	content.WriteString("# Documentation Index\n\n")
	content.WriteString("This documentation was generated from lion comments in the source code.\n\n")

	content.WriteString("## Topics\n\n")
	for _, topic := range topics {
		title := formatTopic(topic)
		if custom, ok := titles[topic]; ok && custom != "" {
			title = custom
		}
		content.WriteString(fmt.Sprintf("- [%s](%s.md)\n", title, topic))
	}

	filename := filepath.Join(outputDir, "index.md")
	if err := os.WriteFile(filename, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}

func formatTopic(topic string) string {
	// Replace hyphens and underscores with spaces
	formatted := strings.ReplaceAll(topic, "-", " ")
	formatted = strings.ReplaceAll(formatted, "_", " ")

	// Capitalize each word
	words := strings.Fields(formatted)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}
