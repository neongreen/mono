// Full article parser using Mozilla Readability
// Compiled to WASM and run via wazero in Go
var Readability = require('@mozilla/readability').Readability;
var JSDOMParser = require('@mozilla/readability/JSDOMParser.js');

// Suppress console.log to avoid polluting stdout
// (JSDOMParser may log parsing warnings)
console.log = function() {};

function readStdin() {
  const chunks = [];
  const chunkSize = 1024;
  const buffer = new Uint8Array(chunkSize);
  while (true) {
    const bytesRead = Javy.IO.readSync(0, buffer);
    if (bytesRead === 0) break;
    if (bytesRead < 0) throw new Error('Read error');
    chunks.push(buffer.slice(0, bytesRead));
  }
  const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
  const result = new Uint8Array(totalLength);
  let offset = 0;
  for (const chunk of chunks) {
    result.set(chunk, offset);
    offset += chunk.length;
  }
  const decoder = new TextDecoder();
  return decoder.decode(result);
}

function writeStdout(text) {
  const encoder = new TextEncoder();
  const data = encoder.encode(text);
  Javy.IO.writeSync(1, data);
}

try {
  const inputText = readStdin();
  const input = JSON.parse(inputText);

  if (!input.html) {
    writeStdout(JSON.stringify({
      success: false,
      error: 'HTML content is required'
    }));
  } else {
    // Parse HTML with JSDOMParser
    var parser = new JSDOMParser();
    var doc = parser.parse(input.html, input.url);

    // Extract article with Readability
    var reader = new Readability(doc);
    var article = reader.parse();

    if (!article) {
      writeStdout(JSON.stringify({
        success: false,
        error: 'Failed to extract article content. The page may not be an article or may not have enough content.'
      }));
    } else {
      // Map Readability output to our Article format
      // Readability returns: title, byline, dir, content, textContent, length, excerpt, siteName, publishedTime
      var result = {
        title: article.title || '',
        author: article.byline || '',
        date_published: article.publishedTime || '',
        dek: article.excerpt || '',
        lead_image_url: '', // Readability doesn't extract this separately
        content: article.content || '',
        excerpt: article.excerpt || '',
        word_count: article.length || 0, // Readability's 'length' is character count
        direction: article.dir || 'ltr',
        url: input.url,
        domain: article.siteName || ''
      };

      // Extract domain from URL if not provided by Readability
      if (!result.domain && input.url) {
        try {
          const urlMatch = input.url.match(/^(?:https?:\/\/)?([^\/]+)/);
          if (urlMatch) {
            result.domain = urlMatch[1];
          }
        } catch (e) {
          // Ignore URL parsing errors
        }
      }

      writeStdout(JSON.stringify({
        success: true,
        data: result
      }));
    }
  }
} catch (error) {
  writeStdout(JSON.stringify({
    success: false,
    error: String(error),
    stack: error.stack
  }));
}
