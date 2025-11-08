package remote

// LoadIngestWatermark loads the ingest watermark from a file
func LoadIngestWatermark(path string) (*IngestWatermark, error) {
	return LoadJSON[IngestWatermark](path)
}

// SaveIngestWatermark saves an ingest watermark to a file
func SaveIngestWatermark(path string, watermark *IngestWatermark) error {
	return SaveJSON(path, watermark)
}

// IsDuplicateError checks if an error is a duplicate key error
func IsDuplicateError(err error) bool {
	// SQLite's duplicate key error contains "UNIQUE constraint failed"
	return err != nil && (
	// modernc.org/sqlite error messages
	containsString(err.Error(), "UNIQUE constraint failed") ||
		containsString(err.Error(), "constraint failed"))
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || containsString(s[1:], substr)))
}
