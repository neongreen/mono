package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// Client represents a persistent connection to a gopls language server.
type Client struct {
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	reader     *bufio.Reader
	nextID     atomic.Int64
	rootURI    string
	mu         sync.Mutex
	pending    map[int64]chan *Response
	openDocs   map[string]int // URI -> version
	serverInfo ServerInfo
}

// ServerInfo contains information about the server.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Request represents a JSON-RPC request.
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Response represents a JSON-RPC response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// ResponseError represents a JSON-RPC error.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Notification represents a JSON-RPC notification.
type Notification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// NewClient creates and initializes a new LSP client connected to gopls.
func NewClient(goplsPath string, workspaceRoot string) (*Client, error) {
	absRoot, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	rootURI := "file://" + absRoot

	slog.Debug("Starting gopls server", "goplsPath", goplsPath, "rootURI", rootURI)

	// Start gopls in serve mode
	cmd := exec.Command(goplsPath, "serve")
	cmd.Dir = absRoot

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start gopls: %w", err)
	}

	client := &Client{
		cmd:      cmd,
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		reader:   bufio.NewReader(stdout),
		rootURI:  rootURI,
		pending:  make(map[int64]chan *Response),
		openDocs: make(map[string]int),
	}

	// Start reading responses in background
	go client.readLoop()

	// Log stderr output
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			slog.Debug("gopls stderr", "output", scanner.Text())
		}
	}()

	// Initialize the server
	if err := client.initialize(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	// Send initialized notification
	if err := client.initialized(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to send initialized notification: %w", err)
	}

	slog.Info("LSP client initialized", "server", client.serverInfo.Name, "version", client.serverInfo.Version)

	return client, nil
}

// initialize sends the initialize request to the server.
func (c *Client) initialize() error {
	params := map[string]interface{}{
		"processId": os.Getpid(),
		"rootUri":   c.rootURI,
		"capabilities": map[string]interface{}{
			"workspace": map[string]interface{}{
				"applyEdit": true,
				"executeCommand": map[string]interface{}{
					"dynamicRegistration": false,
				},
			},
			"textDocument": map[string]interface{}{
				"rename": map[string]interface{}{
					"dynamicRegistration": false,
					"prepareSupport":      true,
				},
				"codeAction": map[string]interface{}{
					"dynamicRegistration": false,
				},
			},
		},
		"initializationOptions": map[string]interface{}{},
	}

	var result struct {
		Capabilities interface{}    `json:"capabilities"`
		ServerInfo   ServerInfo     `json:"serverInfo"`
	}

	if err := c.Call("initialize", params, &result); err != nil {
		return err
	}

	c.serverInfo = result.ServerInfo
	return nil
}

// initialized sends the initialized notification to the server.
func (c *Client) initialized() error {
	return c.Notify("initialized", map[string]interface{}{})
}

// Call sends a request and waits for the response.
func (c *Client) Call(method string, params interface{}, result interface{}) error {
	id := c.nextID.Add(1)

	req := &Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	respChan := make(chan *Response, 1)
	
	c.mu.Lock()
	c.pending[id] = respChan
	c.mu.Unlock()

	if err := c.sendMessage(req); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return err
	}

	resp := <-respChan

	if resp.Error != nil {
		return fmt.Errorf("RPC error: %s (code: %d)", resp.Error.Message, resp.Error.Code)
	}

	if result != nil && len(resp.Result) > 0 {
		if err := json.Unmarshal(resp.Result, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// Notify sends a notification (no response expected).
func (c *Client) Notify(method string, params interface{}) error {
	notif := &Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	return c.sendMessage(notif)
}

// sendMessage sends a JSON-RPC message to the server.
func (c *Client) sendMessage(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := c.stdin.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := c.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	slog.Debug("Sent LSP message", "method", getMethod(msg), "size", len(data))

	return nil
}

// readLoop reads responses and notifications from the server.
func (c *Client) readLoop() {
	for {
		msg, err := c.readMessage()
		if err != nil {
			if err != io.EOF {
				slog.Error("Error reading message", "error", err)
			}
			return
		}

		// Try to parse as response
		var resp Response
		if err := json.Unmarshal(msg, &resp); err == nil && resp.ID != 0 {
			c.mu.Lock()
			respChan, ok := c.pending[resp.ID]
			if ok {
				delete(c.pending, resp.ID)
			}
			c.mu.Unlock()

			if ok {
				respChan <- &resp
			}
			continue
		}

		// Otherwise it's a notification - log it
		var notif Notification
		if err := json.Unmarshal(msg, &notif); err == nil {
			slog.Debug("Received LSP notification", "method", notif.Method)
		}
	}
}

// readMessage reads a single LSP message.
func (c *Client) readMessage() ([]byte, error) {
	var contentLength int

	// Read headers
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = line[:len(line)-2] // Remove \r\n

		if line == "" {
			break // End of headers
		}

		var key string
		var value int
		if _, err := fmt.Sscanf(line, "%s %d", &key, &value); err == nil && key == "Content-Length:" {
			contentLength = value
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	// Read content
	content := make([]byte, contentLength)
	if _, err := io.ReadFull(c.reader, content); err != nil {
		return nil, err
	}

	return content, nil
}

// Close shuts down the LSP server.
func (c *Client) Close() error {
	slog.Debug("Shutting down LSP client")

	// Send shutdown request
	if err := c.Call("shutdown", nil, nil); err != nil {
		slog.Warn("Error during shutdown", "error", err)
	}

	// Send exit notification
	if err := c.Notify("exit", nil); err != nil {
		slog.Warn("Error sending exit", "error", err)
	}

	// Close pipes
	c.stdin.Close()
	c.stdout.Close()
	c.stderr.Close()

	// Wait for process to exit
	if err := c.cmd.Wait(); err != nil {
		return fmt.Errorf("gopls process exited with error: %w", err)
	}

	return nil
}

// getMethod extracts the method name from a message.
func getMethod(msg interface{}) string {
	switch m := msg.(type) {
	case *Request:
		return m.Method
	case *Notification:
		return m.Method
	default:
		return "unknown"
	}
}
