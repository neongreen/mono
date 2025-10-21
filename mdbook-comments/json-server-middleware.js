// Custom middleware for mdbook-comments JSON Server
// Handles the /api/comments/:id/reply endpoint

const fs = require('fs');
const path = require('path');

module.exports = (req, res, next) => {
  // Log all POST requests for debugging
  if (req.method === 'POST') {
    console.log('[Middleware] POST request to:', req.url);
  }

  // Handle POST /api/comments/:id/reply
  if (req.method === 'POST' && req.url.match(/^\/api\/comments\/[^\/]+\/reply$/)) {
    console.log('[Middleware] Handling reply request:', req.url);
    const commentId = req.url.split('/')[3]; // Extract ID from /api/comments/:id/reply
    console.log('[Middleware] Parent comment ID:', commentId);

    try {
      // Use req.body which is already parsed by json-server's body parser
      const replyData = req.body;
      console.log('[Middleware] Reply data:', replyData);

      // Read the database file to get parent comment's paragraph-id
      const dbPath = process.argv[2] || '/tmp/test-db.json'; // Get db path from args
      console.log('[Middleware] Database path:', dbPath);
      const db = JSON.parse(fs.readFileSync(dbPath, 'utf8'));
      console.log('[Middleware] Database has', db.comments.length, 'comments');

      // Find parent comment
      const parentComment = db.comments.find(c => c.id === commentId);
      if (!parentComment) {
        res.statusCode = 404;
        res.end(JSON.stringify({ error: 'Parent comment not found' }));
        return;
      }

      // Create a new comment with parent-id set
      const newComment = {
        id: Date.now().toString(),
        'paragraph-id': parentComment['paragraph-id'],
        metadata: parentComment.metadata || {},
        author: replyData.author || 'Anonymous',
        text: replyData.text || '',
        created: new Date().toISOString(),
        'parent-id': commentId,
        replies: []
      };

      // Save to database
      db.comments.push(newComment);
      fs.writeFileSync(dbPath, JSON.stringify(db, null, 2), 'utf8');

      res.setHeader('Content-Type', 'application/json');
      res.end(JSON.stringify(newComment));

    } catch (error) {
      console.error('Error handling reply:', error);
      res.statusCode = 500;
      res.end(JSON.stringify({ error: 'Internal server error: ' + error.message }));
    }

    return; // Don't call next() for handled routes
  }

  next(); // Continue to JSON Server for other routes
};
