package lsp

import (
	"fmt"
	"log/slog"
	"path/filepath"
)

// Manager manages the lifecycle of LSP clients for different module roots.
type Manager struct {
	clients map[string]*Client // moduleRoot -> client
}

// NewManager creates a new LSP client manager.
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*Client),
	}
}

// GetClient returns an LSP client for the given module root, creating one if necessary.
func (m *Manager) GetClient(goplsPath string, moduleRoot string) (*Client, error) {
	// Normalize the module root path
	absModuleRoot, err := filepath.Abs(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if we already have a client for this module
	if client, ok := m.clients[absModuleRoot]; ok {
		return client, nil
	}

	// Create a new client
	slog.Info("Creating new LSP client", "moduleRoot", absModuleRoot)
	client, err := NewClient(goplsPath, absModuleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to create LSP client: %w", err)
	}

	m.clients[absModuleRoot] = client
	return client, nil
}

// CloseAll closes all LSP clients.
func (m *Manager) CloseAll() error {
	slog.Info("Closing all LSP clients", "count", len(m.clients))
	
	var firstErr error
	for moduleRoot, client := range m.clients {
		if err := client.Close(); err != nil {
			slog.Error("Failed to close LSP client", "moduleRoot", moduleRoot, "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
		delete(m.clients, moduleRoot)
	}

	return firstErr
}
