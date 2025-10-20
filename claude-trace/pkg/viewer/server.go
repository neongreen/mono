package viewer

import (
	"claude-trace/pkg/render"
	"claude-trace/pkg/storage"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Server provides the web viewer for traces
type Server struct {
	trace *storage.Trace
	port  int
}

// NewServer creates a new viewer server for the given trace
func NewServer(trace *storage.Trace, port int) *Server {
	return &Server{
		trace: trace,
		port:  port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/api/trace", s.handleTraceAPI)

	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	log.Printf("Starting viewer at http://%s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleIndex serves the HTML viewer
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(ViewerHTML)
}

// handleTraceAPI serves the trace data as JSON
func (s *Server) handleTraceAPI(w http.ResponseWriter, r *http.Request) {
	// Convert trace to TraceData format
	traceData := render.ToTraceData(s.trace)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(traceData); err != nil {
		http.Error(w, "Failed to encode trace data", http.StatusInternalServerError)
		return
	}
}
