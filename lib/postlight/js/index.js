// Wrapper for Postlight Parser to be compiled to WASM via javy
// This version expects HTML to be provided (fetching is done in Go)
// Uses javy's stdin/stdout model for WASM communication

const Parser = require('@postlight/parser');

// Read input from stdin (javy provides this via readInput())
const inputText = readInput();
const input = JSON.parse(inputText);

// Parse the article - we require HTML to be provided
// The URL is still needed for the parser's context
if (!input.html) {
  const output = JSON.stringify({
    success: false,
    error: 'HTML content is required. Fetch the URL in Go and pass the HTML.'
  });
  console.log(output);
} else {
  Parser.parse(input.url, { html: input.html })
    .then(result => {
      const output = JSON.stringify({
        success: true,
        data: result
      });
      console.log(output);
    })
    .catch(error => {
      const output = JSON.stringify({
        success: false,
        error: error.message || String(error)
      });
      console.log(output);
    });
}
