// parser_wasm.js - JavaScript parser wrapper for WASM compilation
// This file provides a simple HTML parser that can be compiled to WASM

// Simple HTML parser that extracts title and content
function parseHTML(html) {
    const result = {
        title: '',
        content: '',
        word_count: 0
    };
    
    // Extract title
    const titleMatch = html.match(/<title[^>]*>(.*?)<\/title>/i);
    if (titleMatch) {
        result.title = titleMatch[1].trim();
    }
    
    // Extract body content
    const bodyMatch = html.match(/<body[^>]*>(.*?)<\/body>/is);
    if (bodyMatch) {
        result.content = bodyMatch[1];
    } else {
        result.content = html;
    }
    
    // Count words (approximate)
    const text = result.content.replace(/<[^>]+>/g, ' ');
    const words = text.trim().split(/\s+/).filter(w => w.length > 0);
    result.word_count = words.length;
    
    return JSON.stringify(result);
}

// Export for WASM
if (typeof exports !== 'undefined') {
    exports.parseHTML = parseHTML;
}

// For WASI/WASM command-line usage
if (typeof Deno !== 'undefined') {
    const html = Deno.readTextFileSync('/dev/stdin');
    console.log(parseHTML(html));
}
