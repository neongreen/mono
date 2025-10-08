package storage

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Trace represents a Claude Code conversation trace
type Trace struct {
	Path         string            `json:"path"`
	Name         string            `json:"name"`
	Content      string            `json:"content"`
	ModTime      time.Time         `json:"mod_time"`
	Annotations  []Annotation      `json:"annotations"`
	Tags         map[string]bool   `json:"tags"`
	FreeformNote string            `json:"freeform_note"`
}

// Annotation represents a single annotation on a trace
type Annotation struct {
	Timestamp time.Time `json:"timestamp"`
	Tag       string    `json:"tag,omitempty"`
	Note      string    `json:"note,omitempty"`
}

// LoadTraces loads all trace files from the given directories
func LoadTraces(directories []string) ([]*Trace, error) {
	var traces []*Trace

	for _, dir := range directories {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			// Look for log files, JSON files, or text files
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".log" || ext == ".json" || ext == ".txt" || ext == ".md" {
				content, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("failed to read %s: %w", path, err)
				}

				info, err := d.Info()
				if err != nil {
					return fmt.Errorf("failed to get info for %s: %w", path, err)
				}

				trace := &Trace{
					Path:        path,
					Name:        filepath.Base(path),
					Content:     string(content),
					ModTime:     info.ModTime(),
					Annotations: []Annotation{},
					Tags:        make(map[string]bool),
				}

				traces = append(traces, trace)
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error walking directory %s: %w", dir, err)
		}
	}

	// Sort traces by modification time (newest first)
	sort.Slice(traces, func(i, j int) bool {
		return traces[i].ModTime.After(traces[j].ModTime)
	})

	return traces, nil
}

// SaveAnnotations saves the annotations for a trace
func SaveAnnotations(trace *Trace) error {
	annotationPath := trace.Path + ".annotations.json"
	
	data, err := json.MarshalIndent(trace, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal annotations: %w", err)
	}

	err = os.WriteFile(annotationPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write annotations file: %w", err)
	}

	return nil
}

// LoadAnnotations loads existing annotations for a trace if they exist
func LoadAnnotations(trace *Trace) error {
	annotationPath := trace.Path + ".annotations.json"
	
	data, err := os.ReadFile(annotationPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No annotations yet, that's fine
			return nil
		}
		return fmt.Errorf("failed to read annotations file: %w", err)
	}

	var savedTrace Trace
	err = json.Unmarshal(data, &savedTrace)
	if err != nil {
		return fmt.Errorf("failed to unmarshal annotations: %w", err)
	}

	// Copy over the annotations
	trace.Annotations = savedTrace.Annotations
	trace.Tags = savedTrace.Tags
	trace.FreeformNote = savedTrace.FreeformNote

	return nil
}
