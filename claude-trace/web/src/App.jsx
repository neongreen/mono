import React, { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [traceData, setTraceData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    // Fetch trace data from the Go server
    fetch('/api/trace')
      .then(response => {
        if (!response.ok) {
          throw new Error('Failed to load trace data')
        }
        return response.json()
      })
      .then(data => {
        setTraceData(data)
        setLoading(false)
      })
      .catch(err => {
        setError(err.message)
        setLoading(false)
      })
  }, [])

  if (loading) {
    return <div className="loading">Loading trace...</div>
  }

  if (error) {
    return <div className="error">Error: {error}</div>
  }

  if (!traceData) {
    return <div className="error">No trace data available</div>
  }

  return (
    <div className="app">
      <header className="header">
        <h1>Claude Trace Viewer</h1>
      </header>
      
      <main className="main">
        <div className="metadata">
          <h2>{traceData.name}</h2>
          <div className="metadata-grid">
            <div className="metadata-item">
              <span className="label">Path:</span>
              <span className="value">{traceData.path}</span>
            </div>
            <div className="metadata-item">
              <span className="label">Modified:</span>
              <span className="value">{new Date(traceData.mod_time).toLocaleString()}</span>
            </div>
            {traceData.parsed_trace?.session_id && (
              <div className="metadata-item">
                <span className="label">Session ID:</span>
                <span className="value code">{traceData.parsed_trace.session_id}</span>
              </div>
            )}
            {traceData.parsed_trace?.summary && (
              <div className="metadata-item full-width">
                <span className="label">Summary:</span>
                <span className="value">{traceData.parsed_trace.summary}</span>
              </div>
            )}
          </div>

          {traceData.tags && Object.keys(traceData.tags).some(tag => traceData.tags[tag]) && (
            <div className="tags">
              <span className="label">Tags:</span>
              {Object.entries(traceData.tags)
                .filter(([_, active]) => active)
                .map(([tag, _]) => (
                  <span key={tag} className="tag">{tag}</span>
                ))}
            </div>
          )}

          {traceData.freeform_note && (
            <div className="notes">
              <h3>Notes</h3>
              <div className="note-content">{traceData.freeform_note}</div>
            </div>
          )}
        </div>

        <div className="conversation">
          <h3>Conversation</h3>
          {traceData.parsed_trace?.items && traceData.parsed_trace.items.length > 0 ? (
            <div className="conversation-items">
              {traceData.parsed_trace.items.map((item, idx) => (
                <ConversationItem key={idx} item={item} />
              ))}
            </div>
          ) : (
            <pre className="raw-content">{traceData.raw_content || 'No content available'}</pre>
          )}
        </div>

        {traceData.annotations && traceData.annotations.length > 0 && (
          <div className="annotations">
            <h3>Annotation History</h3>
            <div className="annotation-list">
              {traceData.annotations.map((ann, idx) => (
                <div key={idx} className="annotation-item">
                  <span className="annotation-time">
                    {new Date(ann.timestamp).toLocaleString()}
                  </span>
                  {ann.tag && <span className="annotation-tag">{ann.tag}</span>}
                  {ann.note && <span className="annotation-note">{ann.note}</span>}
                </div>
              ))}
            </div>
          </div>
        )}
      </main>
    </div>
  )
}

function ConversationItem({ item }) {
  if (item.type === 'user') {
    return (
      <div className="conversation-item user-message">
        <div className="message-header">
          <span className="message-type">User</span>
          {item.timestamp && (
            <span className="message-time">
              {new Date(item.timestamp).toLocaleString()}
            </span>
          )}
        </div>
        {item.user_message?.cwd && (
          <div className="message-cwd">Working directory: {item.user_message.cwd}</div>
        )}
        <div className="message-content">
          {item.user_message?.content || ''}
        </div>
      </div>
    )
  }

  if (item.type === 'assistant') {
    return (
      <div className="conversation-item assistant-message">
        <div className="message-header">
          <span className="message-type">Assistant</span>
          {item.assistant_message?.model && (
            <span className="message-model">{item.assistant_message.model}</span>
          )}
          {item.timestamp && (
            <span className="message-time">
              {new Date(item.timestamp).toLocaleString()}
            </span>
          )}
        </div>
        {item.assistant_message?.content && (
          <div className="message-blocks">
            {item.assistant_message.content.map((block, idx) => (
              <ContentBlock key={idx} block={block} />
            ))}
          </div>
        )}
      </div>
    )
  }

  if (item.type === 'tool_result') {
    return (
      <div className="conversation-item tool-result">
        <div className="message-header">
          <span className="message-type">Tool Result</span>
          {item.tool_result?.tool_use_id && (
            <span className="tool-id">{item.tool_result.tool_use_id}</span>
          )}
          {item.tool_result?.is_error && (
            <span className="tool-error-badge">Error</span>
          )}
        </div>
        <pre className="tool-content">{item.tool_result?.content || ''}</pre>
      </div>
    )
  }

  return null
}

function ContentBlock({ block }) {
  if (block.type === 'text' && block.text) {
    return <div className="content-block text-block">{block.text}</div>
  }

  if (block.type === 'thinking' && block.thinking) {
    return (
      <div className="content-block thinking-block">
        <div className="thinking-label">Thinking</div>
        <div className="thinking-content">{block.thinking}</div>
      </div>
    )
  }

  if (block.type === 'tool_use' && block.tool_use) {
    return (
      <div className="content-block tool-use-block">
        <div className="tool-header">
          <span className="tool-name">{block.tool_use.name}</span>
          {block.tool_use.id && <span className="tool-id">{block.tool_use.id}</span>}
        </div>
        {block.tool_use.input && (
          <pre className="tool-input">{JSON.stringify(block.tool_use.input, null, 2)}</pre>
        )}
      </div>
    )
  }

  return null
}

export default App
