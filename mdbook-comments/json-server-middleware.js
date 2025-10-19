// Custom middleware for mdbook-comments JSON Server
// Handles the /api/comments/:id/reply endpoint

module.exports = (req, res, next) => {
  // Handle POST /api/comments/:id/reply
  if (req.method === 'POST' && req.url.match(/^\/api\/comments\/[^\/]+\/reply$/)) {
    const commentId = req.url.split('/')[3]; // Extract ID from /api/comments/:id/reply

    // Read the request body
    let body = '';
    req.on('data', chunk => {
      body += chunk.toString();
    });

    req.on('end', () => {
      try {
        const replyData = JSON.parse(body);

        // Create a new comment with parent-id set
        const newComment = {
          id: Date.now().toString(),
          'paragraph-id': replyData['paragraph-id'] || null,
          metadata: replyData.metadata || {},
          author: replyData.author || 'Anonymous',
          text: replyData.text || '',
          created: new Date().toISOString(),
          'parent-id': commentId,
          replies: []
        };

        // For now, just return the new comment structure
        // In a real implementation, this would be saved to the database
        res.setHeader('Content-Type', 'application/json');
        res.end(JSON.stringify(newComment));

      } catch (error) {
        res.statusCode = 400;
        res.end(JSON.stringify({ error: 'Invalid JSON' }));
      }
    });

    return; // Don't call next() for handled routes
  }

  next(); // Continue to JSON Server for other routes
};
