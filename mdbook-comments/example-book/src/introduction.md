# Introduction

Welcome to the example book demonstrating the mdbook-comments plugin!

This book shows how paragraph-level commenting works. Each paragraph, list item, code block, and other content blocks can have comments attached to them.

## How It Works

When you read this book, you'll see small "comment" links at the end of each paragraph. Clicking on these links will expand a comment section where you can:

- Read existing comments from other readers
- Add your own comments
- Reply to comments (one level of nesting)

The comment system is designed to be simple and unobtrusive, following the Real World Haskell approach.

## Features

The system includes several advanced features:

- **Fuzzy Matching**: Comments stay attached to paragraphs even when text is edited slightly
- **Orphaned Comments**: When content is removed or significantly changed, comments are preserved and shown at the end of the chapter
- **Rich Context**: Each comment stores information about its paragraph's location, surrounding content, and more
- **Stateless Matching**: No need for diffs or migration tools - the system figures out where comments belong automatically

Try it out by clicking on the comment links you see throughout the book!
