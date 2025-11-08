import { Readability } from '@mozilla/readability';
import { JSDOM } from 'jsdom';

// Read HTML from stdin
let input = '';
process.stdin.setEncoding('utf8');

process.stdin.on('data', (chunk) => {
    input += chunk;
});

process.stdin.on('end', () => {
    try {
        // Parse with JSDOM
        const dom = new JSDOM(input, { url: 'http://localhost' });
        const reader = new Readability(dom.window.document);
        const article = reader.parse();

        if (article && article.content) {
            // Wrap in clean HTML document
            const cleanHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>${escapeHtml(article.title || 'Article')}</title>
<style>
body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
    max-width: 800px;
    margin: 40px auto;
    padding: 0 20px;
    line-height: 1.6;
    color: #24292e;
}
h1, h2, h3, h4, h5, h6 {
    margin-top: 24px;
    margin-bottom: 16px;
    font-weight: 600;
    line-height: 1.25;
}
p {
    margin-top: 0;
    margin-bottom: 16px;
}
img {
    max-width: 100%;
    height: auto;
}
a {
    color: #0366d6;
    text-decoration: none;
}
a:hover {
    text-decoration: underline;
}
pre, code {
    background-color: #f6f8fa;
    border-radius: 3px;
    padding: 2px 4px;
    font-family: 'Courier New', monospace;
}
pre {
    padding: 16px;
    overflow: auto;
    line-height: 1.45;
}
blockquote {
    border-left: 4px solid #dfe2e5;
    padding-left: 16px;
    color: #6a737d;
    margin-left: 0;
}
ul, ol {
    padding-left: 2em;
}
li {
    margin-bottom: 0.5em;
}
</style>
</head>
<body>
${article.content}
</body>
</html>`;
            process.stdout.write(cleanHTML);
        } else {
            console.error('Failed to extract article content');
            process.exit(1);
        }
    } catch (error) {
        console.error('Error:', error.message);
        process.exit(1);
    }
});

function escapeHtml(text) {
    const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    };
    return text.replace(/[&<>"']/g, m => map[m]);
}
