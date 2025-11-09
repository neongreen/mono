// Test @mozilla/readability with its built-in JSDOMParser
var Readability = require('@mozilla/readability').Readability;
// JSDOMParser is not exported, need to require it directly
var JSDOMParser = require('@mozilla/readability/JSDOMParser.js');

// Suppress console.log to avoid polluting stdout
// (JSDOMParser logs errors to console.log)
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

    writeStdout(JSON.stringify({
      success: true,
      data: article
    }));
  }
} catch (error) {
  writeStdout(JSON.stringify({
    success: false,
    error: String(error),
    stack: error.stack
  }));
}
