// Simplified parser using Javy.IO APIs
// This demonstrates the WASM infrastructure working

// Read from stdin using Javy.IO
function readStdin() {
  const chunks = [];
  const chunkSize = 1024;
  const buffer = new Uint8Array(chunkSize);

  while (true) {
    const bytesRead = Javy.IO.readSync(0, buffer); // 0 is stdin
    if (bytesRead === 0) break;
    if (bytesRead < 0) throw new Error('Read error');
    chunks.push(buffer.slice(0, bytesRead));
  }

  // Concatenate all chunks
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

// Write to stdout using Javy.IO
function writeStdout(text) {
  const encoder = new TextEncoder();
  const data = encoder.encode(text);
  Javy.IO.writeSync(1, data); // 1 is stdout
}

try {
  // Read and parse input
  const inputText = readStdin();
  const input = JSON.parse(inputText);

  if (!input.html) {
    const output = JSON.stringify({
      success: false,
      error: 'HTML content is required'
    });
    writeStdout(output);
  } else {
    // Simple HTML parsing - extract title and basic content
    const html = input.html;

    // Extract title
    let title = '';
    const titleMatch = html.match(/<title[^>]*>([^<]+)<\/title>/i);
    if (titleMatch) {
      title = titleMatch[1].trim();
    }

    // Extract content from body
    let content = '';
    const bodyMatch = html.match(/<body[^>]*>([\s\S]*)<\/body>/i);
    if (bodyMatch) {
      content = bodyMatch[1];
      // Remove scripts and styles
      content = content.replace(/<script[\s\S]*?<\/script>/gi, '');
      content = content.replace(/<style[\s\S]*?<\/style>/gi, '');
    }

    // Count words (simple approximation)
    const textContent = content.replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ').trim();
    const wordCount = textContent.split(/\s+/).filter(w => w.length > 0).length;

    // Extract domain from URL
    let domain = '';
    const urlMatch = input.url.match(/^(?:https?:\/\/)?([^\/]+)/);
    if (urlMatch) {
      domain = urlMatch[1];
    }

    const result = {
      title: title,
      author: '',
      date_published: '',
      dek: '',
      lead_image_url: '',
      content: content,
      excerpt: textContent.substring(0, 150),
      word_count: wordCount,
      direction: 'ltr',
      url: input.url,
      domain: domain
    };

    const output = JSON.stringify({
      success: true,
      data: result
    });
    writeStdout(output);
  }
} catch (error) {
  const output = JSON.stringify({
    success: false,
    error: String(error)
  });
  writeStdout(output);
}
